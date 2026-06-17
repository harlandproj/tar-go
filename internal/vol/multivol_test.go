package vol

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestMultiVolWriteRead(t *testing.T) {
	dir, _ := os.MkdirTemp("", "tar-vol-*")
	defer os.RemoveAll(dir)

	prefix := dir + "/archive"
	mv, err := NewMultiVolWriter(prefix, 512)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 2000)
	for i := range data {
		data[i] = byte(i % 256)
	}
	n, err := mv.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Errorf("wrote %d, want %d", n, len(data))
	}
	mv.Close()
}

func TestMultiVolExactBoundary(t *testing.T) {
	dir, _ := os.MkdirTemp("", "tar-vol-*")
	defer os.RemoveAll(dir)

	prefix := dir + "/archive"
	mv, err := NewMultiVolWriter(prefix, 100)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 100)
	for i := range data {
		data[i] = 0x41
	}
	n, err := mv.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 100 {
		t.Errorf("wrote %d, want 100", n)
	}
	n, err = mv.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 100 {
		t.Errorf("wrote %d, want 100", n)
	}
	mv.Close()
}

func TestMultiVolReader(t *testing.T) {
	dir, _ := os.MkdirTemp("", "tar-vol-*")
	defer os.RemoveAll(dir)

	prefix := dir + "/archive"
	mv, err := NewMultiVolWriter(prefix, 256)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte("hello world, this is a multi-volume archive test")
	n, err := mv.Write(payload)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(payload) {
		t.Errorf("wrote %d, want %d", n, len(payload))
	}
	mv.Close()

	mr, err := NewMultiVolReader(prefix)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, len(payload))
	readN, err := io.ReadFull(mr, buf)
	if err != nil {
		t.Fatal(err)
	}
	if readN != len(payload) {
		t.Errorf("read %d, want %d", readN, len(payload))
	}
	if string(buf) != string(payload) {
		t.Errorf("got %q, want %q", string(buf), string(payload))
	}
	mr.Close()
}

func TestMultiVolSmallSize(t *testing.T) {
	dir, _ := os.MkdirTemp("", "tar-vol-*")
	defer os.RemoveAll(dir)

	prefix := dir + "/archive"
	mv, err := NewMultiVolWriter(prefix, 50)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 300)
	for i := range data {
		data[i] = byte('A' + (i % 26))
	}
	n, err := mv.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Errorf("wrote %d, want %d", n, len(data))
	}
	mv.Close()
}

func TestParseVolName(t *testing.T) {
	tests := []struct {
		input  string
		prefix string
		num    int
	}{
		{"archive", "archive", 1},
		{"archive-2", "archive", 2},
		{"archive-10", "archive", 10},
		{"no-dash", "no-dash", 1},
		{"path/to/file-3", "path/to/file", 3},
		{"abc-xyz", "abc-xyz", 1},
	}
	for _, tt := range tests {
		prefix, num := parseVolName(tt.input)
		if prefix != tt.prefix || num != tt.num {
			t.Errorf("parseVolName(%q) = (%q, %d), want (%q, %d)", tt.input, prefix, num, tt.prefix, tt.num)
		}
	}
}

func TestMultiVolReaderMultiVolume(t *testing.T) {
	dir := t.TempDir()
	prefix := filepath.Join(dir, "archive")

	mv, err := NewMultiVolWriter(prefix, 64)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 200)
	for i := range data {
		data[i] = byte(i % 256)
	}
	n, err := mv.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Errorf("wrote %d, want %d", n, len(data))
	}
	mv.Close()

	mr, err := NewMultiVolReader(prefix)
	if err != nil {
		t.Fatal(err)
	}
	readData, err := io.ReadAll(mr)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if len(readData) != len(data) {
		t.Errorf("read %d bytes, want %d", len(readData), len(data))
	}
	mr.Close()
}

func TestNewMultiVolWriterInvalidSize(t *testing.T) {
	_, err := NewMultiVolWriter("/tmp/archive", 0)
	if err == nil {
		t.Error("expected error for zero maxSize")
	}
}

func TestMultiVolWriterCloseTwice(t *testing.T) {
	dir := t.TempDir()
	prefix := filepath.Join(dir, "archive")
	mv, err := NewMultiVolWriter(prefix, 100)
	if err != nil {
		t.Fatal(err)
	}
	mv.Close()
	if err := mv.Close(); err != nil {
		t.Errorf("second Close should return nil: %v", err)
	}
}

func TestMultiVolReaderCloseTwice(t *testing.T) {
	dir := t.TempDir()
	prefix := filepath.Join(dir, "archive")
	mv, err := NewMultiVolWriter(prefix, 256)
	if err != nil {
		t.Fatal(err)
	}
	mv.Write([]byte("hello"))
	mv.Close()

	mr, err := NewMultiVolReader(prefix)
	if err != nil {
		t.Fatal(err)
	}
	mr.Close()
	if err := mr.Close(); err != nil {
		t.Errorf("second Close should return nil: %v", err)
	}
}

func TestMultiVolReaderNonExistent(t *testing.T) {
	_, err := NewMultiVolReader("/nonexistent/archive")
	if err == nil {
		t.Error("expected error for nonexistent volume")
	}
}

func TestMultiVolWriterWriteAfterClose(t *testing.T) {
	dir := t.TempDir()
	prefix := filepath.Join(dir, "archive")
	mv, err := NewMultiVolWriter(prefix, 100)
	if err != nil {
		t.Fatal(err)
	}
	mv.Close()
	_, err = mv.Write([]byte("data"))
	if err == nil {
		t.Error("expected error writing after close")
	}
}
