package ops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupSimple(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "file.txt")
	os.WriteFile(original, []byte("old"), 0o644)

	backupPath := makeBackup(original, "simple")
	if backupPath == "" {
		t.Fatal("backup path empty")
	}
	if _, err := os.Stat(backupPath); err != nil {
		t.Errorf("backup not created: %v", err)
	}
	if _, err := os.Stat(original); err == nil {
		t.Error("original should have been renamed to backup")
	}
}

func TestBackupNone(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "file.txt")
	os.WriteFile(original, []byte("old"), 0o644)

	backupPath := makeBackup(original, "none")
	if backupPath != "" {
		t.Errorf("expected no backup for 'none', got %q", backupPath)
	}
}

func TestBackupNumbered(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "file.txt")
	os.WriteFile(original, []byte("old"), 0o644)

	backupPath := makeBackup(original, "numbered")
	if backupPath == "" {
		t.Fatal("backup path empty")
	}
	expected := filepath.Join(dir, "file.txt.~1~")
	if backupPath != expected {
		t.Errorf("expected %q, got %q", expected, backupPath)
	}
}

func TestBackupNonExistent(t *testing.T) {
	backupPath := makeBackup("/nonexistent/file.txt", "simple")
	if backupPath != "" {
		t.Errorf("expected empty for nonexistent file, got %q", backupPath)
	}
}
