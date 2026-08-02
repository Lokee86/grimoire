package main

type headerClause struct {
	end     int
	keyword string
	start   int
}

type annotationReference struct {
	end        int
	expression string
	name       string
	start      int
}

func (parser *javaParser) parseTypeRelationships(sourceID, lexicalOwner string, start, end int) {
	for _, clause := range typeHeaderClauses(parser.tokens, start, end) {
		relation := "extends"
		attributes := map[string]any{"clause": clause.keyword}
		if clause.keyword == "implements" {
			relation = "implements"
		} else if clause.keyword == "permits" {
			relation = "references"
			attributes["role"] = "permitted-subtype"
		}
		for _, segment := range splitTopLevel(parser.tokens, clause.start, clause.end, ",") {
			expression := sourceExcerpt(parser.source, parser.tokens, segment[0], segment[1])
			name, valid := headerTypeName(parser.tokens, segment[0], segment[1])
			referenceSpan := parser.tokenSpan(segment[0], segment[1])
			if !valid {
				parser.facts.addUnresolved(sourceID, relation, expression, "unsupported-form", parser.path, referenceSpan, attributes)
				continue
			}
			parser.queueRelationship(sourceID, relation, expression, name, lexicalOwner, referenceSpan, false, attributes)
		}
	}
}

func typeHeaderClauses(tokens []token, start, end int) []headerClause {
	var positions []int
	paren, bracket, angle := 0, 0, 0
	for index := start; index < end; index++ {
		text := tokens[index].text
		if paren == 0 && bracket == 0 && angle == 0 && (text == "extends" || text == "implements" || text == "permits") {
			positions = append(positions, index)
			continue
		}
		switch text {
		case "(":
			paren++
		case ")":
			paren--
		case "[":
			bracket++
		case "]":
			bracket--
		case "<":
			angle++
		case ">":
			if angle > 0 {
				angle--
			}
		}
	}
	result := make([]headerClause, 0, len(positions))
	for index, position := range positions {
		clauseEnd := end
		if index+1 < len(positions) {
			clauseEnd = positions[index+1]
		}
		result = append(result, headerClause{end: clauseEnd, keyword: tokens[position].text, start: position + 1})
	}
	return result
}

func headerTypeName(tokens []token, start, end int) (string, bool) {
	for start < end && tokens[start].text == "@" {
		_, next, valid := annotationName(tokens, start, end)
		if !valid {
			return "", false
		}
		start = next
		if start < end && tokens[start].text == "(" {
			close := matchingToken(tokens, start, end, "(", ")")
			if close < 0 {
				return "", false
			}
			start = close + 1
		}
	}
	var nameTokens []token
	for index := start; index < end; {
		if !identifierToken(tokens[index].text) {
			return "", false
		}
		nameTokens = append(nameTokens, tokens[index])
		index++
		if index < end && tokens[index].text == "<" {
			close := matchingToken(tokens, index, end, "<", ">")
			if close < 0 {
				return "", false
			}
			index = close + 1
		}
		if index == end {
			return qualifiedTokenName(nameTokens)
		}
		if tokens[index].text != "." {
			return "", false
		}
		nameTokens = append(nameTokens, tokens[index])
		index++
	}
	return "", false
}

func (parser *javaParser) queueAnnotations(targetID, lexicalOwner string, start, end int) {
	for _, annotation := range annotationReferences(parser.source, parser.tokens, start, end) {
		parser.queueRelationship(
			targetID, "annotates", annotation.expression, annotation.name, lexicalOwner,
			parser.tokenSpan(annotation.start, annotation.end), true,
			map[string]any{"expression": annotation.expression},
		)
	}
}

func annotationReferences(source string, tokens []token, start, end int) []annotationReference {
	var result []annotationReference
	for index := start; index < end; index++ {
		if tokens[index].text != "@" || index+1 >= end || tokens[index+1].text == "interface" {
			continue
		}
		name, next, valid := annotationName(tokens, index, end)
		if !valid {
			continue
		}
		annotationEnd := next
		if next < end && tokens[next].text == "(" {
			if close := matchingToken(tokens, next, end, "(", ")"); close >= 0 {
				annotationEnd = close + 1
			}
		}
		result = append(result, annotationReference{
			end: annotationEnd, expression: sourceExcerpt(source, tokens, index, annotationEnd),
			name: name, start: index,
		})
		index = annotationEnd - 1
	}
	return result
}

func annotationName(tokens []token, start, end int) (string, int, bool) {
	if start+1 >= end || tokens[start].text != "@" || !identifierToken(tokens[start+1].text) {
		return "", start, false
	}
	nameEnd := start + 2
	for nameEnd+1 < end && tokens[nameEnd].text == "." && identifierToken(tokens[nameEnd+1].text) {
		nameEnd += 2
	}
	name, valid := qualifiedTokenName(tokens[start+1 : nameEnd])
	return name, nameEnd, valid
}
