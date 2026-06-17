package filters

import (
	"fmt"
	"regexp"
	"strings"
)

type Transform struct {
	regex   *regexp.Regexp
	replace string
	global  bool
}

func NewTransform(expr string) (*Transform, error) {
	if expr == "" || len(expr) < 3 || expr[0] != 's' {
		return nil, fmt.Errorf("invalid transform: must be sed-style s/pattern/replace/flags")
	}
	delim := expr[1]
	end := strings.LastIndexByte(expr, byte(delim))
	if end < 4 {
		return nil, fmt.Errorf("invalid transform: malformed expression")
	}
	pattern := expr[2:end]
	rest := expr[end+1:]

	replIdx := strings.LastIndexByte(pattern, byte(delim))
	if replIdx < 0 {
		return nil, fmt.Errorf("invalid transform: missing delimiter in pattern")
	}
	search := pattern[:replIdx]
	replace := unescapeBackrefs(pattern[replIdx+1:])

	flags := rest
	global := false
	caseInsensitive := false
	for _, f := range flags {
		switch f {
		case 'g':
			global = true
		case 'i':
			caseInsensitive = true
		}
	}

	var re *regexp.Regexp
	var err error
	if caseInsensitive {
		re, err = regexp.Compile("(?i)" + search)
	} else {
		re, err = regexp.Compile(search)
	}
	if err != nil {
		return nil, fmt.Errorf("invalid transform regex: %w", err)
	}

	return &Transform{
		regex:   re,
		replace: replace,
		global:  global,
	}, nil
}

func (t *Transform) Apply(name string) string {
	if t == nil {
		return name
	}
	if t.global {
		return t.regex.ReplaceAllString(name, t.replace)
	}
	loc := t.regex.FindStringIndex(name)
	if loc == nil {
		return name
	}
	return name[:loc[0]] + t.regex.ReplaceAllString(name[loc[0]:loc[1]], t.replace) + name[loc[1]:]
}

func unescapeBackrefs(s string) string {
	out := strings.Builder{}
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && s[i+1] >= '1' && s[i+1] <= '9' {
			out.WriteByte('$')
			out.WriteByte(s[i+1])
			i++
		} else {
			out.WriteByte(s[i])
		}
	}
	return out.String()
}
