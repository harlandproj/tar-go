package ops

import (
	"os"
	"testing"
)

func TestResolveFilesEmpty(t *testing.T) {
	result := resolveFiles(nil)
	if len(result) != 1 || result[0] != "." {
		t.Errorf("expected [\".\"], got %v", result)
	}
}

func TestResolveFilesNonEmpty(t *testing.T) {
	result := resolveFiles([]string{"a", "b"})
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}

func TestResolveFilesWithCDirective(t *testing.T) {
	result := resolveFiles([]string{"-C", "/tmp", "a.txt"})
	if len(result) != 3 {
		t.Errorf("expected 3, got %d: %v", len(result), result)
	}
}

func TestReadFileList(t *testing.T) {
	dir := t.TempDir()
	listFile := dir + "/files.txt"
	os.WriteFile(listFile, []byte("a.txt\nb.txt\n# comment\n\nc.txt\n"), 0o644)
	result := readFileList(listFile)
	if len(result) != 3 {
		t.Errorf("expected 3, got %d: %v", len(result), result)
	}
}

func TestReadFileListNonExistent(t *testing.T) {
	result := readFileList("/nonexistent/file.txt")
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}
