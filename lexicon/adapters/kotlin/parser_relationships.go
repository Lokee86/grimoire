package main

func (state *parser) findTypeHeaderEnd() int {
	depth := delimiterDepth{}
	for index := state.index; index < len(state.tokens); index++ {
		current := state.tokens[index]
		if depth.zero() {
			if current.kind == tokenEOF || current.text == "{" || current.text == ";" {
				return index
			}
			if current.kind == tokenNewline {
				next := state.nextNonNewline(index + 1)
				if next < len(state.tokens) && state.tokens[next].text == "{" {
					return next
				}
				previous := previousNonNewline(state.tokens, index-1)
				if previous < 0 || (state.tokens[previous].text != "," && state.tokens[previous].text != ":") {
					return index
				}
			}
		}
		depth.update(current.text)
	}
	return len(state.tokens) - 1
}

func (state *parser) parseSupertypes(start, end int) []supertypeDecl {
	var supertypes []supertypeDecl
	if where := firstTopLevelToken(state.tokens, start, end, "where"); where >= 0 {
		end = where
	}
	for _, bounds := range splitTopLevel(state.tokens, start, end, ",") {
		entryStart := skipTokenKind(state.tokens, bounds[0], bounds[1], tokenNewline)
		entryEnd := trimTokenKind(state.tokens, entryStart, bounds[1], tokenNewline)
		if entryStart >= entryEnd {
			continue
		}
		by := firstTopLevelToken(state.tokens, entryStart, entryEnd, "by")
		typeEnd := entryEnd
		if by >= 0 {
			typeEnd = by
		}
		supertype := supertypeDecl{
			expression: state.tokenText(entryStart, entryEnd),
			span:       state.span(entryStart, entryEnd-1),
			targetName: relationshipTypeName(state.tokens, entryStart, typeEnd),
		}
		if by >= 0 {
			supertype.delegated = true
			supertype.delegateExpression = state.tokenText(by+1, entryEnd)
		}
		supertypes = append(supertypes, supertype)
	}
	return supertypes
}

func relationshipTypeName(tokens []token, start, end int) string {
	var parts []string
	angleDepth := 0
	for index := start; index < end; index++ {
		current := tokens[index]
		if current.kind == tokenNewline {
			continue
		}
		if current.text == "<" {
			angleDepth++
			continue
		}
		if current.text == ">" && angleDepth > 0 {
			angleDepth--
			continue
		}
		if angleDepth > 0 {
			continue
		}
		if current.text == "(" {
			break
		}
		if current.kind != tokenIdentifier && current.text != "." {
			return ""
		}
		if current.kind == tokenIdentifier {
			parts = append(parts, identifierText(current))
		} else {
			parts = append(parts, current.text)
		}
	}
	name := compactTokens(parts)
	if name == "" || name[0] == '.' || name[len(name)-1] == '.' {
		return ""
	}
	return name
}

func previousNonNewline(tokens []token, index int) int {
	for index >= 0 && tokens[index].kind == tokenNewline {
		index--
	}
	return index
}
