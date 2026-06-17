package app

import (
	"os"
	"path/filepath"
	"testing"
)

const testVersion = "0.1.0"

func TestRunHelp(t *testing.T) {
	code := Run([]string{"tar", "--help"}, testVersion)
	if code != 0 {
		t.Errorf("expected 0 for --help, got %d", code)
	}
}

func TestRunVersion(t *testing.T) {
	code := Run([]string{"tar", "--version"}, testVersion)
	if code != 0 {
		t.Errorf("expected 0 for --version, got %d", code)
	}
}

func TestRunNoSubcommand(t *testing.T) {
	code := Run([]string{"tar"}, testVersion)
	if code == 0 {
		t.Error("expected non-zero for no subcommand")
	}
}

func TestRunInvalidOption(t *testing.T) {
	code := Run([]string{"tar", "--bad-option"}, testVersion)
	if code == 0 {
		t.Error("expected non-zero for invalid option")
	}
}

func TestRunCreateAndList(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "test.tar")
	dataFile := filepath.Join(dir, "file.txt")
	os.WriteFile(dataFile, []byte("hello"), 0o644)

	code := Run([]string{"tar", "-cf", archive, "-C", dir, "file.txt"}, testVersion)
	if code != 0 {
		t.Fatalf("create failed with code %d", code)
	}

	code = Run([]string{"tar", "-tf", archive}, testVersion)
	if code != 0 {
		t.Errorf("list failed with code %d", code)
	}
}

func TestRunCreateAndExtract(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "test.tar")
	dataFile := filepath.Join(dir, "file.txt")
	os.WriteFile(dataFile, []byte("hello"), 0o644)

	Run([]string{"tar", "-cf", archive, "-C", dir, "file.txt"}, testVersion)

	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	origDir, _ := os.Getwd()
	code := Run([]string{"tar", "-xf", archive, "-C", outDir}, testVersion)
	os.Chdir(origDir)
	if code != 0 {
		t.Errorf("extract failed with code %d", code)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "file.txt"))
	if err != nil || string(data) != "hello" {
		t.Errorf("extracted content wrong: %q, err=%v", string(data), err)
	}
}

func helperCreateArchive(t *testing.T, dir string) string {
	t.Helper()
	archive := filepath.Join(dir, "test.tar")
	dataFile := filepath.Join(dir, "file.txt")
	os.WriteFile(dataFile, []byte("hello"), 0o644)
	code := Run([]string{"tar", "-cf", archive, "-C", dir, "file.txt"}, testVersion)
	if code != 0 {
		t.Fatalf("create failed with code %d", code)
	}
	return archive
}

func TestRunAppend(t *testing.T) {
	dir := t.TempDir()
	archive := helperCreateArchive(t, dir)
	extraFile := filepath.Join(dir, "extra.txt")
	os.WriteFile(extraFile, []byte("extra"), 0o644)

	code := Run([]string{"tar", "-rf", archive, "-C", dir, "extra.txt"}, testVersion)
	if code != 0 {
		t.Errorf("append failed with code %d", code)
	}
}

func TestRunUpdate(t *testing.T) {
	dir := t.TempDir()
	archive := helperCreateArchive(t, dir)

	code := Run([]string{"tar", "-uf", archive, "-C", dir, "file.txt"}, testVersion)
	if code != 0 {
		t.Errorf("update failed with code %d", code)
	}
}

func TestRunConcat(t *testing.T) {
	dir := t.TempDir()
	archive := helperCreateArchive(t, dir)
	archive2 := filepath.Join(dir, "second.tar")
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644)
	Run([]string{"tar", "-cf", archive2, "-C", dir, "b.txt"}, testVersion)

	code := Run([]string{"tar", "-Af", archive, archive2}, testVersion)
	if code != 0 {
		t.Errorf("concat failed with code %d", code)
	}
}

func TestRunDiff(t *testing.T) {
	dir := t.TempDir()
	archive := helperCreateArchive(t, dir)

	code := Run([]string{"tar", "-df", archive, "-C", dir}, testVersion)
	_ = code
}

func TestRunDelete(t *testing.T) {
	dir := t.TempDir()
	archive := helperCreateArchive(t, dir)

	code := Run([]string{"tar", "--delete", "-f", archive, "file.txt"}, testVersion)
	if code != 0 {
		t.Errorf("delete failed with code %d", code)
	}
}

func TestRunTestLabel(t *testing.T) {
	dir := t.TempDir()
	archive := helperCreateArchive(t, dir)

	code := Run([]string{"tar", "--test-label", "-f", archive}, testVersion)
	_ = code
}

func TestRunCreateNonExistent(t *testing.T) {
	code := Run([]string{"tar", "-cf", "/nonexistent/dir/test.tar", "file.txt"}, testVersion)
	if code == 0 {
		t.Error("expected non-zero for create to nonexistent path")
	}
}

func TestRunListNonExistent(t *testing.T) {
	code := Run([]string{"tar", "-tf", "/nonexistent/archive.tar"}, testVersion)
	if code == 0 {
		t.Error("expected non-zero for list nonexistent archive")
	}
}

func TestRunExtractNonExistent(t *testing.T) {
	code := Run([]string{"tar", "-xf", "/nonexistent/archive.tar"}, testVersion)
	if code == 0 {
		t.Error("expected non-zero for extract nonexistent archive")
	}
}

func TestRunAppendNonExistent(t *testing.T) {
	code := Run([]string{"tar", "-rf", "/nonexistent/archive.tar", "file.txt"}, testVersion)
	if code == 0 {
		t.Error("expected non-zero for append nonexistent archive")
	}
}

func TestRunConcatNonExistent(t *testing.T) {
	code := Run([]string{"tar", "-Af", "/nonexistent/archive.tar", "/nonexistent/other.tar"}, testVersion)
	if code == 0 {
		t.Error("expected non-zero for concat nonexistent archive")
	}
}

func TestRunDiffNonExistent(t *testing.T) {
	code := Run([]string{"tar", "-df", "/nonexistent/archive.tar"}, testVersion)
	if code == 0 {
		t.Error("expected non-zero for diff nonexistent archive")
	}
}

func TestRunDeleteNonExistent(t *testing.T) {
	code := Run([]string{"tar", "--delete", "-f", "/nonexistent/archive.tar", "file.txt"}, testVersion)
	if code == 0 {
		t.Error("expected non-zero for delete nonexistent archive")
	}
}

func TestRunTestLabelNonExistent(t *testing.T) {
	code := Run([]string{"tar", "--test-label", "-f", "/nonexistent/archive.tar"}, testVersion)
	if code == 0 {
		t.Error("expected non-zero for test-label nonexistent archive")
	}
}

func TestRunUpdateNonExistent(t *testing.T) {
	code := Run([]string{"tar", "-uf", "/nonexistent/archive.tar", "file.txt"}, testVersion)
	if code == 0 {
		t.Error("expected non-zero for update nonexistent archive")
	}
}
