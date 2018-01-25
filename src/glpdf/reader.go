package glpdf

import (
	"bufio"
	"io"
	"os"
)

type RandomReader interface {
	Tell() (offset int64, err error)
	Seek(offset int64, whence int) (off int64, err error)
	ReadString(delim byte) (line string, err error)
	ReadBytes(delim byte) (line []byte, err error)
	ReadByte() (c byte, err error)
	Read(p []byte) (n int, err error)
	Peek(n int) ([]byte, error)
	UnreadByte() error
}
type bytesReader struct {
	bytes  []byte
	size   int64
	offset int64
}
type fileReader struct {
	f  *os.File
	br *bufio.Reader
}

func newFileReader(f *os.File) (fr *fileReader) {
	fr = new(fileReader)
	fr.f = f
	fr.br = bufio.NewReader(f)
	return fr
}

func (fr *fileReader) Tell() (offset int64, err error) {
	offset, err = fr.f.Seek(0, os.SEEK_CUR)
	offset = offset - int64(fr.br.Buffered())
	return
}
func (fr *fileReader) Seek(offset int64, whence int) (off int64, err error) {
	off, err = fr.f.Seek(offset, whence)
	fr.br.Reset(fr.f)
	return
}
func (fr *fileReader) ReadString(delim byte) (line string, err error) {
	return fr.br.ReadString(delim)
}
func (fr *fileReader) ReadBytes(delim byte) (line []byte, err error) {
	return fr.br.ReadBytes(delim)
}
func (fr *fileReader) ReadByte() (c byte, err error) {
	return fr.br.ReadByte()
}

func (fr *fileReader) Read(p []byte) (n int, err error) {
	return fr.br.Read(p)
}
func (fr *fileReader) Peek(n int) ([]byte, error) {
	return fr.br.Peek(n)
}

func (fr *fileReader) UnreadByte() error {
	return fr.br.UnreadByte()
}

///// bytesReader
func newBytesReader(bytes []byte) (br *bytesReader) {
	br = new(bytesReader)
	br.bytes = bytes
	br.size = int64(len(bytes))
	return
}

func (br *bytesReader) Tell() (offset int64, err error) {

	return br.offset, nil
}
func (br *bytesReader) Seek(offset int64, whence int) (off int64, err error) {
	switch whence {
	case os.SEEK_SET:
		br.offset = offset
	case os.SEEK_CUR:
		br.offset += offset
	case os.SEEK_END:
		br.offset = br.size + offset
	}
	return br.offset, nil
}
func (br *bytesReader) ReadBytes(delim byte) (line []byte, err error) {
	offset := br.offset
	for br.offset < br.size {
		if br.bytes[br.offset] == delim {
			br.offset++
			line = br.bytes[offset:br.offset]
			return
		}
		br.offset++
	}
	err = io.EOF

	return
}
func (br *bytesReader) ReadString(delim byte) (line string, err error) {
	bytes, err := br.ReadBytes(delim)
	return string(bytes), err
}

func (br *bytesReader) ReadByte() (c byte, err error) {
	if br.offset >= br.size {
		err = io.EOF
	} else {
		c = br.bytes[br.offset]
		br.offset++
	}
	return
}
func (br *bytesReader) Read(p []byte) (n int, err error) {
	leng := int64(len(p))
	left := br.size - br.offset
	if leng < left {
		n = int(leng)
	} else {
		n = int(left)
	}
	offset := br.offset
	br.offset += int64(n)
	copy(p, br.bytes[offset:br.offset])
	return
}
func (br *bytesReader) Peek(n int) (b []byte, err error) {
	left := br.size - br.offset
	var leng int64 = int64(n)
	if leng > left {
		leng = left
		err = io.EOF
	}
	b = br.bytes[br.offset : br.offset+leng]
	return
}
func (br *bytesReader) UnreadByte() error {
	br.offset--
	return nil
}
