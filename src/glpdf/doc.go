package glpdf

import (
	"fmt"
)

type Doc struct {
	pages []*Page
	count int32
	fonts map[Name]*Font
}
type Page struct {
	content []rune
	//	fonts   map[Name]*Font
	res    *Resource
	width  float32
	height float32
}
type Font struct {
	name            string
	baseFont        Name
	encoding        Name
	descendantFonts ObjRef
	subType         Name
	toUnicode       *Cmap
}

func (f *Font) String() string {
	return fmt.Sprint("Font{", f.name, " BaseName: ", f.baseFont, " SubType: ", f.subType, " ToUnicode: ", f.toUnicode, "}")
}
func (p *Page) String() string {
	return fmt.Sprint("Page(", p.width, ",", p.height, ")", p.content)
}

type Resource struct {
	fonts map[Name]*Font
}

func newDoc() (doc *Doc) {
	doc = new(Doc)
	doc.fonts = make(map[Name]*Font, 4)
	return
}
func newPage() (page *Page) {
	page = new(Page)

	return
}

func (f *Font) unicodeStr(pdf *Pdf, str string) (out string) {
	if f.toUnicode != nil {
		unicode, _ := f.toUnicode.Lookup2([]byte(str))
		out = unicodeToStr(unicode)
	} else if f.encoding != "" {
		switch f.encoding {
		case "WinAnsiEncoding":
			out = ""
		case "Identity-H":
			// 查找 DescendantFonts中的CIDSystemInfo，组合出一个cmap文件名称
			desc := pdf.objMap[f.descendantFonts.id]

			dict := desc.data.(Dict)
			cidInfo := dict["CIDSystemInfo"].(Dict)
			cmapName := fmt.Sprintf("%s-%s-UCS2", cidInfo["Registry"].(string), cidInfo["Ordering"].(string))
			cmap := system_cmap[string(cmapName)]
			if cmap != nil {
				unicode, _ := cmap.Lookup2([]byte(str))
				out = unicodeToStr(unicode)
			}
		case "Identity-V":
		}
	}
	return
}

// 解析所有的Page
func loadDoc(pdf *Pdf) {
	root := pdf.objMap[pdf.root]
	pagesid := root.getRefId("Pages")

	doc := newDoc()
	pagesObj := pdf.objMap[pagesid]
	tp := pagesObj.valueOf("Type").(Name)
	if tp == Name("Pages") {
		loadPages(pdf, doc, pagesObj)
		doc.count = pagesObj.valueOf("Count").(int32)
	}
	pdf.doc = doc
}

func loadPages(pdf *Pdf, doc *Doc, pagesObj *PdfObj) {
	pageAry := pagesObj.valueOf("Kids").(Array)

	for _, p := range pageAry {
		//		log("parse ", len(doc.pages), "page")
		id := p.(ObjRef).id
		obj := pdf.objMap[id]
		tp := obj.valueOf("Type").(Name)
		if tp == Name("Pages") {
			loadPages(pdf, doc, obj)
		} else {
			page := loadPage(pdf, doc, obj)
			doc.pages = append(doc.pages, page)
			loge("    PAGE ", len(doc.pages), ":")
			loge("    \t Resource ", page.res.fonts)
			size := len(page.content)
			if size > 10 {
				size = 10
			}
			loge("    \t Content ", string(page.content))

		}

	}
	return
}

// 加载 page的 cotent和 resource
func loadPage(pdf *Pdf, doc *Doc, pageObj *PdfObj) (page *Page) {
	page = newPage()
	// size
	box := pageObj.valueOf("MediaBox").(Array)
	if v, ok := box[2].(int32); ok {
		page.width = float32(v)
	} else {
		page.width = box[2].(float32)
	}
	if v, ok := box[3].(int32); ok {
		page.height = float32(v)
	} else {
		page.height = box[3].(float32)
	}
	// resource /font
	var res *Resource
	if dict, ok := pageObj.valueOf("Resources").(Dict); ok {
		res = loadResource(pdf, doc, dict)
	} else { // RefObj
		resId := pageObj.getRefId("Resources")
		dict = pdf.objMap[resId].data.(Dict)
		res = loadResource(pdf, doc, dict)
	}
	page.res = res
	loge("fonts ", res.fonts)

	content := pageObj.valueOf("Contents")
	if ref, ok := content.(ObjRef); ok {
		contentId := ref.id
		content := pdf.objMap[contentId].stream.stream
		parsePageContent(pdf, page, content, contentId)
	} else { // is array
		ary := content.(Array)
		for _, v := range ary {
			contentId := v.(ObjRef).id
			content := pdf.objMap[contentId].stream.stream
			parsePageContent(pdf, page, content, contentId)
		}
	}

	return
}
func loadResource(pdf *Pdf, doc *Doc, dict Dict) (res *Resource) {
	res = new(Resource)
	res.fonts = make(map[Name]*Font, 4)
	font := dict["Font"].(Dict)
	for k, v := range font {
		fid := v.(ObjRef).id
		if doc.fonts[k] == nil {
			fobj := pdf.objMap[fid]
			font := loadFont(pdf, string(k), fobj)
			res.fonts[k] = font
			doc.fonts[k] = font
		} else {
			res.fonts[k] = doc.fonts[k]
		}
	}
	return
}

//加载字体
func loadFont(pdf *Pdf, name string, obj *PdfObj) (font *Font) {
	dict := obj.data.(Dict)

	//<</BaseFont /HYUHQC+DroidSansFallback /DescendantFonts [10 0 R] /Encoding /Identity-H /Subtype /Type0 /ToUnicode 9 0 R /Type /Font>>
	font = new(Font)
	//	loge("xxxxxxxxxxxx load font ", name, obj)
	font.name = name
	font.baseFont, _ = dict["BaseFont"].(Name)
	if ary, ok := dict["DescendantFonts"].(Array); ok {
		font.descendantFonts = ary[0].(ObjRef)
	}
	font.encoding, _ = dict["Encoding"].(Name)
	font.subType = dict["Subtype"].(Name)
	if r, ok := dict["ToUnicode"].(ObjRef); ok {
		stream := pdf.objMap[r.id].stream.stream
		font.toUnicode, _ = LoadCmapBytes(stream)
	} else {
		loge("")
	}
	return
}

// 解析 ps绘制命令，
func parsePageContent(pdf *Pdf, page *Page, stream []byte, id int32) {

	//	writeToFile(stream, fmt.Sprintf("%d.txt", id))
	br := newBytesReader(stream)
	ops := make([]Operator, 0, 128)
	args := make([]DataType, 0, 4)
	breakfor := false
	for {
		tk := lexer(br)
		//		loge(tk.code, "   ", tk.buf, "   ", tk.n)
		switch tk.code {
		case TK_EOF:
			breakfor = true
		case TK_KEYWORD:
			op := NewOp(tk.buf, args)
			ops = append(ops, op)
			args = make([]DataType, 0, 4)
		case TK_BEGIN_ARRAY:
			ary, err := parseArray(br)
			if err != nil {
				log(err, ary)
			}
			args = append(args, ary)
		case TK_BEGIN_BRACE:
		case TK_END_BRACE:
		default:
			args = append(args, tk.value())
		}
		if breakfor {
			break
		}

	}
	bt := false
	font := ""
	for _, op := range ops {
		name := op.Name()
		switch name {
		case "BT":
			bt = true
		case "ET":
			bt = false
		case "Tf":
			font = op.Args()[0].(string)
		case "Tj":
			str := op.Args()[0].(string)
			f := page.res.fonts[Name(font)]
			if f != nil {
				s := f.unicodeStr(pdf, str)
				logl(s)
				// TODO save content str
				page.content = append(page.content, []rune(s)...)
			} else {
				//				loge("draw ", str, " with font ", font)
			}
		case "TJ":
			strAry := op.Args()[0].(Array)
			for _, v := range strAry {
				if str, ok := v.(string); ok {
					logl(str)
				}
			}

		default:
		}
		if bt {

		}
	}

	logl("\n")
}
