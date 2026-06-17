package compress

import (
	"io"

	"github.com/ulikunitz/xz"
)

type xzCodec struct{}

func (x xzCodec) Name() string        { return "xz" }
func (x xzCodec) Extensions() []string { return []string{"xz", "txz", "tar.xz"} }

func (x xzCodec) NewReader(r io.Reader) (io.ReadCloser, error) {
	reader, err := xz.NewReader(r)
	if err != nil {
		return nil, err
	}
	return nopReadCloser{reader}, nil
}

func (x xzCodec) NewWriter(w io.Writer, level int) (io.WriteCloser, error) {
	return xz.NewWriter(w)
}

func init() {
	Register(xzCodec{})
}
