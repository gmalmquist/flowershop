package dental

import (
  "fmt"
  "strings"
)

func CountIndent(line string, tts int) int {
	if strings.TrimSpace(line) == "" {
		return -1
	}
	count := 0
	for _, c := range line {
		if c == ' ' {
			count += 1
		} else if c == '\t' {
			count += tts
		} else {
			break
		}
	}
	return count
}

func Dedent(text string) string {
	lines := strings.Split(text, "\n")
	indent := -1
	for _, line := range lines {
		count := CountIndent(line, 2)
		if count < 0 {
			continue
		}
		if indent < 0 || count < indent {
			indent = count
		}
	}
	if indent <= 0 {
		return text
	}
	trailingNewline := false
	for i, _ := range lines {
		blank := lines[i] == strings.TrimSpace(lines[i])
		trailingNewline = blank
		if blank {
			lines[i] = ""
			continue
		}
		if len(lines[i]) >= indent {
			lines[i] = lines[i][indent:]
		}
	}
	text = strings.TrimSpace(strings.Join(lines, "\n"))
	if trailingNewline {
		text = fmt.Sprintf("%v\n", text)
	}
	return text
}

