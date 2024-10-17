// main-------------------------------------
// @file      : test_yaegi.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/17 19:04
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/symbols"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"os"
	"path/filepath"
)

const src = `package foo

import "github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"

func Bar(s string) string {
	logger.SlogInfoLocal("system config load begin")
	return s + "-Foo"
}
`

func main() {
	logger.NewLogger()
	// 获取可执行文件的目录
	execPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	execDir := filepath.Dir(execPath)
	fmt.Printf("Executable directory: %s\n", execDir)

	// 设置 GOPATH 环境变量
	goPath := execDir

	// 初始化 yaegi 解释器
	interp := interp.New(interp.Options{
		GoPath: goPath, // 设置 GoPath 为环境变量中的 GOPATH
	})

	// 加载标准库和符号
	interp.Use(stdlib.Symbols)
	interp.Use(symbols.Symbols)

	// 加载插件
	pluginPath := filepath.Join(execDir, "plugins", "demo.go")
	fmt.Printf("Loading plugin from: %s\n", pluginPath) // 打印插件路径以确认
	_, err = interp.EvalPath(pluginPath)
	if err != nil {
		panic(err)
	}

	// 获取 foo.Bar 函数
	v, err := interp.Eval("foo.Bar")
	if err != nil {
		panic(err)
	}

	// 将值转换为函数
	bar := v.Interface().(func(interface{}) string)

	// 调用 Bar 函数
	r := bar("ddddd")
	fmt.Println(r) // 输出: Kung-Foo
}
