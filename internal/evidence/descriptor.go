package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Intent identifies the retrieval purpose a candidate serves.
type Intent string

const (
	IntentDirectLocation Intent = "direct_location"
	IntentMechanism      Intent = "mechanism"
	IntentCallChain      Intent = "call_chain"
	IntentArchitecture   Intent = "architecture"
	IntentMixed          Intent = "mixed"
)

// Role identifies how a candidate contributes to an evidence group.
type Role string

const (
	RolePrimary    Role = "primary"
	RoleSupporting Role = "supporting"
	RoleStructural Role = "structural"
	RoleContext    Role = "context"
)

// Link connects a candidate to another evidence identity.
type Link struct {
	Identity string `json:"identity"`
	Relation string `json:"relation,omitempty"`
	Required bool   `json:"required,omitempty"`
}

// Descriptor is the provider-neutral coordination contract shared by source
// candidates and structural evidence. Producers add the fields they own;
// package assembly consumes the combined descriptor without depending on a
// specific retrieval or structural provider.
type Descriptor struct {
	Identity           string   `json:"identity,omitempty"`
	Intents            []Intent `json:"intents,omitempty"`
	Roles              []Role   `json:"roles,omitempty"`
	GroupIDs           []string `json:"group_ids,omitempty"`
	ExactMatchStrength float64  `json:"exact_match_strength,omitempty"`
	EstimatedTokens    int      `json:"estimated_tokens,omitempty"`
	RedundancyKey      string   `json:"redundancy_key,omitempty"`
	Links              []Link   `json:"links,omitempty"`
}

// RangeIdentity returns a stable identity for one prepared source range.
func RangeIdentity(path string, startLine, endLine int) string {
	return "range:" + normalizedPath(path) + ":" +
		strconv.Itoa(startLine) + ":" + strconv.Itoa(endLine)
}

// StableID returns a deterministic namespaced identifier for ordered parts.
func StableID(namespace string, parts ...string) string {
	hash := sha256.New()
	hash.Write([]byte(strings.TrimSpace(namespace)))
	for _, part := range parts {
		hash.Write([]byte{0})
		hash.Write([]byte(strings.TrimSpace(part)))
	}
	sum := hash.Sum(nil)
	return strings.TrimSpace(namespace) + ":" + hex.EncodeToString(sum[:8])
}

// Merge combines descriptors without discarding metadata contributed by
// another provider. Scalar conflicts use the stronger or more conservative
// value; set-like fields retain stable first-seen order.
func Merge(left, right Descriptor) Descriptor {
	merged := clone(left)
	if merged.Identity == "" {
		merged.Identity = right.Identity
	}
	merged.Intents = appendUnique(merged.Intents, right.Intents...)
	merged.Roles = appendUnique(merged.Roles, right.Roles...)
	merged.GroupIDs = appendUnique(merged.GroupIDs, right.GroupIDs...)
	if right.ExactMatchStrength > merged.ExactMatchStrength {
		merged.ExactMatchStrength = right.ExactMatchStrength
	}
	if right.EstimatedTokens > merged.EstimatedTokens {
		merged.EstimatedTokens = right.EstimatedTokens
	}
	if merged.RedundancyKey == "" {
		merged.RedundancyKey = right.RedundancyKey
	}
	for _, link := range right.Links {
		if !slices.Contains(merged.Links, link) {
			merged.Links = append(merged.Links, link)
		}
	}
	return merged
}

func clone(descriptor Descriptor) Descriptor {
	descriptor.Intents = slices.Clone(descriptor.Intents)
	descriptor.Roles = slices.Clone(descriptor.Roles)
	descriptor.GroupIDs = slices.Clone(descriptor.GroupIDs)
	descriptor.Links = slices.Clone(descriptor.Links)
	return descriptor
}

func appendUnique[T comparable](existing []T, values ...T) []T {
	for _, value := range values {
		if !slices.Contains(existing, value) {
			existing = append(existing, value)
		}
	}
	return existing
}

func normalizedPath(path string) string {
	return strings.ReplaceAll(filepath.ToSlash(filepath.Clean(path)), "\\", "/")
}
