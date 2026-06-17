package ops

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/harlandproj/tar-go/internal/cli"
	"github.com/harlandproj/tar-go/internal/filters"
	"github.com/harlandproj/tar-go/internal/increm"
)

func Create(opts *cli.Options) error {
	files := resolveFiles(opts.FileNames)
	if opts.FilesFrom != "" {
		files = append(files, readFileList(opts.FilesFrom)...)
	}
	if len(files) == 0 {
		files = []string{"."}
	}

	var xform *filters.Transform
	if opts.Transform != "" {
		var err error
		xform, err = filters.NewTransform(opts.Transform)
		if err != nil {
			return fmt.Errorf("invalid --transform: %w", err)
		}
	}

	var snap *increm.Snapshot
	if opts.ListedIncremental != "" {
		var err error
		snap, err = increm.LoadSnapshot(opts.ListedIncremental)
		if err != nil {
			return err
		}
	} else if opts.Incremental {
		snap = &increm.Snapshot{Timestamp: time.Now(), Entries: make(map[string]*increm.SnapshotEntry)}
	}

	w, err := openArchiveWriter(opts)
	if err != nil {
		return err
	}
	defer w.Close()

	var tw *tar.Writer
	if opts.MultiVolume && opts.TapeLength > 0 {
		tw = tar.NewWriter(&volWriter{opts: opts, baseW: w})
	} else {
		tw = tar.NewWriter(w)
	}
	defer tw.Close()

	if opts.VolumeLabel != "" {
		hdr := &tar.Header{
			Name:     opts.VolumeLabel,
			Typeflag: tar.TypeGNULongName,
			Size:     0,
			Mode:     0o644,
			ModTime:  time.Now(),
		}
		tw.WriteHeader(hdr)
	}

	excluder := filters.NewExcluder(opts)

	baseDir, err := os.Getwd()
	if err != nil {
		return err
	}

	i := 0
	for i < len(files) {
		name := files[i]
		if name == "-C" {
			i++
			if i < len(files) {
				baseDir = files[i]
			}
			i++
			continue
		}
		i++

		fullPath := name
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(baseDir, fullPath)
		}

		info, err := os.Lstat(fullPath)
		if err != nil {
			if opts.IgnoreFailedRead {
				fmt.Fprintf(cli.Stderr, "tar: %s: %v\n", name, err)
				continue
			}
			return err
		}

		newerSet := !opts.NewerMtime.IsZero()
		if newerSet && !info.ModTime().After(opts.NewerMtime) {
			continue
		}

		if info.IsDir() {
			err := filepath.Walk(fullPath, func(path string, fi os.FileInfo, err error) error {
				if err != nil {
					if opts.IgnoreFailedRead {
						fmt.Fprintf(cli.Stderr, "tar: %s: %v\n", path, err)
						return nil
					}
					return err
				}
				if opts.OneFileSystem && !isSameDevice(info, fi) {
					return filepath.SkipDir
				}
				rel, _ := filepath.Rel(baseDir, path)
				if excluder.Match(rel) {
					if fi.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
				if snap != nil && !snap.FileChanged(rel, fi.Size(), fi.ModTime()) {
					return nil
				}
				if opts.Dereference && fi.Mode()&os.ModeSymlink != 0 {
					realPath, err := filepath.EvalSymlinks(path)
					if err != nil {
						if opts.IgnoreFailedRead {
							return nil
						}
						return err
					}
					realInfo, err := os.Stat(realPath)
					if err != nil {
						if opts.IgnoreFailedRead {
							return nil
						}
						return err
					}
					return addFileToArchive(tw, path, realInfo, baseDir, opts, xform, snap)
				}
				return addFileToArchive(tw, path, fi, baseDir, opts, xform, snap)
			})
			if err != nil {
				return err
			}
		} else {
			rel, _ := filepath.Rel(baseDir, fullPath)
			if excluder.Match(rel) {
				continue
			}
			if snap != nil && !snap.FileChanged(rel, info.Size(), info.ModTime()) {
				continue
			}
			fi := info
			if opts.Dereference && fi.Mode()&os.ModeSymlink != 0 {
				realInfo, err := os.Stat(fullPath)
				if err != nil {
					if opts.IgnoreFailedRead {
						continue
					}
					return err
				}
				fi = realInfo
			}
			if err := addFileToArchive(tw, fullPath, fi, baseDir, opts, xform, snap); err != nil {
				return err
			}
		}
	}

	if snap != nil && opts.ListedIncremental != "" {
		snap.Save(opts.ListedIncremental)
	}

	return nil
}

func addFileToArchive(tw *tar.Writer, path string, info os.FileInfo, baseDir string, opts *cli.Options, xform *filters.Transform, snap *increm.Snapshot) error {
	link := ""
	if info.Mode()&os.ModeSymlink != 0 {
		link, _ = os.Readlink(path)
	}

	if opts.HardDereference && info.Mode()&os.ModeSymlink != 0 {
		realInfo, err := os.Stat(path)
		if err == nil {
			info = realInfo
			link = ""
		}
	}

	hdr, err := tar.FileInfoHeader(info, link)
	if err != nil {
		return err
	}

	relPath, _ := filepath.Rel(baseDir, path)
	hdr.Name = filepath.ToSlash(relPath)
	if !opts.AbsoluteNames {
		hdr.Name = strings.TrimPrefix(hdr.Name, "/")
	}
	if hdr.Name == "." {
		hdr.Name = ""
	}

	if xform != nil {
		hdr.Name = xform.Apply(hdr.Name)
		if opts.ShowTransformed && opts.Verbose > 0 {
			fmt.Fprintln(cli.Stderr, hdr.Name)
		}
	}

	if opts.Owner != "" || opts.Group != "" {
		if opts.Owner != "" {
			hdr.Uname = opts.Owner
		}
		if opts.Group != "" {
			hdr.Gname = opts.Group
		}
	}

	switch opts.SetMtimeMode {
	case cli.MtimeForce:
		hdr.ModTime = opts.Mtime
	case cli.MtimeClamp:
		if info.ModTime().After(opts.Mtime) {
			hdr.ModTime = opts.Mtime
		}
	}

	if opts.Touch {
		hdr.ModTime = time.Now()
	}

	setArchiveFormat(hdr, opts.ArchiveFormat)

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	if opts.Verbose > 0 {
		displayName := hdr.Name
		if displayName == "" {
			displayName = "."
		}
		fmt.Fprintln(cli.Stdout, displayName)
	}

	if info.IsDir() || hdr.Typeflag != tar.TypeReg {
		if snap != nil {
			snap.Entries[hdr.Name] = &increm.SnapshotEntry{
				Path: hdr.Name, Size: info.Size(), Mtime: info.ModTime(),
			}
		}
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		if opts.IgnoreFailedRead {
			return nil
		}
		return err
	}
	defer f.Close()

	if opts.Sparse {
		if err := writeSparseFile(tw, f, info, hdr.Name, opts); err != nil {
			return err
		}
	} else {
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}
	}

	if snap != nil {
		snap.Entries[hdr.Name] = &increm.SnapshotEntry{
			Path: hdr.Name, Size: info.Size(), Mtime: info.ModTime(),
		}
	}

	if opts.RemoveFiles {
		os.Remove(path)
	}

	return nil
}

func setArchiveFormat(hdr *tar.Header, fmt cli.ArchiveFormat) {
	switch fmt {
	case cli.FormatV7:
		hdr.Format = tar.FormatUnknown
	case cli.FormatOldGNU:
		hdr.Format = tar.FormatGNU
	case cli.FormatUstar:
		hdr.Format = tar.FormatUSTAR
	case cli.FormatGNU:
		hdr.Format = tar.FormatGNU
	case cli.FormatPOSIX:
		hdr.Format = tar.FormatPAX
	}
}

func expandFilesFrom(opts *cli.Options, files []string) []string {
	if opts.FilesFrom == "" {
		return files
	}
	f, err := os.Open(opts.FilesFrom)
	if err != nil {
		fmt.Fprintf(cli.Stderr, "tar: %s: %v\n", opts.FilesFrom, err)
		return files
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			files = append(files, line)
		}
	}
	return files
}

func readFileList(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(cli.Stderr, "tar: %s: %v\n", path, err)
		return nil
	}
	defer f.Close()
	var result []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			result = append(result, line)
		}
	}
	return result
}

type volWriter struct {
	opts   *cli.Options
	baseW  io.WriteCloser
	current *os.File
	num     int
	written int64
	prefix  string
}

func (vw *volWriter) Write(p []byte) (int, error) {
	if vw.current == nil {
		if err := vw.openVol(1); err != nil {
			return 0, err
		}
	}
	total := 0
	for len(p) > 0 {
		space := vw.opts.TapeLength - vw.written
		if space <= 0 {
			vw.num++
			if err := vw.openVol(vw.num); err != nil {
				return total, err
			}
			space = vw.opts.TapeLength
		}
		toWrite := int64(len(p))
		if toWrite > space {
			toWrite = space
		}
		n, err := vw.current.Write(p[:toWrite])
		total += n
		vw.written += int64(n)
		if err != nil {
			return total, err
		}
		p = p[toWrite:]
	}
	return total, nil
}

func (vw *volWriter) openVol(n int) error {
	if vw.current != nil {
		vw.current.Close()
	}
	name := vw.opts.ArchiveNames[0]
	if n > 1 {
		ext := filepath.Ext(name)
		base := name[:len(name)-len(ext)]
		name = fmt.Sprintf("%s-%d%s", base, n, ext)
	}
	fmt.Fprintf(cli.Stderr, "tar: volume #%d = %s\n", n, name)
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	vw.current = f
	vw.written = 0
	return nil
}

func (vw *volWriter) Close() error {
	if vw.current != nil {
		vw.current.Close()
	}
	return nil
}
