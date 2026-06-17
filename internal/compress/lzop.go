package compress

import (
	"bytes"
	"encoding/binary"
	"hash/adler32"
	"hash/crc32"
	"io"

	"github.com/rasky/go-lzo"
)

var lzopMagic = []byte("\x89LZO\x00\r\n\x1a\n")

type lzopCodec struct{}

func (l lzopCodec) Name() string        { return "lzo" }
func (l lzopCodec) Extensions() []string { return []string{"lzo", "tlzo", "tar.lzo"} }

func (l lzopCodec) NewReader(r io.Reader) (io.ReadCloser, error) {
	return newLzopReader(r)
}

func (l lzopCodec) NewWriter(w io.Writer, level int) (io.WriteCloser, error) {
	return newLzopWriter(w, level)
}

func init() {
	Register(lzopCodec{})
}

type lzopReadCloser struct {
	reader *bytes.Reader
}

func newLzopReader(r io.Reader) (*lzopReadCloser, error) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return nil, err
	}
	data := buf.Bytes()

	pos := 0
	if len(data) < 9+2+2+2+1+1+4+4+4+4+4+1+4 {
		return nil, &LzopError{Msg: "header too short"}
	}

	if !bytes.HasPrefix(data[pos:], lzopMagic) {
		return nil, &LzopError{Msg: "invalid lzop magic"}
	}
	pos += 9

	_ = binary.BigEndian.Uint16(data[pos : pos+2]) // version
	pos += 2
	_ = binary.BigEndian.Uint16(data[pos : pos+2]) // lib version
	pos += 2
	_ = binary.BigEndian.Uint16(data[pos : pos+2]) // version needed
	pos += 2
	method := data[pos]
	pos++
	level := data[pos]
	pos++
	flags := binary.BigEndian.Uint32(data[pos : pos+4])
	pos += 4
	// filter
	pos += 4
	// mode
	pos += 4
	// mtime low
	pos += 4
	// mtime high
	pos += 4
	fnameLen := int(data[pos])
	pos++
	// skip filename
	pos += fnameLen
	// skip header checksum
	pos += 4

	if method != 1 && method != 2 && method != 3 {
		return nil, &LzopError{Msg: "unsupported compression method"}
	}

	haveCRC32 := (flags & 0x01000000) != 0
	haveAdler32 := (flags & 0x02000000) != 0

	var result bytes.Buffer

	for {
		if pos+4 > len(data) {
			break
		}
		uncompSize := binary.BigEndian.Uint32(data[pos : pos+4])
		pos += 4

		if uncompSize == 0 {
			break
		}

		if pos+4 > len(data) {
			return nil, &LzopError{Msg: "truncated block header"}
		}
		compSize := binary.BigEndian.Uint32(data[pos : pos+4])
		pos += 4

		if compSize > uint32(len(data)-pos) {
			return nil, &LzopError{Msg: "block data exceeds available data"}
		}

		blockData := data[pos : pos+int(compSize)]
		pos += int(compSize)

		if uncompSize == compSize {
			result.Write(blockData)
		} else {
			decomp, err := lzo.Decompress1X(bytes.NewReader(blockData), int(compSize), int(uncompSize))
			if err != nil {
				return nil, err
			}
			result.Write(decomp)
		}

		if haveCRC32 {
			if pos+4 > len(data) {
				return nil, &LzopError{Msg: "truncated CRC checksum"}
			}
			expectedCRC := binary.BigEndian.Uint32(data[pos : pos+4])
			pos += 4
			actualCRC := crc32.ChecksumIEEE(result.Bytes()[result.Len()-int(uncompSize):])
			if actualCRC != expectedCRC {
				return nil, &LzopError{Msg: "CRC checksum mismatch"}
			}
		} else if haveAdler32 {
			if pos+4 > len(data) {
				return nil, &LzopError{Msg: "truncated Adler checksum"}
			}
			expectedAdler := binary.BigEndian.Uint32(data[pos : pos+4])
			pos += 4
			h := adler32.New()
			h.Write(result.Bytes()[result.Len()-int(uncompSize):])
			if h.Sum32() != expectedAdler {
				return nil, &LzopError{Msg: "Adler checksum mismatch"}
			}
		}

		_ = level
	}

	return &lzopReadCloser{reader: bytes.NewReader(result.Bytes())}, nil
}

func (l *lzopReadCloser) Read(p []byte) (int, error) {
	return l.reader.Read(p)
}

func (l *lzopReadCloser) Close() error {
	return nil
}

type lzopWriteCloser struct {
	w     io.Writer
	buf   bytes.Buffer
	level int
}

func newLzopWriter(w io.Writer, level int) (*lzopWriteCloser, error) {
	if level <= 0 {
		level = 1
	}
	return &lzopWriteCloser{w: w, level: level}, nil
}

func (l *lzopWriteCloser) Write(p []byte) (int, error) {
	return l.buf.Write(p)
}

func (l *lzopWriteCloser) Close() error {
	uncompressed := l.buf.Bytes()

	var header bytes.Buffer
	header.Write(lzopMagic)
	hdr := make([]byte, 2+2+2+1+1+4+4+4+4+4+1)
	binary.BigEndian.PutUint16(hdr[0:2], 0x1030)  // version
	binary.BigEndian.PutUint16(hdr[2:4], 0x2080)  // lib version
	binary.BigEndian.PutUint16(hdr[4:6], 0x0940)  // version needed
	hdr[6] = 1                                     // method (LZO1X-1)
	hdr[7] = byte(l.level)                         // level
	binary.BigEndian.PutUint32(hdr[8:12], 0x03000000) // flags (F_ADLER32_C | F_CRC32_C)
	// filter (4 bytes, zero)
	// mode (4 bytes, zero)
	// mtime low (4 bytes, zero)
	// mtime high (4 bytes, zero)
	hdr[28] = 0 // filename length
	header.Write(hdr)

	// header checksum (Adler-32 of everything up to this point)
	h := adler32.New()
	headerBytes := header.Bytes()
	checksum := h.Sum(headerBytes[:len(headerBytes)-4])
	header.Truncate(header.Len() - 4)
	binary.BigEndian.PutUint32(hdr[28:32], 0)
	header.Write(make([]byte, 4))
	copy(header.Bytes()[header.Len()-4:], checksum[len(checksum)-4:])

	// compute proper header checksum
	header.Reset()
	header.Write(lzopMagic)
	header.Write(hdr)

	h2 := adler32.New()
	h2.Write(header.Bytes())
	cs := h2.Sum32()
	var csBytes [4]byte
	binary.BigEndian.PutUint32(csBytes[:], cs)
	header.Write(csBytes[:])

	if _, err := l.w.Write(header.Bytes()); err != nil {
		return err
	}

	// write block
	var compData []byte
	if len(uncompressed) > 0 {
		compData = lzo.Compress1X(uncompressed)
	}

	blockHdr := make([]byte, 8)
	binary.BigEndian.PutUint32(blockHdr[0:4], uint32(len(uncompressed)))
	if len(compData) < len(uncompressed) {
		binary.BigEndian.PutUint32(blockHdr[4:8], uint32(len(compData)))
	} else {
		compData = uncompressed
		binary.BigEndian.PutUint32(blockHdr[4:8], uint32(len(uncompressed)))
	}

	if _, err := l.w.Write(blockHdr); err != nil {
		return err
	}
	if _, err := l.w.Write(compData); err != nil {
		return err
	}

	// block checksum (CRC32)
	ch := crc32.NewIEEE()
	ch.Write(uncompressed)
	blockCS := make([]byte, 4)
	binary.BigEndian.PutUint32(blockCS, ch.Sum32())
	if _, err := l.w.Write(blockCS); err != nil {
		return err
	}

	// terminator
	term := make([]byte, 4)
	if _, err := l.w.Write(term); err != nil {
		return err
	}

	return nil
}

type LzopError struct {
	Msg string
}

func (e *LzopError) Error() string {
	return "lzop: " + e.Msg
}
