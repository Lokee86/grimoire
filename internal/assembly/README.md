# Assembly package

`internal/assembly` selects a sufficient deterministic evidence pool for an active automatic retrieval policy.

It receives curated source candidates and composed structural evidence in ranked order. It returns the retained evidence plus an inspectable decision containing candidate counts, candidate tokens, structural counts, represented regions and roles, and the stop reason.

## Scope behavior

- Focused: stay near an exact or highest-ranked anchor region.
- Bounded: require at least two represented regions and preserve a deep multi-budget reserve.
- Exploratory: require at least three represented regions and preserve broad source and structural alternatives.

Each scope has hard candidate and structural limits. Coverage may stop selection before the complete curated set, but the compiler still performs final token fitting.

## Boundary

Assembly owns scope-specific coverage and reserve rules. It does not score candidates, infer graph relations, or serialize packages. Fixed positive-budget requests bypass this package.
