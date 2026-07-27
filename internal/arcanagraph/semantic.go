package arcanagraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/structure"
)

type semanticHit struct {
	Score   float64 `json:"score"`
	NodeKey string  `json:"node_key"`
	Kind    string  `json:"kind"`
	Path    string  `json:"path"`
	Name    string  `json:"name"`
}

type semanticMatches struct {
	Matches []semanticHit `json:"matches"`
}

// SemanticSeed preserves Arcana's vector similarity and stable source rank so
// deterministic fusion can compare semantic hits with Lexicon anchors instead
// of discarding the signal before the production seed budget is applied.
type SemanticSeed struct {
	Node  structure.Node
	Score float64
	Rank  int
}

type semanticRun func(context.Context, string, string, string, string, string, int) ([]semanticHit, error)

// SemanticSeeds retrieves graph nodes directly from Arcana's optional semantic
// index. It never builds the index: a missing index is an ordinary no-result
// condition, preserving embedding-free Arcana operation by default.
func (client Client) SemanticSeeds(
	ctx context.Context,
	state string,
	expectedSnapshot string,
	endpoint string,
	query string,
	limit int,
) ([]structure.Node, error) {
	ranked, err := client.RankedSemanticSeeds(ctx, state, expectedSnapshot, endpoint, query, limit)
	if err != nil {
		return nil, err
	}
	seeds := make([]structure.Node, 0, len(ranked))
	for _, seed := range ranked {
		seeds = append(seeds, seed.Node)
	}
	return seeds, nil
}

// RankedSemanticSeeds retrieves the same optional Arcana semantic matches while
// retaining their similarity and stable rank for hybrid reranking.
func (client Client) RankedSemanticSeeds(
	ctx context.Context,
	state string,
	expectedSnapshot string,
	endpoint string,
	query string,
	limit int,
) ([]SemanticSeed, error) {
	if strings.TrimSpace(state) == "" || strings.TrimSpace(query) == "" || limit <= 0 {
		return nil, nil
	}

	run := client.RunSemantic
	if run == nil {
		exists, err := semanticIndexExists(state, expectedSnapshot)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, nil
		}
		run = runSemantic
	}
	command := strings.TrimSpace(client.Command)
	if command == "" {
		command = "arcana"
	}
	hits, err := run(ctx, command, state, expectedSnapshot, endpoint, query, limit)
	if err != nil {
		return nil, err
	}

	seeds := make([]SemanticSeed, 0, len(hits))
	seen := make(map[string]struct{}, len(hits))
	for _, hit := range hits {
		if hit.Name == "" {
			continue
		}
		key := hit.Name + "\x00" + hit.Path
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		seeds = append(seeds, SemanticSeed{
			Node: structure.Node{
				Identity: hit.NodeKey,
				Kind:     hit.Kind,
				Name:     hit.Name,
				Path:     hit.Path,
			},
			Score: hit.Score,
			Rank:  len(seeds) + 1,
		})
	}
	return seeds, nil
}

func semanticIndexExists(state, expectedSnapshot string) (bool, error) {
	value := strings.TrimSpace(expectedSnapshot)
	if value == "" {
		current, err := os.ReadFile(filepath.Join(state, "CURRENT"))
		if os.IsNotExist(err) {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("read Arcana CURRENT for semantic index: %w", err)
		}
		value = strings.TrimSpace(string(current))
	}
	digest, found := strings.CutPrefix(value, "sha256:")
	if !found || len(digest) != 64 {
		return false, fmt.Errorf("invalid Arcana CURRENT value %q", value)
	}
	manifest := filepath.Join(state, "vectors", digest, embedding.Identity(), "manifest.json")
	info, err := os.Stat(manifest)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect Arcana semantic index: %w", err)
	}
	return !info.IsDir(), nil
}

func runSemantic(
	ctx context.Context,
	command string,
	state string,
	expectedSnapshot string,
	endpoint string,
	query string,
	limit int,
) ([]semanticHit, error) {
	arguments := []string{
		"semantic-query",
		"--state", state,
		"--query", query,
		"--limit", strconv.Itoa(limit),
		"--json",
	}
	if strings.TrimSpace(expectedSnapshot) != "" {
		arguments = append(arguments, "--expected-snapshot", expectedSnapshot)
	}
	if strings.TrimSpace(endpoint) != "" {
		arguments = append(arguments, "--endpoint", endpoint)
	}

	var stdout, stderr bytes.Buffer
	process := exec.CommandContext(ctx, command, arguments...)
	process.Stdout = &stdout
	process.Stderr = &stderr
	if err := process.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		if message != "" {
			return nil, fmt.Errorf("run Arcana semantic query: %w: %s", err, message)
		}
		return nil, fmt.Errorf("run Arcana semantic query: %w", err)
	}
	var response semanticMatches
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("decode Arcana semantic query: %w", err)
	}
	return response.Matches, nil
}
