package compress

import (
	"compress/bzip2"
	"io"

	dsnetbzip2 "github.com/dsnet/compress/bzip2"
)

type bzip2Codec struct{}

func (b bzip2Codec) Name() string        { return "bzip2" }
func (b bzip2Codec) Extensions() []string { return []string{"bz2"} }

func (b bzip2Codec) NewReader(r io.Reader) (io.ReadCloser, error) {
	return nopReadCloser{bzip2.NewReader(r)}, nil
}

func (b bzip2Codec) NewWriter(w io.Writer, level int) (io.WriteCloser, error) {
	cfg := &dsnetbzip2.WriterConfig{}
	if level > 0 {
		cfg.Level = level
	}
	return dsnetbzip2.NewWriter(w, cfg)
}

func init() {
	Register(bzip2Codec{})
}
