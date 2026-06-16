package compress

import (
	"fmt"
	"io"
	"strings"
)

type Codec interface {
	Name() string
	Extensions() []string
	NewReader(r io.Reader) (io.ReadCloser, error)
	NewWriter(w io.Writer, level int) (io.WriteCloser, error)
}

var registry = map[string]Codec{}

func Register(c Codec) {
	registry[c.Name()] = c
	for _, ext := range c.Extensions() {
		registry[ext] = c
	}
}

func ByName(name string) (Codec, bool) {
	c, ok := registry[name]
	return c, ok
}

func ByExtension(filename string) (Codec, bool) {
	lower := strings.ToLower(filename)
	for _, ext := range extensionsByPriority {
		suffix := "." + ext
		if strings.HasSuffix(lower, suffix) || strings.HasSuffix(lower, suffix[:len(suffix)-1]) {
			return registry[ext], true
		}
	}
	return nil, false
}

var extensionsByPriority = []string{
	"tar.gz", "tgz",
	"tar.bz2", "tbz2", "tbz",
	"tar.xz", "txz",
	"tar.zst", "tzst",
	"tar.lz", "tlz",
	"tar.lzma", "tlzma",
	"tar.lzo", "tlzo",
}

func StripSuffix(filename string) (string, Codec) {
	for _, ext := range extensionsByPriority {
		suffix := "." + ext
		if strings.HasSuffix(filename, suffix) {
			c, _ := registry[ext]
			return filename[:len(filename)-len(suffix)], c
		}
	}
	return filename, nil
}

func StripAnySuffix(filename string) string {
	for _, ext := range extensionsByPriority {
		suffix := "." + ext
		if strings.HasSuffix(strings.ToLower(filename), suffix) {
			return filename[:len(filename)-len(suffix)]
		}
	}
	return filename
}

func AutoSelectCodec(filename string, defaultCodec string) (Codec, error) {
	c, ok := ByExtension(filename)
	if ok {
		return c, nil
	}
	if defaultCodec != "" {
		c, ok := ByName(defaultCodec)
		if ok {
			return c, nil
		}
	}
	return nil, fmt.Errorf("could not determine compression for %s", filename)
}

func Detect(filename string) bool {
	_, ok := ByExtension(filename)
	return ok
}

type nopReadCloser struct {
	io.Reader
}

func (n nopReadCloser) Close() error { return nil }
