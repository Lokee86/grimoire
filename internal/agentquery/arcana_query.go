package agentquery

import (
	"context"

	"github.com/Lokee86/grimoire/internal/arcanagraph"
	"github.com/Lokee86/grimoire/internal/structure"
)

type arcanaQuery interface {
	Resolve(context.Context, string, string, string, int) ([]structure.Node, error)
	ResolveTyped(context.Context, string, string, string, string, int) ([]structure.Node, error)
	Inspect(context.Context, string, uint32) (structure.Node, error)
	NeighborsBatch(context.Context, string, []uint32, string, []string) (map[uint32][]arcanagraph.QueryNeighbor, error)
	Paths(context.Context, string, uint32, uint32, []string, int, int) ([]arcanagraph.QueryPath, bool, error)
	ImpactQuery(context.Context, string, uint32, string, []string, int, int) ([]arcanagraph.QueryImpact, bool, error)
	Unresolved(context.Context, string, uint32, int) ([]structure.Unresolved, bool, error)
}

func (engine *Engine) openArcanaQuery(ctx context.Context, response *Response) (arcanaQuery, func()) {
	if engine.arcanaSnapshot == "" {
		return nil, func() {}
	}
	if engine.residentArcana != nil && !engine.residentArcana.Closed() {
		return engine.residentArcana, func() {}
	}
	session, err := engine.arcana.OpenSession(ctx, engine.arcanaSnapshot)
	if err != nil {
		response.Warnings = append(response.Warnings, "Arcana protocol session unavailable: "+err.Error())
		return nil, func() {}
	}
	return session, func() {
		if err := session.Close(); err != nil {
			response.Warnings = append(response.Warnings, "close Arcana protocol session: "+err.Error())
		}
	}
}
