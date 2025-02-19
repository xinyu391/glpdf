package glpdf

import (
	"bytes"
	"io"
	"os"
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
	TK_STARTXREF
	TK_XREF
	TK_BEGIN_BRACE
	TK_END_BRACE
)
const (
	EOF = 0xff
)

const (
	NAME_LENGTH      = "Length"
	NAME_NAME        = "Name"
	NAME_FILTER      = Name("Filter")
	NAME_FlateDecode = Name("FlateDecode")
	NAME_DECODEPARMS = Name("DecodeParms")
)

type Token struct {
	code int
	buf  string // str []byte
	n    int32
	r    float32
	hex  bool
}

func (t *Token) is(code ...int) bool {
	for _, v := range code {
		if t.code == v {
			return true
		}
	}
	return false
}
func (t *Token) value() (d DataType) {
	switch t.code {
	case TK_INT:
		d = t.n
	case TK_REAL:
		d = t.r
	case TK_BOOL_FALSE:
		d = false
	case TK_BOOL_TRUE:
		d = true
	case TK_NAME:
		d = t.buf
	case TK_STRING:
		d = t.buf
	default:

	}
	return
}
func (t *Token) isKeyword(str string) (ok bool) {
	return t.code == TK_KEYWORD && t.buf == str
}
func (t *Token) str() (str string, ok bool) {
	if t.code == TK_STRING || t.code == TK_NAME || t.code == TK_KEYWORD {
		return t.buf, true
	} else {
		return "", false
	}
}
func (t *Token) num() (n int32, ok bool) {
	if t.code == TK_INT {
		return t.n, true
	}
	return
}
func (t *Token) real() (r float32, ok bool) {
	if t.code == TK_INT {
		return t.r, true
	}
	return
}

func lexer(fr RandomReader) (tk *Token) {
	tk = new(Token)
	tk.code = TK_NULL

	for {
		c, err := fr.ReadByte()
		if err == io.EOF {
			tk.code = TK_EOF
			return
		}
		//c := c1[0]
		//		logd("\tpeek2 ", c1, isWhite(c))
		switch {
		case isWhite(c):
			fr.UnreadByte()
			skipWhite(fr)
			//			logd("\tskipwhite")
			continue
		case isNumber(c):
			fr.UnreadByte()
			tk.code, tk.n, tk.r, _ = parseNumber(fr)
			//			logd("\tisNumber", tk)
			return
		}
		switch c {
		case EOF:
			tk.code = TK_EOF
			return
		case '%':
			skipComment(fr)
		case '/':
			tk.buf, _ = parseName(fr)
			tk.code = TK_NAME
			//			logd("\tisName", tk)
			return
		case '(':
			tk.buf, _ = parseString1(fr)
			tk.code = TK_STRING
			//			logd("\tisString xxxxxxxxxxxxxxx ", tk.buf)
			return
		case ')':
			logw("\twarning lexical error :unexpected ')'")
		case '<':
			c1, _ := fr.ReadByte()
			if c1 == '<' {
				tk.code = TK_BEGIN_DICT
				//				logd("\tfind begin dict")
				return
			} else { //string
				fr.UnreadByte()
				tk.buf, _ = parseString2(fr)
				tk.code = TK_STRING
				tk.hex = true
				return
			}

		case '>':
			c1, _ := fr.ReadByte()
			if c1 == '>' {
				tk.code = TK_END_DICT
				return
			} else {
				//				logd("\terror ")
				fr.UnreadByte()
			}
		case '[':

			tk.code = TK_BEGIN_ARRAY
			return
		case ']':

			tk.code = TK_END_ARRAY
			return
		case '{':
			panic("{{{{{{")
			tk.code = TK_BEGIN_BRACE
			return
		case '}':
			panic("}}}}}")
			tk.code = TK_BEGIN_BRACE
			return

		default:
			fr.UnreadByte()
			tk.buf, _ = parseName(fr)
			tk.code = typeByName(tk.buf)
			//			logd("\tparseNmae", tk)
			return
		}

	}
	return
}

//Deprecated
func peek(fr RandomReader) (t int, str string, n int32, r float32) {
	t = TK_NULL
	var e error
	for {
		c1, err := fr.Peek(2)
		if err == io.EOF {
			t = TK_EOF
			return
		}
		c := c1[0]
		logd("\tpeek2 ", c1, isWhite(c))
		switch {
		case isWhite(c):
			skipWhite(fr)
			logd("\tskipwhite")
			continue
		case isNumber(c):
			t, n, r, e = parseNumber(fr)
			logd("\tisNumber", n, r, e)

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
			logd("\tTK_name ", str)
			t = TK_NAME
			return
		case '(':
			fr.ReadByte()
			str, _ = parseString1(fr)
			logd("\tTK_string ", str)
			t = TK_STRING
			return
		case ')':
			logw("\twarning lexical error :unexpected ')'")
		case '<':

			if c1[1] == '<' {
				fr.ReadByte()
				fr.ReadByte()
				t = TK_BEGIN_DICT
				logd("\tfind begin dict")
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
				logd("\terror ")
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
			logd("\tparseNmae", str, t)
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
	case "startxref":
		return TK_STARTXREF
	case "xref":
		return TK_XREF
	default:
		return TK_KEYWORD

	}
}

func parseObject(fr RandomReader) (obj *PdfObj, err error) {

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
	//	log("begin obj:", n, g, "obj")
	ref := ObjRef{n, g}
	//read dictonary

	token, str, n, r := peek(fr)
	var data DataType
	switch token {
	case TK_BEGIN_DICT:
		//		log("read dict of obj")
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
func markStream(fr RandomReader) (stream *Stream, err error) {
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
func parseStream(fr RandomReader, pdf *Pdf, obj *PdfObj) {
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
	parseStreamWithLength(fr, obj, length)
}
func parseStreamWithLength(fr RandomReader, obj *PdfObj, length int32) {
	stream := obj.stream
	if stream.offset <= 0 {
		return
	}
	dict, _ := obj.data.(Dict)
	fr.Seek(stream.offset, os.SEEK_SET)
	buf := make([]byte, length)
	fr.Read(buf)
	//	log(length, "read ", nn)
	// 根据filter 解密
	filter := dict[NAME_FILTER]
	param := dict[NAME_DECODEPARMS]
	var err error
	if filter != nil {
		if name, ok := filter.(Name); ok {
			buf, err = decode(buf, name, param)
			if err != nil {
				loge("decode failed with object ", obj.ref.id, err)
			}
		} else {
			//array TODO
			//			for k, v := range filter {
			//buf, err := decode(buf, name, param)
			//			}
		}
	}
	stream.load = true
	stream.stream = buf
}
func parseDict(fr RandomReader) (dict Dict, err error) {
	dict = make(Dict)
	finish := false
	var key Name
	for {
		token, str, n, _ := peek(fr)
	SKIP:
		//log("peek for key:", token, str, n, r)
		if token == TK_END_DICT {
			break
		}
		if token != TK_NAME {
			panic("shoule be name")
		}
		key = Name(str)
		// read value
		token, str, n, _ = peek(fr)
		//log("peek for val:", token, str, n, r)
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
func parseArray(fr RandomReader) (array Array, err error) {
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
func parseNumber(fr RandomReader) (t int, n int32, r float32, err error) {
	var c byte
	isreal := false
	finish := false
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
		r = strToFloat(buf.String())
	} else {
		t = TK_INT
		n = strToInt(buf.String())
	}
	if err != nil {
		loge("parseNum", err)
	}
	return
}
func parseName(fr RandomReader) (name string, err error) {
	rawBuf := make([]byte, 0, 128)
	buf := bytes.NewBuffer(rawBuf)
	var c byte
	for {
		c, err = fr.ReadByte()
		if err != nil {
			return
		}
		//c = c1[0]

		if isWhite(c) || isDelim(c) {
			fr.UnreadByte()
			break
		}
		if c == '#' { // two num (0~9,a~f)
			c1, _ := fr.ReadByte()
			c2, _ := fr.ReadByte()
			var cc = []byte{c1, c2}
			ct, err := hexToInt(cc)
			if err != nil {
				logw("there should be two hex char after # in Name.", err)
			}
			c = byte(ct)
		}
		buf.WriteByte(c)
	}
	name = buf.String()
	return
}
func parseString2(fr RandomReader) (str string, err error) {
	buf, e := fr.ReadBytes('>')
	if e != nil {
		return "", e
	}
	size := len(buf) / 2
	out := make([]byte, size)
	for i := 0; i < size; i++ { // 2个字符 转换成一个byte :'1f'->0x1f
		out[i] = hexToByte(buf[i*2 : i*2+2])
	}
	str = string(out)
	return
}
func parseString1(fr RandomReader) (str string, err error) {
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
			case 'f':
				strBuf.WriteByte('\f')
				//			case '(':
				//				strBuf.WriteByte('(')
				//			case ')':
				//				strBuf.WriteByte(')')
			case '\n':
			case '\r':
				c1, _ := fr.ReadByte()
				if c1 != '\n' {
					fr.UnreadByte()
				}
			case '0':
				fallthrough
			case '1':
				fallthrough
			case '2':
				fallthrough
			case '3':
				fallthrough
			case '4':
				fallthrough
			case '5':
				fallthrough
			case '6':
				fallthrough
			case '7':
				//读取最多三个（<8的）数字
				c1, _ := fr.ReadByte()
				oc := c - '0'
				if isOctalNum(c1) {
					oc = oc*8 + c1 - '0'
					c2, _ := fr.ReadByte()
					if isOctalNum(c2) {
						oc = oc*8 + c2 - '0'
					} else {
						fr.UnreadByte()
					}
				} else {
					fr.UnreadByte()
				}
				strBuf.WriteByte(byte(oc))
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
func isOctalNum(c byte) bool {
	return c >= '0' && c <= '7'
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
func skipWhite(r RandomReader) {
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
func skipComment(r RandomReader) {
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
