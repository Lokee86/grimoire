# Damage resolution

The combat pipeline applies a "shield gate" before health loss. The implementation lives in `internal/damage/resolve.go` and returns `ERR_SHIELD_DEPLETED` when a hit consumes the remaining shield without overflowing into health.
