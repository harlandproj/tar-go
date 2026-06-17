package ops

import (
	"archive/tar"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/harlandproj/tar-go/internal/cli"
)

func TestExtractBasic(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "hello"})

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	data, err := os.ReadFile("file.txt")
	if err != nil {
		t.Fatalf("file not extracted: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("expected 'hello', got %q", string(data))
	}
}

func TestExtractDirectory(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{
		"subdir/":       "",
		"subdir/file.txt": "data",
	})

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if _, err := os.Stat("subdir/file.txt"); err != nil {
		t.Errorf("subdir/file.txt not extracted: %v", err)
	}
}

func TestExtractWithStripComponents(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{
		"a/b/file.txt": "data",
	})

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.StripComponents = 2
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if _, err := os.Stat("file.txt"); err != nil {
		t.Errorf("file.txt not found after strip: %v", err)
	}
}

func TestExtractKeepOldFiles(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "archive"})

	outDir := env.OutDir("out")
	os.WriteFile(filepath.Join(outDir, "file.txt"), []byte("existing"), 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.KeepOldFiles = cli.OldKeepOldFiles
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	data, _ := os.ReadFile("file.txt")
	if string(data) != "existing" {
		t.Errorf("expected existing file preserved, got %q", string(data))
	}
}

func TestExtractWithTransform(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"old.txt": "data"})

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.Transform = "s/old/new/"
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if _, err := os.Stat("new.txt"); err != nil {
		t.Errorf("new.txt not found: %v", err)
	}
}

func TestExtractWithOneTopLevel(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.OneTopLevel = "myprefix"
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if _, err := os.Stat("myprefix/file.txt"); err != nil {
		t.Errorf("myprefix/file.txt not found: %v", err)
	}
}

func TestExtractWithDelayDirRestore(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{
		"subdir/":       "",
		"subdir/file.txt": "data",
	})

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.DelayDirRestore = true
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
}

func TestExtractWithSamePermissions(t *testing.T) {
	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "exec.sh", Typeflag: tar.TypeReg, Size: 4, Mode: 0o755, ModTime: time.Now()})
	tw.Write([]byte("echo"))
	tw.Close()
	f.Close()

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.SamePermissions = true
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
}

func TestExtractWithTouch(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.Touch = true
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
}

func TestExtractWithRecursiveUnlink(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "new"})

	outDir := env.OutDir("out")
	oldDir := filepath.Join(outDir, "file.txt")
	os.MkdirAll(oldDir, 0o755)
	os.WriteFile(filepath.Join(oldDir, "inner.txt"), []byte("old"), 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.RecursiveUnlink = true
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
}
