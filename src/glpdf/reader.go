package glpdf

import (
	"bufio"
	"os"
)

type fileReader struct {
	f  *os.File
	br *bufio.Reader
}

func NewfileReader(f *os.File) (fr *fileReader) {
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
