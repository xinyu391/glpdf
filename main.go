package main

import (
	"fmt"
	"glpdf"
	"strconv"
	"unicode/utf16"
)

func main() {
	//	pdf, err := glpdf.Open("p1.pdf")
	//	fmt.Println("pdf ", pdf, err)
	//	num := pdf.GetPageNum()
	//	page := pdf.GetPage(num - 1)
	//	page.Draw()
	test()
	loge("xxxxx", "dd")
}
func loge(a ...interface{}) {
	b := append([]interface{}{}, "[E]")
	b = append(b, a...)
	fmt.Println(b...)
}
func test() {
	str := `/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CIDSystemInfo <</Registry (Adobe) /Ordering (UCS) /Supplement 0>> def
/CMapName /Adobe-Identity-UCS def
/CMapType 2 def
1 begincodespacerange
<0000> <FFFF>
endcodespacerange
10 beginbfchar
<0001> <0000>
<0002> <0020>
<0003> <0E3F>
<0004> <1100>
<0005> <1101>
<0006> <1102>
<0007> <1103>
<0008> <1104>
<28EF> <5B57>
<39E1> <6C49>
endbfchar
begincidchar
<a0>1
<a9>96
<af>195
<d863ddb9>18942
endcidchar
endcmap CMapName currentdict /CMap defineresource pop end end`
	cmap, _ := glpdf.LoadCmapBytes([]byte(str))
	// 汉字
	cid := "28EF"
	unic, _ := cmap.Lookup2(hexStr2bytes([]byte(cid)))
	r := utf16.Decode([]uint16{uint16(unic)})
	fmt.Println("28ef->", unic, []byte(string(r)))
	//\ud862\udf9c
	s := "df9cd862" // 7FFF FFFF

	i1, _ := strconv.ParseUint(s[:4], 16, 32)
	i2, _ := strconv.ParseUint(s[4:], 16, 32)

	loge(s, i1, i2)
	i, _ := strconv.ParseUint(s, 16, 32)
	n := int32(i)
	loge(s, "-2>", i, i>>16, i&0x0000ffff, n)

	s2 := "𨮜" //\ud862\udf9c
	loge(s2, "->", []byte(s2))
}
func test2() {

	str := "6e61636b616765206d61696E"
	buf := []byte(str)

	size := len(buf) / 2
	out := make([]byte, size)
	fmt.Print("[")
	for i := 0; i < size; i++ {
		v, e := strconv.ParseInt(string(buf[i*2:i*2+2]), 16, 8)
		if e == nil {
			fmt.Print(v, ",")
			out[i] = byte(v) - byte(v)
		}
	}
	fmt.Println("]")
	for i := 0; i < size; i++ {
		out[i] = hex2byte(buf[i*2 : i*2+2])
	}

	str = string(out)
	fmt.Println(out)
}
func hexStr2bytes(b []byte) []byte {
	size := len(b)
	out := make([]byte, size/2)
	j := 0
	for i := 0; i < size; i += 2 {
		out[j] = hex2byte(b[i : i+2])
		j++

	}
	return out
}
func hex2byte(b []byte) byte {
	var a, c byte
	if b[0] <= '9' {
		a = b[0] - '0'
	} else {
		if b[0] >= 'a' {
			a = b[0] - 'a' + 10
		} else {
			a = b[0] - 'A' + 10
		}
	}
	if b[1] <= '9' {
		c = b[1] - '0'
	} else {
		if b[1] >= 'a' {
			c = b[1] - 'a' + 10
		} else {
			c = b[1] - 'A' + 10
		}
	}
	return a*16 + c
	//	n, _ := strconv.ParseUint(string(b), 16, 8)
	//	return byte(n)
}
