package compress

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/adler32"
	"io"
	"strings"
	"testing"
)

func TestByName_Existing(t *testing.T) {
	codecs := []string{"gzip", "bzip2", "xz", "zstd", "lzma", "lzip", "lzo"}
	for _, name := range codecs {
		c, ok := ByName(name)
		if !ok {
			t.Errorf("ByName(%q) not found", name)
		}
		if c.Name() != name {
			t.Errorf("ByName(%q).Name() = %q, want %q", name, c.Name(), name)
		}
	}
}

func TestByName_NonExisting(t *testing.T) {
	_, ok := ByName("nonexistent")
	if ok {
		t.Error("ByName(\"nonexistent\") should return false")
	}
}

func TestByName_ByExtension(t *testing.T) {
	c, ok := ByName("gz")
	if !ok || c.Name() != "gzip" {
		t.Errorf("ByName(\"gz\") = %v, %v; want gzip, true", c, ok)
	}
	c, ok = ByName("bz2")
	if !ok || c.Name() != "bzip2" {
		t.Errorf("ByName(\"bz2\") = %v, %v; want bzip2, true", c, ok)
	}
}

func TestByExtension_TarGz(t *testing.T) {
	c, ok := ByExtension("archive.tar.gz")
	if !ok || c.Name() != "gzip" {
		t.Errorf("ByExtension(\"archive.tar.gz\") = %v, %v", c, ok)
	}
}

func TestByExtension_Tgz(t *testing.T) {
	c, ok := ByExtension("archive.tgz")
	if !ok || c.Name() != "gzip" {
		t.Errorf("ByExtension(\"archive.tgz\") = %v, %v", c, ok)
	}
}

func TestByExtension_TarBz2(t *testing.T) {
	c, ok := ByExtension("archive.tar.bz2")
	if !ok || c.Name() != "bzip2" {
		t.Errorf("ByExtension(\"archive.tar.bz2\") = %v, %v", c, ok)
	}
}

func TestByExtension_Tbz2(t *testing.T) {
	c, ok := ByExtension("archive.tbz2")
	if !ok || c.Name() != "bzip2" {
		t.Errorf("ByExtension(\"archive.tbz2\") = %v, %v", c, ok)
	}
}

func TestByExtension_Tbz(t *testing.T) {
	c, ok := ByExtension("archive.tbz")
	if !ok || c.Name() != "bzip2" {
		t.Errorf("ByExtension(\"archive.tbz\") = %v, %v", c, ok)
	}
}

func TestByExtension_TarXz(t *testing.T) {
	c, ok := ByExtension("archive.tar.xz")
	if !ok || c.Name() != "xz" {
		t.Errorf("ByExtension(\"archive.tar.xz\") = %v, %v", c, ok)
	}
}

func TestByExtension_Txz(t *testing.T) {
	c, ok := ByExtension("archive.txz")
	if !ok || c.Name() != "xz" {
		t.Errorf("ByExtension(\"archive.txz\") = %v, %v", c, ok)
	}
}

func TestByExtension_TarZst(t *testing.T) {
	c, ok := ByExtension("archive.tar.zst")
	if !ok || c.Name() != "zstd" {
		t.Errorf("ByExtension(\"archive.tar.zst\") = %v, %v", c, ok)
	}
}

func TestByExtension_Tzst(t *testing.T) {
	c, ok := ByExtension("archive.tzst")
	if !ok || c.Name() != "zstd" {
		t.Errorf("ByExtension(\"archive.tzst\") = %v, %v", c, ok)
	}
}

func TestByExtension_TarLz(t *testing.T) {
	c, ok := ByExtension("archive.tar.lz")
	if !ok || c.Name() != "lzip" {
		t.Errorf("ByExtension(\"archive.tar.lz\") = %v, %v", c, ok)
	}
}

func TestByExtension_Tlz(t *testing.T) {
	c, ok := ByExtension("archive.tlz")
	if !ok || c.Name() != "lzip" {
		t.Errorf("ByExtension(\"archive.tlz\") = %v, %v", c, ok)
	}
}

func TestByExtension_TarLzma(t *testing.T) {
	c, ok := ByExtension("archive.tar.lzma")
	if !ok || c.Name() != "lzma" {
		t.Errorf("ByExtension(\"archive.tar.lzma\") = %v, %v", c, ok)
	}
}

func TestByExtension_Tlzma(t *testing.T) {
	c, ok := ByExtension("archive.tlzma")
	if !ok || c.Name() != "lzma" {
		t.Errorf("ByExtension(\"archive.tlzma\") = %v, %v", c, ok)
	}
}

func TestByExtension_TarLzo(t *testing.T) {
	c, ok := ByExtension("archive.tar.lzo")
	if !ok || c.Name() != "lzo" {
		t.Errorf("ByExtension(\"archive.tar.lzo\") = %v, %v", c, ok)
	}
}

func TestByExtension_Tlzo(t *testing.T) {
	c, ok := ByExtension("archive.tlzo")
	if !ok || c.Name() != "lzo" {
		t.Errorf("ByExtension(\"archive.tlzo\") = %v, %v", c, ok)
	}
}

func TestByExtension_Unknown(t *testing.T) {
	_, ok := ByExtension("archive.zip")
	if ok {
		t.Error("ByExtension(\"archive.zip\") should return false")
	}
}

func TestByExtension_CaseInsensitive(t *testing.T) {
	c, ok := ByExtension("ARCHIVE.TAR.GZ")
	if !ok || c.Name() != "gzip" {
		t.Errorf("ByExtension(\"ARCHIVE.TAR.GZ\") = %v, %v", c, ok)
	}
}

func TestStripSuffix_Known(t *testing.T) {
	name, c := StripSuffix("archive.tar.gz")
	if name != "archive" || c == nil || c.Name() != "gzip" {
		t.Errorf("StripSuffix(\"archive.tar.gz\") = %q, %v", name, c)
	}
	name, c = StripSuffix("archive.tgz")
	if name != "archive" || c == nil || c.Name() != "gzip" {
		t.Errorf("StripSuffix(\"archive.tgz\") = %q, %v", name, c)
	}
	name, c = StripSuffix("archive.tar.bz2")
	if name != "archive" || c == nil || c.Name() != "bzip2" {
		t.Errorf("StripSuffix(\"archive.tar.bz2\") = %q, %v", name, c)
	}
	name, c = StripSuffix("archive.tar.xz")
	if name != "archive" || c == nil || c.Name() != "xz" {
		t.Errorf("StripSuffix(\"archive.tar.xz\") = %q, %v", name, c)
	}
	name, c = StripSuffix("archive.tar.zst")
	if name != "archive" || c == nil || c.Name() != "zstd" {
		t.Errorf("StripSuffix(\"archive.tar.zst\") = %q, %v", name, c)
	}
	name, c = StripSuffix("archive.tar.lz")
	if name != "archive" || c == nil || c.Name() != "lzip" {
		t.Errorf("StripSuffix(\"archive.tar.lz\") = %q, %v", name, c)
	}
	name, c = StripSuffix("archive.tar.lzma")
	if name != "archive" || c == nil || c.Name() != "lzma" {
		t.Errorf("StripSuffix(\"archive.tar.lzma\") = %q, %v", name, c)
	}
	name, c = StripSuffix("archive.tar.lzo")
	if name != "archive" || c == nil || c.Name() != "lzo" {
		t.Errorf("StripSuffix(\"archive.tar.lzo\") = %q, %v", name, c)
	}
}

func TestStripSuffix_Unknown(t *testing.T) {
	name, c := StripSuffix("archive.zip")
	if name != "archive.zip" || c != nil {
		t.Errorf("StripSuffix(\"archive.zip\") = %q, %v", name, c)
	}
}

func TestStripAnySuffix_Known(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"archive.tar.gz", "archive"},
		{"archive.tgz", "archive"},
		{"archive.tar.bz2", "archive"},
		{"archive.tar.xz", "archive"},
		{"archive.tar.zst", "archive"},
		{"archive.tar.lz", "archive"},
		{"archive.tar.lzma", "archive"},
		{"archive.tar.lzo", "archive"},
	}
	for _, tt := range tests {
		got := StripAnySuffix(tt.input)
		if got != tt.want {
			t.Errorf("StripAnySuffix(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestStripAnySuffix_CaseInsensitive(t *testing.T) {
	got := StripAnySuffix("ARCHIVE.TAR.GZ")
	if got != "ARCHIVE" {
		t.Errorf("StripAnySuffix(\"ARCHIVE.TAR.GZ\") = %q, want \"ARCHIVE\"", got)
	}
}

func TestStripAnySuffix_Unknown(t *testing.T) {
	got := StripAnySuffix("archive.zip")
	if got != "archive.zip" {
		t.Errorf("StripAnySuffix(\"archive.zip\") = %q, want \"archive.zip\"", got)
	}
}

func TestAutoSelectCodec_KnownExtension(t *testing.T) {
	c, err := AutoSelectCodec("archive.tar.gz", "")
	if err != nil || c.Name() != "gzip" {
		t.Errorf("AutoSelectCodec(\"archive.tar.gz\", \"\") = %v, %v", c, err)
	}
}

func TestAutoSelectCodec_DefaultCodec(t *testing.T) {
	c, err := AutoSelectCodec("archive.unknown", "gzip")
	if err != nil || c.Name() != "gzip" {
		t.Errorf("AutoSelectCodec(\"archive.unknown\", \"gzip\") = %v, %v", c, err)
	}
}

func TestAutoSelectCodec_UnknownBoth(t *testing.T) {
	_, err := AutoSelectCodec("archive.unknown", "")
	if err == nil {
		t.Error("AutoSelectCodec with unknown extension and no default should error")
	}
}

func TestAutoSelectCodec_UnknownExtensionBadDefault(t *testing.T) {
	_, err := AutoSelectCodec("archive.unknown", "nonexistent")
	if err == nil {
		t.Error("AutoSelectCodec with unknown extension and bad default should error")
	}
}

func TestDetect_Known(t *testing.T) {
	if !Detect("archive.tar.gz") {
		t.Error("Detect(\"archive.tar.gz\") should be true")
	}
	if !Detect("archive.tgz") {
		t.Error("Detect(\"archive.tgz\") should be true")
	}
}

func TestDetect_Unknown(t *testing.T) {
	if Detect("archive.zip") {
		t.Error("Detect(\"archive.zip\") should be false")
	}
}

func TestNopReadCloser_Close(t *testing.T) {
	n := nopReadCloser{Reader: strings.NewReader("data")}
	if err := n.Close(); err != nil {
		t.Errorf("nopReadCloser.Close() = %v, want nil", err)
	}
}

func TestNopReadCloser_Read(t *testing.T) {
	n := nopReadCloser{Reader: strings.NewReader("hello")}
	buf := make([]byte, 5)
	nr, err := n.Read(buf)
	if nr != 5 || err != nil || string(buf) != "hello" {
		t.Errorf("nopReadCloser.Read() = %d, %v, %q", nr, err, string(buf))
	}
}

func roundTrip(t *testing.T, name string, data []byte) {
	t.Helper()
	c, ok := ByName(name)
	if !ok {
		t.Fatalf("codec %q not found", name)
	}

	var buf bytes.Buffer
	w, err := c.NewWriter(&buf, 6)
	if err != nil {
		t.Fatalf("%s NewWriter: %v", name, err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatalf("%s Write: %v", name, err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("%s Close: %v", name, err)
	}

	r, err := c.NewReader(&buf)
	if err != nil {
		t.Fatalf("%s NewReader: %v", name, err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("%s ReadAll: %v", name, err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("%s Close reader: %v", name, err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("%s round-trip mismatch: got %d bytes, want %d", name, len(got), len(data))
	}
}

func TestGzip_RoundTrip(t *testing.T) {
	roundTrip(t, "gzip", []byte("hello gzip compression test data"))
}

func TestGzip_RoundTrip_Empty(t *testing.T) {
	roundTrip(t, "gzip", []byte{})
}

func TestGzip_RoundTrip_Large(t *testing.T) {
	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	roundTrip(t, "gzip", data)
}

func TestBzip2_RoundTrip(t *testing.T) {
	roundTrip(t, "bzip2", []byte("hello bzip2 compression test data"))
}

func TestBzip2_RoundTrip_Empty(t *testing.T) {
	roundTrip(t, "bzip2", []byte{})
}

func TestBzip2_RoundTrip_Large(t *testing.T) {
	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	roundTrip(t, "bzip2", data)
}

func TestXz_RoundTrip(t *testing.T) {
	roundTrip(t, "xz", []byte("hello xz compression test data"))
}

func TestXz_RoundTrip_Empty(t *testing.T) {
	roundTrip(t, "xz", []byte{})
}

func TestXz_RoundTrip_Large(t *testing.T) {
	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	roundTrip(t, "xz", data)
}

func TestZstd_RoundTrip(t *testing.T) {
	roundTrip(t, "zstd", []byte("hello zstd compression test data"))
}

func TestZstd_RoundTrip_Empty(t *testing.T) {
	roundTrip(t, "zstd", []byte{})
}

func TestZstd_RoundTrip_Large(t *testing.T) {
	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	roundTrip(t, "zstd", data)
}

func TestLzma_RoundTrip(t *testing.T) {
	roundTrip(t, "lzma", []byte("hello lzma compression test data"))
}

func TestLzma_RoundTrip_Empty(t *testing.T) {
	roundTrip(t, "lzma", []byte{})
}

func TestLzma_RoundTrip_Large(t *testing.T) {
	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	roundTrip(t, "lzma", data)
}

func TestLzip_RoundTrip(t *testing.T) {
	roundTrip(t, "lzip", []byte("hello lzip compression test data"))
}

func TestLzip_RoundTrip_Empty(t *testing.T) {
	roundTrip(t, "lzip", []byte{})
}

func TestLzip_RoundTrip_Large(t *testing.T) {
	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	roundTrip(t, "lzip", data)
}

func TestLzip_MultiMember(t *testing.T) {
	part1 := []byte("first member data")
	part2 := []byte("second member data")

	var buf bytes.Buffer
	for i, part := range [][]byte{part1, part2} {
		w, err := newLzipWriter(&buf)
		if err != nil {
			t.Fatalf("member %d newLzipWriter: %v", i, err)
		}
		if _, err := w.Write(part); err != nil {
			t.Fatalf("member %d Write: %v", i, err)
		}
		if err := w.Close(); err != nil {
			t.Fatalf("member %d Close: %v", i, err)
		}
	}

	r, err := newLzipReader(&buf)
	if err != nil {
		t.Fatalf("newLzipReader: %v", err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	r.Close()
	expected := append(part1, part2...)
	if !bytes.Equal(got, expected) {
		t.Errorf("multi-member lzip mismatch: got %d bytes, want %d", len(got), len(expected))
	}
}

func TestLzipError_Error(t *testing.T) {
	e := &LzipError{Msg: "test error"}
	if e.Error() != "lzip: test error" {
		t.Errorf("LzipError.Error() = %q, want %q", e.Error(), "lzip: test error")
	}
}

func TestLzipReader_InvalidMagic(t *testing.T) {
	_, err := newLzipReader(bytes.NewReader([]byte("NOTLZIP")))
	if err == nil {
		t.Error("expected error for invalid lzip magic")
	}
	var le *LzipError
	if !errors.As(err, &le) {
		t.Errorf("expected LzipError, got %T", err)
	}
}

func TestLzipReader_TruncatedHeader(t *testing.T) {
	_, err := newLzipReader(bytes.NewReader([]byte("LZI")))
	if err == nil {
		t.Error("expected error for truncated lzip header")
	}
}

func TestLzop_RoundTrip(t *testing.T) {
	roundTrip(t, "lzo", []byte("hello lzop compression test data"))
}

func TestLzop_RoundTrip_Empty(t *testing.T) {
	roundTrip(t, "lzo", []byte{})
}

func TestLzop_RoundTrip_Large(t *testing.T) {
	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	roundTrip(t, "lzo", data)
}

func TestLzopError_Error(t *testing.T) {
	e := &LzopError{Msg: "test error"}
	if e.Error() != "lzop: test error" {
		t.Errorf("LzopError.Error() = %q, want %q", e.Error(), "lzop: test error")
	}
}

func TestLzopReader_InvalidMagic(t *testing.T) {
	_, err := newLzopReader(bytes.NewReader([]byte("invalid lzop data")))
	if err == nil {
		t.Error("expected error for invalid lzop magic")
	}
	var le *LzopError
	if !errors.As(err, &le) {
		t.Errorf("expected LzopError, got %T", err)
	}
}

func TestLzopReader_TruncatedHeader(t *testing.T) {
	_, err := newLzopReader(bytes.NewReader([]byte{0x89, 'L', 'Z', 'O'}))
	if err == nil {
		t.Error("expected error for truncated lzop header")
	}
}

func TestCodecExtensions(t *testing.T) {
	tests := map[string][]string{
		"gzip":  {"gz", "tgz", "tar.gz"},
		"bzip2": {"bz2", "tbz2", "tbz", "tar.bz2"},
		"xz":    {"xz", "txz", "tar.xz"},
		"zstd":  {"zst", "zstd", "tzst", "tar.zst"},
		"lzma":  {"lzma", "tlzma", "tar.lzma"},
		"lzip":  {"lz", "tlz", "tar.lz"},
		"lzo":   {"lzo", "tlzo", "tar.lzo"},
	}
	for name, wantExts := range tests {
		c, ok := ByName(name)
		if !ok {
			t.Errorf("ByName(%q) not found", name)
			continue
		}
		exts := c.Extensions()
		if len(exts) != len(wantExts) {
			t.Errorf("%s.Extensions() = %v, want %v", name, exts, wantExts)
			continue
		}
		for i, ext := range exts {
			if ext != wantExts[i] {
				t.Errorf("%s.Extensions()[%d] = %q, want %q", name, i, ext, wantExts[i])
			}
		}
	}
}

func TestRegister_Duplicate(t *testing.T) {
	Register(gzipCodec{})
	c, ok := ByName("gzip")
	if !ok || c.Name() != "gzip" {
		t.Error("re-registering gzip should still work")
	}
}

func TestLzipReader_ReadAfterClose(t *testing.T) {
	data := []byte("test data")
	var buf bytes.Buffer
	w, _ := newLzipWriter(&buf)
	w.Write(data)
	w.Close()

	r, err := newLzipReader(&buf)
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
	_, err = r.Read(make([]byte, 10))
	if err == nil {
		t.Error("expected error reading after close/done")
	}
}

func TestLzopWriter_Level(t *testing.T) {
	var buf bytes.Buffer
	w, err := newLzopWriter(&buf, 0)
	if err != nil {
		t.Fatal(err)
	}
	if w.level != 1 {
		t.Errorf("lzopWriter.level = %d, want 1 (default)", w.level)
	}
}

func TestLzopReader_UnsupportedMethod(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(lzopMagic)
	hdr := make([]byte, 2+2+2+1+1+4+4+4+4+4+1+4)
	binary.BigEndian.PutUint16(hdr[0:2], 0x1030)
	binary.BigEndian.PutUint16(hdr[2:4], 0x2080)
	binary.BigEndian.PutUint16(hdr[4:6], 0x0940)
	hdr[6] = 99
	hdr[7] = 1
	binary.BigEndian.PutUint32(hdr[8:12], 0x03000000)
	hdr[28] = 0
	buf.Write(hdr)
	h := adler32.New()
	h.Write(buf.Bytes())
	cs := make([]byte, 4)
	binary.BigEndian.PutUint32(cs, h.Sum32())
	buf.Write(cs)

	_, err := newLzopReader(&buf)
	if err == nil {
		t.Error("expected error for unsupported method")
	}
}

func TestLzipReader_BadVersion(t *testing.T) {
	header := []byte("LZIP")
	header = append(header, 99)
	header = append(header, 7)
	_, err := newLzipReader(bytes.NewReader(header))
	if err == nil {
		t.Error("expected error for bad lzip version")
	}
	var le *LzipError
	if !errors.As(err, &le) {
		t.Errorf("expected LzipError, got %T", err)
	}
}

func TestLzipReader_TruncatedTrailer(t *testing.T) {
	var buf bytes.Buffer
	w, _ := newLzipWriter(&buf)
	w.Write([]byte("data"))
	w.Close()

	compressed := buf.Bytes()
	truncated := compressed[:len(compressed)-10]
	_, err := newLzipReader(bytes.NewReader(truncated))
	if err != nil {
		var le *LzipError
		if !errors.As(err, &le) {
			t.Errorf("expected LzipError for truncated trailer, got %T: %v", err, err)
		}
	}
}

func TestLzipReader_CorruptTrailerCRC(t *testing.T) {
	var buf bytes.Buffer
	w, _ := newLzipWriter(&buf)
	w.Write([]byte("data"))
	w.Close()

	compressed := buf.Bytes()
	compressed[len(compressed)-20] ^= 0xFF
	r, err := newLzipReader(bytes.NewReader(compressed))
	if err != nil {
		t.Skip("reader creation failed")
	}
	_, err = io.ReadAll(r)
	if err == nil {
		t.Error("expected CRC mismatch error")
	}
}

func TestLzopReader_EmptyStream(t *testing.T) {
	_, err := newLzopReader(bytes.NewReader(nil))
	if err == nil {
		t.Error("expected error for empty lzop stream")
	}
}
