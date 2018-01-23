package glpdf

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
)

type FilterParam struct {
	filter Name
}

func decode(buf []byte, param FilterParam) (out []byte, err error) {
	r := bytes.NewReader(buf)

	gr, err := zlib.NewReader(r)
	if err != nil {
		return
	}
	defer gr.Close()
	out, err = ioutil.ReadAll(gr)
	return
}
