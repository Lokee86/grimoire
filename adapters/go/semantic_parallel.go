package main

import (
	"fmt"
	"go/ast"
	"reflect"
	"sort"
	"sync"

	"golang.org/x/tools/go/packages"
)

type semanticFileJob struct {
	pkg    *packages.Package
	file   *ast.File
	rel    string
	weight int
}

func (s *scanner) collectSemanticCallsParallel(
	loaded []*packages.Package,
	targets semanticTargets,
	options ScanOptions,
) error {
	jobs := s.semanticFileJobs(loaded)
	if len(jobs) == 0 {
		return nil
	}
	options = options.normalized(len(jobs))
	shards := partitionSemanticJobs(jobs, options.SemanticShards)
	results := make([]*scanner, len(shards))

	tasks := make(chan int)
	var group sync.WaitGroup
	for worker := 0; worker < options.SemanticWorkers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range tasks {
				local := s.newSemanticShardScanner()
				for _, job := range shards[index] {
					local.collectSemanticFile(job.pkg, job.file, job.rel, targets)
				}
				results[index] = local
			}
		}()
	}
	for index := range shards {
		tasks <- index
	}
	close(tasks)
	group.Wait()

	root, err := reduceSemanticShards(results, options.MergeFanIn)
	if err != nil {
		return err
	}
	return mergeSemanticShard(s, root)
}

func (s *scanner) semanticFileJobs(loaded []*packages.Package) []semanticFileJob {
	jobs := make([]semanticFileJob, 0)
	for _, pkg := range loaded {
		if pkg.TypesInfo == nil || pkg.Fset == nil {
			continue
		}
		for _, file := range pkg.Syntax {
			rel, ok := s.semanticFilePath(pkg.Fset, file)
			if !ok {
				continue
			}
			jobs = append(jobs, semanticFileJob{
				pkg: pkg, file: file, rel: rel, weight: len(file.Decls) + 1,
			})
		}
	}
	sort.SliceStable(jobs, func(left, right int) bool {
		if jobs[left].pkg.ID != jobs[right].pkg.ID {
			return jobs[left].pkg.ID < jobs[right].pkg.ID
		}
		return jobs[left].rel < jobs[right].rel
	})
	return jobs
}

func partitionSemanticJobs(jobs []semanticFileJob, count int) [][]semanticFileJob {
	if count < 1 {
		count = 1
	}
	if count > len(jobs) {
		count = len(jobs)
	}
	ordered := append([]semanticFileJob(nil), jobs...)
	sort.SliceStable(ordered, func(left, right int) bool {
		if ordered[left].weight != ordered[right].weight {
			return ordered[left].weight > ordered[right].weight
		}
		if ordered[left].pkg.ID != ordered[right].pkg.ID {
			return ordered[left].pkg.ID < ordered[right].pkg.ID
		}
		return ordered[left].rel < ordered[right].rel
	})
	shards := make([][]semanticFileJob, count)
	weights := make([]int, count)
	for _, job := range ordered {
		selected := 0
		for index := 1; index < count; index++ {
			if weights[index] < weights[selected] {
				selected = index
			}
		}
		shards[selected] = append(shards[selected], job)
		weights[selected] += job.weight
	}
	return shards
}

func (s *scanner) newSemanticShardScanner() *scanner {
	return &scanner{
		root: s.root, module: s.module, set: s.set,
		nodes: make(map[NodeKey]NodeFact), edges: make(map[string]EdgeFact),
		packages: s.packages, targets: s.targets,
		semanticCalls: make(map[string]semanticCall), semanticIDs: make(map[string][]NodeKey),
		closureKeys: s.closureKeys, callsiteKeys: make(map[string]string),
		fileImports: s.fileImports, packageByFile: s.packageByFile,
		baseNodes: s.nodes, baseSemanticIDs: s.semanticIDs,
	}
}

func reduceSemanticShards(shards []*scanner, fanIn int) (*scanner, error) {
	if len(shards) == 0 {
		return nil, nil
	}
	if fanIn < 2 {
		fanIn = 2
	}
	current := append([]*scanner(nil), shards...)
	for len(current) > 1 {
		next := make([]*scanner, 0, (len(current)+fanIn-1)/fanIn)
		for start := 0; start < len(current); start += fanIn {
			end := start + fanIn
			if end > len(current) {
				end = len(current)
			}
			merged := current[start]
			for index := start + 1; index < end; index++ {
				if err := mergeSemanticShard(merged, current[index]); err != nil {
					return nil, err
				}
			}
			next = append(next, merged)
		}
		current = next
	}
	return current[0], nil
}

func mergeSemanticShard(destination, source *scanner) error {
	if source == nil {
		return nil
	}
	for _, key := range sortedNodeKeys(source.nodes) {
		incoming := source.nodes[key]
		if existing, exists := destination.nodes[key]; exists {
			if !reflect.DeepEqual(existing, incoming) {
				return fmt.Errorf("semantic shard node conflict for %s", key)
			}
			continue
		}
		if destination.baseNodes != nil {
			if existing, exists := destination.baseNodes[key]; exists {
				if !reflect.DeepEqual(existing, incoming) {
					return fmt.Errorf("semantic shard base node conflict for %s", key)
				}
				continue
			}
		}
		destination.nodes[key] = incoming
	}
	for _, key := range sortedStringKeys(source.edges) {
		incoming := source.edges[key]
		if existing, exists := destination.edges[key]; exists {
			if !reflect.DeepEqual(existing, incoming) {
				return fmt.Errorf("semantic shard edge conflict for %s", key)
			}
			continue
		}
		destination.edges[key] = incoming
	}
	for _, id := range sortedSemanticIDs(source.semanticIDs) {
		keys := append([]NodeKey(nil), source.semanticIDs[id]...)
		sort.Slice(keys, func(left, right int) bool { return keys[left] < keys[right] })
		for _, key := range keys {
			destination.registerSemanticID(id, key)
		}
	}
	for _, key := range sortedStringKeys(source.callsiteKeys) {
		incoming := source.callsiteKeys[key]
		if existing, exists := destination.callsiteKeys[key]; exists && existing != incoming {
			return fmt.Errorf("semantic shard callsite conflict for %s", key)
		}
		destination.callsiteKeys[key] = incoming
	}
	for _, key := range sortedStringKeys(source.semanticCalls) {
		destination.mergeSemanticCall(key, source.semanticCalls[key])
	}
	return nil
}

func sortedNodeKeys(values map[NodeKey]NodeFact) []NodeKey {
	keys := make([]NodeKey, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(left, right int) bool { return keys[left] < keys[right] })
	return keys
}

func sortedStringKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedSemanticIDs(values map[string][]NodeKey) []string {
	return sortedStringKeys(values)
}
