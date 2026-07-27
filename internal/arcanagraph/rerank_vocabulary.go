package arcanagraph

var hybridTokenAliases = map[string]string{
	"embedded": "embed", "embedding": "embed", "embeddings": "embed",
	"rendered": "render", "rendering": "render",
	"ranked": "rank", "ranking": "rank", "ordered": "order", "ordering": "order",
	"resolved": "resolve", "resolving": "resolve", "resolution": "resolve",
	"relations": "relation", "relationship": "relation", "relationships": "relation",
	"dependencies": "dependent", "dependency": "dependent", "dependents": "dependent",
	"incoming": "inbound", "outgoing": "outbound",
	"uncertain": "unresolved", "uncertainty": "unresolved",
	"limited": "bound", "bounded": "bound", "limits": "bound",
	"callers": "call", "callee": "call", "callees": "call", "execution": "call",
}

var hybridGenericTokens = map[string]bool{
	"anchor": true, "app": true, "call": true, "embed": true, "entry": true,
	"execute": true, "graph": true, "identity": true, "match": true, "node": true,
	"pack": true, "path": true, "query": true, "relation": true, "resolve": true,
	"search": true, "state": true,
}

var hybridStopWords = map[string]bool{
	"and": true, "are": true, "does": true, "for": true, "from": true,
	"how": true, "into": true, "then": true, "the": true, "this": true,
	"was": true, "were": true, "what": true, "when": true, "where": true,
	"which": true, "with": true, "without": true,
	"internal": true, "src": true, "test": true, "tests": true,
}
