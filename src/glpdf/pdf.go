package glpdf

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Pdf struct {
	version string
	objMap  map[int32]*PdfObj
	root    int32
	info    int32
	size    int32
	doc     *Doc
}

type Name string
type HexString string
type Stream struct {
	offset int64
	stream []byte
	load   bool
}
type Dict map[Name]DataType
type Array []DataType

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

//func (obj *PdfObj) String() string {
//	return fmt.Sprintf("Obj", obj.ref.id, " ", obj.data, " ", obj.stream)
//}

func (obj *PdfObj) valueOf(key Name) DataType {
	return obj.data.(Dict)[key]
}
func (obj *PdfObj) getRefId(key Name) (id int32) {
	return obj.data.(Dict)[key].(ObjRef).id

}
func init() {
	loadSystemCmap()
}
func Open(file string) (pdf *Pdf, err error) {
	pdf = new(Pdf)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	fr := newFileReader(f)

	pdf.version = readVersion(fr)

	if pdf.version == "" {
		return nil, errors.New("Not Pdf File (not find pdf-ver)")
	}

	//	 read startxref
	offset, err := readXrefOffset(fr)

	if err != nil {
		return
	}

	var objRefMap map[int32]*pdfObjRef
	objRefMap, err = readXrefTable(fr, offset)
	if err != nil && err.Error() == "Not find xref Table" {
		log("try to load all obj from front")
	}
	if err != nil {
		log("parse error:", err)
		return nil, err
	}
	readTrailer(pdf, fr)

	pdf.objMap = make(map[int32]*PdfObj)
	// 读取所有的obj
	for k, v := range objRefMap {
		if v.used {
			log("read obj ", k)
			obj, _ := readObject(pdf, fr, v.offset)
			pdf.objMap[k] = obj
			log(obj)
		}
	}
	// 加载所有的stream(读到内存，并且进行flate解码)
	for _, v := range pdf.objMap {
		//		log("v.data", k, v.data)
		if v.stream != nil && v.stream.load == false {

			parseStream(fr, pdf, v)
			//			log("parseStream obj", k, len(v.stream.stream))
		}

	}
	loadDoc(pdf)
	return pdf, nil
}
func (pdf *Pdf) PageNum() int32 {
	return pdf.doc.count

}
func (pdf *Pdf) Page(num int) *Page {
	return nil
}

//读取xref对象索引表
//func readXrefTable2(fr RandomReader, offset int32) (objMap map[int32]*pdfObjRef, err error) {
//	fr.Seek(int64(offset), os.SEEK_SET)
//	tk, _, _, _ := peek(fr)
//	// tk==TK_XREF
//	tk, _, id, _ := peek(fr)
//	//tk==TK_INT
//	tk, _, count, _ := peek(fr)
//	//tk==TK_INT
//	objMap = make(map[int32]*pdfObjRef, count)

//	for {
//		tk, _, offset, _ := peek(fr)
//		tk, _, gen, _ := peek(fr)
//		tk, _, offset, _ := peek(fr)
//		count--
//		if count == 0 {
//			break
//		}
//	}
//	return
//}
func readXrefObj(fr RandomReader) (objMap map[int32]*pdfObjRef, err error) {
	obj, err := parseObject(fr)
	if obj.stream != nil && obj.stream.load == false {
		var length int32 = 0
		var dict Dict
		var ok bool
		if dict, ok = obj.data.(Dict); ok {
			dt := dict[NAME_LENGTH]
			length = dt.(int32)
		}
		parseStreamWithLength(fr, obj, length)
		//		log("parseStream obj", obj.stream.stream)
	}
	//	log(obj)
	return
}
func readXrefTable(fr RandomReader, offset int32) (objMap map[int32]*pdfObjRef, err error) {
	fr.Seek(int64(offset), os.SEEK_SET)
	tk := lexer(fr)
	//	loge("tk ", tk)
	if tk.code == TK_XREF {
		objMap = make(map[int32]*pdfObjRef)
		count := 0
		id := 0
		for {
			l, err := fr.ReadString('\n')
			if err != nil {
				break
			}
			l = strings.TrimSpace(l)
			//			log("line ", l)
			if l == "" {
				continue
			}
			tmp := strings.Split(l, " ")
			if len(tmp) == 2 { // 该段的起始号，和数量
				//				log(" size")
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
				//				log(objMap[int32(id)])
				id++
				count--

			}

			if count == 0 { // 将f重新定位到已读的位置(bufio.Reader会缓存一些）

				break
			}

		}
		return
	} else if tk.code == TK_INT { //read object
		fr.Seek(int64(offset), os.SEEK_SET)

		return readXrefObj(fr)
	} else {
		return nil, errors.New("malformed PDF: Not find xref data")
	}
}
func readVersion(fr RandomReader) (version string) {
	buf := make([]byte, 32)
	n, _ := fr.Read(buf)
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
func readTrailer(pdf *Pdf, fr RandomReader) error {

	tk := lexer(fr)
	if !tk.isKeyword("trailer") {
		loge("not find trailer")
		return errors.New("not find trailer")
	}
	tk = lexer(fr)
	//	loge("code  ", tk.code)
	if tk.is(TK_BEGIN_DICT) {
		dict, err := parseDict(fr)
		if err == nil {
			pdf.root = dict["Root"].(ObjRef).id
			info := dict["Info"]
			if info != nil {
				pdf.info = info.(ObjRef).id
			}
			pdf.size = dict["Size"].(int32)

		}
	}
	return nil
}

//read startxref offset
func readXrefOffset(fr RandomReader) (offset int32, err error) {
	fr.Seek(-32, os.SEEK_END)
	var tk int
	for {
		tk, _, _, _ = peek(fr)
		if tk == TK_STARTXREF || tk == TK_EOF {
			break
		}
	}
	if tk == TK_STARTXREF {
		tk, _, n, _ := peek(fr)
		if tk == TK_INT {
			offset = n
		} else {
			log("should be int")
		}
	} else {
		err = errors.New("not find startxref")
	}
	return
}

//func readRoot(pdf *Pdf, f *os.File) {
//	obj := pdf.objMap[pdf.root]
//	log("read root ", pdf.root, "offset", obj.offset)
//	obj.data, _ = readObject(pdf, f, obj.offset)
//}

func readObject(pdf *Pdf, fr RandomReader, offset int) (obj *PdfObj, err error) {
	//跳转到对象开始位置
	fr.Seek(int64(offset), os.SEEK_SET)
	obj, err = parseObject(fr)
	return
}
