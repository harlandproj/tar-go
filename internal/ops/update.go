package ops

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	added := 0

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

		if !info.ModTime().After(archiveMtime) {
			if opts.Verbose > 0 {
				fmt.Fprintf(os.Stderr, "tar: %s: file is unchanged; not dumped\n", name)
			}
			continue
		}

		if opts.Verbose > 0 {
			fmt.Println(name)
		}
		if err := addFileToArchive(tw, fullPath, info, baseDir, opts); err != nil {
			return err
		}
		added++
	}

	if added == 0 && opts.Verbose > 0 {
		fmt.Fprintln(os.Stderr, "tar: no files to update in archive")
	}
	return nil
}
