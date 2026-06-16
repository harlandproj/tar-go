package ops

import (
	"fmt"
	"io"
	"os"

	"github.com/harlandproj/tar-go/internal/cli"
	"github.com/harlandproj/tar-go/internal/compress"
)

func openArchiveReader(opts *cli.Options) (io.ReadCloser, error) {
	archiveName := "tar.out"
	if len(opts.ArchiveNames) > 0 {
		archiveName = opts.ArchiveNames[0]
	}
	var f io.ReadCloser
	var err error
	if archiveName == "-" {
		f = io.NopCloser(os.Stdin)
	} else {
		f, err = os.Open(archiveName)
		if err != nil {
			return nil, fmt.Errorf("cannot open %s: %w", archiveName, err)
		}
	}
	prog := opts.CompressProgram
	if prog == "" && opts.AutoCompress {
		c, ok := compress.ByExtension(archiveName)
		if ok {
			prog = c.Name()
		}
	}
	if prog != "" {
		c, ok := compress.ByName(prog)
		if ok {
			rc, err := c.NewReader(f)
			if err != nil {
				f.Close()
				return nil, err
			}
			return &multiReadCloser{rc, []io.Closer{rc, f}}, nil
		}
	}
	return f, nil
}

func openArchiveWriter(opts *cli.Options) (io.WriteCloser, error) {
	archiveName := "tar.out"
	if len(opts.ArchiveNames) > 0 {
		archiveName = opts.ArchiveNames[0]
	}
	if opts.AutoCompress && opts.CompressProgram == "" {
		c, ok := compress.ByExtension(archiveName)
		if ok {
			opts.CompressProgram = c.Name()
		}
	}
	f, err := os.Create(archiveName)
	if err != nil {
		return nil, fmt.Errorf("cannot create %s: %w", archiveName, err)
	}
	if opts.CompressProgram != "" {
		c, ok := compress.ByName(opts.CompressProgram)
		if ok {
			w, err := c.NewWriter(f, 6)
			if err != nil {
				f.Close()
				return nil, err
			}
			return &multiWriteCloser{w, []io.Closer{w, f}}, nil
		}
	}
	return f, nil
}

type multiReadCloser struct {
	io.ReadCloser
	closers []io.Closer
}

func (m *multiReadCloser) Close() error {
	var firstErr error
	for _, c := range m.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

type multiWriteCloser struct {
	io.WriteCloser
	closers []io.Closer
}

func (m *multiWriteCloser) Close() error {
	var firstErr error
	for _, c := range m.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func resolveFiles(names []string) []string {
	if len(names) == 0 {
		return []string{"."}
	}
	return names
}

func processCDirectives(fileNames []string) {
	for i := 0; i < len(fileNames); i++ {
		if fileNames[i] == "-C" && i+1 < len(fileNames) {
			os.Chdir(fileNames[i+1])
		}
	}
}
