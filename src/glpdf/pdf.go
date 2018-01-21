package glpdf

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strconv"
	"strings"
)

const (
	DTYPE_STRING1 = iota //()
	DTYPE_STRING2        // <>
	DTYPE_NUMBER
	DTYPE_BOOL
	DTYPE_NAME
	DTYPE_ARRAY
	DTYPE_DICT
	DTYPE_STREAM
	DTYPE_REF // 1 0R
	DTYPE_NULL
)

type Pdf struct {
	version    string
	xrefoffset int
	objMap     map[int]*object
	root       int
	info       int
	size       int
}
type object struct {
	id     int
	offset int
	tiems  int
	used   bool
	data   *objectData
}
type Name string
type Stream []byte
type DataType interface{}

type objectData struct {
	id     int
	dict   map[Name]DataType
	stream Stream
}

func Open(file string) (*Pdf, error) {
	pdf := new(Pdf)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	pdf.version = readVersion(f)

	if pdf.version == "" {
		return nil, errors.New("Not Pdf File")
	}
	// read startxref
	offset := readXrefOffset(f)
	if offset == 0 {
		return nil, errors.New("Not Pdf File")
	}

	pdf.xrefoffset = offset

	pdf.objMap, err = readXrefTable(f, offset)

	if err != nil {
		log("parse error:", err)
		return nil, err
	}
	readTrailer(pdf, f)

	// parse root
	readRoot(pdf, f)

	if pdf.info > 0 {
		readObject(pdf, f, pdf.objMap[pdf.info].offset)
	}
	// pages num
	root := pdf.objMap[pdf.root]
	pagesid := root.data.dict["/Pages"]

	pages, _ := readObject(pdf, f, pdf.objMap[pagesid.(int)].offset)
	log(pages)
	return pdf, nil
}

func (pdf *Pdf) GetPageNum() int {
	return 0
}
func (pdf *Pdf) GetPage(num int) *Page {
	return nil
}

//读取xref对象索引表
func readXrefTable(f *os.File, offset int) (objMap map[int]*object, err error) {
	f.Seek(int64(offset), os.SEEK_SET)
	br := bufio.NewReader(f)

	l, err := br.ReadString('\n')
	if err != nil {
		return
	}
	l = strings.TrimSpace(l)
	if l != "xref" {
		return objMap, errors.New("Not find xref Table")
	}

	objMap = make(map[int]*object)
	count := 0
	id := 0
	for {
		l, err = br.ReadString('\n')
		if err != nil {
			break
		}
		l = strings.TrimSpace(l)
		log("line ", l)
		tmp := strings.Split(l, " ")
		if len(tmp) == 2 { // 该段的起始号，和数量
			log(" size")
			id, _ = strconv.Atoi(tmp[0])
			count, _ = strconv.Atoi(tmp[1])
		} else {

			offset, _ := strconv.Atoi(tmp[0])
			times, _ := strconv.Atoi(tmp[1])
			used := true
			if tmp[2] == "f" {
				used = false
			}
			objMap[id] = &object{id, offset, times, used, nil}
			log(objMap[id])
			id++
			count--

		}

		if count == 0 { // 将f重新定位到已读的位置(bufio.Reader会缓存一些）
			r := br.Buffered()
			f.Seek(int64(-r), os.SEEK_CUR)
			break
		}

	}
	return
}
func readVersion(f *os.File) (version string) {
	buf := make([]byte, 32)
	n, _ := f.Read(buf)
	for i := 0; i < n; i++ {
		c := buf[i]
		if c == '\n' || c == '\r' {
			version = string(buf[5:i])
			break
		}
		//		buf.WriteByte(c)
	}
	return
}
func readTrailer(pdf *Pdf, f *os.File) error {
	br := bufio.NewReader(f)

	l, err := br.ReadString('\n')
	if err != nil {
		return nil
	}
	l = strings.TrimSpace(l)
	if l != "trailer" {
		return errors.New("Not find trailer")
	}
	for {
		l, err = br.ReadString('\n')
		if err != nil {
			return err
		}
		l = strings.TrimSpace(l)
		if l == "startxref" {
			break
		}
		//		log("tralier ", l)

		switch tmp := strings.Split(l, " "); tmp[0] {
		case "/Root":
			id, _ := strconv.Atoi(tmp[1])
			//			log("root id ", id)
			pdf.root = id
		case "/Size":
			id, _ := strconv.Atoi(tmp[1])
			//			log("Size is ", id)
			pdf.size = id
		case "/Info":
			id, _ := strconv.Atoi(tmp[1])
			//			log("Info id ", id)
			pdf.info = id
		case "/Encrypt":
		default:

		}
	}
	return nil
}

//read startxref offset
func readXrefOffset(f *os.File) int {
	f.Seek(-32, os.SEEK_END)
	br := bufio.NewReader(f)
	offset := 0
	for {
		l, err := br.ReadString('\n')

		if err != nil || l == "" {
			break
		}
		l = strings.TrimSpace(l)

		if strings.Compare(l, "startxref") == 0 {

			l, err = br.ReadString('\n')
			if err != nil {
				break
			}
			l = strings.TrimSpace(l)
			offset, err = strconv.Atoi(l)
			//log("find offset", l, offset, err)
		}

	}

	return offset
}

func readRoot(pdf *Pdf, f *os.File) {
	obj := pdf.objMap[pdf.root]
	log("read root ", obj.offset)
	obj.data, _ = readObject(pdf, f, obj.offset)
}
func readObject(pdf *Pdf, f *os.File, offset int) (obj *objectData, err error) {
	//跳转到对象开始位置
	f.Seek(int64(offset), os.SEEK_SET)

	br := bufio.NewReader(f)
	//读取对象号
	line, err := br.ReadString('\n') // n 0 obj [<<]
	if err != nil {
		return
	}
	log(line)
	tmp := strings.Split(line, " ")
	if strings.TrimSpace(tmp[2]) != "obj" {
		panic("it should read 'obj'")
	}

	id, _ := strconv.Atoi(tmp[0])

	started := false
	if len(tmp) == 4 {
		if strings.TrimSpace(tmp[3]) == "<<" {
			started = true
		}
	}
	// 解析对象的词典数据
	dict, _ := readDictonary(br, started)
	// 解析对象的流数据
	w, _ := readWord(br)
	if w == "stream" {
		//stream, _ := readStream(br)
		w, _ = readWord(br)
	}
	if w == "endobj" {
	}
	obj = &objectData{id, dict, nil}
	return
}
func readDictonary(br *bufio.Reader, started bool) (dict map[Name]DataType, err error) {
	dict = make(map[Name]DataType)
	if !started {
		//读取2个字符<<
		num := 0
		var c byte
		for {
			c, _ = br.ReadByte()
			if c == '<' {
				num++
			} else if num == 1 {
				panic("two '<' shoule be togher!")
			}
			if num == 2 {
				started = true
				break
			}
		}

	}

	for {
		if endDict(br) {
			log(dict)
			break
		}
		keyStr, _ := readWord(br)

		key := Name(keyStr)
		skipSpace(br)
		pre, _ := br.Peek(20)
		dtype := decideType(pre)
		log("find key", key, ".type is ", dtype)

		switch dtype {
		case DTYPE_STRING1: // read "until: )
			str, _ := readString1(br)
			log("\t(string)", str)
			dict[key] = str
			break
		case DTYPE_STRING2: //read until >
			str, _ := br.ReadString('>')
			log("\t<string>", str)
			dict[key] = str
		case DTYPE_BOOL:
			bolStr, _ := readWord(br)
			bol, _ := strconv.ParseBool(bolStr)
			dict[key] = bol
			log(bol)
		case DTYPE_NAME:
			name, _ := readWord(br)
			log("\tName value:", name)
			dict[key] = name
		case DTYPE_NUMBER:
			numStr, _ := readWord(br)
			num, _ := strconv.Atoi(numStr)
			log("\t num", num)
			dict[key] = num
		case DTYPE_REF:
			idStr, _ := readWord(br)
			numStr, _ := readWord(br)
			rStr, _ := readWord(br)
			id, _ := strconv.Atoi(idStr)
			dict[key] = id
			log("\tRef value:", idStr, numStr, rStr)
		case DTYPE_ARRAY:
			break
		case DTYPE_DICT:
			readArray(br)
			readDictonary(br, false)
		}

	}

	return
}
func readArray(br *bufio.Reader) (array []DataType) {
	c, _ := br.ReadByte()
	if c != '[' {
		panic("readArray but not start [")
	}
	for {

		str, err := br.ReadString(' ')
		if err != nil {
			log("readArray ", err)
		}
		if str[len(str)-1:] == "]" {
			break
		}
		//dtype := 0
		array = append(array, str)
	}
	return array
}
func readString1(br *bufio.Reader) (str string, err error) {
	c, _ := br.ReadByte()
	if c != '(' {
		panic("readString1 but not start (")
	}
	buf := make([]byte, 0, 32)
	strBuf := bytes.NewBuffer(buf)

	for {
		c, _ = br.ReadByte()
		if c == ')' {
			break
		} else if c == '\\' { // 再读个c进行转义
			c, _ = br.ReadByte()
			switch c {
			case 'n':
				strBuf.WriteByte('\n')
			case 'r':
				strBuf.WriteByte('\r')
			default:
				strBuf.WriteByte(c)
			}
		} else {
			strBuf.WriteByte(c)
		}
	}
	str = strBuf.String()
	return
}
func skipSpace(br *bufio.Reader) (n int, err error) {
	n = 0
	var c []byte
	for {
		c, err = br.Peek(1)

		if c[0] == ' ' || c[0] == '\n' || c[0] == '\r' {
			br.ReadByte()
			n++
		} else {
			break
		}
	}
	return
}
func endDict(br *bufio.Reader) (end bool) {
	skipSpace(br)
	//key, err = br.ReadString(' ')

	c, _ := br.Peek(2)

	if len(c) == 2 {
		if c[0] == '>' && c[1] == '>' {
			br.ReadByte()
			br.ReadByte()
			end = true
		}
	}

	return
}

// /Name, number, true/false, >>
func readWord(br *bufio.Reader) (key string, err error) {
	skipSpace(br)
	//key, err = br.ReadString(' ')
	rawBuf := make([]byte, 0, 128)
	buf := bytes.NewBuffer(rawBuf)
	var c byte
	for {
		c1, _ := br.Peek(1)
		c = c1[0]

		if c == ' ' || c == '\n' || c == '\r' || c == '>' {
			break
		}
		c, err = br.ReadByte()

		buf.WriteByte(c)
	}
	key = buf.String()

	return
}
func decideType(buf []byte) (dtype int) {
	switch {
	case buf[0] == '(':
		dtype = DTYPE_STRING1
	case buf[0] == '[':
		dtype = DTYPE_ARRAY
	case buf[0] == '/':
		dtype = DTYPE_NAME
	case buf[0] == '<' && buf[1] == '<':
		dtype = DTYPE_DICT
	case buf[0] == '<':
		dtype = DTYPE_STRING2
	case buf[0] == 't' || buf[0] == 'f':
		dtype = DTYPE_BOOL
	case (buf[0] >= '0' && buf[0] <= '9') || buf[0] == '+' || buf[0] == '-' || buf[0] == '.': //number TODO
		dtype = numType(buf) //引用包含2个数字和R： 1 0 R
	default:
		dtype = DTYPE_NULL
	}
	return
}

// 10 0 R
//  234232
func numType(buf []byte) int {

	ref := DTYPE_NUMBER
	for i := 1; i < len(buf); i++ {
		if buf[i-1] == ' ' && buf[i] == 'R' {
			ref = DTYPE_REF
		}
	}
	return ref
}
func readStream(r *bufio.Reader) (stream Stream, err error) {
	return
}

func readLine(f *os.File) (line string, err error) {

	return
}
