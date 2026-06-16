package ops

import (
	"archive/tar"
	"fmt"
	"io"
	"os"

	"github.com/harlandproj/tar-go/internal/cli"
)

func Append(opts *cli.Options) error {
	archiveName := "tar.out"
	if len(opts.ArchiveNames) > 0 {
		archiveName = opts.ArchiveNames[0]
	}

	f, err := os.OpenFile(archiveName, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", archiveName, err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.Size() >= 1024 {
		if _, err := f.Seek(-1024, io.SeekEnd); err != nil {
			return err
		}
	}

	tw := tar.NewWriter(f)
	defer tw.Close()

	for _, file := range opts.FileNames {
		info, err := os.Lstat(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tar: %s: %v\n", file, err)
			if !opts.IgnoreFailedRead {
				return err
			}
			continue
		}
		if opts.Verbose > 0 {
			fmt.Println(file)
		}
		if err := addFileToArchive(tw, file, info, ".", opts); err != nil {
			return err
		}
	}
	return nil
}
