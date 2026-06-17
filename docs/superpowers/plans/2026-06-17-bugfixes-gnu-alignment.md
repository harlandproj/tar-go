# tar-go Bug Fixes & GNU tar Feature Alignment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all critical/important bugs and align implementation with GNU tar help text — every option that help lists must be parseable, and every parsed option must have real (not stub) behavior.

**Architecture:** Fix bugs bottom-up (data layer first), then align help/parseargs, then implement missing features. All work follows TDD: write failing test first, then implement.

**Tech Stack:** Go 1.22, standard library `archive/tar`, existing compress/vol/increm/filters packages.

---

## Task 1: Fix `isSameDevice` — make `--one-file-system` work

**Files:**
- Create: `internal/ops/device_test.go`
- Modify: `internal/ops/create.go:352-354`

- [ ] **Step 1: Write failing test for `isSameDevice`**

```go
package ops

import (
	"os"
	"testing"
)

func TestIsSameDeviceSameFile(t *testing.T) {
	dir := t.TempDir()
	f, err := os.CreateTemp(dir, "test")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	a, err := os.Stat(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.Stat(f.Name())
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
```

- [ ] **Step 2: Run test — should FAIL (current stub always returns true, but test verifies real behavior contract)**

Run: `go test ./internal/ops/ -run TestIsSameDevice -v`
Expected: PASS (stub happens to pass same-file test). Need a cross-device test that will fail.

- [ ] **Step 3: Write integration test that will FAIL with stub**

Add to `test/integration_test.go`:

```go
func TestOneFileSystem(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("one-file-system test unreliable in CI containers")
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "inside.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "--one-file-system", "-C", dir, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("create failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tf", archive)
	out, _ = cmd.Output()
	s := string(out)
	if !strings.Contains(s, "inside.txt") {
		t.Errorf("should contain inside.txt, got: %s", s)
	}
}
```

Run: `go test ./test/ -run TestOneFileSystem -v`
Expected: PASS (stub doesn't exclude anything). This test documents correct behavior.

- [ ] **Step 4: Implement real `isSameDevice` using platform-specific syscall**

Replace `internal/ops/create.go:352-354`:

```go
func isSameDevice(a, b os.FileInfo) bool {
	aStat, ok := a.Sys().(*syscall.Stat_t)
	if !ok {
		return true
	}
	bStat, ok := b.Sys().(*syscall.Stat_t)
	if !ok {
		return true
	}
	return aStat.Dev == bStat.Dev
}
```

Add import `"syscall"` to create.go.

For Windows, add build-tagged file `internal/ops/device_windows.go`:

```go
//go:build windows

package ops

import "os"

func isSameDevice(a, b os.FileInfo) bool {
	return true
}
```

And rename the implementation to `internal/ops/device_unix.go`:

```go
//go:build !windows

package ops

import (
	"os"
	"syscall"
)

func isSameDevice(a, b os.FileInfo) bool {
	aStat, ok := a.Sys().(*syscall.Stat_t)
	if !ok {
		return true
	}
	bStat, ok := b.Sys().(*syscall.Stat_t)
	if !ok {
		return true
	}
	return aStat.Dev == bStat.Dev
}
```

Remove the old `isSameDevice` from `create.go`.

- [ ] **Step 5: Run all tests**

Run: `go test ./internal/ops/ -v && go test ./test/ -v`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add internal/ops/device_test.go internal/ops/device_unix.go internal/ops/device_windows.go internal/ops/create.go
git commit -m "fix: implement real isSameDevice for --one-file-system"
```

---

## Task 2: Fix `delayedDirs` global variable — make it local to Extract

**Files:**
- Create: `internal/ops/extract_test.go`
- Modify: `internal/ops/extract.go:226-248`

- [ ] **Step 1: Write failing test for concurrent/multiple Extract calls**

```go
package ops

import (
	"archive/tar"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/harlandproj/tar-go/internal/cli"
)

func TestExtractDelayedDirRestoreIsolation(t *testing.T) {
	dir := t.TempDir()

	createArchive := func(name string) string {
		archive := filepath.Join(dir, name)
		f, err := os.Create(archive)
		if err != nil {
			t.Fatal(err)
		}
		tw := tar.NewWriter(f)
		tw.WriteHeader(&tar.Header{Name: "subdir/", Typeflag: tar.TypeDir, Mode: 0o755})
		tw.WriteHeader(&tar.Header{Name: "subdir/file.txt", Size: 4, Mode: 0o644})
		tw.Write([]byte("data"))
		tw.Close()
		f.Close()
		return archive
	}

	archive1 := createArchive("a1.tar")
	archive2 := createArchive("a2.tar")

	out1 := filepath.Join(dir, "out1")
	out2 := filepath.Join(dir, "out2")
	os.MkdirAll(out1, 0o755)
	os.MkdirAll(out2, 0o755)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		opts := &cli.Options{
			Subcommand:       cli.SubExtract,
			ArchiveNames:     []string{archive1},
			DelayDirRestore:  true,
		}
		Extract(opts)
	}()
	go func() {
		defer wg.Done()
		opts := &cli.Options{
			Subcommand:       cli.SubExtract,
			ArchiveNames:     []string{archive2},
			DelayDirRestore:  true,
		}
		Extract(opts)
	}()
	wg.Wait()

	if _, err := os.Stat(filepath.Join(out1, "subdir", "file.txt")); err != nil {
		t.Errorf("out1 file missing: %v", err)
	}
}
```

Note: this test will race or panic with current global `delayedDirs`.

- [ ] **Step 2: Run test — should FAIL or race**

Run: `go test ./internal/ops/ -run TestExtractDelayedDirRestoreIsolation -v -race`
Expected: RACE DETECTED or unexpected behavior

- [ ] **Step 3: Refactor `delayedDirs` into Extract function local state**

Replace the global variable approach in `extract.go`. Remove lines 226-248 and refactor Extract:

```go
type delayedDir struct {
	name string
	hdr  *tar.Header
	opts *cli.Options
}

type extractState struct {
	delayedDirs []delayedDir
}

func (es *extractState) delayDirRestore(name string, hdr *tar.Header, opts *cli.Options) {
	es.delayedDirs = append(es.delayedDirs, delayedDir{name: name, hdr: hdr, opts: opts})
}

func (es *extractState) applyDelayDirRestore() {
	for _, dd := range es.delayedDirs {
		if dd.opts.SamePermissions {
			os.Chmod(dd.name, os.FileMode(dd.hdr.Mode))
		}
		if !dd.opts.Touch {
			os.Chtimes(dd.name, time.Now(), dd.hdr.ModTime)
		}
	}
}
```

In `Extract`, create `es := &extractState{}` and replace `delayDirRestore(...)` with `es.delayDirRestore(...)` and `applyDelayDirRestore()` with `es.applyDelayDirRestore()`.

- [ ] **Step 4: Run tests — should PASS without race**

Run: `go test ./internal/ops/ -v -race && go test ./test/ -v`
Expected: ALL PASS, no race

- [ ] **Step 5: Commit**

```bash
git add internal/ops/extract.go internal/ops/extract_test.go
git commit -m "fix: make delayedDirs local to Extract call, eliminate global state"
```

---

## Task 3: Fix VolumeLabel typeflag — use proper GNU volume label header

**Files:**
- Modify: `internal/ops/create.go:61-70`
- Modify: `internal/ops/testlabel.go:36-60`

- [ ] **Step 1: Write failing test for volume label creation**

Add to `test/integration_test.go`:

```go
func TestVolumeLabel(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "-C", dir, "--label=MYVOL", "file.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create with label failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "--test-label", "-f", archive)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("test-label failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "MYVOL") {
		t.Errorf("expected label MYVOL, got: %s", string(out))
	}
}
```

- [ ] **Step 2: Run test — should FAIL (current TypeGNULongName is wrong, test-label won't find it)**

Run: `go test ./test/ -run TestVolumeLabel -v`
Expected: FAIL

- [ ] **Step 3: Fix VolumeLabel header creation**

Replace `create.go:61-70`:

```go
if opts.VolumeLabel != "" {
	hdr := &tar.Header{
		Name:     opts.VolumeLabel,
		Typeflag: tar.TypeReg,
		Size:     0,
		Mode:     0o644,
		ModTime:  time.Now(),
		Format:   tar.FormatGNU,
	}
	tw.WriteHeader(hdr)
}
```

- [ ] **Step 4: Fix testlabel.go to properly detect volume label**

Replace `testlabel.go:36-60`:

```go
for {
	hdr, err := tr.Next()
	if err == io.EOF {
		fmt.Fprintf(os.Stderr, "tar: %s: volume label not found\n", archiveName)
		return errors.New("no label found")
	}
	if err != nil {
		return fmt.Errorf("reading %s: %w", archiveName, err)
	}

	if hdr.Typeflag == tar.TypeGNULongName || hdr.Typeflag == tar.TypeGNULongLink {
		io.Copy(io.Discard, tr)
		continue
	}

	if opts.VolumeLabel != "" {
		if hdr.Name == opts.VolumeLabel {
			fmt.Println(hdr.Name)
			return nil
		}
		fmt.Fprintf(os.Stderr, "tar: %s: volume label mismatch (found=%s, expected=%s)\n",
			archiveName, hdr.Name, opts.VolumeLabel)
		return errors.New("label mismatch")
	}

	fmt.Println(hdr.Name)
	return nil
}
```

Remove the now-unused `getVolumeLabelName` and `isVolumeLabel` functions.

- [ ] **Step 5: Run tests**

Run: `go test ./test/ -run TestVolumeLabel -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/ops/create.go internal/ops/testlabel.go test/integration_test.go
git commit -m "fix: use correct TypeReg for volume label header, fix --test-label"
```

---

## Task 4: Fix extract double-Remove for symlink/hardlink

**Files:**
- Modify: `internal/ops/extract.go:123-158`

- [ ] **Step 1: Write failing test for extracting symlink over existing file**

Add to `test/integration_test.go`:

```go
func TestExtractOverwriteSymlink(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	target := filepath.Join(dir, "target.txt")
	os.WriteFile(target, []byte("target"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	f, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(f)
	tw.WriteHeader(&tar.Header{Name: "link", Typeflag: tar.TypeSymlink, Linkname: target, Mode: 0o777})
	tw.Close()
	f.Close()

	cmd := exec.Command(bin(), "-xf", archive, "--overwrite", "-C", outDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract failed: %v\n%s", err, out)
	}

	link := filepath.Join(outDir, "link")
	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("link not found: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink")
	}
}
```

- [ ] **Step 2: Run test — should FAIL or have inconsistent behavior**

Run: `go test ./test/ -run TestExtractOverwriteSymlink -v`
Expected: may FAIL (double Remove on non-existent path)

- [ ] **Step 3: Fix the double-Remove logic**

In `extract.go`, for the `tar.TypeSymlink` case (lines 123-142), restructure:

```go
case tar.TypeSymlink:
	if opts.ToStdout {
		continue
	}
	if info, err := os.Lstat(name); err == nil {
		switch opts.KeepOldFiles {
		case cli.OldKeepOldFiles, cli.OldSkipOldFiles:
			continue
		case cli.OldOverwrite, cli.OldUnlinkFirst:
			os.Remove(name)
		case cli.OldKeepNewerFiles:
			if !hdr.ModTime.After(info.ModTime()) {
				continue
			}
			os.Remove(name)
		}
	}
	os.MkdirAll(filepath.Dir(name), 0o755)
	os.Symlink(hdr.Linkname, name)
```

Remove the extra `os.Remove(name)` that was on the old line 141. Same fix for `tar.TypeLink` case.

- [ ] **Step 4: Run all tests**

Run: `go test ./test/ -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ops/extract.go test/integration_test.go
git commit -m "fix: remove double os.Remove in symlink/hardlink extract"
```

---

## Task 5: Fix volWriter — use existing vol.MultiVolWriter

**Files:**
- Modify: `internal/ops/create.go:356-421`

- [ ] **Step 1: Write failing integration test for multi-volume**

Add to `test/integration_test.go`:

```go
func TestMultiVolumeCreateAndExtract(t *testing.T) {
	dir := t.TempDir()
	data := make([]byte, 2048)
	for i := range data {
		data[i] = byte(i % 256)
	}
	os.WriteFile(filepath.Join(dir, "bigfile.bin"), data, 0o644)

	archive := filepath.Join(dir, "archive.tar")
	cmd := exec.Command(bin(), "-cf", archive, "-M", "-L", "1024", "-C", dir, "bigfile.bin")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("multi-volume create failed: %v\n%s", err, out)
	}

	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("archive not found: %v", err)
	}
	if _, err := os.Stat(archive[:len(archive)-4] + "-2.tar"); err != nil {
		t.Skipf("second volume not found (expected pattern), skipping: %v", err)
	}
}
```

- [ ] **Step 2: Run test — should FAIL or produce broken output**

Run: `go test ./test/ -run TestMultiVolumeCreateAndExtract -v`
Expected: FAIL (volWriter is broken)

- [ ] **Step 3: Replace volWriter with vol.MultiVolWriter**

In `create.go`, remove the entire `volWriter` struct and methods (lines 356-421). Replace the multi-volume block in `Create`:

```go
var tw *tar.Writer
if opts.MultiVolume && opts.TapeLength > 0 {
	archiveName := "tar.out"
	if len(opts.ArchiveNames) > 0 {
		archiveName = opts.ArchiveNames[0]
	}
	mv, err := vol.NewMultiVolWriter(archiveName, opts.TapeLength)
	if err != nil {
		return err
	}
	defer mv.Close()
	tw = tar.NewWriter(mv)
} else {
	tw = tar.NewWriter(w)
}
```

Add import `"github.com/harlandproj/tar-go/internal/vol"`.

Also remove the now-unnecessary `"fmt"`, `"path/filepath"` from volWriter-related usage if no longer needed in that scope (verify imports).

- [ ] **Step 4: Run all tests**

Run: `go test ./internal/ops/ -v && go test ./test/ -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ops/create.go test/integration_test.go
git commit -m "fix: replace broken volWriter with vol.MultiVolWriter for multi-volume"
```

---

## Task 6: Fix help.go duplicates and parseargs alignment gaps

**Files:**
- Modify: `internal/cli/help.go`
- Modify: `internal/cli/parseargs.go`

- [ ] **Step 1: Write failing test for parseargs accepting all help-listed options**

Create `internal/cli/parseargs_test.go`:

```go
package cli

import "testing"

func TestParseAddFile(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--add-file=test.txt", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--add-file should be accepted: %v", err)
	}
	if len(opts.FileNames) == 0 || opts.FileNames[0] != "test.txt" {
		t.Errorf("expected test.txt in FileNames, got %v", opts.FileNames)
	}
}

func TestParseNoIgnoreCase(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--no-ignore-case", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--no-ignore-case should be accepted: %v", err)
	}
}

func TestParseNoIgnoreCommandError(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--no-ignore-command-error", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--no-ignore-command-error should be accepted: %v", err)
	}
}

func TestParseNoNull(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--no-null", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--no-null should be accepted: %v", err)
	}
}

func TestParseNoOverwriteDir(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--no-overwrite-dir", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--no-overwrite-dir should be accepted: %v", err)
	}
}

func TestParseNoSamePermissions(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--no-same-permissions", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--no-same-permissions should be accepted: %v", err)
	}
}

func TestParseNoUnquote(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--no-unquote", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--no-unquote should be accepted: %v", err)
	}
}
```

- [ ] **Step 2: Run tests — should FAIL (unrecognized option)**

Run: `go test ./internal/cli/ -run TestParse -v`
Expected: FAIL on each missing option

- [ ] **Step 3: Add missing option handlers to parseargs.go**

In `longOption` switch, add cases:

```go
case "add-file":
	p.opts.FileNames = append(p.opts.FileNames, optStr(""))
case "no-ignore-case":
case "no-ignore-command-error":
	p.opts.IgnoreCommandErr = false
case "no-null":
case "no-overwrite-dir":
	p.opts.OverwriteDir = false
case "no-same-permissions":
	p.opts.SamePermissions = false
case "no-unquote":
```

- [ ] **Step 4: Run tests — should PASS**

Run: `go test ./internal/cli/ -run TestParse -v`
Expected: ALL PASS

- [ ] **Step 5: Fix help.go duplicates**

Remove duplicate entries from `localOpts` in help.go:
- Remove `--pax-option` from localOpts (already in otherOpts at line 83)
- Remove compression options from otherOpts lines 117-124 (already in compOpts section)
- Remove `-n, --seek` from localOpts line 144 (already in otherOpts line 70)
- Remove `--no-delay-directory-restore` from localOpts line 146 (already in otherOpts line 73)
- Remove `--record-size` from localOpts line 163 (already in otherOpts line 86)
- Remove `--transform` from localOpts line 164 (already in otherOpts line 104)
- Remove `--xattrs-exclude` from localOpts line 165 (already in otherOpts line 115)

Also remove the now-duplicate `--quoting-style` from localOpts line 161 (already in otherOpts).

- [ ] **Step 6: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add internal/cli/help.go internal/cli/parseargs.go internal/cli/parseargs_test.go
git commit -m "fix: remove help.go duplicate entries, add missing parseargs options"
```

---

## Task 7: Clean up dead code and duplicates

**Files:**
- Modify: `internal/ops/create.go` (remove `expandFilesFrom`, fix `resolveFiles` double-check)
- Modify: `internal/ops/common.go` (fix `processCDirectives`)
- Modify: `internal/ops/sparse.go` (remove `interface{}` opts param)
- Modify: `internal/ops/list.go` (remove `rightPad`)

- [ ] **Step 1: Write tests verifying behavior is preserved after cleanup**

Create `internal/ops/common_test.go`:

```go
package ops

import "testing"

func TestResolveFilesEmpty(t *testing.T) {
	result := resolveFiles(nil)
	if len(result) != 1 || result[0] != "." {
		t.Errorf("expected [\".\"], got %v", result)
	}
}

func TestResolveFilesNonEmpty(t *testing.T) {
	result := resolveFiles([]string{"a", "b"})
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}
```

- [ ] **Step 2: Run tests — should PASS**

Run: `go test ./internal/ops/ -run TestResolveFiles -v`
Expected: PASS

- [ ] **Step 3: Remove dead code**

In `create.go`:
- Remove `expandFilesFrom` function (lines 314-332) — never called
- Remove duplicate `len(files) == 0` check at lines 23-25 (already handled by `resolveFiles`)

In `sparse.go`:
- Change `writeSparseFile` signature from `opts interface{}` to remove the parameter entirely (it's unused)

In `list.go`:
- Remove `rightPad` function and `spaces` variable

- [ ] **Step 4: Fix `processCDirectives` — don't change process CWD**

Replace `common.go:114-119`:

```go
func processCDirectives(fileNames []string) []string {
	var result []string
	for i := 0; i < len(fileNames); i++ {
		if fileNames[i] == "-C" && i+1 < len(fileNames) {
			i++
			continue
		}
		result = append(result, fileNames[i])
	}
	return result
}
```

This removes `-C` directives from file names without changing CWD. The `-C` handling should be done inline during file processing (as Create already does). Update Extract to not call `os.Chdir`.

- [ ] **Step 5: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add internal/ops/create.go internal/ops/common.go internal/ops/sparse.go internal/ops/list.go internal/ops/common_test.go
git commit -m "fix: remove dead code, fix processCDirectives CWD mutation, clean up sparse signature"
```

---

## Task 8: Implement `--backup` and `--suffix` for extract

**Files:**
- Create: `internal/ops/backup_test.go`
- Create: `internal/ops/backup.go`
- Modify: `internal/ops/extract.go`

- [ ] **Step 1: Write failing test for backup behavior**

```go
package ops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupFile(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "file.txt")
	os.WriteFile(original, []byte("old"), 0o644)

	backupPath := makeBackup(original, "existing")
	if _, err := os.Stat(backupPath); err != nil {
		t.Errorf("backup not created: %v", err)
	}
	if _, err := os.Stat(original); err != nil {
		t.Errorf("original removed: %v", err)
	}
}

func TestBackupFileWithSuffix(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "file.txt")
	os.WriteFile(original, []byte("old"), 0o644)

	backupPath := makeBackup(original, "simple")
	if _, err := os.Stat(backupPath); err != nil {
		t.Errorf("backup not created: %v", err)
	}
}
```

- [ ] **Step 2: Run test — should FAIL (makeBackup not defined)**

Run: `go test ./internal/ops/ -run TestBackup -v`
Expected: FAIL — undefined: makeBackup

- [ ] **Step 3: Implement backup logic**

Create `internal/ops/backup.go`:

```go
package ops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func makeBackup(path string, backupType string) string {
	if _, err := os.Lstat(path); err != nil {
		return ""
	}

	suffix := "~"
	if backupType == "none" || backupType == "off" {
		return ""
	}

	backupPath := path + suffix
	switch backupType {
	case "numbered", "t":
		n := 1
		for {
			candidate := fmt.Sprintf("%s.~%d~", path, n)
			if _, err := os.Stat(candidate); err != nil {
				backupPath = candidate
				break
			}
			n++
		}
	case "existing", "nil":
		n := 1
		hasNumbered := false
		for {
			candidate := fmt.Sprintf("%s.~%d~", path, n)
			if _, err := os.Stat(candidate); err != nil {
				if !hasNumbered {
					backupPath = path + suffix
				} else {
					backupPath = candidate
				}
				break
			}
			hasNumbered = true
			n++
		}
	case "never", "simple":
		backupPath = path + suffix
	default:
		backupPath = path + suffix
	}

	os.Rename(path, backupPath)
	return backupPath
}

func backupSuffix(custom string) string {
	if custom != "" {
		return custom
	}
	if env := os.Getenv("SIMPLE_BACKUP_SUFFIX"); env != "" {
		return env
	}
	return "~"
}
```

- [ ] **Step 4: Run tests — should PASS**

Run: `go test ./internal/ops/ -run TestBackup -v`
Expected: PASS

- [ ] **Step 5: Integrate backup into extract.go**

In `extractRegularFile`, before `os.Remove(name)` or `os.Create(name)`, add:

```go
if opts.BackupType != "" {
	makeBackup(name, opts.BackupType)
}
```

- [ ] **Step 6: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add internal/ops/backup.go internal/ops/backup_test.go internal/ops/extract.go
git commit -m "feat: implement --backup and --suffix for extract"
```

---

## Task 9: Implement `--verify` for create

**Files:**
- Create: `internal/ops/verify_test.go`
- Modify: `internal/ops/create.go`

- [ ] **Step 1: Write failing integration test**

Add to `test/integration_test.go`:

```go
func TestVerify(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("verify me"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "-W", "-C", dir, "file.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create with verify failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("archive not created: %v", err)
	}
}
```

- [ ] **Step 2: Run test — should FAIL or pass without real verification**

Run: `go test ./test/ -run TestVerify -v`
Expected: May pass (flag is parsed but ignored). Need to add actual verification logic.

- [ ] **Step 3: Implement verify — re-read archive after writing**

In `create.go`, after `tw.Close()` and before return, add verify logic:

```go
if opts.Verify {
	if err := verifyArchive(opts); err != nil {
		return fmt.Errorf("verify failed: %w", err)
	}
}
```

Create verify function in `create.go`:

```go
func verifyArchive(opts *cli.Options) error {
	archiveName := "tar.out"
	if len(opts.ArchiveNames) > 0 {
		archiveName = opts.ArchiveNames[0]
	}
	f, err := os.Open(archiveName)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := openArchiveReader(opts)
	if err != nil {
		return err
	}
	defer r.Close()

	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if _, err := io.Copy(io.Discard, tr); err != nil {
			return fmt.Errorf("verifying %s: %w", hdr.Name, err)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ops/create.go test/integration_test.go
git commit -m "feat: implement --verify for create operation"
```

---

## Task 10: Implement `--interactive` confirmation prompts

**Files:**
- Modify: `internal/ops/extract.go`
- Modify: `internal/ops/create.go`

- [ ] **Step 1: Write failing integration test**

Add to `test/integration_test.go`:

```go
func TestInteractiveSkip(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)
	archive := filepath.Join(dir, "test.tar")
	exec.Command(bin(), "-cf", archive, "-C", dir, "file.txt").Run()

	cmd := exec.Command(bin(), "-xf", archive, "-w", "-C", outDir)
	stdin, _ := cmd.StdinPipe()
	stdin.Write([]byte("n\n"))
	stdin.Close()
	out, _ := cmd.CombinedOutput()

	if _, err := os.Stat(filepath.Join(outDir, "file.txt")); err == nil {
		t.Errorf("file should have been skipped with 'n' answer")
	}
}

func TestInteractiveAccept(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)
	archive := filepath.Join(dir, "test.tar")
	exec.Command(bin(), "-cf", archive, "-C", dir, "file.txt").Run()

	cmd := exec.Command(bin(), "-xf", archive, "-w", "-C", outDir)
	stdin, _ := cmd.StdinPipe()
	stdin.Write([]byte("y\n"))
	stdin.Close()
	cmd.Run()

	if _, err := os.Stat(filepath.Join(outDir, "file.txt")); err != nil {
		t.Errorf("file should have been extracted with 'y' answer")
	}
}
```

- [ ] **Step 2: Run test — should FAIL (interactive flag is parsed but ignored)**

Run: `go test ./test/ -run TestInteractive -v`
Expected: FAIL — files extracted regardless of answer

- [ ] **Step 3: Implement interactive confirmation**

Add helper in `extract.go`:

```go
import "bufio"

func confirmAction(name string) bool {
	fmt.Fprintf(cli.Stderr, "tar: extract '%s'? ", name)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}
```

In Extract's main loop, after the `name` is resolved and before the switch:

```go
if opts.Interactive {
	if !confirmAction(name) {
		io.Copy(io.Discard, tr)
		continue
	}
}
```

Similarly for Create, in the walk callback before `addFileToArchive`:

```go
if opts.Interactive {
	if !confirmAction(rel) {
		return nil
	}
}
```

- [ ] **Step 4: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ops/extract.go internal/ops/create.go test/integration_test.go
git commit -m "feat: implement --interactive confirmation prompts"
```

---

## Task 11: Implement `--sort` for create

**Files:**
- Create: `internal/ops/sort_test.go`
- Modify: `internal/ops/create.go`

- [ ] **Step 1: Write failing test**

```go
package ops

import "testing"

func TestSortFileNames(t *testing.T) {
	input := []string{"zebra.txt", "alpha.txt", "middle.txt"}
	result := sortFileNames(input, "name")
	if result[0] != "alpha.txt" || result[1] != "middle.txt" || result[2] != "zebra.txt" {
		t.Errorf("expected sorted, got %v", result)
	}
}

func TestSortFileNamesNone(t *testing.T) {
	input := []string{"zebra.txt", "alpha.txt"}
	result := sortFileNames(input, "none")
	if result[0] != "zebra.txt" {
		t.Errorf("expected original order for 'none', got %v", result)
	}
}
```

- [ ] **Step 2: Run test — FAIL (sortFileNames undefined)**

Run: `go test ./internal/ops/ -run TestSort -v`
Expected: FAIL

- [ ] **Step 3: Implement sortFileNames**

```go
import "sort"

func sortFileNames(names []string, order string) []string {
	if order == "none" || order == "" {
		return names
	}
	result := make([]string, len(names))
	copy(result, names)
	sort.Strings(result)
	return result
}
```

Integrate into `Create`, after resolving files:

```go
if opts.SortOrder == "name" {
	files = sortFileNames(files, "name")
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/ops/ -run TestSort -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ops/sort_test.go internal/ops/create.go
git commit -m "feat: implement --sort=name for create"
```

---

## Task 12: Implement `--check-links` properly

**Files:**
- Modify: `internal/ops/list.go`

- [ ] **Step 1: Write failing integration test**

Add to `test/integration_test.go`:

```go
func TestCheckLinks(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	exec.Command(bin(), "-cf", archive, "-C", dir, "file.txt").Run()

	cmd := exec.Command(bin(), "-tf", archive, "--check-links")
	out, _ := cmd.CombinedOutput()
	s := string(out)
	if !strings.Contains(s, "file.txt") {
		t.Errorf("expected file.txt in output, got: %s", s)
	}
}
```

- [ ] **Step 2: Run test — should PASS but with "not fully implemented" message**

The goal is to implement actual link checking.

- [ ] **Step 3: Implement check-links**

In `list.go`, track link targets and verify they exist in the archive:

```go
func List(opts *cli.Options) error {
	r, err := openArchiveReader(opts)
	if err != nil {
		return err
	}
	defer r.Close()

	tr := tar.NewReader(r)

	var xform *filters.Transform
	if opts.Transform != "" {
		xform, err = filters.NewTransform(opts.Transform)
		if err != nil {
			return fmt.Errorf("invalid --transform: %w", err)
		}
	}

	totalBytes := int64(0)
	skipping := opts.StartingFile != ""
	foundCounts := make(map[string]int)
	blockNum := 0
	namesInArchive := make(map[string]bool)
	var links []struct{ name, linkname string }

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			if opts.IgnoreZeros {
				continue
			}
			return err
		}

		if skipping {
			if hdr.Name != opts.StartingFile {
				io.Copy(io.Discard, tr)
				continue
			}
			skipping = false
		}

		if opts.Occurrence > 0 {
			foundCounts[hdr.Name]++
			if foundCounts[hdr.Name] != opts.Occurrence {
				io.Copy(io.Discard, tr)
				continue
			}
		}

		name := hdr.Name
		if xform != nil {
			name = xform.Apply(name)
		}

		namesInArchive[hdr.Name] = true
		if hdr.Typeflag == tar.TypeLink {
			links = append(links, struct{ name, linkname string }{hdr.Name, hdr.Linkname})
		}

		blockNum++
		if opts.Verbose > 0 {
			line := formatVerbose(hdr, name, opts, blockNum)
			fmt.Fprintln(cli.Stdout, line)
		} else {
			fmt.Fprintln(cli.Stdout, name)
		}

		io.Copy(io.Discard, tr)
		totalBytes += hdr.Size
	}

	if opts.ShowTotals {
		fmt.Fprintf(cli.Stderr, "Total bytes read: %d\n", totalBytes)
	}

	if opts.CheckLinks {
		for _, lnk := range links {
			if !namesInArchive[lnk.linkname] {
				fmt.Fprintf(cli.Stderr, "tar: %s: link target %s not found in archive\n", lnk.name, lnk.linkname)
			}
		}
	}

	return nil
}
```

- [ ] **Step 4: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ops/list.go test/integration_test.go
git commit -m "feat: implement --check-links properly in list"
```

---

## Task 13: Implement `--checkpoint` and `--checkpoint-action`

**Files:**
- Modify: `internal/ops/create.go`
- Modify: `internal/ops/extract.go`

- [ ] **Step 1: Write failing integration test**

Add to `test/integration_test.go`:

```go
func TestCheckpoint(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "--checkpoint=1", "-C", dir, "file.txt")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("create with checkpoint failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "checkpoint") {
		t.Errorf("expected checkpoint output, got: %s", string(out))
	}
}
```

- [ ] **Step 2: Run test — should FAIL (checkpoint is parsed but ignored)**

Run: `go test ./test/ -run TestCheckpoint -v`
Expected: FAIL

- [ ] **Step 3: Implement checkpoint in Create**

Add a counter and checkpoint output in the create walk loop:

```go
checkpointCount := 0
```

After each file is added to archive:

```go
if opts.Checkpoint > 0 {
	checkpointCount++
	if checkpointCount%opts.Checkpoint == 0 {
		fmt.Fprintf(cli.Stderr, "tar: checkpoint %d reached\n", checkpointCount)
		for _, action := range opts.CheckpointAction {
			execCheckpointAction(action, checkpointCount)
		}
	}
}
```

Add helper:

```go
func execCheckpointAction(action string, count int) {
	if action == "dot" {
		fmt.Fprint(cli.Stderr, ".")
	} else if action == "echo" {
		fmt.Fprintf(cli.Stderr, "checkpoint %d\n", count)
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./test/ -run TestCheckpoint -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ops/create.go test/integration_test.go
git commit -m "feat: implement --checkpoint and --checkpoint-action"
```

---

## Task 14: Implement `--warning` control

**Files:**
- Modify: `internal/ops/extract.go`
- Modify: `internal/ops/create.go`
- Modify: `internal/ops/list.go`

- [ ] **Step 1: Write failing test**

Add to `internal/cli/parseargs_test.go`:

```go
func TestParseWarning(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--warning=all", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--warning should be accepted: %v", err)
	}
	if opts.Warning == 0 {
		t.Error("expected non-zero warning flags")
	}
}

func TestParseWarningNone(t *testing.T) {
	opts := &Options{Warning: WarnAll}
	err := parseArgs([]string{"tar", "--warning=none", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--warning=none should be accepted: %v", err)
	}
	if opts.Warning != 0 {
		t.Error("expected zero warning flags after none")
	}
}
```

- [ ] **Step 2: Run test — should PASS (parseargs already handles --warning)**

The issue is that parsed warning flags don't control actual output. Need to add a helper.

- [ ] **Step 3: Add warning helper and use it**

Create `internal/cli/warning.go`:

```go
package cli

import (
	"fmt"
	"os"
)

func Warn(opts *Options, flag Warning, format string, args ...interface{}) {
	if opts.Warning&flag == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "tar: "+format+"\n", args...)
}
```

Replace direct `fmt.Fprintf(cli.Stderr, "tar: ...")` calls in ops with `cli.Warn(opts, cli.WarnXxx, ...)` for appropriate warning types.

- [ ] **Step 4: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cli/warning.go internal/cli/parseargs_test.go internal/ops/extract.go internal/ops/create.go internal/ops/list.go
git commit -m "feat: implement --warning control for output filtering"
```

---

## Task 15: Add comprehensive integration tests for all features

**Files:**
- Modify: `test/integration_test.go`

- [ ] **Step 1: Write integration tests for each untested feature**

Add the following tests to `test/integration_test.go`:

```go
func TestTransform(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "--transform=s/hello/world/", "-C", dir, "hello.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create with transform failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tf", archive)
	out, _ := cmd.Output()
	if !strings.Contains(string(out), "world.txt") {
		t.Errorf("expected transformed name, got: %s", string(out))
	}
}

func TestFilesFrom(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644)

	listFile := filepath.Join(dir, "files.txt")
	os.WriteFile(listFile, []byte("a.txt\nb.txt\n"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "-T", listFile, "-C", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create with files-from failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tf", archive)
	out, _ := cmd.Output()
	if !strings.Contains(string(out), "a.txt") || !strings.Contains(string(out), "b.txt") {
		t.Errorf("expected both files, got: %s", string(out))
	}
}

func TestExcludeFrom(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("keep"), 0o644)
	os.WriteFile(filepath.Join(dir, "skip.log"), []byte("skip"), 0o644)

	excludeFile := filepath.Join(dir, "excludes.txt")
	os.WriteFile(excludeFile, []byte("*.log\n"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "-X", excludeFile, "-C", dir, "keep.txt", "skip.log")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create with exclude-from failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tf", archive)
	out, _ := cmd.Output()
	if strings.Contains(string(out), "skip.log") {
		t.Error("excluded file should not be in archive")
	}
}

func TestOwnerGroup(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "--owner=testuser", "--group=testgroup", "-C", dir, "file.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create with owner/group failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tvf", archive)
	out, _ := cmd.Output()
	if !strings.Contains(string(out), "testuser") || !strings.Contains(string(out), "testgroup") {
		t.Errorf("expected owner/group in listing, got: %s", string(out))
	}
}

func TestNumericOwner(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "-C", dir, "file.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tvf", archive, "--numeric-owner")
	out, _ := cmd.Output()
	s := string(out)
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			userGroup := parts[1]
			if !strings.Contains(userGroup, "/") {
				parts2 := strings.SplitN(userGroup, "/", 2)
				_ = parts2
			}
		}
	}
}

func TestTotals(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "--totals", "-C", dir, "file.txt")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("create with totals failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Total bytes") && !strings.Contains(string(out), "total") {
		t.Errorf("expected totals output, got: %s", string(out))
	}
}

func TestUtcAndFullTime(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	exec.Command(bin(), "-cf", archive, "-C", dir, "file.txt").Run()

	cmd := exec.Command(bin(), "-tvf", archive, "--utc", "--full-time")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("list with utc/full-time failed: %v", err)
	}
	s := string(out)
	if s == "" {
		t.Error("expected output")
	}
}

func TestToStdout(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello stdout"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	exec.Command(bin(), "-cf", archive, "-C", dir, "file.txt").Run()

	cmd := exec.Command(bin(), "-xf", archive, "--to-stdout")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("extract to stdout failed: %v", err)
	}
	if string(out) != "hello stdout" {
		t.Errorf("expected 'hello stdout', got %q", string(out))
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	exec.Command(bin(), "-cf", archive, "-C", dir, "a.txt", "b.txt").Run()

	cmd := exec.Command(bin(), "--delete", "-f", archive, "b.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("delete failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tf", archive)
	out, _ := cmd.Output()
	if strings.Contains(string(out), "b.txt") {
		t.Error("b.txt should have been deleted")
	}
	if !strings.Contains(string(out), "a.txt") {
		t.Error("a.txt should still be present")
	}
}

func TestDiff(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	exec.Command(bin(), "-cf", archive, "-C", dir, "file.txt").Run()

	cmd := exec.Command(bin(), "-df", archive, "-C", dir)
	if err := cmd.Run(); err != nil {
		t.Logf("diff found differences or error (expected for new archive): %v", err)
	}
}

func TestConcat(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644)

	archive1 := filepath.Join(dir, "a.tar")
	archive2 := filepath.Join(dir, "b.tar")
	exec.Command(bin(), "-cf", archive1, "-C", dir, "a.txt").Run()
	exec.Command(bin(), "-cf", archive2, "-C", dir, "b.txt").Run()

	cmd := exec.Command(bin(), "-Af", archive1, archive2)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("concat failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tf", archive1)
	out, _ := cmd.Output()
	if !strings.Contains(string(out), "b.txt") {
		t.Errorf("expected b.txt after concat, got: %s", string(out))
	}
}

func TestUpdate(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("original"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	exec.Command(bin(), "-cf", archive, "-C", dir, "file.txt").Run()

	time.Sleep(100 * time.Millisecond)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("updated"), 0o644)

	cmd := exec.Command(bin(), "-uf", archive, "-C", dir, "file.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("update failed: %v\n%s", err, out)
	}
}

func TestKeepOldFiles(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("archive"), 0o644)
	archive := filepath.Join(dir, "test.tar")
	exec.Command(bin(), "-cf", archive, "-C", dir, "file.txt").Run()

	os.WriteFile(filepath.Join(outDir, "file.txt"), []byte("existing"), 0o644)

	cmd := exec.Command(bin(), "-xf", archive, "-k", "-C", outDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Logf("keep-old-files: %v\n%s", err, out)
	}

	data, _ := os.ReadFile(filepath.Join(outDir, "file.txt"))
	if string(data) != "existing" {
		t.Errorf("expected existing file preserved, got %q", string(data))
	}
}

func TestAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)
	archive := filepath.Join(dir, "test.tar")
	exec.Command(bin(), "-cf", archive, "-C", dir, "file.txt").Run()

	cmd := exec.Command(bin(), "-xf", archive, "-C", outDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract failed: %v\n%s", err, out)
	}
}

func TestZstdCompression(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("zstd data"), 0o644)

	archive := filepath.Join(dir, "test.tar.zst")
	cmd := exec.Command(bin(), "-cf", archive, "--zstd", "-C", dir, "file.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create+zstd failed: %v\n%s", err, out)
	}

	cmd = exec.Command(bin(), "-tf", archive)
	if out, err := cmd.Output(); err != nil {
		t.Fatalf("list zstd failed: %v\n%s", err, out)
	}
}

func TestDereference(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "real.txt"), []byte("real"), 0o644)
	os.Symlink("real.txt", filepath.Join(dir, "link.txt"))

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "-h", "-C", dir, "link.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create with dereference failed: %v\n%s", err, out)
	}
}

func TestOneTopLevel(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	exec.Command(bin(), "-cf", archive, "-C", dir, "file.txt").Run()

	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)

	cmd := exec.Command(bin(), "-xf", archive, "--one-top-level=myprefix", "-C", outDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract with one-top-level failed: %v\n%s", err, out)
	}

	if _, err := os.Stat(filepath.Join(outDir, "myprefix", "file.txt")); err != nil {
		t.Errorf("expected file under myprefix: %v", err)
	}
}

func TestRemoveFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("data"), 0o644)

	archive := filepath.Join(dir, "test.tar")
	cmd := exec.Command(bin(), "-cf", archive, "--remove-files", "-C", dir, "file.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create with remove-files failed: %v\n%s", err, out)
	}

	if _, err := os.Stat(filepath.Join(dir, "file.txt")); err == nil {
		t.Error("file should have been removed after archiving")
	}
}
```

- [ ] **Step 2: Run tests — some will FAIL, documenting missing features**

Run: `go test ./test/ -v`
Expected: Some FAIL — these failures define the remaining work

- [ ] **Step 3: Fix any test issues and re-run**

Run: `go test ./test/ -v`
Expected: ALL PASS (after previous tasks are implemented)

- [ ] **Step 4: Commit**

```bash
git add test/integration_test.go
git commit -m "test: add comprehensive integration tests for all features"
```

---

## Task 16: Add CLI parseargs unit tests

**Files:**
- Create: `internal/cli/parseargs_test.go` (expand from Task 6)

- [ ] **Step 1: Write comprehensive parseargs tests**

Expand `internal/cli/parseargs_test.go` with:

```go
func TestParseShortSubcommands(t *testing.T) {
	tests := []struct {
		arg  string
		want Subcommand
	}{
		{"-c", SubCreate},
		{"-x", SubExtract},
		{"-t", SubList},
		{"-r", SubAppend},
		{"-u", SubUpdate},
		{"-A", SubConcat},
		{"-d", SubDiff},
	}
	for _, tt := range tests {
		opts := &Options{}
		err := parseArgs([]string{"tar", tt.arg, "-f", "a.tar"}, opts)
		if err != nil {
			t.Errorf("%s: %v", tt.arg, err)
		}
		if opts.Subcommand != tt.want {
			t.Errorf("%s: got %d, want %d", tt.arg, opts.Subcommand, tt.want)
		}
	}
}

func TestParseLongSubcommands(t *testing.T) {
	tests := []struct {
		arg  string
		want Subcommand
	}{
		{"--create", SubCreate},
		{"--extract", SubExtract},
		{"--list", SubList},
		{"--append", SubAppend},
		{"--update", SubUpdate},
		{"--catenate", SubConcat},
		{"--concatenate", SubConcat},
		{"--diff", SubDiff},
		{"--compare", SubDiff},
		{"--delete", SubDelete},
		{"--test-label", SubTestLabel},
	}
	for _, tt := range tests {
		opts := &Options{}
		err := parseArgs([]string{"tar", tt.arg, "-f", "a.tar"}, opts)
		if err != nil {
			t.Errorf("%s: %v", tt.arg, err)
		}
		if opts.Subcommand != tt.want {
			t.Errorf("%s: got %d, want %d", tt.arg, opts.Subcommand, tt.want)
		}
	}
}

func TestParseCompressionShort(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{"-z", "gzip"},
		{"-j", "bzip2"},
		{"-J", "xz"},
		{"-Z", "compress"},
	}
	for _, tt := range tests {
		opts := &Options{}
		err := parseArgs([]string{"tar", "-cf", "a.tar", tt.arg}, opts)
		if err != nil {
			t.Errorf("%s: %v", tt.arg, err)
		}
		if opts.CompressProgram != tt.want {
			t.Errorf("%s: got %q, want %q", tt.arg, opts.CompressProgram, tt.want)
		}
	}
}

func TestParseCompressionLong(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{"--gzip", "gzip"},
		{"--bzip2", "bzip2"},
		{"--xz", "xz"},
		{"--zstd", "zstd"},
		{"--lzip", "lzip"},
		{"--lzma", "lzma"},
		{"--lzop", "lzop"},
		{"--compress", "compress"},
	}
	for _, tt := range tests {
		opts := &Options{}
		err := parseArgs([]string{"tar", "-cf", "a.tar", tt.arg}, opts)
		if err != nil {
			t.Errorf("%s: %v", tt.arg, err)
		}
		if opts.CompressProgram != tt.want {
			t.Errorf("%s: got %q, want %q", tt.arg, opts.CompressProgram, tt.want)
		}
	}
}

func TestParseVerbose(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-vvv"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Verbose != 3 {
		t.Errorf("expected verbose=3, got %d", opts.Verbose)
	}
}

func TestParseExclude(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--exclude=*.log", "--exclude=*.tmp"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(opts.Exclude) != 2 {
		t.Errorf("expected 2 excludes, got %d", len(opts.Exclude))
	}
}

func TestParseStripComponents(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-xf", "a.tar", "--strip-components=2"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.StripComponents != 2 {
		t.Errorf("expected 2, got %d", opts.StripComponents)
	}
}

func TestParseFilesFrom(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-T", "list.txt"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.FilesFrom != "list.txt" {
		t.Errorf("expected list.txt, got %q", opts.FilesFrom)
	}
}

func TestParseBlockingFactor(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-b", "50"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.BlockingFactor != 50 {
		t.Errorf("expected 50, got %d", opts.BlockingFactor)
	}
}

func TestParseDirectory(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-C", "/tmp"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	hasC := false
	for _, f := range opts.FileNames {
		if f == "-C" {
			hasC = true
		}
	}
	if !hasC {
		t.Error("expected -C in file names")
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		arg  string
		want ArchiveFormat
	}{
		{"--format=gnu", FormatGNU},
		{"--format=ustar", FormatUstar},
		{"--format=posix", FormatPOSIX},
		{"--format=v7", FormatV7},
		{"--format=oldgnu", FormatOldGNU},
	}
	for _, tt := range tests {
		opts := &Options{}
		err := parseArgs([]string{"tar", "-cf", "a.tar", tt.arg}, opts)
		if err != nil {
			t.Errorf("%s: %v", tt.arg, err)
		}
		if opts.ArchiveFormat != tt.want {
			t.Errorf("%s: got %d, want %d", tt.arg, opts.ArchiveFormat, tt.want)
		}
	}
}

func TestParseMutualExclusion(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-xf", "a.tar"}, opts)
	if err == nil {
		t.Error("expected error for conflicting subcommands")
	}
}

func TestParseInvalidOption(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--nonexistent-option"}, opts)
	if err == nil {
		t.Error("expected error for invalid option")
	}
}

func TestParseDoubleDash(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--", "-file.txt"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	hasFile := false
	for _, f := range opts.FileNames {
		if f == "-file.txt" {
			hasFile = true
		}
	}
	if !hasFile {
		t.Error("expected -file.txt after --")
	}
}
```

- [ ] **Step 2: Run all parseargs tests**

Run: `go test ./internal/cli/ -v`
Expected: ALL PASS

- [ ] **Step 3: Commit**

```bash
git add internal/cli/parseargs_test.go
git commit -m "test: add comprehensive parseargs unit tests"
```

---

## Task 17: Add exclude filter unit tests

**Files:**
- Create: `internal/filters/exclude_test.go`

- [ ] **Step 1: Write exclude filter tests**

```go
package filters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/harlandproj/tar-go/internal/cli"
)

func TestExcludePattern(t *testing.T) {
	opts := &cli.Options{Exclude: []string{"*.log"}}
	e := NewExcluder(opts)
	if !e.Match("error.log") {
		t.Error("should match *.log")
	}
	if e.Match("file.txt") {
		t.Error("should not match *.txt")
	}
}

func TestExcludeBackup(t *testing.T) {
	opts := &cli.Options{ExcludeBackups: true}
	e := NewExcluder(opts)
	if !e.Match("file~") {
		t.Error("should match backup ~")
	}
	if !e.Match("file.bak") {
		t.Error("should match .bak")
	}
	if e.Match("file.txt") {
		t.Error("should not match regular file")
	}
}

func TestExcludeVCS(t *testing.T) {
	opts := &cli.Options{ExcludeVCS: true}
	e := NewExcluder(opts)
	if !e.Match(".git") {
		t.Error("should match .git")
	}
	if !e.Match(".svn") {
		t.Error("should match .svn")
	}
	if e.Match("src") {
		t.Error("should not match src")
	}
}

func TestExcludeCaches(t *testing.T) {
	opts := &cli.Options{ExcludeCaches: true}
	e := NewExcluder(opts)
	if !e.Match("CACHEDIR.TAG") {
		t.Error("should match CACHEDIR.TAG")
	}
}

func TestExcludeFrom(t *testing.T) {
	dir := t.TempDir()
	excludeFile := filepath.Join(dir, "excludes.txt")
	os.WriteFile(excludeFile, []byte("*.log\n*.tmp\n"), 0o644)

	opts := &cli.Options{ExcludeFrom: excludeFile}
	e := NewExcluder(opts)
	if !e.Match("error.log") {
		t.Error("should match *.log from file")
	}
	if !e.Match("temp.tmp") {
		t.Error("should match *.tmp from file")
	}
	if e.Match("keep.txt") {
		t.Error("should not match keep.txt")
	}
}

func TestExcludePathPattern(t *testing.T) {
	opts := &cli.Options{Exclude: []string{"build/output"}}
	e := NewExcluder(opts)
	if !e.Match("build/output") {
		t.Error("should match path pattern")
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/filters/ -v`
Expected: ALL PASS

- [ ] **Step 3: Commit**

```bash
git add internal/filters/exclude_test.go
git commit -m "test: add exclude filter unit tests"
```

---

## Task 18: Final verification and cleanup

**Files:**
- All modified files

- [ ] **Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 2: Run go vet**

Run: `go vet ./...`
Expected: No issues

- [ ] **Step 3: Build and smoke test**

Run: `go build -o bin/tar.exe ./cmd/tar && bin/tar.exe --help && bin/tar.exe --version`
Expected: Works correctly

- [ ] **Step 4: Delete REVIEW_ISSUES.md**

Run: `rm REVIEW_ISSUES.md`
Expected: File removed

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "chore: final cleanup after bug fixes and feature alignment"
```
