package glpdf

type Doc struct {
	pages []*Page
	count int32
	fonts map[Name]*Font
}
type Page struct {
	content string
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

}
func loadPages(pdf *Pdf, doc *Doc, pagesObj *PdfObj) {
	pageAry := pagesObj.valueOf("Kids").(Array)

	for i, p := range pageAry {
		log("parse ", i+1, "page")
		id := p.(ObjRef).id
		obj := pdf.objMap[id]
		tp := obj.valueOf("Type").(Name)
		if tp == Name("Pages") {
			loadPages(pdf, doc, obj)
		} else {
			page := loadPage(pdf, doc, obj)
			doc.pages = append(doc.pages, page)
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

	content := pageObj.valueOf("Contents")
	if ref, ok := content.(ObjRef); ok {
		contentId := ref.id
		content := pdf.objMap[contentId].stream.stream
		parsePageContent(page, content)
	} else { // is array
		ary := content.(Array)
		for _, v := range ary {
			contentId := v.(ObjRef).id
			content := pdf.objMap[contentId].stream.stream
			parsePageContent(page, content)
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
			font := loadFont(pdf, fobj)
			res.fonts[k] = font
			doc.fonts[k] = font
		} else {
			res.fonts[k] = doc.fonts[k]
		}
	}
	return
}

//加载字体
func loadFont(pdf *Pdf, obj *PdfObj) (font *Font) {
	dict := obj.data.(Dict)

	//<</BaseFont /HYUHQC+DroidSansFallback /DescendantFonts [10 0 R] /Encoding /Identity-H /Subtype /Type0 /ToUnicode 9 0 R /Type /Font>>
	font = new(Font)
	font.baseFont, _ = dict["BaseFont"].(Name)
	font.descendantFonts, _ = dict["DescendantFonts"].(ObjRef)
	font.encoding, _ = dict["Encoding"].(Name)
	font.subType = dict["Subtype"].(Name)
	if r, ok := dict["ToUnicode"].(ObjRef); ok {
		stream := pdf.objMap[r.id].stream.stream
		font.toUnicode, _ = LoadCmapBytes(stream)
	}
	return
}

// 解析 ps绘制命令，
func parsePageContent(page *Page, stream []byte) {
	loge("parsePageContent", string(stream))
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
			if f != nil && f.toUnicode != nil {
				unicode, _ := f.toUnicode.Lookup2([]byte(str))
				loge(unicodeToStr(unicode))
			} else {
				loge("draw ", str, " with font ", font)
			}
		case "TJ":
			strAry := op.Args()[0].(Array)
			for _, v := range strAry {
				if str, ok := v.(string); ok {
					logl(str)
				}
			}
			logl("\n")
		default:
		}
		if bt {

		}
	}
}
