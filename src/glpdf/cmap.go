package glpdf

import (
	"errors"
	"fmt"
	"os"
)

type CmapManager struct {
	list map[string]*Cmap
}
type Cmap struct {
	name         string
	wmode        int
	usecmap_name string
	codespace    []codeSpace
	cmap         []cid2uinc
}
type codeSpace struct {
	n    int
	low  uint
	high uint
}
type cid2uinc struct {
	low  uint
	high uint
	unic int32 //rune
}

func NewCmap() (cmap *Cmap) {
	cmap = new(Cmap)
	cmap.codespace = make([]codeSpace, 0, 40)
	cmap.cmap = make([]cid2uinc, 0, 128)
	return
}

func (cmap *Cmap) mapOne(src uint32, dst int32) {
	//	cmap.cmap = append(cmap.cmap, cid2uinc{src, src, dst})
}
func (cmap *Cmap) mapOneToMany(lo, hi uint, n int) {
}
func (cmap *Cmap) addRange(lo, hi uint, out int32, n int) {
	//TODO add_range
	cmap.cmap = append(cmap.cmap, cid2uinc{lo, hi, out})
}
func (cmap *Cmap) addCodespace(lo, hi uint, n int) {
	cmap.codespace = append(cmap.codespace, codeSpace{n, lo, hi})
}
func (cmap *Cmap) Lookup2(cid []byte) (unicode int32, err error) {
	code := codeFromString(cid)
	log("code", code)
	return cmap.lookup(code)
}
func (cmap *Cmap) lookup(cid uint) (unicode int32, err error) {
	size := len(cmap.cmap)
	//TODO 简单实现，顺序搜索
	for n := 0; n < size; n++ {
		if cid >= cmap.cmap[n].low && cid <= cmap.cmap[n].high {
			if cid == cmap.cmap[n].low {
				unicode = cmap.cmap[n].unic

			} else {
				unicode = cmap.cmap[n].unic + int32(cid-cmap.cmap[n].low)
			}
			return
		}
	}

	err = errors.New(fmt.Sprint("Not found ", cid, " in ", cmap.name))
	return
}

func (cmap *CmapManager) load(data []byte) (err error) {
	return
}
func (cmap *CmapManager) lookup(cid []byte, name string) (unicode []byte, err error) {
	return
}

func LoadCmapBytes(data []byte) (cmap *Cmap, err error) {

	br := newBytesReader(data)
	return loadCmap(br)
}
func loadCmapFile(path string) (cmap *Cmap, err error) {

	f, err := os.Open(path)
	fr := newFileReader(f)
	return loadCmap(fr)
}
func loadCmap(fr RandomReader) (cmap *Cmap, err error) {
	cmap = NewCmap()
	key := ""
	for {
		tk, str, _, _ := peek(fr)
		if tk == TK_EOF {
			break
		}
		log(str)
		if tk == TK_NAME {
			switch str {
			case "CMapName":
				tk, str, _, _ = peek(fr)
				if tk == TK_NAME {
					cmap.name = str
				} else {
					logw("expected Name after CMapName in cmap")
				}
			case "WMode":
				tk, _, n, _ := peek(fr)
				if tk == TK_INT {
					cmap.wmode = int(n)
				} else {
					logw("expected number after CMapName in cmap")
				}
			default:
				key = str
			}
		} else if tk == TK_KEYWORD {
			switch str {
			case "endcmp":
			case "usecmap":
				cmap.usecmap_name = key
			case "begincodespacerange":
				parseCodespaceRange(fr, cmap)
			case "beginbfchar":
				parseBfChar(fr, cmap)
			case "begincidchar":
				parseCidChar(fr, cmap)
			case "beginbfrange":
				parseBfRange(fr, cmap)
			case "begincidrange":
				parseCidRange(fr, cmap)
			default:
			}
		}
		// ignore others

	}
	// sort
	log(cmap)
	return
}
func parseCidChar(fr RandomReader, cmap *Cmap) (err error) {
	for {
		tk := lexer(fr)
		if tk.is(TK_EOF) { // tk.code ==TK_EOF
			break
		} else if tk.isKeyword("endcidchar") {
			break
		}
		if src, ok := tk.str(); ok {

			tk = lexer(fr)
			if dst, ok := tk.num(); ok {
				//				log("cid char", hexStrToInt(src), dst)
				cmap.mapOne(codeFromHex(src), dst)
			}
		}
	}
	return
}
func parseBfRange(fr RandomReader, cmap *Cmap) (err error) {
	return
}
func parseCidRange(fr RandomReader, cmap *Cmap) (err error) {
	return
}
func parseBfChar(fr RandomReader, cmap *Cmap) (err error) {
	for {
		tk := lexer(fr)
		if tk.is(TK_EOF) {
			break
		}
		if tk.isKeyword("endbfchar") {
			return
		}
		if src, ok := tk.str(); ok {
			tk = lexer(fr)
			if dst, ok := tk.str(); ok {
				cmap.mapOne(codeFromHex(src), int32(codeFromString([]byte(dst))))
			} else {
				err = errors.New("")
			}
		} else {
			err = errors.New("")
		}
	}
	return
}
func parseBfChar2(fr RandomReader, cmap *Cmap) (err error) {
	for {
		tk, str, _, _ := peek(fr)
		if tk == TK_KEYWORD && str == "endbfchar" {
			return
		} else if tk == TK_STRING {
			tk2, str2, _, _ := peek(fr)
			if tk2 == TK_STRING {

				src := codeFromString([]byte(str))
				var dst uint = 0
				n := 1
				if len([]byte(str2)) == 2 {
					dst = codeFromString([]byte(str2))
				} else {
					dst = codeFromString2([]byte(str2))
					n = 2
				}
				log("parseBfChar  ", src, dst, n)
				cmap.addRange(src, src, int32(dst), n)
			} else {
				break
			}
		} else {
			break
		}

	}
	return
}
func parseCodespaceRange(fr RandomReader, cmap *Cmap) (err error) {

	for {
		tk, str, _, _ := peek(fr)
		if tk == TK_KEYWORD && str == "endcodespacerange" {
			return
		} else if tk == TK_STRING {
			tk2, str2, _, _ := peek(fr)
			if tk2 == TK_STRING {
				lo := codeFromString([]byte(str))
				hi := codeFromString([]byte(str2))
				cmap.addCodespace(lo, hi, 0)
				logw("parseCodespaceRange ", str, str2, lo, hi)
			} else {
				break
			}
		} else {
			break
		}

	}
	logw("expected endcodespacerange ")
	return
}
