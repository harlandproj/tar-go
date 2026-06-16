package ops

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/harlandproj/tar-go/internal/cli"
)

func Delete(opts *cli.Options) error {
	archiveName := "tar.out"
	if len(opts.ArchiveNames) > 0 {
		archiveName = opts.ArchiveNames[0]
	}

	deleteSet := make(map[string]bool, len(opts.FileNames))
	for _, name := range opts.FileNames {
		deleteSet[filepath.ToSlash(filepath.Clean(name))] = true
	}

	tmpName := archiveName + ".tmp"
	src, err := os.Open(archiveName)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", archiveName, err)
	}
	defer src.Close()

	tmp, err := os.Create(tmpName)
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", tmpName, err)
	}
	defer tmp.Close()
	defer os.Remove(tmpName)

	tw := tar.NewWriter(tmp)
	tr := tar.NewReader(src)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading %s: %w", archiveName, err)
		}

		if deleteSet[hdr.Name] {
			if opts.Verbose > 0 {
				fmt.Fprintf(os.Stderr, "tar: deleting '%s'\n", hdr.Name)
			}
			if hdr.Size > 0 {
				io.Copy(io.Discard, tr)
			}
			continue
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if hdr.Size > 0 {
			if _, err := io.Copy(tw, tr); err != nil {
				return err
			}
		}
	}

	tw.Close()
	tmp.Close()
	src.Close()

	if err := os.Rename(tmpName, archiveName); err != nil {
		return fmt.Errorf("cannot replace %s: %w", archiveName, err)
	}

	return nil
}
