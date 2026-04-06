package regu

import (
  "regexp"
  "strings"
)

func RegMatch(p *regexp.Regexp, text string) map[string]string {
	return RegMatchNoTrim(p, strings.TrimSpace(text))
}

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

