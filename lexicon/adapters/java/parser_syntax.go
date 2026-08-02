package main

import "sort"

func skipPrefix(tokens []token, start, end int) int {
	index := start
	for index < end {
		if tokens[index].text == "@" && index+1 < end && tokens[index+1].text != "interface" {
			index += 2
			for index+1 < end && tokens[index].text == "." && identifierToken(tokens[index+1].text) {
				index += 2
			}
			if index < end && tokens[index].text == "(" {
				close := matchingToken(tokens, index, end, "(", ")")
				if close < 0 {
					return index
				}
				index = close + 1
			}
			continue
		}
		if _, modifier := javaModifiers[tokens[index].text]; modifier {
			index++
			continue
		}
		if index+2 < end && tokens[index].text == "non" && tokens[index+1].text == "-" && tokens[index+2].text == "sealed" {
			index += 3
			continue
		}
		break
	}
	return index
}

func modifierList(tokens []token, start, end int) []string {
	values := make(map[string]struct{})
	for index := start; index < end; index++ {
		if _, modifier := javaModifiers[tokens[index].text]; modifier {
			values[tokens[index].text] = struct{}{}
		}
		if index+2 < end && tokens[index].text == "non" && tokens[index+1].text == "-" && tokens[index+2].text == "sealed" {
			values["non-sealed"] = struct{}{}
		}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func typeKeywordAt(tokens []token, index int) (string, string, bool) {
	if index < 0 || index >= len(tokens) {
		return "", "", false
	}
	switch tokens[index].text {
	case "class":
		return "class", "class", true
	case "interface":
		return "interface", "interface", true
	case "enum":
		return "enum", "enum", true
	case "record":
		return "record", "record", true
	case "@":
		if index+1 < len(tokens) && tokens[index+1].text == "interface" {
			return "annotation", "annotation", true
		}
	}
	return "", "", false
}

func variableName(tokens []token, start, end int) int {
	limit := end
	paren, bracket, braces, angle := 0, 0, 0, 0
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
		case "<":
			angle++
		case ">":
			if angle > 0 {
				angle--
			}
		case "=":
			if paren == 0 && bracket == 0 && braces == 0 && angle == 0 {
				limit = index
				index = end
			}
		}
	}
	if limit > start && tokens[limit-1].text == "this" {
		return limit - 1
	}
	for index := limit - 1; index >= start; index-- {
		if identifierToken(tokens[index].text) {
			return index
		}
	}
	return -1
}

func splitTopLevel(tokens []token, start, end int, separator string) [][2]int {
	if start >= end {
		return nil
	}
	var result [][2]int
	segmentStart := start
	paren, bracket, braces, angle := 0, 0, 0, 0
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
		case "<":
			angle++
		case ">":
			if angle > 0 {
				angle--
			}
		default:
			if tokens[index].text == separator && paren == 0 && bracket == 0 && braces == 0 && angle == 0 {
				result = append(result, [2]int{segmentStart, index})
				segmentStart = index + 1
			}
		}
	}
	if segmentStart < end {
		result = append(result, [2]int{segmentStart, end})
	}
	return result
}
