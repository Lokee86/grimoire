# Assembly package

`internal/assembly` selects a sufficient deterministic evidence pool for an active automatic retrieval policy.

It receives curated source candidates and composed structural evidence in ranked order. It returns the retained evidence plus an inspectable decision containing candidate counts, candidate tokens, structural counts, represented regions, roles, query facets, and the stop reason.

Coverage-aware planning is the production strategy. Decomposed query intents receive stable facet identities during retrieval, and assembly reserves three distinct ranked candidates per facet before spending the remaining pool on repeated evidence. A candidate may match several facets, but it claims only its strongest still-open facet during coverage planning so one generic chunk cannot satisfy an entire multi-part query.

## Scope behavior

- Focused: stay near an exact or highest-ranked anchor region.
- Bounded: require at least two represented regions and preserve a deep multi-budget reserve.
- Exploratory: require at least three represented regions and preserve broad source and structural alternatives.

Each scope has hard candidate and structural limits. Facet reservation is applied before the normal scope coverage checks. Coverage may stop selection before the complete curated set, but the compiler still performs final token fitting.

## Boundary

Assembly owns scope-specific coverage, distinct facet claims, and reserve rules. It consumes provider-assigned facet ranks but does not score candidates, infer graph relations, or serialize packages. Fixed positive-budget requests bypass this package.
