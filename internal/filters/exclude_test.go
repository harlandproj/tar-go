package filters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/harlandproj/tar-go/internal/cli"
)

func TestExcludePattern(t *testing.T) {
	opts := &cli.Options{Exclude: []string{"*.log"}}
	e := NewExcluder(opts)
	if !e.Match("error.log") {
		t.Error("should match *.log")
	}
	if e.Match("file.txt") {
		t.Error("should not match *.txt")
	}
}

func TestExcludeBackup(t *testing.T) {
	opts := &cli.Options{ExcludeBackups: true}
	e := NewExcluder(opts)
	if !e.Match("file~") {
		t.Error("should match backup ~")
	}
	if !e.Match("file.bak") {
		t.Error("should match .bak")
	}
	if e.Match("file.txt") {
		t.Error("should not match regular file")
	}
}

func TestExcludeVCS(t *testing.T) {
	opts := &cli.Options{ExcludeVCS: true}
	e := NewExcluder(opts)
	if !e.Match(".git") {
		t.Error("should match .git")
	}
	if !e.Match(".svn") {
		t.Error("should match .svn")
	}
	if e.Match("src") {
		t.Error("should not match src")
	}
}

func TestExcludeCaches(t *testing.T) {
	opts := &cli.Options{ExcludeCaches: true}
	e := NewExcluder(opts)
	if !e.Match("CACHEDIR.TAG") {
		t.Error("should match CACHEDIR.TAG")
	}
}

func TestExcludeFrom(t *testing.T) {
	dir := t.TempDir()
	excludeFile := filepath.Join(dir, "excludes.txt")
	os.WriteFile(excludeFile, []byte("*.log\n*.tmp\n"), 0o644)

	opts := &cli.Options{ExcludeFrom: excludeFile}
	e := NewExcluder(opts)
	if !e.Match("error.log") {
		t.Error("should match *.log from file")
	}
	if !e.Match("temp.tmp") {
		t.Error("should match *.tmp from file")
	}
	if e.Match("keep.txt") {
		t.Error("should not match keep.txt")
	}
}

func TestExcludePathPattern(t *testing.T) {
	opts := &cli.Options{Exclude: []string{"build/output"}}
	e := NewExcluder(opts)
	if !e.Match("build/output") {
		t.Error("should match path pattern")
	}
}
