package glpdf

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
)

type Pdf struct {
	version    string
	xrefoffset int
	objMap     map[int32]*PdfObj
	root       int
	info       int
	size       int
}

type Name string
type Stream struct {
	offset int64
	stream []byte
	load   bool
}
type Dict map[Name]DataType

type ObjRef struct {
	id  int32
	gen int32
}
type DataType interface{}

type pdfObjRef struct {
	offset int
	used   bool
	//	data   *PdfObj
}

type PdfObj struct {
	ref    ObjRef
	data   DataType
	stream *Stream
}

func Open(file string) (*Pdf, error) {
	pdf := new(Pdf)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	pdf.version = readVersion(f)

	if pdf.version == "" {
		return nil, errors.New("Not Pdf File (not find pdf-ver)")
	}
	// read startxref
	offset := readXrefOffset(f)
	if offset == 0 {
		return nil, errors.New("Not Pdf File(not find xref")
	}

	pdf.xrefoffset = offset

	var objRefMap map[int32]*pdfObjRef
	objRefMap, err = readXrefTable(f, offset)

	if err != nil {
		log("parse error:", err)
		return nil, err
	}
	readTrailer(pdf, f)

	fr := NewfileReader(f)
	pdf.objMap = make(map[int32]*PdfObj)
	for k, v := range objRefMap {
		log("read obj ", k)
		if v.used {
			obj, _ := readObject(pdf, fr, v.offset)
			pdf.objMap[k] = obj
			log(obj)
		}
	}
	for k, v := range pdf.objMap {
		log("v.data", k, v.data)
		if v.stream != nil && v.stream.load == false {

			parseStream(fr, pdf, v)
			log("parseStream obj", k, len(v.stream.stream))
		}

	}
	return pdf, nil
}

func (pdf *Pdf) GetPageNum() int {
	return 0
}
func (pdf *Pdf) GetPage(num int) *Page {
	return nil
}

//读取xref对象索引表
func readXrefTable(f *os.File, offset int) (objMap map[int32]*pdfObjRef, err error) {
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

	objMap = make(map[int32]*pdfObjRef)
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
			//			times, _ := strconv.Atoi(tmp[1])
			used := true
			if tmp[2] == "f" {
				used = false
			}
			objMap[int32(id)] = &pdfObjRef{offset, used}
			log(objMap[int32(id)])
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

//func readRoot(pdf *Pdf, f *os.File) {
//	obj := pdf.objMap[pdf.root]
//	log("read root ", pdf.root, "offset", obj.offset)
//	obj.data, _ = readObject(pdf, f, obj.offset)
//}

func readObject(pdf *Pdf, fr *fileReader, offset int) (obj *PdfObj, err error) {
	//跳转到对象开始位置
	fr.Seek(int64(offset), os.SEEK_SET)
	obj, err = parseObject(fr)
	return
}
