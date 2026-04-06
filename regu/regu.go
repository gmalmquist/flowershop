// Convenience functions for working with regexes.
package regu

import (
  "regexp"
  "strings"
)

// Returns a map of capture group names to values if the provided regex matches the given text.
//
// This ignores leading and trailing whitespace around the text.
func RegMatch(p *regexp.Regexp, text string) map[string]string {
	return RegMatchNoTrim(p, strings.TrimSpace(text))
}

// Returns a map of capture group names to values if the provided regex matches the given text.
func RegMatchNoTrim(p *regexp.Regexp, text string) map[string]string {
	m := p.FindStringSubmatch(text)
	if m == nil {
		return nil
	}
	res := map[string]string{
    "0": m[0],
  }
	for i, name := range p.SubexpNames() {
		res[name] = m[i]
	}
	return res
}

