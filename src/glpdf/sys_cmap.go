package glpdf

import (
	"encoding/gob"
	"os"
)

///
var extra_Adobe_Korea1_2 = []byte{9} ///
var extra_Adobe_Korea1_1 = []byte{9}

type extra_cmap struct {
	name     string
	cmap_bin []byte
}

var system_cmap_bin = []extra_cmap{{"Adobe-Korea1-2", extra_Adobe_Korea1_2}, {"Adobe-Korea1-1", extra_Adobe_Korea1_1}}

var system_cmap map[string]*Cmap

func loadSystemCmap() {
	log("loadSystemCmap...")
	if system_cmap != nil {
		return
	}
	system_cmap = make(map[string]*Cmap)
	// 反序列化
	cmap, err := loadCmapFile("/home/xinyu391/proj/mupdf-1.12.0-source/resources/cmaps/cjk/Adobe-Korea1-UCS2")
	if err != nil {
		return
	}
	//	log("cmap is ", cmap)
	system_cmap[cmap.Name] = cmap
	log("loadSystemCmap finish ")
}
func LoadSystemCmap() {
	cmap, err := loadCmapFile("/home/xinyu391/proj/mupdf-1.12.0-source/resources/cmaps/extra/Adobe-Korea1-2")
	if err != nil {
		return
	}
	log("cmap is ", cmap)
	f, err := os.Create("test.bin")

	w := gob.NewEncoder(f)
	err = w.Encode(cmap)
	if err != nil {
		log("encode ", err)
	}

	f.Close()
	var cp Cmap
	f2, err := os.Open("test.bin")
	if err == nil {
		r := gob.NewDecoder(f2)
		err = r.Decode(&cp)
		if err != nil {
			log("decode failed", err)
		}
		log("cmap are ", cp)
		f2.Close()
	}
}
