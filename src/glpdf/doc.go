package glpdf

type Doc struct {
	pages []*Page
	count int32
	fonts map[Name]*Font
}
type Page struct {
	content string
	fonts   map[Name]*Font
	width   int32
	height  int32
}
type Font struct {
	name            string
	baseFont        Name
	encoding        Name
	descendantFonts ObjRef
	subType         Name
	toUnicode       *Cmap
}

func newDoc() (doc *Doc) {
	doc = new(Doc)
	doc.fonts = make(map[Name]*Font, 4)
	return
}
func newPage() (page *Page) {
	page = new(Page)
	page.fonts = make(map[Name]*Font, 4)
	return
}
func loadDoc(pdf *Pdf) {
	root := pdf.objMap[pdf.root]
	pagesid := root.getRefId("Pages")

	doc := newDoc()
	pagesObj := pdf.objMap[pagesid]
	pageAry := pagesObj.valueOf("Kids").(Array)
	doc.count = pagesObj.valueOf("Count").(int32)
	for i, p := range pageAry {
		log("parse ", i+1, "page")
		id := p.(ObjRef).id
		page := loadPage(pdf, doc, id)
		doc.pages = append(doc.pages, page)
	}
}

// 加载 page的 cotent和 resource
func loadPage(pdf *Pdf, doc *Doc, id int32) (page *Page) {
	pageObj := pdf.objMap[id]
	page = newPage()
	// size
	box := pageObj.valueOf("MediaBox").(Array)
	width := box[2].(int32)
	height := box[3].(int32)

	page.width = width
	page.height = height
	// resource /font
	res := pageObj.valueOf("Resources").(Dict)
	font := res["Font"].(Dict)
	for k, v := range font {
		fid := v.(ObjRef).id
		if doc.fonts[k] == nil {
			fobj := pdf.objMap[fid]
			font := loadFont(pdf, fobj)
			page.fonts[k] = font
			doc.fonts[k] = font
		} else {
			page.fonts[k] = doc.fonts[k]
		}

	}

	contentId := pageObj.getRefId("Contents")
	content := pdf.objMap[contentId].stream.stream
	resX := pageObj.valueOf("Resources")
	if dict, ok := resX.(Dict); ok {
		log(dict)
	}
	//parsePageResource()
	parsePageContent(page, content)

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
	log(string(stream))
	br := newBytesReader(stream)
	ops := make([]Operator, 0, 128)
	args := make([]DataType, 0, 4)
	for {
		tk := lexer(br)
		//		loge(tk.code, "   ", tk.buf, "   ", tk.n)
		if tk.is(TK_EOF) {
			break
		}
		if tk.is(TK_KEYWORD) {
			op := NewOp(tk.buf, args)
			ops = append(ops, op)
			args = make([]DataType, 0, 4)

		} else {
			args = append(args, tk.value())
		}
	}
	//	log(ops)
	bt := false
	font := ""
	for _, op := range ops {
		name := op.Name()
		if name == "BT" {
			bt = true
		}
		if bt {
			if name == "ET" {
				bt = false
			}
			if name == "Tf" {
				font = op.Args()[0].(string)
			}
			if name == "Tj" {
				str := op.Args()[0].(string)
				f := page.fonts[Name(font)]
				if f != nil && f.toUnicode != nil {
					unicode, _ := f.toUnicode.Lookup2([]byte(str))
					loge(" str:", unicodeToStr(unicode))
				}
				//				loge("draw ", str, " with font ", font)
			}

		}
	}
}
