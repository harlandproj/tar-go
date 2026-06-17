package misc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestZapSlashes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello/", "hello"},
		{"hello//", "hello"},
		{"/", "/"},
		{"", ""},
		{"no-trailing", "no-trailing"},
		{"a/b/c/", "a/b/c"},
	}
	for _, tt := range tests {
		got := ZapSlashes(tt.input)
		if got != tt.want {
			t.Errorf("ZapSlashes(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	result := NormalizePath("a/b/c")
	if result != "a/b/c" {
		t.Errorf("NormalizePath = %q, want %q", result, "a/b/c")
	}
}

func TestSafePath(t *testing.T) {
	result := SafePath("/base", "file.txt")
	expected := filepath.Join("/base", "file.txt")
	if result != expected {
		t.Errorf("SafePath = %q, want %q", result, expected)
	}
	result = SafePath("/base", "../escape")
	base := filepath.Clean("/base")
	if !(strings.HasPrefix(filepath.Clean(result), base+string(filepath.Separator)) || filepath.Clean(result) == base) {
		t.Errorf("SafePath should keep path under base: got %q", result)
	}
}

func TestDirExists(t *testing.T) {
	dir := t.TempDir()
	if !DirExists(dir) {
		t.Error("DirExists should return true for existing dir")
	}
	if DirExists("/nonexistent/dir/abc") {
		t.Error("DirExists should return false for nonexistent dir")
	}
	f := filepath.Join(dir, "file.txt")
	os.WriteFile(f, []byte("x"), 0o644)
	if DirExists(f) {
		t.Error("DirExists should return false for a file")
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "file.txt")
	os.WriteFile(f, []byte("x"), 0o644)
	if !FileExists(f) {
		t.Error("FileExists should return true for existing file")
	}
	if FileExists("/nonexistent/file/abc") {
		t.Error("FileExists should return false for nonexistent file")
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "sub", "dst.txt")
	os.WriteFile(src, []byte("copy me"), 0o644)
	if err := CopyFile(dst, src, 0o644); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst failed: %v", err)
	}
	if string(data) != "copy me" {
		t.Errorf("CopyFile content = %q, want %q", string(data), "copy me")
	}
}

func TestCopyFileMissingSrc(t *testing.T) {
	err := CopyFile("/tmp/dst", "/nonexistent/src", 0o644)
	if err == nil {
		t.Error("expected error for missing src")
	}
}

func TestLeadingDots(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"..file", 2},
		{".hidden", 1},
		{"normal", 0},
		{"...triple", 3},
		{"", 0},
	}
	for _, tt := range tests {
		got := LeadingDots(tt.input)
		if got != tt.want {
			t.Errorf("LeadingDots(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestHasFilePathPrefix(t *testing.T) {
	if !HasFilePathPrefix("a/b/c", "a/b") {
		t.Error("expected true for prefix match with /")
	}
	if !HasFilePathPrefix("a/b", "a/b") {
		t.Error("expected true for exact match")
	}
	if HasFilePathPrefix("abc", "ab") {
		t.Error("expected false for partial match without /")
	}
	if HasFilePathPrefix("a/b", "a/c") {
		t.Error("expected false for different prefix")
	}
}

func TestQuoteArg(t *testing.T) {
	if QuoteArg("simple") != "simple" {
		t.Error("simple arg should not be quoted")
	}
	quoted := QuoteArg("has space")
	if quoted[0] != '\'' {
		t.Error("arg with space should be single-quoted")
	}
	quoted = QuoteArg("it's")
	if quoted == "it's" {
		t.Error("arg with special char should be quoted")
	}
}

func TestQuoteNameLiteral(t *testing.T) {
	if QuoteName("hello", "literal") != "hello" {
		t.Error("literal style should return unchanged")
	}
}

func TestQuoteNameShell(t *testing.T) {
	if QuoteName("hello", "shell") != "hello" {
		t.Error("shell style should not quote simple names")
	}
	quoted := QuoteName("hello world", "shell")
	if quoted[0] != '\'' {
		t.Error("shell style should quote names with spaces")
	}
}

func TestQuoteNameShellAlways(t *testing.T) {
	quoted := QuoteName("hello", "shell-always")
	if quoted[0] != '\'' {
		t.Error("shell-always should always quote")
	}
}

func TestQuoteNameC(t *testing.T) {
	quoted := QuoteName("hello", "c")
	if quoted[0] != '"' {
		t.Error("c style should use double quotes")
	}
}

func TestQuoteNameCMaybe(t *testing.T) {
	quoted := QuoteName("hello", "c-maybe")
	if quoted[0] != '"' {
		t.Error("c-maybe style should use double quotes")
	}
}

func TestQuoteNameEscape(t *testing.T) {
	result := escapeName("hello\nworld")
	if result != "hello\\nworld" {
		t.Errorf("escapeName = %q, want %q", result, "hello\\nworld")
	}
	result = escapeName("tab\there")
	if result != "tab\\there" {
		t.Errorf("escapeName = %q, want %q", result, "tab\\there")
	}
	result = escapeName("back\\slash")
	if result != "back\\\\slash" {
		t.Errorf("escapeName = %q, want %q", result, "back\\\\slash")
	}
	result = escapeName("normal")
	if result != "normal" {
		t.Errorf("escapeName = %q, want %q", result, "normal")
	}
}

func TestQuoteNameDefault(t *testing.T) {
	if QuoteName("hello", "unknown") != "hello" {
		t.Error("unknown style should return unchanged")
	}
}

func TestNeedsShellQuoting(t *testing.T) {
	if needsShellQuoting("simple") {
		t.Error("simple name should not need quoting")
	}
	if !needsShellQuoting("has space") {
		t.Error("name with space should need quoting")
	}
	if !needsShellQuoting("has*glob") {
		t.Error("name with glob should need quoting")
	}
}

func TestQuotingStyles(t *testing.T) {
	if len(QuotingStyles) != 8 {
		t.Errorf("expected 8 styles, got %d", len(QuotingStyles))
	}
}
