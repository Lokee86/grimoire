package retrieve

import (
	"regexp"
	"strings"
)

type exactSignal struct {
	value, kind, label string
	weight             float64
}

var (
	versionSignal = regexp.MustCompile(`^(?:v|V|go)?[0-9]+(?:\.[0-9]+)+(?:[-+][A-Za-z0-9.-]+)?$`)
	errorSignal   = regexp.MustCompile(`^(?:[0-9]{3}|0[xX][0-9A-Fa-f]+|(?:ERR|ERROR)[-_0-9A-Z]+|E(?:[0-9_][A-Za-z0-9_-]*|[A-Z][A-Z0-9_]+))$`)
	keyPart       = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)
)

func exactReason(signal exactSignal, location string) string {
	label := signal.label
	if label == "" {
		label = signal.value
	}
	return signal.kind + " \"" + label + "\" matches " + location
}

func exactSignals(query string) []exactSignal {
	plain := []byte(query)
	var signals []exactSignal
	for i := 0; i < len(plain); i++ {
		if !strings.ContainsRune("\"'`", rune(plain[i])) {
			continue
		}
		quote, end := plain[i], i+1
		for end < len(plain) && plain[end] != quote {
			end++
		}
		if end == len(plain) {
			continue
		}
		if value := strings.TrimSpace(string(plain[i+1 : end])); value != "" {
			signals = addSignal(signals, exactSignal{value: value, kind: "quoted phrase", weight: 100})
		}
		for j := i; j <= end; j++ {
			plain[j] = ' '
		}
		i = end
	}
	for _, field := range strings.Fields(string(plain)) {
		field = strings.Trim(field, "()[]{}<>,;.!?")
		if field == "" {
			continue
		}
		if at := strings.IndexByte(field, '='); at > 0 {
			signals = addConfigSignals(signals, strings.TrimPrefix(field[:at], "--"))
			continue
		}
		if at := strings.IndexByte(field, ':'); at > 0 && !strings.ContainsAny(field[:at], "/\\") && isConfigKey(field[:at]) {
			signals = addConfigSignals(signals, field[:at])
			continue
		}
		if signal, ok := classifySignal(field); ok {
			signals = addSignal(signals, signal)
			if signal.kind == "configuration key" {
				signals = addTerminalSignal(signals, signal.value)
			}
		}
	}
	return signals
}

func addConfigSignals(signals []exactSignal, value string) []exactSignal {
	if !isConfigKey(value) {
		return signals
	}
	signals = addSignal(signals, exactSignal{value: value, kind: "configuration key", weight: 80})
	return addTerminalSignal(signals, value)
}

func addTerminalSignal(signals []exactSignal, value string) []exactSignal {
	at := strings.LastIndexByte(value, '.')
	if at < 1 || !keyPart.MatchString(value[at+1:]) {
		return signals
	}
	return addSignal(signals, exactSignal{value: value[at+1:], kind: "configuration key", label: value, weight: 79})
}

func addSignal(signals []exactSignal, signal exactSignal) []exactSignal {
	for _, existing := range signals {
		if existing.value == signal.value && existing.kind == signal.kind {
			return signals
		}
	}
	return append(signals, signal)
}

func classifySignal(value string) (exactSignal, bool) {
	value = strings.Trim(value, "()[]{}<>,;.!?")
	switch {
	case versionSignal.MatchString(value):
		return exactSignal{value: value, kind: "version string", weight: 85}, true
	case errorSignal.MatchString(value):
		return exactSignal{value: value, kind: "error code", weight: 90}, true
	case isConfigKey(value):
		return exactSignal{value: value, kind: "configuration key", weight: 80}, true
	case isPath(value):
		kind := "filename"
		if strings.ContainsAny(value, "/\\") {
			kind = "path"
		}
		return exactSignal{value: value, kind: kind, weight: 75}, true
	case isIdentifier(value):
		return exactSignal{value: value, kind: "identifier", weight: 70}, true
	default:
		return exactSignal{}, false
	}
}

func isPath(value string) bool {
	if strings.ContainsAny(value, "/\\") {
		return true
	}
	dot := strings.LastIndexByte(value, '.')
	return dot > 0 && dot < len(value)-1 && keyPart.MatchString(value[dot+1:])
}

func isConfigKey(value string) bool {
	if value == "" || strings.ContainsAny(value, "/\\") {
		return false
	}
	parts := strings.Split(value, ".")
	if len(parts) < 2 || knownExtension(parts[len(parts)-1]) {
		return false
	}
	for _, part := range parts {
		if !keyPart.MatchString(part) {
			return false
		}
	}
	return true
}

func knownExtension(value string) bool {
	return strings.Contains(",c,cc,cpp,css,go,h,hpp,html,java,js,json,md,mod,py,rs,sum,toml,ts,txt,xml,yaml,yml,", ","+strings.ToLower(value)+",")
}

func isIdentifier(value string) bool {
	return keyPart.MatchString(value) && strings.ContainsAny(value, "_ABCDEFGHIJKLMNOPQRSTUVWXYZ")
}

func exactContains(text, value, kind string) bool {
	if kind == "quoted phrase" || kind == "path" || kind == "filename" {
		return strings.Contains(text, value)
	}
	for offset := 0; offset < len(text); {
		at := strings.Index(text[offset:], value)
		if at < 0 {
			return false
		}
		start := offset + at
		if exactBoundary(text, start, start+len(value)) {
			return true
		}
		offset = start + 1
	}
	return false
}

func exactBoundary(text string, start, end int) bool {
	word := func(b byte) bool {
		return b == '_' || b >= '0' && b <= '9' || b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z'
	}
	return (start == 0 || !word(text[start-1])) && (end >= len(text) || !word(text[end]))
}
