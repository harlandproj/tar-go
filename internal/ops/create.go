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

func Create(opts *cli.Options) error {
	w, err := openArchiveWriter(opts)
	if err != nil {
		return err
	}
	defer w.Close()

	tw := tar.NewWriter(w)
	defer tw.Close()

	excluder := filters.NewExcluder(opts)
	files := resolveFiles(opts.FileNames)

	baseDir, err := os.Getwd()
	if err != nil {
		return err
	}

	i := 0
	for i < len(files) {
		name := files[i]
		if name == "-C" {
			i++
			if i < len(files) {
				baseDir = files[i]
			}
			i++
			continue
		}
		i++

		if !filepath.IsAbs(name) {
			name = filepath.Join(baseDir, name)
		}

		info, err := os.Lstat(name)
		if err != nil {
			if opts.IgnoreFailedRead {
				fmt.Fprintf(cli.Stderr, "tar: %s: %v\n", name, err)
				continue
			}
			return err
		}

		if info.IsDir() {
			err := filepath.Walk(name, func(path string, fi os.FileInfo, err error) error {
				if err != nil {
					if opts.IgnoreFailedRead {
						fmt.Fprintf(cli.Stderr, "tar: %s: %v\n", path, err)
						return nil
					}
					return err
				}
				rel, _ := filepath.Rel(baseDir, path)
				if excluder.Match(rel) {
					if fi.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
				if opts.Dereference && fi.Mode()&os.ModeSymlink != 0 {
					realPath, err := filepath.EvalSymlinks(path)
					if err != nil {
						if opts.IgnoreFailedRead {
							return nil
						}
						return err
					}
					realInfo, err := os.Stat(realPath)
					if err != nil {
						if opts.IgnoreFailedRead {
							return nil
						}
						return err
					}
					return addFileToArchive(tw, path, realInfo, baseDir, opts)
				}
				return addFileToArchive(tw, path, fi, baseDir, opts)
			})
			if err != nil {
				return err
			}
		} else {
			rel, _ := filepath.Rel(baseDir, name)
			if excluder.Match(rel) {
				continue
			}
			fi := info
			if opts.Dereference && fi.Mode()&os.ModeSymlink != 0 {
				realInfo, err := os.Stat(name)
				if err != nil {
					if opts.IgnoreFailedRead {
						continue
					}
					return err
				}
				fi = realInfo
			}
			if err := addFileToArchive(tw, name, fi, baseDir, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func addFileToArchive(tw *tar.Writer, path string, info os.FileInfo, baseDir string, opts *cli.Options) error {
	link := ""
	if info.Mode()&os.ModeSymlink != 0 {
		link, _ = os.Readlink(path)
	}
	hdr, err := tar.FileInfoHeader(info, link)
	if err != nil {
		return err
	}

	relPath, _ := filepath.Rel(baseDir, path)
	hdr.Name = filepath.ToSlash(relPath)
	if !opts.AbsoluteNames {
		hdr.Name = strings.TrimPrefix(hdr.Name, "/")
	}
	if hdr.Name == "." {
		hdr.Name = ""
	}

	if opts.Touch {
		hdr.ModTime = time.Now()
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	if opts.Verbose > 0 {
		displayName := hdr.Name
		if displayName == "" {
			displayName = "."
		}
		fmt.Fprintln(cli.Stdout, displayName)
	}

	if info.IsDir() || hdr.Typeflag != tar.TypeReg {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		if opts.IgnoreFailedRead {
			return nil
		}
		return err
	}
	defer f.Close()
	_, err = io.Copy(tw, f)

	if err == nil && opts.RemoveFiles {
		os.Remove(path)
	}

	return err
}
