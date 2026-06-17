package ops

import (
	"archive/tar"
	"fmt"
	"io"
	"os"

	"github.com/harlandproj/tar-go/internal/cli"
)

func Concat(opts *cli.Options) error {
	archiveName := "tar.out"
	if len(opts.ArchiveNames) > 0 {
		archiveName = opts.ArchiveNames[0]
	}

	f, err := os.OpenFile(archiveName, os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", archiveName, err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	pos := fi.Size()
	if pos >= 1024 {
		pos -= 1024
		if _, err := f.Seek(pos, io.SeekStart); err != nil {
			return err
		}
		if err := f.Truncate(pos); err != nil {
			return err
		}
	}

	tw := tar.NewWriter(f)
	defer tw.Close()

	for _, srcName := range opts.FileNames {
		if opts.Verbose > 0 {
			fmt.Fprintf(os.Stderr, "tar: %s\n", srcName)
		}
		if err := copyArchiveEntries(tw, srcName); err != nil {
			return err
		}
	}
	return nil
}

func copyArchiveEntries(tw *tar.Writer, srcName string) error {
	src, err := os.Open(srcName)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", srcName, err)
	}
	defer src.Close()

	tr := tar.NewReader(src)
	buf := make([]byte, 32*1024)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading %s: %w", srcName, err)
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if hdr.Size > 0 {
			if _, err := io.CopyBuffer(tw, tr, buf); err != nil {
				return fmt.Errorf("copying entry %s from %s: %w", hdr.Name, srcName, err)
			}
		}
	}
	return nil
}
