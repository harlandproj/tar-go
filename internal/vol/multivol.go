package vol

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type MultiVolWriter struct {
	prefix     string
	maxSize    int64
	current    *os.File
	currentNum int
	written    int64
}

func NewMultiVolWriter(prefix string, maxSize int64) (*MultiVolWriter, error) {
	if maxSize <= 0 {
		return nil, fmt.Errorf("maxSize must be positive")
	}
	mv := &MultiVolWriter{
		prefix:     prefix,
		maxSize:    maxSize,
		currentNum: 1,
	}
	if err := mv.openNext(); err != nil {
		return nil, err
	}
	return mv, nil
}

func (mv *MultiVolWriter) volName() string {
	if mv.currentNum == 1 {
		return mv.prefix
	}
	return fmt.Sprintf("%s-%d", mv.prefix, mv.currentNum)
}

func (mv *MultiVolWriter) openNext() error {
	if mv.current != nil {
		mv.current.Close()
	}
	var err error
	mv.current, err = os.Create(mv.volName())
	if err != nil {
		return err
	}
	mv.written = 0
	return nil
}

func (mv *MultiVolWriter) Write(p []byte) (int, error) {
	total := 0
	for len(p) > 0 {
		if mv.current == nil {
			return total, fmt.Errorf("writer is closed")
		}
		space := mv.maxSize - mv.written
		if space <= 0 {
			mv.currentNum++
			if err := mv.openNext(); err != nil {
				return total, err
			}
			space = mv.maxSize
		}
		toWrite := int64(len(p))
		if toWrite > space {
			toWrite = space
		}
		n, err := mv.current.Write(p[:toWrite])
		total += n
		mv.written += int64(n)
		if err != nil {
			return total, err
		}
		p = p[toWrite:]
	}
	return total, nil
}

func (mv *MultiVolWriter) Close() error {
	if mv.current == nil {
		return nil
	}
	err := mv.current.Close()
	mv.current = nil
	return err
}

type MultiVolReader struct {
	prefix     string
	current    *os.File
	currentNum int
}

func NewMultiVolReader(prefix string) (*MultiVolReader, error) {
	mr := &MultiVolReader{
		prefix:     prefix,
		currentNum: 1,
	}
	if err := mr.openCurrent(); err != nil {
		return nil, err
	}
	return mr, nil
}

func (mr *MultiVolReader) volName() string {
	if mr.currentNum == 1 {
		return mr.prefix
	}
	return fmt.Sprintf("%s-%d", mr.prefix, mr.currentNum)
}

func (mr *MultiVolReader) openCurrent() error {
	if mr.current != nil {
		mr.current.Close()
	}
	var err error
	mr.current, err = os.Open(mr.volName())
	return err
}

func (mr *MultiVolReader) Read(p []byte) (int, error) {
	for {
		if mr.current == nil {
			return 0, io.EOF
		}
		n, err := mr.current.Read(p)
		if err == io.EOF {
			mr.current.Close()
			mr.current = nil
			mr.currentNum++
			if err := mr.openCurrent(); err != nil {
				return n, io.EOF
			}
			if n > 0 {
				return n, nil
			}
			continue
		}
		if err != nil {
			return n, err
		}
		return n, nil
	}
}

func (mr *MultiVolReader) Close() error {
	if mr.current == nil {
		return nil
	}
	err := mr.current.Close()
	mr.current = nil
	return err
}

func parseVolName(name string) (string, int) {
	lastDash := strings.LastIndex(name, "-")
	if lastDash < 0 {
		return name, 1
	}
	suffix := name[lastDash+1:]
	if n, err := strconv.Atoi(suffix); err == nil {
		return name[:lastDash], n
	}
	return name, 1
}
