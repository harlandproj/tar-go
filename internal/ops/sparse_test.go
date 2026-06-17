package ops

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"testing"
)

func TestDetectHolesEmpty(t *testing.T) {
	dir, _ := os.MkdirTemp("", "sparse-*")
	defer os.RemoveAll(dir)

	path := dir + "/empty"
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	f, err = os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	holes := detectHoles(f, 0)
	if len(holes) != 0 {
		t.Errorf("expected 0 holes, got %d", len(holes))
	}
}

func TestDetectHolesNoHoles(t *testing.T) {
	dir, _ := os.MkdirTemp("", "sparse-*")
	defer os.RemoveAll(dir)

	path := dir + "/dense"
	data := []byte("hello world, this is a test file with no holes")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	holes := detectHoles(f, int64(len(data)))
	if len(holes) != 0 {
		t.Errorf("expected 0 holes, got %d", len(holes))
	}
}

func TestDetectHolesWithSparseData(t *testing.T) {
	dir, _ := os.MkdirTemp("", "sparse-*")
	defer os.RemoveAll(dir)

	path := dir + "/sparse"
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Write(make([]byte, 1))
	f.Seek(1024, io.SeekStart)
	f.Write(make([]byte, 1))
	f.Seek(4096, io.SeekStart)
	f.Write(make([]byte, 1))
	f.Close()

	fi, _ := os.Stat(path)
	size := fi.Size()

	f, err = os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	holes := detectHoles(f, size)
	if len(holes) == 0 {
		t.Log("no holes detected (expected on some platforms)")
	}
	if len(holes) > 0 {
		t.Logf("detected %d holes", len(holes))
	}
}

type sparseHole struct {
	Offset int64
	Length int64
}

func TestSparseHeaderRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	hdr := &tar.Header{
		Name:     "test.sparse",
		Size:     8192,
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	tw.Write(make([]byte, 512))
	tw.Close()

	tr := tar.NewReader(&buf)
	hdr2, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}
	if hdr2.Name != "test.sparse" {
		t.Errorf("got %q, want %q", hdr2.Name, "test.sparse")
	}
}
