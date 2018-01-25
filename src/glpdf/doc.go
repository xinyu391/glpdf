package glpdf

type Doc struct {
	pages []*Page
	count int32
}
type Page struct {
	content string
	fonts   []*Font
}
type Font struct {
	name      string
	toUnicode *Cmap
}

func parseDoc(pdf *Pdf) {
	root := pdf.objMap[pdf.root]
	pagesid := root.getRefId("Pages")

	doc := new(Doc)
	pages := pdf.objMap[pagesid]
	pageAry := pages.data.(Dict)["Kids"].(Array)
	doc.count = pages.valueOf("Count").(int32)
	for i, p := range pageAry {
		log("parse ", i+1, "page")
		id := p.(ObjRef).id
		pageObj := pdf.objMap[id]
		contentId := pageObj.getRefId("Contents")
		content := pdf.objMap[contentId].stream.stream
		parsePageContent(content)
		page := new(Page)
		doc.pages = append(doc.pages, page)
	}
}
func (obj *PdfObj) valueOf(key Name) DataType {
	return obj.data.(Dict)[key]
}
func (obj *PdfObj) getRefId(key Name) (id int32) {
	return obj.data.(Dict)[key].(ObjRef).id

}

func parsePageContent(stream []byte) {
	log(string(stream))
	br := newBytesReader(stream)
	for {
		tk := lexer(br)
		loge(tk.code, "   ", tk.buf, "   ", tk.n)
		if tk.is(TK_EOF) {
			break
		}
	}
}
