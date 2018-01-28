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
	Name        string
	Wmode       int
	UsecmapName string
	Codespace   []codeSpace
	Cmap        []cid2uinc
}
type codeSpace struct {
	Low  uint32
	High uint32
	N    int
}
type cid2uinc struct {
	Low  uint32
	High uint32
	Unic int32 //rune
}

func NewCmap() (cmap *Cmap) {
	cmap = new(Cmap)
	cmap.Codespace = make([]codeSpace, 0, 40)
	cmap.Cmap = make([]cid2uinc, 0, 128)
	return
}

func (cmap *Cmap) mapOne(src uint32, dst uint32) {
	cmap.Cmap = append(cmap.Cmap, cid2uinc{src, src, int32(dst)})
}
func (cmap *Cmap) mapBfRange(low, high uint32, dst int32) {
	cmap.Cmap = append(cmap.Cmap, cid2uinc{low, high, dst})
}
func (cmap *Cmap) mapBfRangeToArray(low, high uint32, dst []int32) {
	j := 0
	size := len(dst)
	for low <= high && j < size {
		cmap.Cmap = append(cmap.Cmap, cid2uinc{low, high, dst[j]})
		low++
		j++
	}
}
func (cmap *Cmap) mapOneToMany(lo, hi uint, n int) {
}
func (cmap *Cmap) addRange(lo, hi uint32, out int32, n int) {
	//TODO add_range
	cmap.Cmap = append(cmap.Cmap, cid2uinc{lo, hi, out})
}
func (cmap *Cmap) addCodespace(lo, hi uint32, n int) {
	cmap.Codespace = append(cmap.Codespace, codeSpace{lo, hi, n})
}
func (cmap *Cmap) Lookup2(cid []byte) (unicode []int32, err error) {

	size := len(cid)
	unicode = make([]int32, 0, size/2)

	offset := 0
	for offset < size {
		find := false
		for _, space := range cmap.Codespace {
			if offset+space.N > size {
				break
			}
			v := bytesToInt(cid[offset : offset+space.N])
			if v >= space.Low && v <= space.High {
				// search
				offset += space.N
				unic, err := cmap.lookup(v)
				if err != nil {
					loge("can not find map in ", cmap.Name, " skip ", space.N, " byte to continue")
				}
				unicode = append(unicode, unic)
				find = true
				break
			}
		}
		if find == false {
			//?? 找不到，就跳过一个字节继续
			loge("can not find map in ", cmap.Name, " skip one byte to continue")
			offset += 1
			return
		}
	}
	return
}

//
func (cmap *Cmap) Lookup(cid []byte) (unicode int32, n int) {
	// 取1个字节还是2个字节
	for _, space := range cmap.Codespace {
		v := bytesToInt(cid[:space.N])
		if v >= space.Low && v <= space.High {
			// search
			unic, err := cmap.lookup(v)
			if err != nil {
				return 0, 0
			} else {
				return unic, space.N
			}
		}
	}
	return
}
func (cmap *Cmap) lookup(cid uint32) (unicode int32, err error) {
	size := len(cmap.Cmap)
	//TODO 简单实现，顺序搜索
	for n := 0; n < size; n++ {
		if cid >= cmap.Cmap[n].Low && cid <= cmap.Cmap[n].High {
			if cid == cmap.Cmap[n].Low {
				unicode = cmap.Cmap[n].Unic

			} else {
				unicode = cmap.Cmap[n].Unic + int32(cid-cmap.Cmap[n].Low)
			}
			return
		}
	}

	err = errors.New(fmt.Sprint("Not found ", cid, " in ", cmap.Name))
	return
}

func (cmap *CmapManager) init() (err error) {
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
		//		log(str)
		if tk == TK_NAME {
			switch str {
			case "CMapName":
				tk, str, _, _ = peek(fr)
				if tk == TK_NAME {
					cmap.Name = str
				} else {
					logw("expected Name after CMapName in cmap")
				}
			case "WMode":
				tk, _, n, _ := peek(fr)
				if tk == TK_INT {
					cmap.Wmode = int(n)
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
				cmap.UsecmapName = key
			case "begincodespacerange":
				parseCodespaceRange(fr, cmap)
			case "beginbfchar": // bfchar  charcode -> cid
				parseBfChar(fr, cmap)
			case "begincidchar": // cidchar cid->unicode
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
	//	log(cmap)
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
				cmap.mapOne(bytesToInt([]byte(src)), uint32(dst))
			}
		}
	}
	return
}
func parseCidRange(fr RandomReader, cmap *Cmap) (err error) {
	for {
		tk := lexer(fr)
		if tk.is(TK_EOF) {
			break
		}
		if tk.isKeyword("endcidrange") {
			return
		}
		src, _ := tk.str()
		low := bytesToInt([]byte(src))

		tk = lexer(fr)
		src, _ = tk.str()
		high := bytesToInt([]byte(src))

		tk = lexer(fr)
		switch tk.code {
		case TK_INT:
			dst, _ := tk.num()
			cmap.mapBfRange(low, high, dst)
		case TK_STRING:
			src, _ = tk.str()
			dst := bytesToInt([]byte(src))
			cmap.mapBfRange(low, high, int32(dst))
		case TK_BEGIN_ARRAY:
			ary, _ := parseArray(fr)
			dst := make([]int32, len(ary))
			for k, v := range ary {
				dst[k] = v.(int32)
			}
			cmap.mapBfRangeToArray(low, high, dst)
		}
	}
	return
}
func parseBfRange(fr RandomReader, cmap *Cmap) (err error) {
	for {
		tk := lexer(fr)
		if tk.is(TK_EOF) {
			break
		}
		if tk.isKeyword("endbfrange") {
			return
		}
		src, _ := tk.str()
		low := bytesToInt([]byte(src))

		tk = lexer(fr)
		src, _ = tk.str()
		high := bytesToInt([]byte(src))

		tk = lexer(fr)
		switch tk.code {
		case TK_INT:
			dst, _ := tk.num()
			cmap.mapBfRange(low, high, dst)
		case TK_STRING:
			src, _ = tk.str()
			dst := bytesToInt([]byte(src))
			cmap.mapBfRange(low, high, int32(dst))
		case TK_BEGIN_ARRAY:
			ary, _ := parseArray(fr)
			dst := make([]int32, len(ary))
			for k, v := range ary {
				dst[k] = v.(int32)
			}
			cmap.mapBfRangeToArray(low, high, dst)
		}
	}
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
				cmap.mapOne(bytesToInt([]byte(src)), bytesToInt([]byte(dst)))
			} else {
				err = errors.New("")
			}
		} else {
			err = errors.New("")
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
				lo := bytesToInt([]byte(str))
				hi := bytesToInt([]byte(str2))
				cmap.addCodespace(lo, hi, len(str))
				//				logw("parseCodespaceRange ", str, str2, lo, hi)
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
