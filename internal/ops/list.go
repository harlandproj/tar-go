package ops

import (
	"archive/tar"
	"fmt"
	"io"
	"os/user"
	"strconv"
	"time"

	"github.com/harlandproj/tar-go/internal/cli"
)

func List(opts *cli.Options) error {
	r, err := openArchiveReader(opts)
	if err != nil {
		return err
	}
	defer r.Close()

	tr := tar.NewReader(r)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		hdr.ModTime = hdr.ModTime.In(time.Local)

		if opts.Verbose > 0 {
			line := formatVerbose(hdr, opts)
			fmt.Fprintln(cli.Stdout, line)
		} else {
			fmt.Fprintln(cli.Stdout, hdr.Name)
		}

		if _, err := io.Copy(io.Discard, tr); err != nil {
			return err
		}
	}
	return nil
}

func formatVerbose(hdr *tar.Header, opts *cli.Options) string {
	perm := hdr.FileInfo().Mode().String()
	if hdr.Typeflag == tar.TypeDir && perm[0] != 'd' {
		perm = "d" + perm[1:]
	} else if hdr.Typeflag == tar.TypeSymlink {
		perm = "L" + perm[1:]
	}

	if perm == "-rwxrwxrwx" {
		perm = "----------"
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

	gname := hdr.Gname
	if gname == "" {
		g, err := user.LookupGroupId(strconv.Itoa(hdr.Gid))
		if err == nil {
			gname = g.Name
		} else {
			gname = strconv.Itoa(hdr.Gid)
		}
	}

	size := strconv.FormatInt(hdr.Size, 10)

	t := hdr.ModTime
	if opts.Utc {
		t = t.UTC()
	}
	dateStr := t.Format("2006-01-02 15:04")

	name := hdr.Name
	if hdr.Typeflag == tar.TypeSymlink {
		name = name + " -> " + hdr.Linkname
	} else if hdr.Typeflag == tar.TypeLink {
		name = name + " link to " + hdr.Linkname
	}

	return fmt.Sprintf("%s %s/%-8s %s %s %s",
		perm, uname, gname, rightPad(size, 8), dateStr, name)
}

func rightPad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + spaces[:width-len(s)]
}

var spaces = "                "
