package compress

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

type zstdCodec struct{}

func (z zstdCodec) Name() string        { return "zstd" }
func (z zstdCodec) Extensions() []string { return []string{"zst", "zstd"} }

func (z zstdCodec) NewReader(r io.Reader) (io.ReadCloser, error) {
	dec, err := zstd.NewReader(r)
	if err != nil {
		return nil, err
	}
	return dec.IOReadCloser(), nil
}

func (z zstdCodec) NewWriter(w io.Writer, level int) (io.WriteCloser, error) {
	return zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(level)))
}

func init() {
	Register(zstdCodec{})
}
