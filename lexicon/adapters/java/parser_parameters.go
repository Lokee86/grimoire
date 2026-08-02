package main

type parameterDeclaration struct {
	end          int
	name         string
	receiverType string
	start        int
	typeName     string
	varargs      bool
}

func parseParameters(tokens []token, start, end int) []parameterDeclaration {
	if start >= end {
		return []parameterDeclaration{}
	}
	segments := splitTopLevel(tokens, start, end, ",")
	result := make([]parameterDeclaration, 0, len(segments))
	for _, segment := range segments {
		core := skipPrefix(tokens, segment[0], segment[1])
		nameIndex := variableName(tokens, core, segment[1])
		if nameIndex < 0 || nameIndex == core {
			return nil
		}
		typeTokens := make([]token, 0, nameIndex-core)
		for index := core; index < segment[1]; index++ {
			if index != nameIndex {
				typeTokens = append(typeTokens, tokens[index])
			}
		}
		typeName := normalizedTokens(typeTokens)
		if typeName == "" {
			return nil
		}
		result = append(result, parameterDeclaration{
			end: segment[1], name: tokens[nameIndex].text,
			receiverType: receiverDeclaredType(tokens, core, nameIndex),
			start:        segment[0], typeName: typeName,
			varargs: containsToken(tokens, core, nameIndex, "..."),
		})
	}
	return result
}

func callableParameters(tokens []token, start, end int) (int, int) {
	bestOpen, bestClose := -1, -1
	for index := start; index < end; index++ {
		if tokens[index].text != "(" || index == start || !identifierToken(tokens[index-1].text) {
			continue
		}
		close := matchingToken(tokens, index, end, "(", ")")
		if close >= 0 {
			bestOpen, bestClose = index, close
			index = close
		}
	}
	return bestOpen, bestClose
}

func findMemberDelimiter(tokens []token, start, end int) (int, string) {
	paren, bracket := 0, 0
	hasEquals := false
	for index := start; index < end; index++ {
		switch tokens[index].text {
		case "(":
			paren++
		case ")":
			paren--
		case "[":
			bracket++
		case "]":
			bracket--
		case "=":
			if paren == 0 && bracket == 0 {
				hasEquals = true
			}
		case "{":
			if paren == 0 && bracket == 0 {
				if hasEquals {
					close := matchingToken(tokens, index, end, "{", "}")
					if close < 0 {
						return index, "{"
					}
					index = close
					continue
				}
				return index, "{"
			}
		case ";":
			if paren == 0 && bracket == 0 {
				return index, ";"
			}
		}
	}
	return -1, ""
}

func findTypeBody(tokens []token, start, end int) int {
	paren, bracket := 0, 0
	for index := start; index < end; index++ {
		switch tokens[index].text {
		case "(":
			paren++
		case ")":
			paren--
		case "[":
			bracket++
		case "]":
			bracket--
		case "{":
			if paren == 0 && bracket == 0 {
				return index
			}
		case ";":
			if paren == 0 && bracket == 0 {
				return -1
			}
		}
	}
	return -1
}

func enumMemberStart(tokens []token, start, end int) int {
	paren, bracket, braces := 0, 0, 0
	for index := start; index < end; index++ {
		switch tokens[index].text {
		case "(":
			paren++
		case ")":
			paren--
		case "[":
			bracket++
		case "]":
			bracket--
		case "{":
			braces++
		case "}":
			braces--
		case ";":
			if paren == 0 && bracket == 0 && braces == 0 {
				return index + 1
			}
		}
	}
	return end
}

func nextTopLevelTerminator(tokens []token, start, end int) int {
	paren, bracket, braces := 0, 0, 0
	for index := start; index < end; index++ {
		switch tokens[index].text {
		case "(":
			paren++
		case ")":
			paren--
		case "[":
			bracket++
		case "]":
			bracket--
		case "{":
			braces++
		case "}":
			if braces == 0 {
				return index + 1
			}
			braces--
		case ";":
			if paren == 0 && bracket == 0 && braces == 0 {
				return index + 1
			}
		}
	}
	return end
}

func matchingToken(tokens []token, open, end int, left, right string) int {
	depth := 0
	for index := open; index < end; index++ {
		if tokens[index].text == left {
			depth++
		} else if tokens[index].text == right {
			depth--
			if depth == 0 {
				return index
			}
		}
	}
	return -1
}
