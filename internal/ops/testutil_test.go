package ops

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/harlandproj/tar-go/internal/cli"
)

type testEnv struct {
	Dir     string
	Archive string
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	dir := t.TempDir()
	return &testEnv{
		Dir:     dir,
		Archive: filepath.Join(dir, "test.tar"),
	}
}

func (e *testEnv) WriteFile(name, content string) string {
	path := filepath.Join(e.Dir, name)
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte(content), 0o644)
	return path
}

func (e *testEnv) Mkdir(name string) string {
	path := filepath.Join(e.Dir, name)
	os.MkdirAll(path, 0o755)
	return path
}

func (e *testEnv) Symlink(target, link string) string {
	linkPath := filepath.Join(e.Dir, link)
	os.MkdirAll(filepath.Dir(linkPath), 0o755)
	os.Symlink(target, linkPath)
	return linkPath
}

func (e *testEnv) CreateArchive(files map[string]string) {
	f, err := os.Create(e.Archive)
	if err != nil {
		panic(err)
	}
	tw := tar.NewWriter(f)
	for name, content := range files {
		hdr := &tar.Header{
			Name:    name,
			Mode:    0o644,
			Size:    int64(len(content)),
			ModTime: time.Now(),
		}
		if strings.HasSuffix(name, "/") {
			hdr.Typeflag = tar.TypeDir
			hdr.Size = 0
			tw.WriteHeader(hdr)
		} else {
			tw.WriteHeader(hdr)
			tw.Write([]byte(content))
		}
	}
	tw.Close()
	f.Close()
}

func (e *testEnv) CreateArchiveWithLinks(links map[string]struct {
	Linkname string
	Typeflag  byte
}) {
	f, err := os.Create(e.Archive)
	if err != nil {
		panic(err)
	}
	tw := tar.NewWriter(f)
	for name, link := range links {
		hdr := &tar.Header{
			Name:     name,
			Typeflag: link.Typeflag,
			Linkname: link.Linkname,
			Mode:     0o644,
			ModTime:  time.Now(),
		}
		tw.WriteHeader(hdr)
	}
	tw.Close()
	f.Close()
}

func (e *testEnv) ReadArchive() []string {
	f, err := os.Open(e.Archive)
	if err != nil {
		return nil
	}
	defer f.Close()
	tr := tar.NewReader(f)
	var names []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return names
		}
		names = append(names, hdr.Name)
	}
	return names
}

func (e *testEnv) ReadArchiveContents() map[string]string {
	f, err := os.Open(e.Archive)
	if err != nil {
		return nil
	}
	defer f.Close()
	tr := tar.NewReader(f)
	result := make(map[string]string)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result
		}
		if hdr.Typeflag == tar.TypeReg {
			buf, _ := io.ReadAll(tr)
			result[hdr.Name] = string(buf)
		}
	}
	return result
}

func (e *testEnv) BaseOpts(sub cli.Subcommand) *cli.Options {
	return &cli.Options{
		Subcommand:    sub,
		ArchiveNames:  []string{e.Archive},
		BlockingFactor: 20,
		RecordSize:    10240,
		ArchiveFormat: cli.FormatGNU,
	}
}

func (e *testEnv) OutDir(subdir string) string {
	out := filepath.Join(e.Dir, subdir)
	os.MkdirAll(out, 0o755)
	return out
}
