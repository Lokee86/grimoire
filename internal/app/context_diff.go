package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/Lokee86/grimoire/internal/diffcontext"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

type contextDiffResult struct {
	PackageQuery   string
	RetrievalQuery string
	Candidates     []retrieve.Candidate
	Evidence       []structure.Evidence
}

func prepareContextDiff(
	ctx context.Context,
	snapshot index.Snapshot,
	root, spec, query string,
	limit int,
) (contextDiffResult, error) {
	query = strings.TrimSpace(query)
	if strings.TrimSpace(spec) == "" {
		return contextDiffResult{PackageQuery: query, RetrievalQuery: query}, nil
	}
	changes, err := diffcontext.Collect(ctx, root, spec)
	if err != nil {
		return contextDiffResult{}, err
	}
	if len(changes) == 0 {
		return contextDiffResult{}, fmt.Errorf("Git diff %q contains no changed source spans", strings.TrimSpace(spec))
	}
	return buildContextDiff(snapshot, query, changes, limit), nil
}

func buildContextDiff(
	snapshot index.Snapshot,
	query string,
	changes []diffcontext.Change,
	limit int,
) contextDiffResult {
	packageQuery := diffcontext.EffectiveQuery(query)
	candidateLimit := min(limit, diffcontext.DefaultCandidateLimit)
	evidenceLimit := min(limit, diffcontext.DefaultEvidenceLimit)
	candidates := diffcontext.Candidates(snapshot, changes, candidateLimit)
	return contextDiffResult{
		PackageQuery:   packageQuery,
		RetrievalQuery: diffcontext.RetrievalQuery(packageQuery, changes, candidates),
		Candidates:     candidates,
		Evidence:       diffcontext.Evidence(changes, evidenceLimit),
	}
}
