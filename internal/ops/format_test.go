package ops

import (
	"archive/tar"
	"testing"
	"time"

	"github.com/harlandproj/tar-go/internal/cli"
)

func TestSetArchiveFormat(t *testing.T) {
	tests := []struct {
		fmt   cli.ArchiveFormat
		want  tar.Format
	}{
		{cli.FormatV7, tar.FormatUnknown},
		{cli.FormatOldGNU, tar.FormatGNU},
		{cli.FormatUstar, tar.FormatUSTAR},
		{cli.FormatGNU, tar.FormatGNU},
		{cli.FormatPOSIX, tar.FormatPAX},
	}
	for _, tt := range tests {
		hdr := &tar.Header{}
		setArchiveFormat(hdr, tt.fmt)
		if hdr.Format != tt.want {
			t.Errorf("format %d: got %d, want %d", tt.fmt, hdr.Format, tt.want)
		}
	}
}

func TestIsSparseType(t *testing.T) {
	if !isSparseType(&tar.Header{Typeflag: tar.TypeGNUSparse}) {
		t.Error("expected true for TypeGNUSparse")
	}
	if isSparseType(&tar.Header{Typeflag: tar.TypeReg}) {
		t.Error("expected false for TypeReg")
	}
}

func TestFormatVerboseDir(t *testing.T) {
	hdr := &tar.Header{
		Name:     "dir/",
		Typeflag: tar.TypeDir,
		Mode:     0o755,
		Size:     0,
		ModTime:  time.Now(),
		Uid:      0,
		Gid:      0,
	}
	opts := &cli.Options{}
	result := formatVerbose(hdr, "dir/", opts, 1)
	if len(result) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestFormatVerboseSymlink(t *testing.T) {
	hdr := &tar.Header{
		Name:     "link",
		Typeflag: tar.TypeSymlink,
		Linkname: "target",
		Mode:     0o777,
		Size:     0,
		ModTime:  time.Now(),
		Uid:      0,
		Gid:      0,
	}
	opts := &cli.Options{}
	result := formatVerbose(hdr, "link", opts, 1)
	if len(result) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestFormatVerboseNumericOwner(t *testing.T) {
	hdr := &tar.Header{
		Name:     "file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0o644,
		Size:     100,
		ModTime:  time.Now(),
		Uid:      1000,
		Gid:      1000,
	}
	opts := &cli.Options{NumericOwner: true}
	result := formatVerbose(hdr, "file.txt", opts, 1)
	if len(result) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestFormatVerboseUtc(t *testing.T) {
	hdr := &tar.Header{
		Name:     "file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0o644,
		Size:     100,
		ModTime:  time.Now(),
		Uid:      0,
		Gid:      0,
	}
	opts := &cli.Options{Utc: true}
	result := formatVerbose(hdr, "file.txt", opts, 1)
	if len(result) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestFormatVerboseFullTime(t *testing.T) {
	hdr := &tar.Header{
		Name:     "file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0o644,
		Size:     100,
		ModTime:  time.Now(),
		Uid:      0,
		Gid:      0,
	}
	opts := &cli.Options{FullTime: true}
	result := formatVerbose(hdr, "file.txt", opts, 1)
	if len(result) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestFormatVerboseBlockNumber(t *testing.T) {
	hdr := &tar.Header{
		Name:     "file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0o644,
		Size:     100,
		ModTime:  time.Now(),
		Uid:      0,
		Gid:      0,
	}
	opts := &cli.Options{BlockNumber: true}
	result := formatVerbose(hdr, "file.txt", opts, 5)
	if len(result) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestFormatVerboseHardLink(t *testing.T) {
	hdr := &tar.Header{
		Name:     "hardlink",
		Typeflag: tar.TypeLink,
		Linkname: "original",
		Mode:     0o644,
		Size:     0,
		ModTime:  time.Now(),
		Uid:      0,
		Gid:      0,
	}
	opts := &cli.Options{}
	result := formatVerbose(hdr, "hardlink", opts, 1)
	if len(result) == 0 {
		t.Error("expected non-empty output")
	}
}
