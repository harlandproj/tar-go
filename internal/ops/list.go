package ops

import (
	"archive/tar"
	"fmt"
	"io"
	"os/user"
	"strconv"

	"github.com/harlandproj/tar-go/internal/cli"
	"github.com/harlandproj/tar-go/internal/filters"
)

func List(opts *cli.Options) error {
	r, err := openArchiveReader(opts)
	if err != nil {
		return err
	}
	defer r.Close()

	tr := tar.NewReader(r)

	var xform *filters.Transform
	if opts.Transform != "" {
		xform, err = filters.NewTransform(opts.Transform)
		if err != nil {
			return fmt.Errorf("invalid --transform: %w", err)
		}
	}

	totalBytes := int64(0)
	skipping := opts.StartingFile != ""
	foundCounts := make(map[string]int)
	blockNum := 0
	namesInArchive := make(map[string]bool)
	var hardLinks []struct {
		name     string
		linkname string
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			if opts.IgnoreZeros {
				continue
			}
			return err
		}

		if skipping {
			if hdr.Name != opts.StartingFile {
				io.Copy(io.Discard, tr)
				continue
			}
			skipping = false
		}

		if opts.Occurrence > 0 {
			foundCounts[hdr.Name]++
			if foundCounts[hdr.Name] != opts.Occurrence {
				io.Copy(io.Discard, tr)
				continue
			}
		}

		name := hdr.Name
		if xform != nil {
			name = xform.Apply(name)
		}

		namesInArchive[hdr.Name] = true
		if hdr.Typeflag == tar.TypeLink {
			hardLinks = append(hardLinks, struct {
				name     string
				linkname string
			}{hdr.Name, hdr.Linkname})
		}

		blockNum++
		if opts.Verbose > 0 {
			line := formatVerbose(hdr, name, opts, blockNum)
			fmt.Fprintln(cli.Stdout, line)
		} else {
			fmt.Fprintln(cli.Stdout, name)
		}

		io.Copy(io.Discard, tr)
		totalBytes += hdr.Size
	}

	if opts.ShowTotals {
		fmt.Fprintf(cli.Stderr, "Total bytes read: %d\n", totalBytes)
	}

	if opts.CheckLinks {
		for _, lnk := range hardLinks {
			if !namesInArchive[lnk.linkname] {
				fmt.Fprintf(cli.Stderr, "tar: %s: link target %s not found in archive\n", lnk.name, lnk.linkname)
			}
		}
	}

	return nil
}

func formatVerbose(hdr *tar.Header, name string, opts *cli.Options, blockNum int) string {
	perm := hdr.FileInfo().Mode().String()
	switch hdr.Typeflag {
	case tar.TypeDir:
		if perm[0] != 'd' {
			perm = "d" + perm[1:]
		}
	case tar.TypeSymlink:
		perm = "l" + perm[1:]
	case tar.TypeLink:
		perm = "h" + perm[1:]
	}

	uname := hdr.Uname
	if uname == "" {
		u, err := user.LookupId(strconv.Itoa(hdr.Uid))
		if err == nil {
			uname = u.Username
		} else {
			uname = strconv.Itoa(hdr.Uid)
		}
	}
	if opts.NumericOwner {
		uname = strconv.Itoa(hdr.Uid)
	}

	gname := hdr.Gname
	if gname == "" {
		g, err := user.LookupGroupId(strconv.Itoa(hdr.Gid))
		if err == nil {
			gname = g.Name
		} else {
			gname = strconv.Itoa(hdr.Gid)
		}
	}
	if opts.NumericOwner {
		gname = strconv.Itoa(hdr.Gid)
	}

	size := strconv.FormatInt(hdr.Size, 10)

	t := hdr.ModTime
	if opts.Utc {
		t = t.UTC()
	}
	var dateStr string
	if opts.FullTime {
		dateStr = t.Format("2006-01-02 15:04:05.999999999 -0700")
	} else {
		dateStr = t.Format("2006-01-02 15:04")
	}

	displayName := name
	if hdr.Typeflag == tar.TypeSymlink {
		displayName = name + " -> " + hdr.Linkname
	} else if hdr.Typeflag == tar.TypeLink {
		displayName = name + " link to " + hdr.Linkname
	}

	result := fmt.Sprintf("%s %s/%s %8s %s %s",
		perm, uname, gname, size, dateStr, displayName)

	if opts.BlockNumber {
		result = fmt.Sprintf("block %d: %s", blockNum, result)
	}

	return result
}
