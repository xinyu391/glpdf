package glpdf

import (
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
}
type codeSpace struct {
	n    int
	low  uint
	high uint
}

func NewCmap() (cmap *Cmap) {
	cmap = new(Cmap)
	cmap.codespace = make([]codeSpace, 0, 40)
	return
}

func (cmap *CmapManager) load(data []byte) (err error) {
	return
}
func (cmap *CmapManager) lookup(cid []byte, name string) (unicode []byte, err error) {
	return
}

func loadCmap(path string) (cmap *Cmap, err error) {
	cmap = NewCmap()
	f, err := os.Open(path)
	fr := NewfileReader(f)
	key := ""
	for {
		tk, str, _, _ := peek(fr)
		if tk == TK_EOF {
			break
		}
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
			case "beginbfrange":
			case "begincidrange":
			default:
			}
		}
		// ignore others
		log(key)
	}
	// sort
	return
}
func parseBfChar(fr *fileReader, cmap *Cmap) (err error) {
	for {
		tk, str, _, _ := peek(fr)
		if tk == TK_KEYWORD && str == "endbfchar" {
			return
		} else if tk == TK_STRING {
			tk2, str2, _, _ := peek(fr)
			if tk2 == TK_STRING {
				lo := codeFromString(str)
				hi := codeFromString(str2)
				cmap.addCodespace(lo, hi, 0)
			} else {
				break
			}
		} else {
			break
		}

	}
	return
}
func parseCodespaceRange(fr *fileReader, cmap *Cmap) (err error) {

	for {
		tk, str, _, _ := peek(fr)
		if tk == TK_KEYWORD && str == "endcodespacerange" {
			return
		} else if tk == TK_STRING {
			tk2, _, _, _ := peek(fr)
			if tk2 == TK_STRING {
				//				lo := codeFromString(str)
				//				hi := codeFromString(str2)
				//				cmap.mapOneToMany(lo, hi, 0)
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
func (cmap *Cmap) mapOneToMany(lo, hi uint, n int) {
}
func (cmap *Cmap) addCodespace(lo, hi uint, n int) {
	cmap.codespace = append(cmap.codespace, codeSpace{n, lo, hi})
}
func codeFromString(str string) (code uint) {
	var a uint = 0
	bytes := []byte(str)
	for n := 0; n < len(str); n++ {
		a = (a << 8) | uint(bytes[n])
	}
	return a
}
