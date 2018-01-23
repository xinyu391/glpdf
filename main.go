package main

import (
	"fmt"
	"glpdf"
	"strconv"
)

func main() {
	fmt.Println("pdf.go", len("位"))
	pdf, err := glpdf.Open("sample.pdf")
	fmt.Println("pdf ", pdf, err)
	num := pdf.GetPageNum()
	page := pdf.GetPage(num - 1)
	page.Draw()
	test()
}
func test() {
	str := "6d79737472696e67"
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
