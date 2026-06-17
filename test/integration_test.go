package test

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const tarBin = "../bin/tar"

func bin() string {
	if fi, _ := os.Stat("../bin/tar.exe"); fi != nil {
		return "../bin/tar.exe"
	}
	return "../bin/tar"
}

func TestHelpOutput(t *testing.T) {
	cmd := exec.Command(bin(), "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}
	if len(out) < 500 {
		t.Errorf("help too short: %d bytes", len(out))
	}
	if !strings.Contains(string(out), "--create") {
		t.Error("help missing --create")
	}
	if !strings.Contains(string(out), "--extract") {
		t.Error("help missing --extract")
	}
}

func TestVersion(t *testing.T) {
	cmd := exec.Command(bin(), "--version")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}
	if !strings.Contains(string(out), "0.1.0") {
		t.Error("version output wrong")
	}
}

func TestUsage(t *testing.T) {
	cmd := exec.Command(bin(), "--usage")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--usage failed: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Error("usage missing")
	}
}

func TestCreateAndList(t *testing.T) {
	dir := t.TempDir()

	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	os.WriteFile(a, []byte("hello"), 0o644)
	os.WriteFile(b, []byte("world"), 0o644)

	archive := filepath.Join(dir, "test.tar")

	cmd := exec.Command(bin(), "-cf", archive, "-C", dir, "a.txt", "b.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tf", archive)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("list failed: %v\n%s", err, out)
	}

	s := string(out)
	if !strings.Contains(s, "a.txt") || !strings.Contains(s, "b.txt") {
		t.Errorf("list missing files: %s", s)
	}
}

func TestCreateAndExtract(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	data := []byte("test content")
	os.WriteFile(filepath.Join(dir, "input.txt"), data, 0o644)

	archive := filepath.Join(dir, "test.tar")

	cmd := exec.Command(bin(), "-cf", archive, "-C", dir, "input.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-xf", archive, "-C", outDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract failed: %v\n%s", err, out)
	}

	extracted, err := os.ReadFile(filepath.Join(outDir, "input.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(extracted) != string(data) {
		t.Errorf("content mismatch: got %q, want %q", extracted, data)
	}
}

func TestCreateGzipAndExtract(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	data := []byte("gzip compressed data")
	os.WriteFile(filepath.Join(dir, "file.txt"), data, 0o644)

	archive := filepath.Join(dir, "test.tar.gz")

	cmd := exec.Command(bin(), "-czvf", archive, "-C", dir, "file.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create+gzip failed: %v\n%s", err, out)
	}

	f, err := os.Open(archive)
	if err != nil {
		t.Fatal(err)
	}
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	hdr, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}
	if hdr.Name != "file.txt" {
		t.Errorf("wrong name: %s", hdr.Name)
	}
	buf, _ := io.ReadAll(tr)
	if string(buf) != string(data) {
		t.Errorf("content mismatch in tar: got %q, want %q", buf, data)
	}
	gz.Close()
	f.Close()

	cmd = exec.Command(bin(), "-xzvf", archive, "-C", outDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract+gzip failed: %v\n%s", err, out)
	}

	extracted, _ := os.ReadFile(filepath.Join(outDir, "file.txt"))
	if string(extracted) != string(data) {
		t.Errorf("extracted mismatch: got %q, want %q", extracted, data)
	}
}

func TestCreateBzip2AndList(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("bzip2 data"), 0o644)

	archive := filepath.Join(dir, "test.tar.bz2")

	cmd := exec.Command(bin(), "-cjvf", archive, "-C", dir, "file.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create+bzip2 failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tvf", archive)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(string(out), "file.txt") {
		t.Errorf("list missing file: %s", string(out))
	}
}

func TestCreateXZAndList(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("xz data"), 0o644)

	archive := filepath.Join(dir, "test.tar.xz")

	cmd := exec.Command(bin(), "-cJvf", archive, "-C", dir, "file.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create+xz failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tvf", archive)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("list failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "file.txt") {
		t.Errorf("list missing file: %s", string(out))
	}
}

func TestMutualExclusionSubcommand(t *testing.T) {
	cmd := exec.Command(bin(), "-xc")
	err := cmd.Run()
	if err == nil {
		t.Error("expected error for -xc, got none")
	}
}

func TestEmpty(t *testing.T) {
	cmd := exec.Command(bin())
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("expected exit error")
	}
	if !strings.Contains(string(out), "specify one of the") {
		t.Error("wrong error message")
	}
}

func TestAppend(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("first"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("second"), 0o644)

	archive := filepath.Join(dir, "test.tar")

	cmd := exec.Command(bin(), "-cf", archive, "-C", dir, "a.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-rf", archive, "-C", dir, "b.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("append failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-xf", archive, "-C", outDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract failed: %v\n%s", err, out)
	}

	a, _ := os.ReadFile(filepath.Join(outDir, "a.txt"))
	b, _ := os.ReadFile(filepath.Join(outDir, "b.txt"))
	if string(a) != "first" || string(b) != "second" {
		t.Errorf("append failed: a=%q b=%q", a, b)
	}
}

func TestStripComponents(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	nested := filepath.Join(dir, "a", "b", "c")
	os.MkdirAll(nested, 0o755)
	os.WriteFile(filepath.Join(nested, "file.txt"), []byte("deep"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "-C", dir, "a")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-xf", archive, "-C", outDir, "--strip-components=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract failed: %v\n%s", err, out)
	}

	data, _ := os.ReadFile(filepath.Join(outDir, "b", "c", "file.txt"))
	if string(data) != "deep" {
		t.Errorf("strip-components failed: got %q", data)
	}
}

func TestExclude(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("keep"), 0o644)
	os.WriteFile(filepath.Join(dir, "skip.log"), []byte("skip"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "-C", dir, "--exclude=*.log", "keep.txt", "skip.log")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tf", archive)
	out, _ := cmd.Output()
	if strings.Contains(string(out), "skip.log") {
		t.Error("excluded file still in archive")
	}
	if !strings.Contains(string(out), "keep.txt") {
		t.Error("included file missing")
	}
}

func TestVolumeLabel(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "-C", dir, "--label=MYVOL", "file.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create with label failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "--test-label", "-f", archive)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("test-label failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "MYVOL") {
		t.Errorf("expected label MYVOL, got: %s", string(out))
	}
}
