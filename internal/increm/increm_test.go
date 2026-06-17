package increm

import (
	"os"
	"testing"
	"time"
)

func TestSnapshotSaveLoad(t *testing.T) {
	dir, _ := os.MkdirTemp("", "snap-*")
	defer os.RemoveAll(dir)
	path := dir + "/snapshot"

	snap := &Snapshot{
		Timestamp: time.Now(),
		Entries:   make(map[string]*SnapshotEntry),
	}
	snap.Entries["/test/file"] = &SnapshotEntry{
		Path: "/test/file", Size: 100,
		Mtime: time.Now(),
	}

	if err := snap.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadSnapshot(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := loaded.Entries["/test/file"]; !ok {
		t.Error("entry not found after reload")
	}
}

func TestSnapshotEmpty(t *testing.T) {
	dir, _ := os.MkdirTemp("", "snap-*")
	defer os.RemoveAll(dir)
	path := dir + "/snapshot"

	snap := &Snapshot{
		Timestamp: time.Now(),
		Entries:   make(map[string]*SnapshotEntry),
	}
	if err := snap.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadSnapshot(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(loaded.Entries))
	}
}

func TestSnapshotMultipleEntries(t *testing.T) {
	dir, _ := os.MkdirTemp("", "snap-*")
	defer os.RemoveAll(dir)
	path := dir + "/snapshot"

	snap := &Snapshot{
		Timestamp: time.Now(),
		Entries:   make(map[string]*SnapshotEntry),
	}
	now := time.Now()
	snap.Entries["/a"] = &SnapshotEntry{Path: "/a", Size: 1, Mtime: now}
	snap.Entries["/b"] = &SnapshotEntry{Path: "/b", Size: 2, Mtime: now}
	snap.Entries["/c"] = &SnapshotEntry{Path: "/c", Size: 3, Mtime: now}

	if err := snap.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadSnapshot(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(loaded.Entries))
	}
	if loaded.Entries["/b"].Size != 2 {
		t.Errorf("expected size 2, got %d", loaded.Entries["/b"].Size)
	}
}

func TestSnapshotFileChanged(t *testing.T) {
	dir, _ := os.MkdirTemp("", "snap-*")
	defer os.RemoveAll(dir)
	_ = dir + "/snapshot" // path reserved for future use

	oldTime := time.Now().Add(-time.Hour)
	snap := &Snapshot{
		Timestamp: time.Now(),
		Entries:   make(map[string]*SnapshotEntry),
	}
	snap.Entries["/test"] = &SnapshotEntry{Path: "/test", Size: 100, Mtime: oldTime}

	if snap.FileChanged("/test", 100, oldTime) {
		t.Error("expected unchanged file")
	}
	if !snap.FileChanged("/test", 200, oldTime) {
		t.Error("expected changed file (size differs)")
	}
	if !snap.FileChanged("/test", 100, time.Now()) {
		t.Error("expected changed file (mtime differs)")
	}
	if !snap.FileChanged("/missing", 0, time.Now()) {
		t.Error("new file should be changed")
	}
}

func TestLoadSnapshotMissing(t *testing.T) {
	snap, err := LoadSnapshot("/nonexistent/path")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(snap.Entries))
	}
}
