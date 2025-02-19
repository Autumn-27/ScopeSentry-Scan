// main-------------------------------------
// @file      : testGetPdfContent.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/2/19 22:16
// -------------------------------------------

package main

import (
	"fmt"
	"rsc.io/pdf"
)

func main() {
	c := GetPdfContent("C:\\Users\\autumn\\Desktop\\PHPOK 5.3 Getshell.pdf")
	fmt.Println(c)
}

func GetPdfContent(filePath string) string {
	f, err := pdf.Open(filePath)
	if err != nil {
	}

	// 获取 PDF 页数
	numPages := f.NumPage()

	for i := 1; i <= numPages; i++ {
		page := f.Page(i)

		fmt.Println(page.Content().Text)
	}

	return ""
}
