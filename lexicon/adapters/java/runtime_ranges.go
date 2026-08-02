package main

func ignoredRuntimeTokens(callable *callableDeclaration) map[int]bool {
	ignored := make(map[int]bool)
	tokens := callable.tokens
	for index := callable.bodyStart; index < callable.bodyEnd; index++ {
		if localTypeKeyword(tokens, index) {
			open := findTypeBody(tokens, index+1, callable.bodyEnd)
			if open >= 0 {
				if close := matchingToken(tokens, open, callable.bodyEnd, "{", "}"); close >= 0 {
					markIgnored(ignored, index, close+1)
					index = close
					continue
				}
			}
		}
		if tokens[index].text == "{" && anonymousClassBody(tokens, callable.bodyStart, index) {
			if close := matchingToken(tokens, index, callable.bodyEnd, "{", "}"); close >= 0 {
				markIgnored(ignored, index, close+1)
				index = close
				continue
			}
		}
		if index+1 < callable.bodyEnd && tokens[index].text == "-" && tokens[index+1].text == ">" {
			end := lambdaBodyEnd(tokens, index+2, callable.bodyEnd)
			markIgnored(ignored, index+2, end)
			if end > index+2 {
				index = end - 1
			}
		}
	}
	return ignored
}

func localTypeKeyword(tokens []token, index int) bool {
	if index < 0 || index >= len(tokens) {
		return false
	}
	if index > 0 && tokens[index-1].text == "." {
		return false
	}
	text := tokens[index].text
	return text == "class" || text == "interface" || text == "enum" || text == "record" ||
		(text == "@" && index+1 < len(tokens) && tokens[index+1].text == "interface")
}

func anonymousClassBody(tokens []token, bodyStart, brace int) bool {
	if brace <= bodyStart || tokens[brace-1].text != ")" {
		return false
	}
	open := matchingOpenToken(tokens, brace-1, bodyStart, "(", ")")
	if open <= bodyStart || !identifierToken(tokens[open-1].text) {
		return false
	}
	start := qualifiedChainStart(tokens, open-1)
	return start > bodyStart && tokens[start-1].text == "new"
}

func matchingOpenToken(tokens []token, close, start int, left, right string) int {
	depth := 0
	for index := close; index >= start; index-- {
		if tokens[index].text == right {
			depth++
		} else if tokens[index].text == left {
			depth--
			if depth == 0 {
				return index
			}
		}
	}
	return -1
}

func lambdaBodyEnd(tokens []token, start, end int) int {
	if start < end && tokens[start].text == "{" {
		if close := matchingToken(tokens, start, end, "{", "}"); close >= 0 {
			return close + 1
		}
	}
	paren, bracket, braces := 0, 0, 0
	for index := start; index < end; index++ {
		switch tokens[index].text {
		case "(":
			paren++
		case ")":
			if paren == 0 && bracket == 0 && braces == 0 {
				return index
			}
			paren--
		case "[":
			bracket++
		case "]":
			bracket--
		case "{":
			braces++
		case "}":
			braces--
		case ",", ";":
			if paren == 0 && bracket == 0 && braces == 0 {
				return index
			}
		}
	}
	return end
}

func markIgnored(ignored map[int]bool, start, end int) {
	for index := start; index < end; index++ {
		ignored[index] = true
	}
}
