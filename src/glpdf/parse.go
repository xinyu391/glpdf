package glpdf

import (
	"bytes"
	"os"
	"strconv"
)

const (
	TK_EOF = iota
	TK_NAME
	//	TK_NUMBER
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

const (
	NAME_LENGTH      = "Length"
	NAME_NAME        = "Name"
	NAME_FILTER      = Name("Filter")
	NAME_FlateDecode = Name("FlateDecode")
)

func peek(fr *fileReader) (t int, str string, n int32, r float32) {
	t = TK_NULL
	var e error
	for {
		c1, _ := fr.Peek(2)
		c := c1[0]
		log("\tpeek2 ", c1, isWhite(c))
		switch {
		case isWhite(c):
			skipWhite(fr)
			log("\tskipwhite")
			continue
		case isNumber(c):
			t, n, r, e = parseNumber(fr)
			log("\tisNumber", n, r, e)

			return
		}
		switch c {
		case EOF:
			t = TK_EOF
			return
		case '%':
			skipComment(fr)
		case '/':
			fr.ReadByte()
			str, _ = parseName(fr)
			log("\tTK_name ", str)
			t = TK_NAME
			return
		case '(':
			fr.ReadByte()
			str, _ = parseString1(fr)
			log("\tTK_string ", str)
			t = TK_STRING
			return
		case ')':
			log("\twarning lexical error :unexpected ')'")
		case '<':

			if c1[1] == '<' {
				fr.ReadByte()
				fr.ReadByte()
				t = TK_BEGIN_DICT
				log("\tfind begin dict")
				return
			} else { //string
				fr.ReadByte()
				str, _ = parseString2(fr)
				t = TK_STRING
				return
			}

		case '>':
			if c1[1] == '>' {
				fr.ReadByte()
				fr.ReadByte()
				t = TK_END_DICT
				return
			} else {
				log("\terror ")
			}
			fr.ReadByte()
		case '[':
			fr.ReadByte()
			t = TK_BEGIN_ARRAY
			return
		case ']':
			fr.ReadByte()
			t = TK_END_ARRAY
			return
		case '{':
			panic("{{{{{{")
		case '}':
			panic("}}}}")

		default:
			str, _ = parseName(fr)

			t = typeByName(str)
			log("\tparseNmae", str, t)
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

func parseObject(fr *fileReader) (obj *PdfObj, err error) {

	//read 1 0 R
	token, _, n, _ := peek(fr)
	if token != TK_INT {
		panic("shoule be number(int)")
	}
	token, _, g, _ := peek(fr)
	if token != TK_INT {
		panic("shoule be number(int)")
	}
	token, _, _, _ = peek(fr)
	if token != TK_BEGIN_OBJ {
		panic("shoule be 'obj'")
	}
	log("begin obj:", n, g, "obj")
	ref := ObjRef{n, g}
	//read dictonary

	token, str, n, r := peek(fr)
	var data DataType
	switch token {
	case TK_BEGIN_DICT:
		log("read dict of obj")
		data, _ = parseDict(fr)

	case TK_BEGIN_ARRAY:
		data, _ = parseArray(fr)
	case TK_NAME:
		data = Name(str)
	case TK_REAL:
		data = r
	case TK_STRING:
		data = str
	case TK_BOOL_TRUE:
		data = true
	case TK_BOOL_FALSE:
		data = false
	case TK_NULL:
		data = nil
	case TK_INT:
		token2, _, n2, _ := peek(fr)
		if token2 == TK_BEGIN_STREAM || token2 == TK_END_OBJ {
			token = token2
			data = n
		} else if token2 == TK_INT {
			token3, _, _, _ := peek(fr)
			if token3 == TK_R {
				data = ObjRef{n, n2}
			} else {
				panic("what the hell:two int in obj?")
			}

		}
	case TK_END_OBJ:
		data = nil
	default:
		panic("what the hell ?")
	}
	if token != TK_BEGIN_STREAM && token != TK_END_OBJ {
		token, _, _, _ = peek(fr)
	}
	var stream *Stream
	if token == TK_BEGIN_STREAM {
		stream, _ = markStream(fr)
	}
	obj = &PdfObj{ref, data, stream}
	return
}

// 仅做mark
func markStream(fr *fileReader) (stream *Stream, err error) {
	c, _ := fr.ReadByte()
	for c == ' ' {
		c, _ = fr.ReadByte()
	}
	if c == '\r' {
		c1, _ := fr.Peek(1)
		if c1[0] != '\n' {
			log("warning \\n should after \\r after 'stream'")
		} else {
			c, _ = fr.ReadByte()
		}
	}
	if c != '\n' {
		panic("shoule be \\n after 'stream'")
	}

	offset, _ := fr.Tell()
	stream = &Stream{offset, nil, false}

	return
}

func parseStream(fr *fileReader, pdf *Pdf, obj *PdfObj) {
	stream := obj.stream
	if stream.offset <= 0 {
		return
	}
	var length int32 = 0
	var dict Dict
	var ok bool
	if dict, ok = obj.data.(Dict); ok {
		dt := dict[NAME_LENGTH]
		switch dt.(type) {
		case int32:
			length = dt.(int32)
		case ObjRef:
			rdt := pdf.objMap[dt.(ObjRef).id]
			length = rdt.data.(int32)
		default:
		}
	}
	if length <= 0 {
		log("length is wrontg :", length)
		return
	}
	fr.Seek(stream.offset, os.SEEK_SET)
	buf := make([]byte, length)
	nn, _ := fr.Read(buf)
	log(length, "read ", nn)
	// 根据filter 解密
	filter := dict[NAME_FILTER]
	if filter != nil {
		var param FilterParam
		if n, ok := filter.(Name); ok {
			param = FilterParam{n}
		}
		out, err := decode(buf, param)
		if err == nil {
			buf = out
		}
	}
	stream.load = true
	stream.stream = buf
}
func parseDict(fr *fileReader) (dict Dict, err error) {
	dict = make(Dict)
	finish := false
	var key Name
	for {
		token, str, n, r := peek(fr)
	SKIP:
		log("peek for key:", token, str, n, r)
		if token == TK_END_DICT {
			break
		}
		if token != TK_NAME {
			panic("shoule be name")
		}
		key = Name(str)
		// read value
		token, str, n, r = peek(fr)
		log("peek for val:", token, str, n, r)
		switch token {
		case TK_NAME:
			dict[key] = Name(str)
		case TK_INT:
			// 是一个单独的number，还是个obj ref
			token2, str2, n2, _ := peek(fr)
			if token2 == TK_END_DICT || token2 == TK_NAME || (token2 == TK_KEYWORD && str2 != "ID") {
				dict[key] = n
				token = token2
				str = str2
				goto SKIP
			}
			if token2 == TK_INT {
				t3, _, _, _ := peek(fr)
				if t3 == TK_R {
					dict[key] = ObjRef{n, n2}
				} else {
					panic("what the hell")
				}
			}
		case TK_STRING:
			dict[key] = str

		case TK_BEGIN_ARRAY:
			ary, _ := parseArray(fr)
			dict[key] = ary
		case TK_BEGIN_DICT:
			dic, _ := parseDict(fr)
			dict[key] = dic
		case TK_END_DICT:
			finish = true
		}
		if finish {
			break
		}
	}
	return
}
func parseArray(fr *fileReader) (array []DataType, err error) {
	var a, b int32
	nc := 0
	for {
		token, str, n, r := peek(fr)
		if token != TK_INT && token != TK_R {
			if nc > 0 {
				array = append(array, a)
			}
			if nc > 1 {
				array = append(array, b)
			}
			nc = 0
		}
		if token == TK_INT && nc == 2 {
			array = append(array, a)
			a = b
			nc--
		}
		finish := false
		switch token {
		case TK_END_ARRAY:
			finish = true
		case TK_INT:
			if nc == 0 {
				a = n
			} else if nc == 1 {
				b = n
			}
			nc++
		case TK_R:
			array = append(array, ObjRef{a, b})
			nc = 0
		case TK_BEGIN_ARRAY:
			ary, _ := parseArray(fr)
			array = append(array, ary)
		case TK_BEGIN_DICT:
			dict, _ := parseDict(fr)
			array = append(array, dict)
		case TK_NAME:
			array = append(array, Name(str))
		case TK_REAL:
			array = append(array, r)
		case TK_STRING:
			array = append(array, str)
		case TK_BOOL_TRUE:
			array = append(array, true)
		case TK_BOOL_FALSE:
			array = append(array, false)
		case TK_NULL:
			array = append(array, nil)
		default:
			panic("what hell in array?")
		}
		if finish {
			break
		}
	}
	return
}
func parseNumber(fr *fileReader) (t int, n int32, r float32, err error) {
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
		c1, _ := fr.Peek(1)
		c = c1[0]
		if isWhite(c) || isDelim(c) {
			break
		}
		fr.ReadByte()
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
		t = TK_REAL
		r64, err = strconv.ParseFloat(buf.String(), 64)
		if err == nil {
			r = float32(r64)
		}
	} else {
		t = TK_INT
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
func parseName(fr *fileReader) (name string, err error) {
	rawBuf := make([]byte, 0, 128)
	buf := bytes.NewBuffer(rawBuf)
	var c byte
	for {
		c1, _ := fr.Peek(1)
		c = c1[0]

		if isWhite(c) || isDelim(c) {
			break
		}
		if c == '#' {
			panic("# in name ??")
		}
		c, err = fr.ReadByte()

		buf.WriteByte(c)
	}
	name = buf.String()
	return
}
func parseString2(fr *fileReader) (str string, err error) {
	return fr.ReadString('>')
}
func parseString1(fr *fileReader) (str string, err error) {
	bal := 1
	finish := false
	buf := make([]byte, 0, 32)
	strBuf := bytes.NewBuffer(buf)

	for {
		c, _ := fr.ReadByte()
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
			c, _ = fr.ReadByte()
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
		default:
			strBuf.WriteByte(c)
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
func skipWhite(r *fileReader) {
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
func skipComment(r *fileReader) {
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
