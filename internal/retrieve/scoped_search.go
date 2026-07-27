package retrieve

import "github.com/Lokee86/grimoire/internal/index"

// SearchManyInPathsWithConfig runs ordinary chunk BM25, then keeps only
// candidates inside the supplied discovery scope. No structural signal enters
// this stage.
func SearchManyInPathsWithConfig(
	snapshot index.Snapshot,
	queries []string,
	paths []string,
	limit int,
	config Config,
) [][]Candidate {
	results := make([][]Candidate, len(queries))
	if len(paths) == 0 || limit <= 0 {
		return results
	}
	allowed := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		allowed[path] = struct{}{}
	}
	return searchManyWithConfig(snapshot, queries, limit, config, allowed)
}
