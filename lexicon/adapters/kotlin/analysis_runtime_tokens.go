package main

type runtimeInvocation struct {
	arity       int
	calleeStart int
	close       int
	expression  string
	fluent      bool
	name        string
	open        int
	qualifier   string
	span        sourceSpan
	unsupported bool
}

func runtimeInvocations(file *parsedKotlinFile, ranges ...tokenRange) []runtimeInvocation {
	var result []runtimeInvocation
	for _, bounds := range ranges {
		if bounds.end <= bounds.start {
			continue
		}
		ignored := runtimeNestedRanges(file.tokens, bounds)
		for index := bounds.start; index < bounds.end && index < len(file.tokens); index++ {
			if runtimeTokenIgnored(index, ignored) {
				continue
			}
			current := file.tokens[index]
			if current.kind != tokenIdentifier {
				continue
			}
			open := nextRuntimeToken(file.tokens, index+1, bounds.end)
			if open >= bounds.end || file.tokens[open].text != "(" || runtimeControlCall(current.text) {
				continue
			}
			close := matchingRuntimeDelimiter(file.tokens, open, bounds.end, "(", ")")
			if close < 0 {
				continue
			}
			start, unsupported := runtimeCalleeStart(file.tokens, index, bounds.start)
			previous := previousRuntimeToken(file.tokens, start-1, bounds.start)
			if previous >= bounds.start && (file.tokens[previous].text == "fun" || file.tokens[previous].text == "@") {
				continue
			}
			qualifier := ""
			if dot := lastToken(file.tokens, start, index, "."); dot >= 0 {
				qualifier = compactTokenRange(file.tokens, start, dot)
			}
			result = append(result, runtimeInvocation{
				arity: runtimeArgumentArity(file.tokens, open+1, close), calleeStart: start,
				close: close, expression: compactTokenRange(file.tokens, start, close+1),
				fluent: runtimeCallResultIsReceiver(file.tokens, close, bounds.end),
				name:   identifierText(current), open: open, qualifier: qualifier,
				span: runtimeTokenSpan(file.path, file.tokens[start], file.tokens[close]), unsupported: unsupported,
			})
		}
	}
	return result
}

func runtimeCallResultIsReceiver(tokens []token, close, upper int) bool {
	next := nextRuntimeToken(tokens, close+1, upper)
	if next >= upper {
		return false
	}
	if tokens[next].text == "." {
		return true
	}
	if tokens[next].text != "?" {
		return false
	}
	dot := nextRuntimeToken(tokens, next+1, upper)
	return dot < upper && tokens[dot].text == "."
}

func runtimeCalleeStart(tokens []token, name, lower int) (int, bool) {
	start := name
	unsupported := false
	for {
		dot := previousRuntimeToken(tokens, start-1, lower)
		if dot < lower || tokens[dot].text != "." {
			break
		}
		owner := previousRuntimeToken(tokens, dot-1, lower)
		if owner < lower || tokens[owner].kind != tokenIdentifier {
			unsupported = true
			break
		}
		question := previousRuntimeToken(tokens, owner-1, lower)
		if question >= lower && tokens[question].text == "?" {
			unsupported = true
			break
		}
		start = owner
	}
	previous := previousRuntimeToken(tokens, start-1, lower)
	if previous >= lower && (tokens[previous].text == "." || tokens[previous].text == "?") {
		unsupported = true
	}
	return start, unsupported
}

func runtimeArgumentArity(tokens []token, start, end int) int {
	count := 0
	for _, segment := range splitTopLevel(tokens, start, end, ",") {
		left := skipTokenKind(tokens, segment[0], segment[1], tokenNewline)
		right := trimTokenKind(tokens, left, segment[1], tokenNewline)
		if left < right {
			count++
		}
	}
	return count
}

func nextRuntimeToken(tokens []token, index, upper int) int {
	for index < upper && tokens[index].kind == tokenNewline {
		index++
	}
	return index
}

func previousRuntimeToken(tokens []token, index, lower int) int {
	for index >= lower && tokens[index].kind == tokenNewline {
		index--
	}
	return index
}

func matchingRuntimeDelimiter(tokens []token, open, upper int, opening, closing string) int {
	depth := 0
	for index := open; index < upper; index++ {
		switch tokens[index].text {
		case opening:
			depth++
		case closing:
			depth--
			if depth == 0 {
				return index
			}
		}
	}
	return -1
}

func runtimeTokenSpan(path string, first, last token) sourceSpan {
	return sourceSpan{
		EndColumn: last.endColumn, EndLine: last.endLine, Path: path,
		StartColumn: first.startColumn, StartLine: first.startLine,
	}
}

func runtimeControlCall(name string) bool {
	switch name {
	case "catch", "for", "if", "val", "var", "when", "while":
		return true
	}
	return false
}

func runtimeNestedRanges(tokens []token, bounds tokenRange) []tokenRange {
	var result []tokenRange
	for index := bounds.start; index < bounds.end && index < len(tokens); index++ {
		if tokens[index].text == "fun" || tokens[index].text == "class" || tokens[index].text == "object" {
			if nested := runtimeNestedDeclarationRange(tokens, index, bounds.end); nested.end > nested.start {
				result = append(result, nested)
				index = nested.end - 1
				continue
			}
		}
		if tokens[index].text != "{" || !runtimeLambdaBrace(tokens, index, bounds.start, bounds.end) {
			continue
		}
		if close := matchingRuntimeDelimiter(tokens, index, bounds.end, "{", "}"); close >= 0 {
			result = append(result, tokenRange{start: index, end: close + 1})
			index = close
		}
	}
	return result
}

func runtimeNestedDeclarationRange(tokens []token, start, upper int) tokenRange {
	for index := start + 1; index < upper; index++ {
		if tokens[index].text == "{" {
			if close := matchingRuntimeDelimiter(tokens, index, upper, "{", "}"); close >= 0 {
				return tokenRange{start: start, end: close + 1}
			}
			return tokenRange{}
		}
		if tokens[index].text == "=" {
			end := index + 1
			for end < upper && tokens[end].kind != tokenNewline && tokens[end].text != ";" {
				end++
			}
			return tokenRange{start: start, end: end}
		}
	}
	return tokenRange{}
}

func runtimeLambdaBrace(tokens []token, open, lower, upper int) bool {
	if close := matchingRuntimeDelimiter(tokens, open, upper, "{", "}"); close >= 0 {
		for index := open + 1; index < close; index++ {
			if tokens[index].text == "-" {
				next := nextRuntimeToken(tokens, index+1, close)
				if next < close && tokens[next].text == ">" {
					return true
				}
			}
		}
	}
	previous := previousRuntimeToken(tokens, open-1, lower)
	if previous < lower {
		return false
	}
	if tokens[previous].text == "=" {
		return true
	}
	if tokens[previous].kind == tokenIdentifier {
		switch tokens[previous].text {
		case "do", "else", "finally", "try", "when":
			return false
		default:
			return true
		}
	}
	return false
}

func runtimeTokenIgnored(index int, ranges []tokenRange) bool {
	for _, bounds := range ranges {
		if index >= bounds.start && index < bounds.end {
			return true
		}
	}
	return false
}
