package interstack

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"
)

func syntheticPath(category, identity string) string {
	digest := sha256.Sum256([]byte(identity))
	return "@interstack/" + category + "/" + hex.EncodeToString(digest[:8])
}

func camelize(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == '_' || r == '-' || r == '/' })
	for index, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		parts[index] = string(runes)
	}
	return strings.Join(parts, "")
}

func lastIdentifier(value string) string {
	value = strings.TrimSpace(value)
	if index := strings.Index(value, "("); index >= 0 {
		value = value[:index]
	}
	value = strings.TrimSpace(value)
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_')
	})
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
