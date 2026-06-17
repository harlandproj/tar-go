package compress

import (
	"bytes"
	"encoding/binary"
	"hash"
	"hash/crc32"
	"io"

	"github.com/ulikunitz/xz/lzma"
)

const (
	lzipMagic   = "LZIP"
	lzipVersion = 1
)

type lzipCodec struct{}

func (l lzipCodec) Name() string        { return "lzip" }
func (l lzipCodec) Extensions() []string { return []string{"lz", "tlz", "tar.lz"} }

func (l lzipCodec) NewReader(r io.Reader) (io.ReadCloser, error) {
	return newLzipReader(r)
}

func (l lzipCodec) NewWriter(w io.Writer, level int) (io.WriteCloser, error) {
	return newLzipWriter(w)
}

func init() {
	Register(lzipCodec{})
}

type lzipReadCloser struct {
	raw      io.Reader
	cur      io.Reader
	crc      hash.Hash32
	dataCnt  uint64
	buf      bytes.Buffer
	done     bool
}

func newLzipReader(r io.Reader) (*lzipReadCloser, error) {
	lr := &lzipReadCloser{raw: r}
	if err := lr.readHeader(); err != nil {
		return nil, err
	}
	return lr, nil
}

func (lr *lzipReadCloser) readHeader() error {
	lr.crc = crc32.NewIEEE()
	lr.dataCnt = 0
	lr.cur = nil

	header := make([]byte, 6)
	if _, err := io.ReadFull(lr.raw, header); err != nil {
		if err == io.EOF {
			lr.done = true
			return nil
		}
		return err
	}

	if string(header[:4]) != lzipMagic {
		return &LzipError{Msg: "invalid lzip magic"}
	}
	if header[4] != lzipVersion {
		return &LzipError{Msg: "unsupported lzip version"}
	}
	_ = header[5] // dictionary size byte, used by lzma.Reader

	r, err := lzma.NewReader(lr.raw)
	if err != nil {
		return err
	}
	lr.cur = r
	return nil
}

func (lr *lzipReadCloser) readTrailer() error {
	trailer := make([]byte, 20)
	if _, err := io.ReadFull(lr.raw, trailer); err != nil {
		return &LzipError{Msg: "failed to read trailer"}
	}

	expectedCRC := binary.BigEndian.Uint32(trailer[0:4])
	dataSize := binary.BigEndian.Uint64(trailer[4:12])
	_ = binary.BigEndian.Uint64(trailer[12:20]) // member size

	if lr.crc.Sum32() != expectedCRC {
		return &LzipError{Msg: "CRC mismatch"}
	}
	if lr.dataCnt != dataSize {
		return &LzipError{Msg: "data size mismatch"}
	}
	return nil
}

func (lr *lzipReadCloser) Read(p []byte) (int, error) {
	if lr.done {
		return 0, io.EOF
	}

	if lr.cur == nil {
		if err := lr.readHeader(); err != nil {
			return 0, err
		}
		if lr.done {
			return 0, io.EOF
		}
	}

	n, err := lr.cur.Read(p)
	if n > 0 {
		lr.crc.Write(p[:n])
		lr.dataCnt += uint64(n)
	}

	if err == io.EOF {
		lr.cur = nil
		if trailerErr := lr.readTrailer(); trailerErr != nil {
			return n, trailerErr
		}
		peekHeader := make([]byte, 4)
		// Attempt to peek: if the next bytes start with LZIP magic, it's a multi-member archive
		if _, peekErr := io.ReadFull(lr.raw, peekHeader); peekErr != nil {
			lr.done = true
			return n, io.EOF
		}
		if string(peekHeader) == lzipMagic {
			// Restore the peeked bytes
			lr.buf.Reset()
			lr.buf.Write(peekHeader)
			lr.raw = io.MultiReader(&lr.buf, lr.raw)
			if headerErr := lr.readHeader(); headerErr != nil {
				return n, headerErr
			}
			return n, nil
		}
		lr.done = true
		lr.buf.Reset()
		lr.buf.Write(peekHeader)
		lr.raw = io.MultiReader(&lr.buf, lr.raw)
		return n, io.EOF
	}

	return n, err
}

func (lr *lzipReadCloser) Close() error {
	return nil
}

type lzipWriteCloser struct {
	w      io.Writer
	crc    hash.Hash32
	dCount uint64
	buf    bytes.Buffer
	lzmaW  *lzma.Writer
}

func newLzipWriter(w io.Writer) (*lzipWriteCloser, error) {
	lw := &lzipWriteCloser{
		w:   w,
		crc: crc32.NewIEEE(),
	}

	// Write header
	header := []byte(lzipMagic)
	header = append(header, lzipVersion)
	header = append(header, 7) // 1<<(9+7) = 64KiB dictionary size

	if _, err := w.Write(header); err != nil {
		return nil, err
	}

	lzmaWriter, err := lzma.NewWriter(&lw.buf)
	if err != nil {
		return nil, err
	}
	lw.lzmaW = lzmaWriter

	return lw, nil
}

func (lw *lzipWriteCloser) Write(p []byte) (int, error) {
	lw.crc.Write(p)
	lw.dCount += uint64(len(p))
	return lw.lzmaW.Write(p)
}

func (lw *lzipWriteCloser) Close() error {
	if err := lw.lzmaW.Close(); err != nil {
		return err
	}

	lzmaData := lw.buf.Bytes()
	if _, err := lw.w.Write(lzmaData); err != nil {
		return err
	}

	trailer := make([]byte, 20)
	binary.BigEndian.PutUint32(trailer[0:4], lw.crc.Sum32())
	binary.BigEndian.PutUint64(trailer[4:12], lw.dCount)
	memberSize := uint64(6 + len(lzmaData) + 20)
	binary.BigEndian.PutUint64(trailer[12:20], memberSize)

	_, err := lw.w.Write(trailer)
	return err
}

type LzipError struct {
	Msg string
}

func (e *LzipError) Error() string {
	return "lzip: " + e.Msg
}
