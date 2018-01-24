package main

import (
	"fmt"
	"glpdf"
	"strconv"
)

func main() {
	pdf, err := glpdf.Open("sample.pdf")
	fmt.Println("pdf ", pdf, err)
	num := pdf.GetPageNum()
	page := pdf.GetPage(num - 1)
	page.Draw()
	test()
	loge("xxxxx", "dd")
}
func loge(a ...interface{}) {
	b := append([]interface{}{}, "[E]")
	b = append(b, a...)
	fmt.Println(b...)
}
func test() {
	s := "한"
	fmt.Println([]byte(s))
	str := "7061636b616765206d61696e"
	buf := []byte(str)

	size := len(buf) / 2
	out := make([]byte, size)
	for i := 0; i < size; i++ {
		v, e := strconv.ParseInt(string(buf[i*2:i*2+2]), 16, 8)
		if e == nil {
			fmt.Print(v, ",")
			out[i] = byte(v) - byte(v)
		}
	}
	for i := 0; i < size; i++ {
		out[i] = hex2byte(buf[i*2 : i*2+2])
	}

	str = string(out)
	fmt.Println(str)
}
func hex2byte(b []byte) byte {
	var a, c byte
	if b[0] <= '9' {
		a = b[0] - '0'
	} else {
		a = b[0] - 'a' + 10
	}
	if b[1] <= '9' {
		c = b[1] - '0'
	} else {
		c = b[1] - 'a' + 10
	}
	return a*16 + c
}
