package ops

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/harlandproj/tar-go/internal/cli"
)

func TestLabel(opts *cli.Options) error {
	archiveName := "tar.out"
	if len(opts.ArchiveNames) > 0 {
		archiveName = opts.ArchiveNames[0]
	}

	f, err := os.Open(archiveName)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", archiveName, err)
	}
	defer f.Close()

	tr := tar.NewReader(f)
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
}
