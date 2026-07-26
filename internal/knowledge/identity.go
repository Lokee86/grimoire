package knowledge

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// Identity returns a stable content identity for the indexed documentation and
// its source-link hints. Repository location and Git metadata are excluded.
func Identity(index Index) string {
	hash := sha256.New()
	_, _ = fmt.Fprintf(hash, "%d\x00", index.Version)
	for _, document := range index.Documents {
		_, _ = fmt.Fprintf(hash, "%s\x00%s\x00%s\x00", document.Path, document.Kind, document.Hash)
		for _, section := range document.Sections {
			_, _ = fmt.Fprintf(hash, "%s\x00%s\x00", section.ID, section.Hash)
			links := append([]CodeLink(nil), section.CodeLinks...)
			sort.Slice(links, func(i, j int) bool {
				if links[i].Kind != links[j].Kind {
					return links[i].Kind < links[j].Kind
				}
				if links[i].Value != links[j].Value {
					return links[i].Value < links[j].Value
				}
				if links[i].SourcePath != links[j].SourcePath {
					return links[i].SourcePath < links[j].SourcePath
				}
				return links[i].Evidence < links[j].Evidence
			})
			for _, link := range links {
				_, _ = fmt.Fprintf(hash, "%s\x00%s\x00%s\x00%s\x00", link.Kind, link.Value, link.SourcePath, link.Evidence)
			}
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}
