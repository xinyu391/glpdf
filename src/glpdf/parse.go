package glpdf

import (
	"bufio"
	"bytes"
	"strconv"
)

type PdfObj struct {
	dict map[Name]DataType
}

const (
	TK_EOF = iota
	TK_NAME
	TK_NUMBER
	TK_COMMENT
	TK_STRING
	TK_BEGIN_OBJ //5
	TK_END_OBJ
	TK_BEGIN_STRING
	TK_END_STRING
	TK_BEGIN_ARRAY
	TK_END_ARRAY //10
	TK_BOOL_TRUE
	TK_BOOL_FALSE
	TK_BEGIN_STREAM
	TK_END_STREAM
	TK_BEGIN_DICT //15
	TK_END_DICT
	TK_R
	TK_INT
	TK_REAL
	TK_KEYWORD //20
	TK_NULL
)
const (
	EOF = 0xff
)

func peek(br *bufio.Reader) (t int, str string, n int32, r float32) {
	t = TK_NULL
	var e error
	for {
		c1, _ := br.Peek(2)
		c := c1[0]
		log("peek2 ", c1, isWhite(c))
		switch {
		case isWhite(c):
			skipWhite(br)
			log("skipwhite")
			continue
		case isNumber(c):
			n, r, e = parseNumber(br)
			log("isNumber", n, r, e)
			t = TK_NUMBER
			return
		}
		switch c {
		case EOF:
			t = TK_EOF
			return
		case '%':
			skipComment(br)
		case '/':
			str, _ = parseName(br)

			t = TK_NAME
			return
		case '(':
			br.ReadByte()
			str, _ = parseString1(br)
			t = TK_EOF
			return
		case ')':
			log("warning lexical error :unexpected ')'")
		case '<':

			if c1[1] == '<' {
				br.ReadByte()
				br.ReadByte()
				t = TK_BEGIN_DICT
				return
			} else { //string
				br.ReadByte()
				str, _ = parseString2(br)
				t = TK_STRING
				return
			}

		case '>':
			if c1[1] == '>' {
				br.ReadByte()
				br.ReadByte()
				t = TK_END_DICT
			} else {
				log("error ")
			}
			br.ReadByte()
		case '[':
			t = TK_BEGIN_ARRAY
			return
		case ']':
			t = TK_END_ARRAY
		case '{':
			panic("{{{{{{")
		case '}':
			panic("}}}}")

		default:
			str, _ = parseName(br)
			log("parseNmae", str)
			t = typeByName(str)
			return
		}

	}
	return
}
func typeByName(name string) int {
	switch name {
	case "R":
		return TK_R
	case "true":
		return TK_BOOL_TRUE
	case "false":
		return TK_BOOL_FALSE
	case "null":
		return TK_NULL
	case "obj":
		return TK_BEGIN_OBJ
	case "endobj":
		return TK_END_OBJ
	case "stream":
		return TK_BEGIN_STREAM
	case "endstream":
		return TK_END_STREAM
	default:
		return TK_KEYWORD

	}
}

func parseObject(br *bufio.Reader) (obj *PdfObj, err error) {
	finish := false
	for {
		token, name, n, r := peek(br)
		log("peek ", token, name, n, r)

		if token == 1 {
			break
		}
		log(token, name, n, r)
		switch token {
		case TK_NAME:
			//k, _ := parseName(br)
			//log(k)
		case TK_END_OBJ:
			finish = true
		case TK_STRING:
			//k, _ := parseString1(br)
			//log(k)
		}
		if finish {
			break
		}
	}
	return
}
func parseNumber(br *bufio.Reader) (n int32, r float32, err error) {
	var c byte
	isreal := false
	finish := false
	var n64 int64
	var r64 float64
	n = -1
	r = -1
	rawBuf := make([]byte, 0, 128)
	buf := bytes.NewBuffer(rawBuf)
	for {
		c1, _ := br.Peek(1)
		c = c1[0]
		if isWhite(c) || isDelim(c) {
			break
		}
		br.ReadByte()
		switch c {
		case '.':
			isreal = true
		case '-':
			buf.WriteByte(c)

		case EOF:
			finish = true
		default:
			buf.WriteByte(c)
		}
		if finish {
			break
		}
	}
	if isreal {
		r64, err = strconv.ParseFloat(buf.String(), 64)
		if err == nil {
			r = float32(r64)
		}

	} else {
		n64, err = strconv.ParseInt(buf.String(), 10, 32)
		if err == nil {
			n = int32(n64)
		}
	}
	if err != nil {
		log("parseNum", err)
	}
	return
}
func parseName(br *bufio.Reader) (name string, err error) {
	rawBuf := make([]byte, 0, 128)
	buf := bytes.NewBuffer(rawBuf)
	var c byte
	for {
		c1, _ := br.Peek(1)
		c = c1[0]

		if isWhite(c) || isDelim(c) {
			break
		}
		if c == '#' {
			panic("# in name ??")
		}
		c, err = br.ReadByte()

		buf.WriteByte(c)
	}
	name = buf.String()
	return
}
func parseString2(br *bufio.Reader) (str string, err error) {
	return br.ReadString('>')
}
func parseString1(br *bufio.Reader) (str string, err error) {
	bal := 1
	finish := false
	buf := make([]byte, 0, 32)
	strBuf := bytes.NewBuffer(buf)
	c, _ := br.ReadByte() // read first (
	for {
		c, _ = br.ReadByte()
		switch c {
		case '(':
			bal++
			strBuf.WriteByte(c)
		case ')':
			bal--
			if bal == 0 {
				finish = true
			} else {
				strBuf.WriteByte(c)
			}
		case '\\':
			c, _ = br.ReadByte()
			switch c {
			case EOF:
				finish = true
			case 'n':
				strBuf.WriteByte('\n')
			case 'r':
				strBuf.WriteByte('\r')
			case 't':
				strBuf.WriteByte('\t')
			case 'b':
				strBuf.WriteByte('\b')
				//			case '(':
				//				strBuf.WriteByte('(')
				//			case ')':
				//				strBuf.WriteByte(')')
				//			case '\n':
				//				strBuf.WriteByte('\n')
			default:
				strBuf.WriteByte(c)
			}
		}
		if finish {
			break
		}
	}
	str = strBuf.String()
	return
}
func isNumber(c byte) bool {
	if c == '+' || c == '-' || c == '.' || (c >= '0' && c <= '9') {
		return true
	}
	return false
}
func isDelim(c byte) bool {
	if c == '(' || c == ')' || c == '<' || c == '>' || c == '[' || c == ']' || c == '{' || c == '}' || c == '/' || c == '%' {
		return true
	}
	return false
}
func isWhite(c byte) bool {
	if c == '\x00' || c == '\x09' || c == '\x0a' || c == '\x0c' || c == '\x0d' || c == '\x20' {
		return true
	}
	return false
}
func skipWhite(r *bufio.Reader) {
	for {
		c1, err := r.Peek(1)
		if err != nil {
			break
		}
		c := c1[0]
		if c <= 32 && isWhite(c) {
			r.ReadByte()
		} else {
			break
		}
	}
}
func skipComment(r *bufio.Reader) {
	for {
		c, err := r.ReadByte()
		if err != nil {
			break
		}

		if c == '\012' || c == '\015' || c == EOF {
			break
		}
	}
}
