package ops

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/harlandproj/tar-go/internal/cli"
)

func Extract(opts *cli.Options) error {
	r, err := openArchiveReader(opts)
	if err != nil {
		return err
	}
	defer r.Close()

	processCDirectives(opts.FileNames)

	tr := tar.NewReader(r)

	oneTopLevel := opts.OneTopLevel

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := hdr.Name
		if !opts.AbsoluteNames {
			name = strings.TrimPrefix(name, "/")
		}
		if opts.StripComponents > 0 {
			parts := strings.SplitN(name, "/", opts.StripComponents+1)
			if len(parts) <= opts.StripComponents {
				continue
			}
			name = parts[opts.StripComponents]
		}
		if oneTopLevel != "" {
			if oneTopLevel == "-" {
				fmt.Fprintln(cli.Stdout, name)
				continue
			}
			name = filepath.Join(oneTopLevel, name)
		}

		if name == "" {
			continue
		}

		if opts.Verbose > 0 {
			fmt.Fprintln(cli.Stdout, name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(name, 0755); err != nil {
				return err
			}
			if opts.SamePermissions {
				os.Chmod(name, os.FileMode(hdr.Mode))
			}

		case tar.TypeReg, tar.TypeRegA:
			if opts.ToStdout {
				if _, err := io.Copy(cli.Stdout, tr); err != nil {
					return err
				}
				continue
			}
			if err := extractRegularFile(name, hdr, tr, opts); err != nil {
				return err
			}

		case tar.TypeSymlink:
			if opts.ToStdout {
				continue
			}
			if info, err := os.Lstat(name); err == nil {
				if opts.KeepOldFiles == cli.OldKeepOldFiles {
					continue
				}
				if opts.KeepOldFiles == cli.OldSkipOldFiles {
					continue
				}
				if opts.KeepOldFiles == cli.OldOverwrite {
					os.Remove(name)
				}
				_ = info
			}
			os.MkdirAll(filepath.Dir(name), 0755)
			os.Remove(name)
			if err := os.Symlink(hdr.Linkname, name); err != nil {
				return err
			}

		case tar.TypeLink:
			if opts.ToStdout {
				continue
			}
			if info, err := os.Lstat(name); err == nil {
				if opts.KeepOldFiles == cli.OldKeepOldFiles {
					continue
				}
				if opts.KeepOldFiles == cli.OldSkipOldFiles {
					continue
				}
				if opts.KeepOldFiles == cli.OldOverwrite {
					os.Remove(name)
				}
				_ = info
			}
			os.MkdirAll(filepath.Dir(name), 0755)
			os.Remove(name)
			if err := os.Link(hdr.Linkname, name); err != nil {
				return err
			}

		case tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
			fmt.Fprintf(cli.Stderr, "tar: %s: skipping special file\n", name)
			continue

		default:
			fmt.Fprintf(cli.Stderr, "tar: %s: unknown file type %c\n", name, hdr.Typeflag)
			continue
		}
	}

	return nil
}

func extractRegularFile(name string, hdr *tar.Header, tr *tar.Reader, opts *cli.Options) error {
	if info, err := os.Lstat(name); err == nil {
		if opts.KeepOldFiles == cli.OldKeepOldFiles {
			_, _ = io.Copy(io.Discard, tr)
			return nil
		}
		if opts.KeepOldFiles == cli.OldSkipOldFiles {
			_, _ = io.Copy(io.Discard, tr)
			return nil
		}
		if opts.KeepOldFiles == cli.OldOverwrite || opts.KeepOldFiles == cli.OldUnlinkFirst {
			os.Remove(name)
		}
		if opts.KeepOldFiles == cli.OldKeepNewerFiles {
			if !hdr.ModTime.After(info.ModTime()) {
				_, _ = io.Copy(io.Discard, tr)
				return nil
			}
			os.Remove(name)
		}
		_ = info
	}

	os.MkdirAll(filepath.Dir(name), 0755)

	f, err := os.Create(name)
	if err != nil {
		_, _ = io.Copy(io.Discard, tr)
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, tr); err != nil {
		return err
	}

	if opts.SamePermissions {
		os.Chmod(name, os.FileMode(hdr.Mode))
	}
	if !opts.Touch {
		os.Chtimes(name, time.Now(), hdr.ModTime)
	}

	return nil
}
