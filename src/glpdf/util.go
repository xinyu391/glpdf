package glpdf

import (
	"io/ioutil"
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

// 需求
// Name解析中，遇到#xx 进行2位hex数字解析,  []byte -> uint32
func hexToInt(bytes []byte) (uint32, error) {
	i, err := strconv.ParseInt(string(bytes), 16, 32)
	return uint32(i), err
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

// [0x01,0x01] ->257
//func bstrToInt(str string) uint32 {
//	bytes := []byte(str)
//	return bytesToInt(bytes)
//}

// [0x01,0x01] ->257
func bytesToInt(bytes []byte) uint32 {
	var a uint32 = 0
	for _, b := range bytes {
		a = a<<8 | uint32(b)
	}
	return a
}
func strToFloat(str string) float32 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic("strToFloat")
	}
	return float32(f)
}

// 'f' -> 15
func _hexTobyte(b byte) byte {
	if b >= '0' && b <= '9' {
		return b & 0x0f
	}
	if (b >= 'A' && b <= 'Z') || (b >= 'a' && b < 'z') {
		return b&0x0f + 9
	}
	return 0xff
}

func HexToByte(b []byte) byte {
	return hexToByte(b)
}
func hexToByte(b []byte) byte {
	h := _hexTobyte(b[0])
	l := _hexTobyte(b[1])
	return h<<4 | l
}

//func hexTobyte(b []byte) byte {
//	var a, c byte
//	if b[0] <= '9' {
//		a = b[0] - '0'
//	} else {
//		if b[0] >= 'a' {
//			a = b[0] - 'a' + 10
//		} else {
//			a = b[0] - 'A' + 10
//		}
//	}
//	if b[1] <= '9' {
//		c = b[1] - '0'
//	} else {
//		if b[1] >= 'a' {
//			c = b[1] - 'a' + 10
//		} else {
//			c = b[1] - 'A' + 10
//		}
//	}
//	return a*16 + c
//	//	n, _ := strconv.ParseUint(string(b), 16, 8)
//	//	return byte(n)
//}

//func codeFromString2(bytes []byte) (code uint) {
//	a := codeFromString(bytes[:2])
//	b := codeFromString(bytes[2:])
//	if a >= 0xd800 && a <= 0xdbff && b >= 0xdc00 && b <= 0xdfff {
//		return (a-0xd800)<<10 + (b - 0xdc00) + 0x10000
//	}
//	return
//}
//func codeFromHex(str string) (code uint32) {
//	//<00>  2
//	//<dfac> 4
//	//<d863ddb9> 8

//	i, err := strconv.ParseUint(str, 16, 32)
//	if err != nil {
//		loge("ss", str)
//		panic("not hex string")
//	}
//	code = uint32(i)
//	return
//}
//func codeFromString(bytes []byte) (code uint) {
//	//	loge("bytes", bytes)
//	var a uint = 0
//	for n := 0; n < len(bytes); n++ {
//		a = (a << 8) | uint(bytes[n])
//	}
//	return a
//}
func unicodeToStr(unicode []int32) (str string) {
	str = string(unicode)
	return
}

func writeToFile(data []byte, name string) {
	ioutil.WriteFile(name, data, 0666)
}
