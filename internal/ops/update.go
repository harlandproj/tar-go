package ops

import (
	"archive/tar"
	"fmt"
	"io"
	"os"

	"github.com/harlandproj/tar-go/internal/cli"
)

func Update(opts *cli.Options) error {
	archiveName := "tar.out"
	if len(opts.ArchiveNames) > 0 {
		archiveName = opts.ArchiveNames[0]
	}

	archiveInfo, err := os.Stat(archiveName)
	if err != nil {
		return fmt.Errorf("cannot stat %s: %w", archiveName, err)
	}
	archiveMtime := archiveInfo.ModTime()

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

	added := 0
	for _, file := range opts.FileNames {
		info, err := os.Lstat(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tar: %s: %v\n", file, err)
			if !opts.IgnoreFailedRead {
				return err
			}
			continue
		}

		if !info.ModTime().After(archiveMtime) {
			if opts.Verbose > 0 {
				fmt.Fprintf(os.Stderr, "tar: %s: file is unchanged; not dumped\n", file)
			}
			continue
		}

		if opts.Verbose > 0 {
			fmt.Println(file)
		}
		if err := addFileToArchive(tw, file, info, ".", opts); err != nil {
			return err
		}
		added++
	}

	if added == 0 && opts.Verbose > 0 {
		fmt.Fprintln(os.Stderr, "tar: no files to update in archive")
	}
	return nil
}
