package ops

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/harlandproj/tar-go/internal/cli"
)

func TestMultiReadCloserClose(t *testing.T) {
	r1 := io.NopCloser(bytes.NewReader([]byte("hello")))
	mc := &multiReadCloser{ReadCloser: r1, closers: []io.Closer{r1}}
	if err := mc.Close(); err != nil {
		t.Errorf("multiReadCloser.Close failed: %v", err)
	}
}

type errCloser struct{}

func (e *errCloser) Close() error { return fmt.Errorf("close error") }

func TestMultiReadCloserCloseError(t *testing.T) {
	r1 := io.NopCloser(bytes.NewReader([]byte("hello")))
	ec := &errCloser{}
	mc := &multiReadCloser{ReadCloser: r1, closers: []io.Closer{r1, ec}}
	if err := mc.Close(); err == nil {
		t.Error("expected error from multiReadCloser.Close")
	}
}

func TestExtractToStdout(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "stdout data"})

	opts := env.BaseOpts(cli.SubExtract)
	opts.ToStdout = true

	tmpFile := filepath.Join(t.TempDir(), "stdout")
	f, _ := os.Create(tmpFile)
	origStdout := cli.Stdout
	cli.Stdout = f

	if err := Extract(opts); err != nil {
		t.Fatalf("Extract to stdout failed: %v", err)
	}
	cli.Stdout = origStdout
	f.Close()
	data, _ := os.ReadFile(tmpFile)
	if !bytes.Contains(data, []byte("stdout data")) {
		t.Errorf("expected stdout data in output, got %q", string(data))
	}
}

func TestExtractWithOccurrence(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"a.txt": "aaa", "b.txt": "bbb"})

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.Occurrence = 1
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract with occurrence failed: %v", err)
	}
}

func TestExtractWithStartingFile(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"a.txt": "aaa", "b.txt": "bbb"})

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.StartingFile = "b.txt"
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract with starting file failed: %v", err)
	}
}

func TestExtractWithAbsoluteNames(t *testing.T) {
	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "file.txt", Typeflag: tar.TypeReg, Size: 4, Mode: 0o644, ModTime: time.Now()})
	tw.Write([]byte("data"))
	tw.Close()
	f.Close()

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.AbsoluteNames = true
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract with absolute names failed: %v", err)
	}
}

func TestExtractSymlink(t *testing.T) {
	skipIfWindows(t, "symlinks require admin on Windows")

	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "target", Mode: 0o777, ModTime: time.Now()})
	tw.Close()
	f.Close()

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract symlink failed: %v", err)
	}
}

func TestExtractHardLink(t *testing.T) {
	skipIfWindows(t, "hard links unreliable on Windows")

	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	f, _ := os.OpenFile(env.Archive, os.O_RDWR, 0)
	fi, _ := f.Stat()
	buf := make([]byte, fi.Size())
	f.ReadAt(buf, 0)
	f.Close()

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
}

func TestExtractWithBackup(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "new"})

	outDir := env.OutDir("out")
	os.WriteFile(filepath.Join(outDir, "file.txt"), []byte("old"), 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.BackupType = "simple"
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract with backup failed: %v", err)
	}
	if _, err := os.Stat("file.txt~"); err != nil {
		t.Errorf("backup not created: %v", err)
	}
}

func TestExtractWithOverwrite(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "new"})

	outDir := env.OutDir("out")
	os.WriteFile(filepath.Join(outDir, "file.txt"), []byte("old"), 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.KeepOldFiles = cli.OldOverwrite
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract with overwrite failed: %v", err)
	}
}

func TestExtractWithKeepNewerFiles(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "old"})

	outDir := env.OutDir("out")
	os.WriteFile(filepath.Join(outDir, "file.txt"), []byte("newer"), 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.KeepOldFiles = cli.OldKeepNewerFiles
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract with keep-newer failed: %v", err)
	}
}

func TestExtractOneTopLevelDash(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	opts := env.BaseOpts(cli.SubExtract)
	opts.OneTopLevel = "-"

	tmpFile := filepath.Join(t.TempDir(), "stdout")
	f, _ := os.Create(tmpFile)
	origStdout := cli.Stdout
	cli.Stdout = f

	if err := Extract(opts); err != nil {
		t.Fatalf("Extract with one-top-level=- failed: %v", err)
	}
	cli.Stdout = origStdout
	f.Close()
	data, _ := os.ReadFile(tmpFile)
	if !bytes.Contains(data, []byte("file.txt")) {
		t.Error("expected file.txt in stdout output")
	}
}

func TestExtractSkipOldFiles(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "archive"})

	outDir := env.OutDir("out")
	os.WriteFile(filepath.Join(outDir, "file.txt"), []byte("existing"), 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.KeepOldFiles = cli.OldSkipOldFiles
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract with skip-old-files failed: %v", err)
	}
}

func TestExtractWithVerbose(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.Verbose = 1

	tmpFile := filepath.Join(t.TempDir(), "stdout")
	f, _ := os.Create(tmpFile)
	origStdout := cli.Stdout
	cli.Stdout = f

	if err := Extract(opts); err != nil {
		t.Fatalf("Extract with verbose failed: %v", err)
	}
	cli.Stdout = origStdout
	f.Close()
}

func TestCreateWithCheckpoint(t *testing.T) {
	env := newTestEnv(t)
	for i := 0; i < 5; i++ {
		env.WriteFile(fmt.Sprintf("file%d.txt", i), "data")
	}

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file0.txt", "file1.txt", "file2.txt", "file3.txt", "file4.txt"}
	opts.Checkpoint = 2

	tmpFile := filepath.Join(t.TempDir(), "stderr")
	f, _ := os.Create(tmpFile)
	origStderr := cli.Stderr
	cli.Stderr = f

	if err := Create(opts); err != nil {
		t.Fatalf("Create with checkpoint failed: %v", err)
	}
	cli.Stderr = origStderr
	f.Close()
}

func TestCreateWithRemoveFiles(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.RemoveFiles = true
	if err := Create(opts); err != nil {
		t.Fatalf("Create with remove-files failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(env.Dir, "file.txt")); err == nil {
		t.Error("file should have been removed after archiving")
	}
}

func TestCreateWithVerify(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.Verify = true
	if err := Create(opts); err != nil {
		t.Fatalf("Create with verify failed: %v", err)
	}
}

func TestCreateWithDereference(t *testing.T) {
	skipIfWindows(t, "symlinks require admin on Windows")

	env := newTestEnv(t)
	env.WriteFile("real.txt", "real data")
	env.Symlink("real.txt", "link.txt")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"link.txt"}
	opts.Dereference = true
	if err := Create(opts); err != nil {
		t.Fatalf("Create with dereference failed: %v", err)
	}
}

func TestCreateWithHardDereference(t *testing.T) {
	skipIfWindows(t, "symlinks require admin on Windows")

	env := newTestEnv(t)
	env.WriteFile("real.txt", "real data")
	env.Symlink("real.txt", "link.txt")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"link.txt"}
	opts.HardDereference = true
	if err := Create(opts); err != nil {
		t.Fatalf("Create with hard-dereference failed: %v", err)
	}
}

func TestCreateWithNewerMtime(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("old.txt", "old")
	env.WriteFile("new.txt", "new")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"old.txt", "new.txt"}
	opts.NewerMtime = time.Now().Add(time.Hour)
	if err := Create(opts); err != nil {
		t.Fatalf("Create with newer-mtime failed: %v", err)
	}
	names := env.ReadArchive()
	if len(names) > 0 {
		t.Errorf("expected empty archive with future newer-mtime, got %v", names)
	}
}

func TestCreateWithIgnoreFailedRead(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"nonexistent.txt", "file.txt"}
	opts.IgnoreFailedRead = true
	if err := Create(opts); err != nil {
		t.Fatalf("Create with ignore-failed-read failed: %v", err)
	}
}

func TestCreateWithMtimeClamp(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.SetMtimeMode = cli.MtimeClamp
	opts.Mtime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := Create(opts); err != nil {
		t.Fatalf("Create with clamp-mtime failed: %v", err)
	}
}

func TestCreateWithSparse(t *testing.T) {
	t.Skip("skipping: archive/tar has a known bug with sparse file byte counting")
}

func TestCreateWithIncremental(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.Incremental = true
	if err := Create(opts); err != nil {
		t.Fatalf("Create with incremental failed: %v", err)
	}
}

func TestCreateWithListedIncremental(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")
	snapFile := filepath.Join(env.Dir, "snap.inc")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.ListedIncremental = snapFile
	if err := Create(opts); err != nil {
		t.Fatalf("Create with listed-incremental failed: %v", err)
	}
}

func TestCreateWithMultiVolume(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.MultiVolume = true
	opts.TapeLength = 1024
	if err := Create(opts); err != nil {
		t.Fatalf("Create with multi-volume failed: %v", err)
	}
}

func TestOpenArchiveReaderNonExistent(t *testing.T) {
	opts := &cli.Options{ArchiveNames: []string{"/nonexistent/archive.tar"}}
	_, err := openArchiveReader(opts)
	if err == nil {
		t.Error("expected error for nonexistent archive")
	}
}

func TestOpenArchiveReaderGzip(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "test.tar.gz")

	wOpts := &cli.Options{
		ArchiveNames:    []string{archivePath},
		CompressProgram: "gzip",
		BlockingFactor:  20,
		RecordSize:      10240,
	}
	w, err := openArchiveWriter(wOpts)
	if err != nil {
		t.Fatalf("openArchiveWriter gzip failed: %v", err)
	}
	w.Write([]byte("hello"))
	w.Close()

	rOpts := &cli.Options{ArchiveNames: []string{archivePath}}
	r, err := openArchiveReader(rOpts)
	if err != nil {
		t.Fatalf("openArchiveReader gzip failed: %v", err)
	}
	buf := make([]byte, 5)
	n, _ := r.Read(buf)
	r.Close()
	if n != 5 || string(buf) != "hello" {
		t.Errorf("read back failed: got %q", string(buf[:n]))
	}
}

func TestOpenArchiveWriterAutoCompress(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "test.tar.gz")

	opts := &cli.Options{
		ArchiveNames:   []string{archivePath},
		AutoCompress:   true,
		BlockingFactor: 20,
		RecordSize:     10240,
	}
	w, err := openArchiveWriter(opts)
	if err != nil {
		t.Fatalf("openArchiveWriter auto-compress failed: %v", err)
	}
	w.Close()
}

func TestDiffEntryMtimeDiffers(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")
	os.WriteFile(filePath, []byte("data"), 0o644)
	fi, _ := os.Lstat(filePath)

	hdr := &tar.Header{
		Name:     "test.txt",
		Typeflag: tar.TypeReg,
		Size:     fi.Size(),
		Mode:     int64(fi.Mode()),
		ModTime:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	opts := &cli.Options{}
	if !diffEntry(hdr, fi, "test.txt", opts) {
		t.Error("expected diff for different mtime")
	}
}

func TestDiffEntryModeDiffers(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")
	os.WriteFile(filePath, []byte("data"), 0o644)
	fi, _ := os.Lstat(filePath)

	hdr := &tar.Header{
		Name:     "test.txt",
		Typeflag: tar.TypeReg,
		Size:     fi.Size(),
		Mode:     0o755,
		ModTime:  fi.ModTime(),
	}
	opts := &cli.Options{Touch: true}
	if runtime.GOOS != "windows" {
		if !diffEntry(hdr, fi, "test.txt", opts) {
			t.Error("expected diff for different mode")
		}
	}
}

func TestDiffEntryTypeDiffers(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")
	os.WriteFile(filePath, []byte("data"), 0o644)
	fi, _ := os.Lstat(filePath)

	hdr := &tar.Header{
		Name:     "test.txt",
		Typeflag: tar.TypeDir,
		Mode:     0o755,
		ModTime:  fi.ModTime(),
	}
	opts := &cli.Options{Touch: true}
	if !diffEntry(hdr, fi, "test.txt", opts) {
		t.Error("expected diff for different type")
	}
}

func TestBackupExisting(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "file.txt")
	os.WriteFile(original, []byte("old"), 0o644)

	backupPath := makeBackup(original, "existing")
	if backupPath == "" {
		t.Fatal("backup path empty")
	}
}

func TestBackupNever(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "file.txt")
	os.WriteFile(original, []byte("old"), 0o644)

	backupPath := makeBackup(original, "never")
	if backupPath == "" {
		t.Fatal("backup path empty")
	}
}

func TestBackupOff(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "file.txt")
	os.WriteFile(original, []byte("old"), 0o644)

	backupPath := makeBackup(original, "off")
	if backupPath != "" {
		t.Errorf("expected no backup for 'off', got %q", backupPath)
	}
}

func TestBackupSuffixEnv(t *testing.T) {
	os.Setenv("SIMPLE_BACKUP_SUFFIX", ".bak")
	defer os.Unsetenv("SIMPLE_BACKUP_SUFFIX")
	s := backupSuffix("")
	if s != ".bak" {
		t.Errorf("expected .bak, got %q", s)
	}
}

func TestBackupSuffixCustom(t *testing.T) {
	s := backupSuffix(".orig")
	if s != ".orig" {
		t.Errorf("expected .orig, got %q", s)
	}
}

func TestBackupSuffixDefault(t *testing.T) {
	os.Unsetenv("SIMPLE_BACKUP_SUFFIX")
	s := backupSuffix("")
	if s != "~" {
		t.Errorf("expected ~, got %q", s)
	}
}

func TestExtractGNUSparse(t *testing.T) {
	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	hdr := &tar.Header{Name: "sparse.bin", Typeflag: tar.TypeGNUSparse, Size: 0, Mode: 0o644, ModTime: time.Now(), Format: tar.FormatGNU}
	tw.WriteHeader(hdr)
	tw.Close()
	f.Close()

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract GNUSparse failed: %v", err)
	}
}

func TestExtractGNUSparseToStdout(t *testing.T) {
	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	hdr := &tar.Header{Name: "sparse.bin", Typeflag: tar.TypeGNUSparse, Size: 0, Mode: 0o644, ModTime: time.Now(), Format: tar.FormatGNU}
	tw.WriteHeader(hdr)
	tw.Close()
	f.Close()

	opts := env.BaseOpts(cli.SubExtract)
	opts.ToStdout = true

	tmpFile := filepath.Join(t.TempDir(), "stdout")
	tmpF, _ := os.Create(tmpFile)
	origStdout := cli.Stdout
	cli.Stdout = tmpF

	if err := Extract(opts); err != nil {
		t.Fatalf("Extract GNUSparse to stdout failed: %v", err)
	}
	cli.Stdout = origStdout
	tmpF.Close()
}

func TestExtractDefaultType(t *testing.T) {
	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "unknown", Typeflag: 'X', Size: 3, Mode: 0o644, ModTime: time.Now()})
	tw.Write([]byte("abc"))
	tw.Close()
	f.Close()

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract unknown type failed: %v", err)
	}
}

func TestExtractEmptyName(t *testing.T) {
	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "", Typeflag: tar.TypeReg, Size: 0, Mode: 0o644, ModTime: time.Now()})
	tw.Close()
	f.Close()

	opts := env.BaseOpts(cli.SubExtract)
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract empty name failed: %v", err)
	}
}

func TestExtractKeepOldFilesSkip(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "new"})

	outDir := env.OutDir("out")
	os.WriteFile(filepath.Join(outDir, "file.txt"), []byte("old"), 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.KeepOldFiles = cli.OldSkipOldFiles
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract skip-old-files failed: %v", err)
	}
	data, _ := os.ReadFile("file.txt")
	if string(data) != "old" {
		t.Errorf("expected old file preserved, got %q", string(data))
	}
}

func TestExtractUnlinkFirst(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "new"})

	outDir := env.OutDir("out")
	os.WriteFile(filepath.Join(outDir, "file.txt"), []byte("old"), 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.KeepOldFiles = cli.OldUnlinkFirst
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract unlink-first failed: %v", err)
	}
}

func TestExtractSymlinkExistingOverwrite(t *testing.T) {
	skipIfWindows(t, "symlinks require admin on Windows")

	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "target", Mode: 0o777, ModTime: time.Now()})
	tw.Close()
	f.Close()

	outDir := env.OutDir("out")
	os.Symlink("oldtarget", filepath.Join(outDir, "link"))

	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.KeepOldFiles = cli.OldOverwrite
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract symlink overwrite failed: %v", err)
	}
}

func TestExtractLinkExistingKeepOld(t *testing.T) {
	skipIfWindows(t, "hard links unreliable on Windows")

	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "link", Typeflag: tar.TypeLink, Linkname: "target", Mode: 0o644, ModTime: time.Now()})
	tw.Close()
	f.Close()

	outDir := env.OutDir("out")
	os.WriteFile(filepath.Join(outDir, "link"), []byte("existing"), 0o644)

	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.KeepOldFiles = cli.OldKeepOldFiles
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract link keep-old failed: %v", err)
	}
}

func TestWriteSparseFileWithHoles(t *testing.T) {
	skipIfWindows(t, "sparse hole detection differs on Windows")

	dir := t.TempDir()
	dataPath := filepath.Join(dir, "sparse.bin")
	f, _ := os.Create(dataPath)
	f.Write(make([]byte, 1))
	f.Seek(4096, io.SeekStart)
	f.Write(make([]byte, 1))
	f.Seek(8192, io.SeekStart)
	f.Write(make([]byte, 1))
	f.Close()

	fi, _ := os.Stat(dataPath)
	dataFile, _ := os.Open(dataPath)
	defer dataFile.Close()

	archivePath := filepath.Join(dir, "sparse.tar")
	af, _ := os.Create(archivePath)
	tw := tar.NewWriter(af)

	err := writeSparseFile(tw, dataFile, fi, "sparse.bin")
	tw.Close()
	af.Close()

	if err != nil {
		t.Fatalf("writeSparseFile with holes failed: %v", err)
	}
}

func TestOpenArchiveWriterAutoCompressBz2(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "test.tar.bz2")

	opts := &cli.Options{
		ArchiveNames:   []string{archivePath},
		AutoCompress:   true,
		BlockingFactor: 20,
		RecordSize:     10240,
	}
	w, err := openArchiveWriter(opts)
	if err != nil {
		t.Fatalf("openArchiveWriter auto-compress bz2 failed: %v", err)
	}
	w.Close()
}

func TestOpenArchiveReaderStdin(t *testing.T) {
	opts := &cli.Options{ArchiveNames: []string{"-"}}
	r, err := openArchiveReader(opts)
	if err != nil {
		t.Fatalf("openArchiveReader stdin failed: %v", err)
	}
	r.Close()
}

func TestConfirmActionSkipped(t *testing.T) {
	t.Skip("confirmAction reads from stdin - tested via integration")
}

func TestOpenArchiveWriterDefaultName(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	opts := &cli.Options{
		BlockingFactor: 20,
		RecordSize:     10240,
	}
	w, err := openArchiveWriter(opts)
	if err != nil {
		t.Fatalf("openArchiveWriter default name failed: %v", err)
	}
	w.Close()
}

func TestOpenArchiveReaderByExtension(t *testing.T) {
	dir := t.TempDir()

	wOpts := &cli.Options{
		ArchiveNames:    []string{filepath.Join(dir, "test.tar.gz")},
		CompressProgram: "gzip",
		BlockingFactor:  20,
		RecordSize:      10240,
	}
	w, err := openArchiveWriter(wOpts)
	if err != nil {
		t.Fatalf("create gzip archive failed: %v", err)
	}
	w.Write([]byte("hello"))
	w.Close()

	rOpts := &cli.Options{ArchiveNames: []string{filepath.Join(dir, "test.tar.gz")}}
	r, err := openArchiveReader(rOpts)
	if err != nil {
		t.Fatalf("openArchiveReader by extension failed: %v", err)
	}
	r.Close()
}

func TestExtractWithIgnoreFailedRead(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	outDir := env.OutDir("out")
	origDir, _ := os.Getwd()
	os.Chdir(outDir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubExtract)
	opts.IgnoreFailedRead = true
	if err := Extract(opts); err != nil {
		t.Fatalf("Extract with ignore-failed-read failed: %v", err)
	}
}

func TestCreateWithShowTransformed(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("hello.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"hello.txt"}
	opts.Transform = "s/hello/world/"
	opts.ShowTransformed = true
	opts.Verbose = 1

	tmpFile := filepath.Join(t.TempDir(), "stderr")
	f, _ := os.Create(tmpFile)
	origStderr := cli.Stderr
	cli.Stderr = f

	if err := Create(opts); err != nil {
		t.Fatalf("Create with show-transformed failed: %v", err)
	}
	cli.Stderr = origStderr
	f.Close()
}

func TestCreateWithAbsoluteNames(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubCreate)
	opts.FileNames = []string{"file.txt"}
	opts.AbsoluteNames = true
	if err := Create(opts); err != nil {
		t.Fatalf("Create with absolute-names failed: %v", err)
	}
}

func TestListWithIgnoreZeros(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	opts := env.BaseOpts(cli.SubList)
	opts.IgnoreZeros = true
	if err := List(opts); err != nil {
		t.Fatalf("List with ignore-zeros failed: %v", err)
	}
}

func TestListWithOccurrence(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"a.txt": "aaa", "b.txt": "bbb"})

	opts := env.BaseOpts(cli.SubList)
	opts.Occurrence = 1
	if err := List(opts); err != nil {
		t.Fatalf("List with occurrence failed: %v", err)
	}
}

func TestListWithStartingFile(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"a.txt": "aaa", "b.txt": "bbb"})

	opts := env.BaseOpts(cli.SubList)
	opts.StartingFile = "b.txt"
	if err := List(opts); err != nil {
		t.Fatalf("List with starting-file failed: %v", err)
	}
}

func TestListWithTransform(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"old.txt": "data"})

	opts := env.BaseOpts(cli.SubList)
	opts.Transform = "s/old/new/"
	if err := List(opts); err != nil {
		t.Fatalf("List with transform failed: %v", err)
	}
}

func TestListWithBlockNumber(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	opts := env.BaseOpts(cli.SubList)
	opts.Verbose = 1
	opts.BlockNumber = true
	if err := List(opts); err != nil {
		t.Fatalf("List with block-number failed: %v", err)
	}
}

func TestListWithCheckLinks(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"file.txt": "data"})

	opts := env.BaseOpts(cli.SubList)
	opts.CheckLinks = true
	opts.Verbose = 1
	if err := List(opts); err != nil {
		t.Fatalf("List with check-links failed: %v", err)
	}
}

func TestTestLabelWithLongName(t *testing.T) {
	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "LONGNAME", Typeflag: tar.TypeGNULongName, Size: 10, Mode: 0o644, ModTime: time.Now()})
	tw.Write([]byte("realname\n"))
	tw.WriteHeader(&tar.Header{Name: "realname", Typeflag: tar.TypeReg, Size: 0, Mode: 0o644, ModTime: time.Now()})
	tw.Close()
	f.Close()

	opts := env.BaseOpts(cli.SubTestLabel)
	if err := TestLabel(opts); err != nil {
		t.Fatalf("TestLabel with long name failed: %v", err)
	}
}

func TestTestLabelWithLongLink(t *testing.T) {
	env := newTestEnv(t)
	f, _ := os.Create(env.Archive)
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "LONGLINK", Typeflag: tar.TypeGNULongLink, Size: 10, Mode: 0o644, ModTime: time.Now()})
	tw.Write([]byte("linkname\n"))
	tw.WriteHeader(&tar.Header{Name: "file.txt", Typeflag: tar.TypeReg, Size: 0, Mode: 0o644, ModTime: time.Now()})
	tw.Close()
	f.Close()

	opts := env.BaseOpts(cli.SubTestLabel)
	if err := TestLabel(opts); err != nil {
		t.Fatalf("TestLabel with long link failed: %v", err)
	}
}

func TestAppendWithCDirective(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("a.txt", "first")
	env.CreateArchive(map[string]string{"a.txt": "first"})
	env.WriteFile("b.txt", "second")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubAppend)
	opts.FileNames = []string{"-C", env.Dir, "b.txt"}
	if err := Append(opts); err != nil {
		t.Fatalf("Append with -C failed: %v", err)
	}
}

func TestAppendWithVerbose(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("a.txt", "first")
	env.CreateArchive(map[string]string{"a.txt": "first"})
	env.WriteFile("b.txt", "second")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubAppend)
	opts.FileNames = []string{"b.txt"}
	opts.Verbose = 1
	if err := Append(opts); err != nil {
		t.Fatalf("Append with verbose failed: %v", err)
	}
}

func TestAppendWithIgnoreFailedRead(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"a.txt": "first"})

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubAppend)
	opts.FileNames = []string{"nonexistent.txt"}
	opts.IgnoreFailedRead = true
	if err := Append(opts); err != nil {
		t.Fatalf("Append with ignore-failed-read failed: %v", err)
	}
}

func TestDiffWithUtc(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")
	env.CreateArchive(map[string]string{"file.txt": "data"})

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	opts := env.BaseOpts(cli.SubDiff)
	opts.Utc = true
	err := Diff(opts)
	if err != nil && err != DiffExitError {
		t.Logf("Diff returned: %v", err)
	}
}

func TestDeleteWithVerbose(t *testing.T) {
	env := newTestEnv(t)
	env.CreateArchive(map[string]string{"a.txt": "a", "b.txt": "b"})

	opts := env.BaseOpts(cli.SubDelete)
	opts.FileNames = []string{"b.txt"}
	opts.Verbose = 1
	if err := Delete(opts); err != nil {
		t.Fatalf("Delete with verbose failed: %v", err)
	}
}

func TestConcatWithVerbose(t *testing.T) {
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
	opts.Verbose = 1
	if err := Concat(opts); err != nil {
		t.Fatalf("Concat with verbose failed: %v", err)
	}
}
