package ops

import (
	"archive/tar"
	"os"
	"path/filepath"
	"testing"

	"github.com/harlandproj/tar-go/internal/cli"
)

func TestVerifyArchiveValid(t *testing.T) {
	env := newTestEnv(t)
	env.WriteFile("file.txt", "data")

	origDir, _ := os.Getwd()
	os.Chdir(env.Dir)
	defer os.Chdir(origDir)

	createOpts := env.BaseOpts(cli.SubCreate)
	createOpts.FileNames = []string{"file.txt"}
	if err := Create(createOpts); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	opts := env.BaseOpts(cli.SubCreate)
	if err := verifyArchive(opts); err != nil {
		t.Fatalf("verifyArchive failed: %v", err)
	}
}

func TestVerifyArchiveNonExistent(t *testing.T) {
	opts := &cli.Options{ArchiveNames: []string{"/nonexistent/archive.tar"}}
	if err := verifyArchive(opts); err == nil {
		t.Error("expected error for nonexistent archive")
	}
}

func TestConfirmAction(t *testing.T) {
	t.Skip("confirmAction reads stdin - tested via integration")
}

func TestWriteSparseFileNoHoles(t *testing.T) {
	dir := t.TempDir()
	dataPath := filepath.Join(dir, "dense.bin")
	data := []byte("no holes here")
	os.WriteFile(dataPath, data, 0o644)

	archivePath := filepath.Join(dir, "sparse.tar")
	af, _ := os.Create(archivePath)
	tw := tar.NewWriter(af)

	f, _ := os.Open(dataPath)
	fi, _ := f.Stat()
	err := writeSparseFile(tw, f, fi, "dense.bin")
	f.Close()
	tw.Close()
	af.Close()

	if err != nil {
		t.Fatalf("writeSparseFile failed: %v", err)
	}

	rf, _ := os.Open(archivePath)
	tr := tar.NewReader(rf)
	hdr, err := tr.Next()
	if err != nil {
		t.Fatalf("read sparse archive failed: %v", err)
	}
	if hdr.Name != "dense.bin" {
		t.Errorf("expected dense.bin, got %q", hdr.Name)
	}
	rf.Close()
}

func TestWriteSparseFileEmpty(t *testing.T) {
	dir := t.TempDir()
	dataPath := filepath.Join(dir, "empty.bin")
	os.WriteFile(dataPath, []byte{}, 0o644)

	archivePath := filepath.Join(dir, "sparse.tar")
	af, _ := os.Create(archivePath)
	tw := tar.NewWriter(af)

	f, _ := os.Open(dataPath)
	fi, _ := f.Stat()
	err := writeSparseFile(tw, f, fi, "empty.bin")
	f.Close()
	tw.Close()
	af.Close()

	if err != nil {
		t.Fatalf("writeSparseFile empty failed: %v", err)
	}
}

func TestTarTypeChar(t *testing.T) {
	tests := []struct {
		tf   byte
		want string
	}{
		{tar.TypeReg, "file"},
		{tar.TypeLink, "link"},
		{tar.TypeSymlink, "symlink"},
		{tar.TypeChar, "char"},
		{tar.TypeBlock, "block"},
		{tar.TypeDir, "dir"},
		{tar.TypeFifo, "fifo"},
		{tar.TypeGNULongName, "unknown"},
	}
	for _, tt := range tests {
		got := tarTypeChar(tt.tf)
		if got != tt.want {
			t.Errorf("tarTypeChar(%d) = %q, want %q", tt.tf, got, tt.want)
		}
	}
}

func TestFileTypeFromMode(t *testing.T) {
	tests := []struct {
		name string
		mode os.FileMode
		want byte
	}{
		{"dir", os.ModeDir, tar.TypeDir},
		{"symlink", os.ModeSymlink, tar.TypeSymlink},
		{"device", os.ModeDevice, tar.TypeBlock},
		{"char", os.ModeCharDevice, tar.TypeChar},
		{"pipe", os.ModeNamedPipe, tar.TypeFifo},
		{"regular", 0, tar.TypeReg},
	}
	for _, tt := range tests {
		got := fileTypeFromMode(tt.mode)
		if got != tt.want {
			t.Errorf("fileTypeFromMode(%s) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestDiffEntry(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(filePath, []byte("data"), 0o644)
	fi, _ := os.Lstat(filePath)

	hdr := &tar.Header{
		Name:     "test.txt",
		Typeflag: tar.TypeReg,
		Size:     fi.Size(),
		Mode:     int64(fi.Mode()),
		ModTime:  fi.ModTime(),
	}
	opts := &cli.Options{Touch: true}
	if diffEntry(hdr, fi, "test.txt", opts) {
		t.Error("expected no diff for same file with touch")
	}
}

func TestDiffEntrySizeDiffers(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(filePath, []byte("different length"), 0o644)
	fi, _ := os.Lstat(filePath)

	hdr := &tar.Header{
		Name:     "test.txt",
		Typeflag: tar.TypeReg,
		Size:     4,
		Mode:     0o644,
		ModTime:  fi.ModTime(),
	}
	opts := &cli.Options{}
	if !diffEntry(hdr, fi, "test.txt", opts) {
		t.Error("expected diff for different size")
	}
}

func TestOpenArchiveWriter(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "test.tar")
	opts := &cli.Options{
		ArchiveNames:   []string{archivePath},
		BlockingFactor: 20,
		RecordSize:     10240,
	}
	w, err := openArchiveWriter(opts)
	if err != nil {
		t.Fatalf("openArchiveWriter failed: %v", err)
	}
	w.Close()
	if _, err := os.Stat(archivePath); err != nil {
		t.Errorf("archive not created: %v", err)
	}
}

func TestOpenArchiveWriterGzip(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "test.tar.gz")
	opts := &cli.Options{
		ArchiveNames:   []string{archivePath},
		CompressProgram: "gzip",
		BlockingFactor: 20,
		RecordSize:     10240,
	}
	w, err := openArchiveWriter(opts)
	if err != nil {
		t.Fatalf("openArchiveWriter gzip failed: %v", err)
	}
	w.Close()
}

func TestOpenArchiveReader(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "test.tar")
	os.WriteFile(archivePath, []byte{}, 0o644)
	opts := &cli.Options{ArchiveNames: []string{archivePath}}
	r, err := openArchiveReader(opts)
	if err != nil {
		t.Fatalf("openArchiveReader failed: %v", err)
	}
	r.Close()
}

func TestProcessCDirectives(t *testing.T) {
	names := []string{"-C", "/tmp", "file.txt"}
	processCDirectives(names)
}
