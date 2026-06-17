package ops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsSameDeviceSameFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	a, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if !isSameDevice(a, b) {
		t.Error("same file should be same device")
	}
}

func TestIsSameDeviceSameDir(t *testing.T) {
	dir := t.TempDir()
	a, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !isSameDevice(a, b) {
		t.Error("same dir should be same device")
	}
}

func TestIsSameDeviceSameFilesystem(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a")
	p2 := filepath.Join(dir, "b")
	if err := os.WriteFile(p1, []byte("1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p2, []byte("2"), 0644); err != nil {
		t.Fatal(err)
	}
	a, err := os.Stat(p1)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.Stat(p2)
	if err != nil {
		t.Fatal(err)
	}
	if !isSameDevice(a, b) {
		t.Error("files on same filesystem should be same device")
	}
}
