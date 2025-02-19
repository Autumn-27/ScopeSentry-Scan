// main-------------------------------------
// @file      : testGetPdfContent.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/2/19 22:16
// -------------------------------------------

package main

import (
	"fmt"
)

func main() {
	c := GetPdfContent("C:\\Users\\autumn\\Desktop\\")
	fmt.Println(c)
}

func GetPdfContent(filePath string) string {
	return ""
	//f, r, err := pdf.Open(filePath)
	//// remember close file
	//defer f.Close()
	//if err != nil {
	//	logger.SlogWarn(fmt.Sprintf("GetPdfContent error: %v", err))
	//	return ""
	//}
	//var buf bytes.Buffer
	//b, err := r.GetPlainText()
	//if err != nil {
	//	logger.SlogWarn(fmt.Sprintf("GetPdfContent GetPlainText error: %v", err))
	//	return ""
	//}
	//buf.ReadFrom(b)
	//
	//return buf.String()
}
