package main

import "strings"

func maskComments(content, language string) string {
	var output strings.Builder
	output.Grow(len(content))
	blockEnd := ""
	quote := byte(0)
	escaped := false

	for index := 0; index < len(content); {
		if blockEnd != "" {
			if strings.HasPrefix(content[index:], blockEnd) {
				output.WriteString(strings.Repeat(" ", len(blockEnd)))
				index += len(blockEnd)
				blockEnd = ""
				continue
			}
			if content[index] == '\n' || content[index] == '\r' {
				output.WriteByte(content[index])
			} else {
				output.WriteByte(' ')
			}
			index++
			continue
		}

		current := content[index]
		if quote != 0 {
			output.WriteByte(current)
			if escaped {
				escaped = false
			} else if current == '\\' {
				escaped = true
			} else if current == quote {
				quote = 0
			}
			index++
			continue
		}
		if current == '\'' || current == '"' || current == '`' {
			quote = current
			output.WriteByte(current)
			index++
			continue
		}

		if usesCComments(language) && strings.HasPrefix(content[index:], "/*") {
			output.WriteString("  ")
			index += 2
			blockEnd = "*/"
			continue
		}
		if language == "ps1" && strings.HasPrefix(content[index:], "<#") {
			output.WriteString("  ")
			index += 2
			blockEnd = "#>"
			continue
		}
		if language == "lua" && strings.HasPrefix(content[index:], "--[[") {
			output.WriteString("    ")
			index += 4
			blockEnd = "]]"
			continue
		}
		if usesSlashLineComments(language) && strings.HasPrefix(content[index:], "//") {
			index = maskLine(content, index, &output)
			continue
		}
		if usesDashLineComments(language) && strings.HasPrefix(content[index:], "--") {
			index = maskLine(content, index, &output)
			continue
		}
		if usesHashLineComments(language) && current == '#' {
			index = maskLine(content, index, &output)
			continue
		}
		output.WriteByte(current)
		index++
	}
	return output.String()
}

func maskLine(content string, index int, output *strings.Builder) int {
	for index < len(content) && content[index] != '\n' && content[index] != '\r' {
		output.WriteByte(' ')
		index++
	}
	return index
}

func maskStrings(content string) string {
	var output strings.Builder
	output.Grow(len(content))
	quote := byte(0)
	escaped := false
	for index := 0; index < len(content); index++ {
		current := content[index]
		if quote == 0 {
			if current == '\'' || current == '"' || current == '`' {
				quote = current
				output.WriteByte(' ')
			} else {
				output.WriteByte(current)
			}
			continue
		}
		if current == '\n' || current == '\r' {
			output.WriteByte(current)
			continue
		}
		output.WriteByte(' ')
		if escaped {
			escaped = false
			continue
		}
		if current == '\\' {
			escaped = true
			continue
		}
		if current == quote {
			if index+1 < len(content) && content[index+1] == quote {
				output.WriteByte(' ')
				index++
				continue
			}
			quote = 0
		}
	}
	return output.String()
}

func usesCComments(language string) bool {
	switch language {
	case "c", "cc", "cpp", "h", "hh", "hpp", "cs", "java", "kt", "kts", "swift", "m", "mm", "dart", "groovy", "scala", "sol", "proto", "php", "sql", "v", "sv", "zig":
		return true
	default:
		return false
	}
}

func usesSlashLineComments(language string) bool {
	switch language {
	case "c", "cc", "cpp", "h", "hh", "hpp", "cs", "java", "kt", "kts", "swift", "m", "mm", "dart", "groovy", "scala", "sol", "proto", "php", "v", "sv", "zig":
		return true
	default:
		return false
	}
}

func usesDashLineComments(language string) bool {
	switch language {
	case "lua", "sql", "hs", "lhs", "elm":
		return true
	default:
		return false
	}
}

func usesHashLineComments(language string) bool {
	switch language {
	case "sh", "bash", "fish", "ps1", "pl", "pm", "r", "rb", "cr", "nim", "nims":
		return true
	default:
		return false
	}
}
