package compress

import (
	"io"

	"github.com/ulikunitz/xz/lzma"
)

type lzmaCodec struct{}

func (l lzmaCodec) Name() string        { return "lzma" }
func (l lzmaCodec) Extensions() []string { return []string{"lzma"} }

func (l lzmaCodec) NewReader(r io.Reader) (io.ReadCloser, error) {
	reader, err := lzma.NewReader(r)
	if err != nil {
		return nil, err
	}
	return nopReadCloser{reader}, nil
}

func (l lzmaCodec) NewWriter(w io.Writer, level int) (io.WriteCloser, error) {
	return lzma.NewWriter(w)
}

func init() {
	Register(lzmaCodec{})
}
