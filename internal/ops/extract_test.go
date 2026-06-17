package ops

import (
	"archive/tar"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/harlandproj/tar-go/internal/cli"
)

func TestExtractDelayedDirRestoreBasic(t *testing.T) {
	dir := t.TempDir()

	archive := filepath.Join(dir, "test.tar")
	f, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "subdir/", Typeflag: tar.TypeDir, Mode: 0o755, ModTime: time.Now()})
	tw.WriteHeader(&tar.Header{Name: "subdir/file.txt", Typeflag: tar.TypeReg, Size: 4, Mode: 0o644})
	tw.Write([]byte("data"))
	tw.Close()
	f.Close()

	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := &cli.Options{
		Subcommand:      cli.SubExtract,
		ArchiveNames:    []string{archive},
		DelayDirRestore: true,
	}
	if err := Extract(opts); err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "subdir")); err != nil {
		t.Errorf("subdir not created: %v", err)
	}
}

func TestExtractNoGlobalStateLeak(t *testing.T) {
	dir := t.TempDir()

	archive := filepath.Join(dir, "test.tar")
	f, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "dir/", Typeflag: tar.TypeDir, Mode: 0o755})
	tw.Close()
	f.Close()

	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := &cli.Options{
		Subcommand:      cli.SubExtract,
		ArchiveNames:    []string{archive},
		DelayDirRestore: true,
	}
	if err := Extract(opts); err != nil {
		t.Fatalf("extract 1 failed: %v", err)
	}
}
