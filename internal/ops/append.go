package ops

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/harlandproj/tar-go/internal/cli"
)

func Append(opts *cli.Options) error {
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

	files := resolveFiles(opts.FileNames)
	baseDir, _ := os.Getwd()

	for i := 0; i < len(files); i++ {
		name := files[i]
		if name == "-C" {
			i++
			if i < len(files) {
				baseDir = files[i]
			}
			continue
		}

		fullPath := name
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(baseDir, fullPath)
		}

		info, err := os.Lstat(fullPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tar: %s: %v\n", name, err)
			if !opts.IgnoreFailedRead {
				return err
			}
			continue
		}
		if opts.Verbose > 0 {
			fmt.Println(name)
		}
		if err := addFileToArchive(tw, fullPath, info, baseDir, opts); err != nil {
			return err
		}
	}
	return nil
}
