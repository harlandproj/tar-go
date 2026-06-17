package vol

import (
	"io"
	"os"
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
