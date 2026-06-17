package increm

import (
	"encoding/gob"
	"os"
	"time"
)

type SnapshotEntry struct {
	Path  string
	Size  int64
	Mtime time.Time
}

type Snapshot struct {
	Timestamp time.Time
	Level     int
	Entries   map[string]*SnapshotEntry
}

func (s *Snapshot) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	return enc.Encode(s)
}

func LoadSnapshot(path string) (*Snapshot, error) {
	snap := &Snapshot{
		Timestamp: time.Now(),
		Entries:   make(map[string]*SnapshotEntry),
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return snap, nil
		}
		return snap, err
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	if err := dec.Decode(snap); err != nil {
		return snap, err
	}
	if snap.Entries == nil {
		snap.Entries = make(map[string]*SnapshotEntry)
	}
	return snap, nil
}

func (s *Snapshot) FileChanged(path string, size int64, mtime time.Time) bool {
	entry, ok := s.Entries[path]
	if !ok {
		return true
	}
	if entry.Size != size {
		return true
	}
	if !entry.Mtime.Equal(mtime) {
		return true
	}
	return false
}
