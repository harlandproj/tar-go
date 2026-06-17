package ops

import (
	"archive/tar"
	"io"
	"os"
)

type SparseHole struct {
	Offset int64
	Length int64
}

func detectHoles(f *os.File, size int64) []SparseHole {
	if size <= 0 {
		return nil
	}
	var holes []SparseHole
	buf := make([]byte, 4096)
	var pos int64
	var inHole bool
	var holeStart int64

	for pos < size {
		toRead := int64(len(buf))
		if pos+toRead > size {
			toRead = size - pos
		}
		n, err := f.ReadAt(buf[:toRead], pos)
		if err != nil && err != io.EOF {
			break
		}
		if n == 0 {
			break
		}
		isZero := true
		for i := 0; i < n; i++ {
			if buf[i] != 0 {
				isZero = false
				break
			}
		}
		if isZero {
			if !inHole {
				inHole = true
				holeStart = pos
			}
		} else {
			if inHole {
				holes = append(holes, SparseHole{Offset: holeStart, Length: pos - holeStart})
				inHole = false
			}
		}
		pos += int64(n)
	}
	if inHole {
		holes = append(holes, SparseHole{Offset: holeStart, Length: size - holeStart})
	}
	return holes
}

func writeSparseFile(tw *tar.Writer, f *os.File, info os.FileInfo, name string) error {
	holes := detectHoles(f, info.Size())
	if len(holes) == 0 {
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = name
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		f.Seek(0, io.SeekStart)
		_, err = io.Copy(tw, f)
		return err
	}

	hdr := &tar.Header{
		Name:     name,
		Size:     info.Size(),
		Mode:     int64(info.Mode()),
		ModTime:  info.ModTime(),
		Typeflag: tar.TypeGNUSparse,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	var pos int64
	for _, hole := range holes {
		if hole.Offset > pos {
			if _, err := io.CopyN(tw, io.NewSectionReader(f, pos, hole.Offset-pos), hole.Offset-pos); err != nil {
				return err
			}
		}
		pos = hole.Offset + hole.Length
	}
	if pos < info.Size() {
		if _, err := io.CopyN(tw, io.NewSectionReader(f, pos, info.Size()-pos), info.Size()-pos); err != nil {
			return err
		}
	}
	return nil
}

func isSparseType(hdr *tar.Header) bool {
	return hdr.Typeflag == tar.TypeGNUSparse
}
