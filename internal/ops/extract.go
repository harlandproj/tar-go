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
	"github.com/harlandproj/tar-go/internal/filters"
)

func Extract(opts *cli.Options) error {
	r, err := openArchiveReader(opts)
	if err != nil {
		return err
	}
	defer r.Close()

	processCDirectives(opts.FileNames)

	tr := tar.NewReader(r)

	var xform *filters.Transform
	if opts.Transform != "" {
		xform, err = filters.NewTransform(opts.Transform)
		if err != nil {
			return fmt.Errorf("invalid --transform: %w", err)
		}
	}

	oneTopLevel := opts.OneTopLevel
	skipping := opts.StartingFile != ""
	foundCounts := make(map[string]int)
	es := &extractState{}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := hdr.Name

		if skipping {
			if name != opts.StartingFile {
				io.Copy(io.Discard, tr)
				continue
			}
			skipping = false
		}

		if opts.Occurrence > 0 {
			foundCounts[name]++
			if foundCounts[name] != opts.Occurrence {
				io.Copy(io.Discard, tr)
				continue
			}
		}

		if xform != nil {
			name = xform.Apply(name)
		}

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
			if err := os.MkdirAll(name, 0o755); err != nil {
				if !opts.IgnoreFailedRead {
					return err
				}
			}
			if opts.SamePermissions {
				os.Chmod(name, os.FileMode(hdr.Mode))
			}
			if !opts.DelayDirRestore {
				if !opts.Touch {
					os.Chtimes(name, time.Now(), hdr.ModTime)
				}
			} else {
				es.delayDirRestore(name, hdr, opts)
			}

		case tar.TypeReg, tar.TypeRegA:
			if opts.ToStdout {
				io.Copy(cli.Stdout, tr)
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
				switch opts.KeepOldFiles {
				case cli.OldKeepOldFiles, cli.OldSkipOldFiles:
					continue
				case cli.OldOverwrite, cli.OldUnlinkFirst:
					os.Remove(name)
				case cli.OldKeepNewerFiles:
					if !hdr.ModTime.After(info.ModTime()) {
						continue
					}
					os.Remove(name)
				default:
					os.Remove(name)
				}
			}
			os.MkdirAll(filepath.Dir(name), 0o755)
			os.Symlink(hdr.Linkname, name)

		case tar.TypeLink:
			if opts.ToStdout {
				continue
			}
			if _, err := os.Lstat(name); err == nil {
				switch opts.KeepOldFiles {
				case cli.OldKeepOldFiles, cli.OldSkipOldFiles:
					continue
				case cli.OldOverwrite, cli.OldUnlinkFirst:
					os.Remove(name)
				default:
					os.Remove(name)
				}
			}
			os.MkdirAll(filepath.Dir(name), 0o755)
			os.Link(hdr.Linkname, name)

		case tar.TypeGNUSparse:
			if opts.ToStdout {
				io.Copy(io.Discard, tr)
				continue
			}
			if err := extractRegularFile(name, hdr, tr, opts); err != nil {
				return err
			}

		default:
			io.Copy(io.Discard, tr)
		}
	}

	es.applyDelayDirRestore()

	return nil
}

func extractRegularFile(name string, hdr *tar.Header, tr *tar.Reader, opts *cli.Options) error {
	if info, err := os.Lstat(name); err == nil {
		switch opts.KeepOldFiles {
		case cli.OldKeepOldFiles:
			io.Copy(io.Discard, tr)
			return nil
		case cli.OldSkipOldFiles:
			io.Copy(io.Discard, tr)
			return nil
		case cli.OldKeepNewerFiles:
			if !hdr.ModTime.After(info.ModTime()) {
				io.Copy(io.Discard, tr)
				return nil
			}
			os.Remove(name)
		case cli.OldOverwrite, cli.OldUnlinkFirst:
			os.Remove(name)
		}
	}

	if opts.RecursiveUnlink {
		os.RemoveAll(name)
	}

	os.MkdirAll(filepath.Dir(name), 0o755)

	f, err := os.Create(name)
	if err != nil {
		io.Copy(io.Discard, tr)
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

type delayedDir struct {
	name string
	hdr  *tar.Header
	opts *cli.Options
}

type extractState struct {
	delayedDirs []delayedDir
}

func (es *extractState) delayDirRestore(name string, hdr *tar.Header, opts *cli.Options) {
	es.delayedDirs = append(es.delayedDirs, delayedDir{name: name, hdr: hdr, opts: opts})
}

func (es *extractState) applyDelayDirRestore() {
	for _, dd := range es.delayedDirs {
		if dd.opts.SamePermissions {
			os.Chmod(dd.name, os.FileMode(dd.hdr.Mode))
		}
		if !dd.opts.Touch {
			os.Chtimes(dd.name, time.Now(), dd.hdr.ModTime)
		}
	}
}
