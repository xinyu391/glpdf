package glpdf

import (
	"fmt"
)

func log(a ...interface{}) {
	fmt.Println(a...)
}

func loge(a ...interface{}) {
	b := append([]interface{}{}, "[E]")
	b = append(b, a...)
	fmt.Println(b...)
}
func logd(a ...interface{}) {
	//	b := append([]interface{}{}, "[D]")
	//	b = append(b, a...)
	//	fmt.Println(b...)
}
func logw(a ...interface{}) {
	b := append([]interface{}{}, "[W]")
	b = append(b, a...)
	fmt.Println(b...)
}

func logl(a ...interface{}) {
	fmt.Print(a...)
}
