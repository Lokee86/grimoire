package spandiscovery

type bracePair struct {
	openOffset  int
	closeOffset int
	openLine    int
	closeLine   int
	depth       int
}

type braceFrame struct {
	offset int
	line   int
	depth  int
}

type scanState int

const (
	scanNormal scanState = iota
	scanLineComment
	scanBlockComment
	scanSingleQuote
	scanDoubleQuote
	scanBacktick
)

func scanBracePairs(content string) ([]bracePair, []int) {
	pairs := make([]bracePair, 0)
	stack := make([]braceFrame, 0)
	lineDepths := []int{0}
	line := 1
	depth := 0
	state := scanNormal
	escaped := false

	for index := 0; index < len(content); index++ {
		current := content[index]
		if current == '\n' {
			if state == scanLineComment {
				state = scanNormal
			}
			line++
			lineDepths = append(lineDepths, depth)
			if state == scanSingleQuote || state == scanDoubleQuote {
				state = scanNormal
			}
			escaped = false
			continue
		}

		switch state {
		case scanLineComment:
			continue
		case scanBlockComment:
			if current == '*' && index+1 < len(content) && content[index+1] == '/' {
				state = scanNormal
				index++
			}
			continue
		case scanSingleQuote, scanDoubleQuote, scanBacktick:
			if escaped {
				escaped = false
				continue
			}
			if current == '\\' {
				escaped = true
				continue
			}
			if (state == scanSingleQuote && current == '\'') ||
				(state == scanDoubleQuote && current == '"') ||
				(state == scanBacktick && current == '`') {
				state = scanNormal
			}
			continue
		}

		if current == '/' && index+1 < len(content) {
			switch content[index+1] {
			case '/':
				state = scanLineComment
				index++
				continue
			case '*':
				state = scanBlockComment
				index++
				continue
			}
		}
		switch current {
		case '"':
			state = scanDoubleQuote
		case '`':
			state = scanBacktick
		case '\'':
			if isCharacterLiteral(content, index) {
				state = scanSingleQuote
			}
		case '{':
			stack = append(stack, braceFrame{offset: index, line: line, depth: depth})
			depth++
		case '}':
			if len(stack) == 0 {
				continue
			}
			depth--
			frame := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			pairs = append(pairs, bracePair{
				openOffset: frame.offset, closeOffset: index,
				openLine: frame.line, closeLine: line, depth: frame.depth,
			})
		}
	}
	return pairs, lineDepths
}

func isCharacterLiteral(content string, start int) bool {
	limit := min(start+8, len(content))
	escaped := false
	for index := start + 1; index < limit; index++ {
		current := content[index]
		if current == '\n' {
			return false
		}
		if escaped {
			escaped = false
			continue
		}
		if current == '\\' {
			escaped = true
			continue
		}
		if current == '\'' {
			return true
		}
	}
	return false
}
