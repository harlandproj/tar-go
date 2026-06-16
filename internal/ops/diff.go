package ops

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/harlandproj/tar-go/internal/cli"
)

var DiffExitError = &cli.ExitError{Code: 1, Message: "differences found"}

func Diff(opts *cli.Options) error {
	rc, err := openArchiveReader(opts)
	if err != nil {
		return err
	}
	defer rc.Close()

	tr := tar.NewReader(rc)

	diffFound := false
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading archive: %w", err)
		}

		name := filepath.FromSlash(hdr.Name)
		fi, err := os.Lstat(name)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "tar: %s: does not exist\n", name)
				diffFound = true
			} else {
				fmt.Fprintf(os.Stderr, "tar: %s: cannot stat: %v\n", name, err)
				diffFound = true
			}
			continue
		}

		if diffEntry(hdr, fi, name, opts) {
			diffFound = true
		}
	}

	if diffFound {
		return DiffExitError
	}
	return nil
}

func diffEntry(hdr *tar.Header, fi os.FileInfo, name string, opts *cli.Options) bool {
	different := false

	diskType := fileTypeFromMode(fi.Mode())
	archiveType := hdr.Typeflag
	if archiveType != diskType {
		fmt.Fprintf(os.Stderr, "tar: %s: type differs (archive=%s, disk=%s)\n",
			name, tarTypeChar(archiveType), tarTypeChar(diskType))
		different = true
	}

	if archiveType == tar.TypeReg || archiveType == tar.TypeRegA {
		if hdr.Size != fi.Size() {
			fmt.Fprintf(os.Stderr, "tar: %s: size differs (archive=%d, disk=%d)\n",
				name, hdr.Size, fi.Size())
			different = true
		}
	}

	if !opts.Touch {
		archiveMtime := hdr.ModTime
		diskMtime := fi.ModTime()
		if opts.Utc {
			archiveMtime = archiveMtime.UTC()
			diskMtime = diskMtime.UTC()
		}
		secDiff := archiveMtime.Unix() - diskMtime.Unix()
		if secDiff < 0 {
			secDiff = -secDiff
		}
		if secDiff > 1 && !hdr.ModTime.IsZero() {
			fmt.Fprintf(os.Stderr, "tar: %s: mod time differs (archive=%s, disk=%s)\n",
				name, hdr.ModTime.Format("2006-01-02 15:04:05"),
				fi.ModTime().Format("2006-01-02 15:04:05"))
			different = true
		}
	}

	if int64(hdr.Mode)&0o777 != int64(fi.Mode())&0o777 {
		fmt.Fprintf(os.Stderr, "tar: %s: mode differs (archive=%o, disk=%o)\n",
			name, hdr.Mode&0o777, fi.Mode()&0o777)
		different = true
	}

	return different
}

func fileTypeFromMode(mode os.FileMode) byte {
	switch {
	case mode&os.ModeDir != 0:
		return tar.TypeDir
	case mode&os.ModeSymlink != 0:
		return tar.TypeSymlink
	case mode&os.ModeNamedPipe != 0:
		return tar.TypeFifo
	case mode&os.ModeCharDevice != 0:
		return tar.TypeChar
	case mode&os.ModeDevice != 0:
		return tar.TypeBlock
	default:
		return tar.TypeReg
	}
}

func tarTypeChar(t byte) string {
	switch t {
	case tar.TypeReg, tar.TypeRegA:
		return "file"
	case tar.TypeLink:
		return "link"
	case tar.TypeSymlink:
		return "symlink"
	case tar.TypeChar:
		return "char"
	case tar.TypeBlock:
		return "block"
	case tar.TypeDir:
		return "dir"
	case tar.TypeFifo:
		return "fifo"
	default:
		return "unknown"
	}
}
