package main

import (
	"fmt"

	"glpdf"
)

func main() {
	fmt.Println("pdf.go", len("位"))
	pdf, err := glpdf.Open("sample.pdf")
	fmt.Println("pdf ", pdf, err)
	num := pdf.GetPageNum()
	page := pdf.GetPage(num - 1)
	page.Draw()

}
