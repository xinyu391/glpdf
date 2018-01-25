package glpdf

import (
	"strconv"
)

// '000f' -> 15
func hexStrToInt(str string) int32 {
	i, err := strconv.ParseInt(str, 16, 32)
	if err != nil {
		loge("hexStrToInt", str)
		//		panic("hexStrToInt")
	}
	return int32(i)
}

// '12' ->12
func strToInt(str string) int32 {
	i, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		loge("strToInt", str)
		//		panic("strToInt")
	}
	return int32(i)
}

func strToFloat(str string) float32 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic("strToFloat")
	}
	return float32(f)
}

// '0f' -> 15
func hexTobyte(b []byte) byte {
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

func codeFromString2(bytes []byte) (code uint) {
	a := codeFromString(bytes[:2])
	b := codeFromString(bytes[2:])
	if a >= 0xd800 && a <= 0xdbff && b >= 0xdc00 && b <= 0xdfff {
		return (a-0xd800)<<10 + (b - 0xdc00) + 0x10000
	}
	return
}
func codeFromHex(str string) (code uint32) {
	//<00>  2
	//<dfac> 4
	//<d863ddb9> 8

	i, err := strconv.ParseUint(str, 16, 32)
	if err != nil {
		loge("ss", str)
		panic("not hex string")
	}
	code = uint32(i)
	return
}
func codeFromString(bytes []byte) (code uint) {
	//	loge("bytes", bytes)
	var a uint = 0
	for n := 0; n < len(bytes); n++ {
		a = (a << 8) | uint(bytes[n])
	}
	return a
}
