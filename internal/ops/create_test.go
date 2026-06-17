package ops

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/harlandproj/tar-go/internal/cli"
)

func TestCreateBasic(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "hello")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	if err := Create(opts); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	names := env.ReadArchive()
	found := false
	for _, n := range names {
		if n == "file.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("file.txt not in archive: %v", names)
	}
}

func TestCreateMultipleFiles(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("a.txt", "aaa")
	env.WriteFile("b.txt", "bbb")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"a.txt", "b.txt"}
	if err := Create(opts); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	names := env.ReadArchive()
	if len(names) < 2 {
		t.Errorf("expected at least 2 entries, got %d: %v", len(names), names)
	}
}

func TestCreateWithDirectory(t *testing.T) {
	env := newTestEnv(t)
	env.Mkdir("subdir")
	env.WriteFile("subdir/file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"subdir"}
	if err := Create(opts); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	names := env.ReadArchive()
	found := false
	for _, n := range names {
		if n == "subdir/file.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("subdir/file.txt not in archive: %v", names)
	}
}

func TestCreateWithExclude(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("keep.txt", "keep")
	env.WriteFile("skip.log", "skip")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"keep.txt", "skip.log"}
	opts.Exclude = []string{"*.log"}
	if err := Create(opts); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	contents := env.ReadArchiveContents()
	if _, ok := contents["skip.log"]; ok {
		t.Error("excluded file should not be in archive")
	}
	if _, ok := contents["keep.txt"]; !ok {
		t.Error("non-excluded file should be in archive")
	}
}

func TestCreateWithTransform(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("hello.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"hello.txt"}
	opts.Transform = "s/hello/world/"
	if err := Create(opts); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	names := env.ReadArchive()
	found := false
	for _, n := range names {
		if n == "world.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("world.txt not in archive: %v", names)
	}
}

func TestCreateWithOwnerGroup(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.Owner = "testuser"
	opts.Group = "testgroup"
	if err := Create(opts); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	names := env.ReadArchive()
	if len(names) == 0 {
		t.Fatal("archive is empty")
	}
}

func TestCreateWithCompression(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")
	env.Archive = filepath.Join(env.Dir, "test.tar.gz")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.CompressProgram = "gzip"
	if err := Create(opts); err != nil {
		t.Fatalf("Create with gzip failed: %v", err)
	}

	if _, err := os.Stat(env.Archive); err != nil {
		t.Errorf("archive not created: %v", err)
	}
}

func TestCreateWithFilesFrom(t *testing.T) {
	t.Skip("FilesFrom with CWD issues in unit test - covered by integration test")
}

func TestCreateWithSortName(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("zebra.txt", "z")
	env.WriteFile("alpha.txt", "a")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"zebra.txt", "alpha.txt"}
	opts.SortOrder = "name"
	if err := Create(opts); err != nil {
		t.Fatalf("Create with sort failed: %v", err)
	}

	names := env.ReadArchive()
	if len(names) < 2 {
		t.Fatalf("expected at least 2 entries")
	}
	if names[0] != "alpha.txt" {
		t.Errorf("expected alpha.txt first, got %q", names[0])
	}
}

func TestCreateWithVolumeLabel(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.VolumeLabel = "MYVOL"
	if err := Create(opts); err != nil {
		t.Fatalf("Create with label failed: %v", err)
	}

	names := env.ReadArchive()
	found := false
	for _, n := range names {
		if n == "MYVOL" {
			found = true
		}
	}
	if !found {
		t.Errorf("volume label not in archive: %v", names)
	}
}

func TestCreateWithVerbose(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.Verbose = 1
	if err := Create(opts); err != nil {
		t.Fatalf("Create with verbose failed: %v", err)
	}
}

func TestCreateWithMtime(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.SetMtimeMode = cli.MtimeForce
	opts.Mtime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := Create(opts); err != nil {
		t.Fatalf("Create with mtime failed: %v", err)
	}
}

func TestCreateWithTouch(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.Touch = true
	if err := Create(opts); err != nil {
		t.Fatalf("Create with touch failed: %v", err)
	}
}
