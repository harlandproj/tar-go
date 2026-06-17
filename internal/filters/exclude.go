package filters

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/harlandproj/tar-go/internal/cli"
)

type Excluder struct {
	patterns       []string
	excludeCaches  bool
	excludeBackups bool
	excludeVCS     bool
}

var vcsDirs = map[string]bool{
	".git": true, ".svn": true, ".hg": true, ".bzr": true, "CVS": true,
}

func NewExcluder(opts *cli.Options) *Excluder {
	e := &Excluder{
		patterns:       opts.Exclude,
		excludeCaches:  opts.ExcludeCaches,
		excludeBackups: opts.ExcludeBackups,
		excludeVCS:     opts.ExcludeVCS,
	}
	if opts.ExcludeFrom != "" {
		e.patterns = append(e.patterns, loadExcludeFile(opts.ExcludeFrom)...)
	}
	return e
}

func loadExcludeFile(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}
	return patterns
}

func (e *Excluder) Match(path string) bool {
	base := filepath.Base(path)

	if e.excludeBackups && (strings.HasSuffix(base, "~") || strings.HasSuffix(base, ".bak")) {
		return true
	}
	if e.excludeVCS && vcsDirs[base] {
		return true
	}
	if e.excludeCaches && base == "CACHEDIR.TAG" {
		return true
	}

	for _, pat := range e.patterns {
		if matched, _ := filepath.Match(pat, base); matched {
			return true
		}
		if strings.Contains(pat, "/") {
			if matched, _ := filepath.Match(pat, path); matched {
				return true
			}
		}
	}
	return false
}
