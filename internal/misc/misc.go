package misc

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ZapSlashes(name string) string {
	for len(name) > 1 && name[len(name)-1] == '/' {
		name = name[:len(name)-1]
	}
	return name
}

func NormalizePath(name string) string {
	return filepath.ToSlash(name)
}

func SafePath(base, name string) string {
	cleaned := filepath.Clean(string(filepath.Separator) + name)
	joined := filepath.Join(base, cleaned)
	if !strings.HasPrefix(joined, base+string(filepath.Separator)) && joined != base {
		return filepath.Join(base, filepath.Base(name))
	}
	return joined
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func CopyFile(dst, src string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func LeadingDots(name string) int {
	i := 0
	for i < len(name) && name[i] == '.' {
		i++
	}
	return i
}

func HasFilePathPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix) &&
		(len(s) == len(prefix) || s[len(prefix)] == '/')
}

func QuoteArg(arg string) string {
	if !strings.ContainsAny(arg, " \t\n\r|&;<>()$`\\\"'*?[]#~=%") {
		return arg
	}
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}

var QuotingStyles = []string{
	"literal",
	"shell",
	"shell-always",
	"c",
	"c-maybe",
	"escape",
	"locale",
	"clocale",
}

func QuoteName(name string, style string) string {
	switch style {
	case "literal":
		return name
	case "shell":
		if needsShellQuoting(name) {
			return "'" + strings.ReplaceAll(name, "'", "'\\''") + "'"
		}
		return name
	case "shell-always":
		return "'" + strings.ReplaceAll(name, "'", "'\\''") + "'"
	case "c", "c-maybe":
		return strconv.Quote(name)
	case "escape":
		return escapeName(name)
	default:
		return name
	}
}

func needsShellQuoting(name string) bool {
	return strings.ContainsAny(name, " \t\n*?[]|&;<>()$`\\\"'#~")
}

func escapeName(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch r {
		case '\n':
			b.WriteString("\\n")
		case '\t':
			b.WriteString("\\t")
		case '\r':
			b.WriteString("\\r")
		case '\\':
			b.WriteString("\\\\")
		default:
			if r < 0x20 || r >= 0x7f {
				b.WriteString("\\")
				b.WriteString(strings.TrimLeft(strconv.FormatUint(uint64(r), 8), "0"))
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}
