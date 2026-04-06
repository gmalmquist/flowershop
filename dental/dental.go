// String manipulation for text indentation.
package dental

import (
  "fmt"
  "strings"
)

var DefaultTabsToSpaces int = 2

// Contains methods for indenting and dedenting text
type Dental struct {
  tts int // how many spaces are in a single indentation level. default is 2
}

// Creates a new Dental struct for indenting and dedenting text.
//
// `tts`: The tabs-to-spaces count, i.e. the number of spaces per indentation level.
func New(tts int) *Dental {
  if tts <= 0 {
    tts = DefaultTabsToSpaces
  }
  return &Dental{
    tts: tts,
  }
}

// The tabs-to-spaces count, i.e. the number of spaces per indentation level.
func (d *Dental) TTS() int {
  if d.tts <= 0 {
    return DefaultTabsToSpaces
  }
  return d.tts
}

// Counts the leading spaces of a single line.
// If the line is blank, returns -1.
func (d *Dental) LineLeadingSpaces(line string) int {
  if strings.TrimSpace(line) == "" {
    return -1
  }
  tts := d.TTS()
  spaces := 0
  for _, c := range line {
    if c == ' ' {
      spaces += 1
    } else if c == '\t' {
      spaces += tts
    } else {
      break
    }
  }
  return spaces
}

// Gets the minimum leading spaces of the given lines, in spaces.
//
// Empty lines are disregarded.
// Lines containing the '\n' rune will be split.
// If no non-empty lines are given, returns -1
func (d *Dental) MinLeadingSpaces(lines []string) int {
  min := -1
  for _, xline := range lines {
    spaces := -1
    if strings.Contains(xline, "\n") {
      spaces = d.MinLeadingSpaces(strings.Split(xline, "\n"))
    } else {
      spaces = d.LineLeadingSpaces(xline)
    }
    if spaces < 0 {
      continue
    }
    if min < 0 || spaces < min {
      min = spaces
    }
  }
	return min
}

// Gets the minimum indentation level of all given lines.
//
// Empty lines are disregarded.
// Lines containing the '\n' rune will be split.
// If no non-empty lines are given, returns -1
//
// The second return value is the remainder of the minimum number of leading spaces is not a multiple of the configured spaces per indentation level.
func (d *Dental) IndentationLevel(lines []string) (int, int) {
  spaces := d.MinLeadingSpaces(lines)
  if spaces < 0 {
    return -1, -1
  }
  tts := d.TTS()
  return spaces / tts, spaces % tts
}

// Dedents the given lines *in-place* such that the minimum indentation is zero.
func (d *Dental) DedentLines(lines []string) {
  indent := Spaces(d.TTS())
  spaces := d.MinLeadingSpaces(lines)
  trim := Spaces(spaces)
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			lines[i] = ""
			continue
		}
    line = strings.ReplaceAll(line, "\t", indent)
    line = strings.TrimPrefix(line, trim)
    lines[i] = line
	}
}

func (d *Dental) DedentBlock(block string) string {
  lines := strings.Split(block, "\n")
  d.DedentLines(lines)
  return strings.Join(lines, "\n")
}

// Indents the given lines *in-place* by the given indentation level.
//
// Negative values will dedent the lines by the given amount.
func (d *Dental) IndentLines(lines []string, level int) {
  if level == 0 {
    return
  }
  spaces := level * d.TTS()
  if level < 0 {
    // we're actually *dedenting*
    prevSpaces := d.MinLeadingSpaces(lines)
    if prevSpaces + spaces < 0 {
      spaces = -prevSpaces
    }
  }
  if spaces == 0 {
    return // nothing to do
  }
  if spaces > 0 {
    // indent
    indentation := Spaces(spaces)
    for i, line := range lines {
      if strings.TrimSpace(line) == "" {
        lines[i] = ""
        continue
      }
      lines[i] = fmt.Sprintf("%v%v", indentation, line)
    }
  } else {
    // dedent
    for i, line := range lines {
      if strings.TrimSpace(line) == "" {
        lines[i] = ""
        continue
      }
      lines[i] = line[-spaces:]
    }
  }
}

// Indents the given block by the given indentation level.
//
// Negative values will dedent the block by the given amount.
func (d *Dental) IndentBlock(block string, level int) string {
  lines := strings.Split(block, "\n")
  d.IndentLines(lines, level)
  return strings.Join(lines, "\n")
}

// Sets the indentation of the given lines *in-place*.
func (d *Dental) SetLinesIndentation(lines []string, level int) {
  d.DedentLines(lines)
  d.IndentLines(lines, level)
}

// Sets the indentation of the given block.
func (d *Dental) SetBlockIndentation(block string, level int) string {
  lines := strings.Split(block, "\n")
  d.SetLinesIndentation(lines, level)
  return strings.Join(lines, "\n")
}

func Spaces(count int) string {
  var b strings.Builder
  for i := 0; i < count; i++ {
    b.WriteRune(' ')
  }
  return b.String()
}
