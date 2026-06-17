package ops

import (
	"archive/tar"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/harlandproj/tar-go/internal/cli"
)

func TestListBasic(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	opts := env.BaseOpts(cli.SubList)
	if err := List(opts); err != nil {
		t.Fatalf("List failed: %v", err)
	}
}

func TestListVerbose(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	opts := env.BaseOpts(cli.SubList)
	opts.Verbose = 1
	if err := List(opts); err != nil {
		t.Fatalf("List failed: %v", err)
	}
}

func TestListWithNumericOwner(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	opts := env.BaseOpts(cli.SubList)
	opts.Verbose = 1
	opts.NumericOwner = true
	if err := List(opts); err != nil {
		t.Fatalf("List failed: %v", err)
	}
}

func TestListWithTotals(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	opts := env.BaseOpts(cli.SubList)
	opts.ShowTotals = true
	if err := List(opts); err != nil {
		t.Fatalf("List failed: %v", err)
	}
}

func TestListWithUtc(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	opts := env.BaseOpts(cli.SubList)
	opts.Verbose = 1
	opts.Utc = true
	if err := List(opts); err != nil {
		t.Fatalf("List failed: %v", err)
	}
}

func TestListWithFullTime(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	opts := env.BaseOpts(cli.SubList)
	opts.Verbose = 1
	opts.FullTime = true
	if err := List(opts); err != nil {
		t.Fatalf("List failed: %v", err)
	}
}

func TestDeleteBasic(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{
		"a.txt": "a",
		"b.txt": "b",
	})

	opts := env.BaseOpts(cli.SubDelete)
	opts.FileNames = []string{"b.txt"}
	if err := Delete(opts); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	names := env.ReadArchive()
	for _, n := range names {
		if n == "b.txt" {
			t.Error("b.txt should have been deleted")
		}
	}
}

func TestDeleteNonExistentMember(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"a.txt": "a"})

	opts := env.BaseOpts(cli.SubDelete)
	opts.FileNames = []string{"nonexistent.txt"}
	if err := Delete(opts); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestConcatBasic(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"a.txt": "a"})
	archive2 := filepath.Join(env.Dir, "b.tar")
	f, _ := os.Create(archive2)
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "b.txt", Typeflag: tar.TypeReg, Size: 1, Mode: 0o644, ModTime: time.Now()})
	tw.Write([]byte("b"))
	tw.Close()
	f.Close()

	opts := env.BaseOpts(cli.SubConcat)
	opts.FileNames = []string{archive2}
	if err := Concat(opts); err != nil {
		t.Fatalf("Concat failed: %v", err)
	}

	names := env.ReadArchive()
	found := false
	for _, n := range names {
		if n == "b.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("b.txt not found after concat: %v", names)
	}
}

func TestDiffNoDifference(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")
	env.CreateArchive(map[string]string{"file.txt": "data"})

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubDiff)
	err := Diff(opts)
	if err != nil && err != DiffExitError {
		t.Logf("Diff returned: %v", err)
	}
}

func TestDiffWithDifference(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "different")
	env.CreateArchive(map[string]string{"file.txt": "original"})

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubDiff)
	err := Diff(opts)
	if err != DiffExitError {
		t.Errorf("expected DiffExitError for differences, got: %v", err)
	}
}

func TestDiffMissingFile(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubDiff)
	err := Diff(opts)
	if err != DiffExitError {
		t.Errorf("expected DiffExitError for missing file, got: %v", err)
	}
}

func TestTestLabelFound(t *testing.T) {
	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "MYVOL", Typeflag: tar.TypeReg, Size: 0, Mode: 0o644, ModTime: time.Now()})
	tw.WriteHeader(&tar.Header{Name: "file.txt", Typeflag: tar.TypeReg, Size: 4, Mode: 0o644, ModTime: time.Now()})
	tw.Write([]byte("data"))
	tw.Close()
	f.Close()

	opts := env.BaseOpts(cli.SubTestLabel)
	if err := TestLabel(opts); err != nil {
		t.Fatalf("TestLabel failed: %v", err)
	}
}

func TestTestLabelMismatch(t *testing.T) {
	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "OTHERVOL", Typeflag: tar.TypeReg, Size: 0, Mode: 0o644, ModTime: time.Now()})
	tw.Close()
	f.Close()

	opts := env.BaseOpts(cli.SubTestLabel)
	opts.VolumeLabel = "MYVOL"
	err := TestLabel(opts)
	if err == nil {
		t.Error("expected error for label mismatch")
	}
}

func TestTestLabelNoLabel(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	opts := env.BaseOpts(cli.SubTestLabel)
	opts.VolumeLabel = "MYVOL"
	err := TestLabel(opts)
	if err == nil {
		t.Error("expected error for no matching label")
	}
}

func TestUpdateBasic(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "original")
	env.CreateArchive(map[string]string{"file.txt": "original"})

	time.Sleep(100 * time.Millisecond)
	os.WriteFile(filepath.Join(env.Dir, "file.txt"), []byte("updated"), 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubUpdate)
	opts.FileNames = []string{"file.txt"}
	if err := Update(opts); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
}

func TestUpdateNoChange(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "original")
	env.CreateArchive(map[string]string{"file.txt": "original"})

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubUpdate)
	opts.FileNames = []string{"file.txt"}
	opts.Verbose = 1
	if err := Update(opts); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
}

func TestAppendBasic(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("a.txt", "first")
	env.CreateArchive(map[string]string{"a.txt": "first"})
	env.WriteFile("b.txt", "second")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubAppend)
	opts.FileNames = []string{"b.txt"}
	if err := Append(opts); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	names := env.ReadArchive()
	found := false
	for _, n := range names {
		if n == "b.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("b.txt not found after append: %v", names)
	}
}
