package glpdf

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
)

func decode(buf []byte, name Name, param DataType) (out []byte, err error) {
	if name == "FlateDecode" {
		out, err = fliterFlate(buf)
	}

	if param != nil {
		dict := param.(Dict)
		pre := dict["Predictor"]
		col := dict["Columns"]
		if pre != nil {
			c := col.(int32)
			p := pre.(int32)
			log(param)
			out, err = pngFilter(out, int(c), int(p), 1, 8)
		}
	}
	return
}

//ASCIIHexDecode  2HEX, '20' ->  ' '
func filterASCIIHexDecode() {

}
func filterASCII85Decode() {

}

//zlib
func fliterFlate(buf []byte) (out []byte, err error) {
	r := bytes.NewReader(buf)
	gr, err := zlib.NewReader(r)
	if err != nil {
		return
	}
	defer gr.Close()
	out, err = ioutil.ReadAll(gr)
	return
}
func pngFilter(buf []byte, column, predictor, colors, bpp int) (out []byte, err error) {

	width := column * colors * bpp / 8
	height := len(buf) / (width + 1)
	out = make([]byte, len(buf)-height)
	hist := make([]byte, width+1)
	offset := 0
	loge("size", width, height)
	tmp := out[:]
	for offset < len(buf) {
		for i := 0; i < width+1; i++ {
			hist[i] += buf[offset]
			offset++
		}
		m := copy(tmp, hist[1:])
		tmp = tmp[m:]

	}
	//loge(string(out))

	return
}
