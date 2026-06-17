package compress

import (
	"compress/gzip"
	"io"
)

type gzipCodec struct{}

func (g gzipCodec) Name() string                { return "gzip" }
func (g gzipCodec) Extensions() []string         { return []string{"gz", "tgz", "tar.gz"} }

func (g gzipCodec) NewReader(r io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}

func (g gzipCodec) NewWriter(w io.Writer, level int) (io.WriteCloser, error) {
	return gzip.NewWriterLevel(w, level)
}

func init() {
	Register(gzipCodec{})
}
