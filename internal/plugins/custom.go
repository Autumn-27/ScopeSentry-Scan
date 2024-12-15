// plugins-------------------------------------
// @file      : custom.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/11/15 21:10
// -------------------------------------------

package plugins

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/options"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/symbols"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/customplugin"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"os"
	"path/filepath"
	"strings"
)

func GetCustomPlugin() ([]interfaces.Plugin, error) {
	var plugins []interfaces.Plugin

	// 使用 WalkDir 遍历 global.PluginDir 目录
	err := filepath.WalkDir(global.PluginDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Error walking path", path, ":", err))
		}

		// 如果是目录，则获取其名字并遍历该子目录
		if d.IsDir() {
		} else {
			dir := filepath.Dir(path)
			moduleName := filepath.Base(dir)

			filename := filepath.Base(path)
			extension := filepath.Ext(filename)              // 获取文件的扩展名
			plgId := strings.TrimSuffix(filename, extension) // 去掉扩展名
			plugin, err := LoadCustomPlugin(path, moduleName, plgId)
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("module %v plg load error: %v", moduleName, plgId, err))
			}
			plugins = append(plugins, plugin)
		}
		return nil
	})

	if err != nil {
		// 遍历过程中如果有错误，返回错误
		return nil, fmt.Errorf("error walking the path %v: %v", global.PluginDir, err)
	}

	return plugins, nil
}

func LoadCustomPlugin(path string, modlue string, plgId string) (interfaces.Plugin, error) {
	logger.SlogInfoLocal(fmt.Sprintf("Load custom plugin: %v", path))
	// 初始化 yaegi 解释器
	interp := interp.New(interp.Options{})
	// 加载标准库和符号
	err := interp.Use(stdlib.Symbols)
	if err != nil {
		return nil, err
	}
	err = interp.Use(symbols.Symbols)
	if err != nil {
		return nil, err
	}
	_, err = interp.EvalPath(path)
	if err != nil {
		return nil, err
	}
	// 获取Execute
	v, err := interp.Eval("plugin.Execute")
	if err != nil {
		return nil, err
	}
	executeFunc := v.Interface().(func(input interface{}, op options.PluginOption) (interface{}, error))

	v, err = interp.Eval("plugin.GetName")
	if err != nil {
		return nil, err
	}
	getNameFunc := v.Interface().(func() string)

	v, err = interp.Eval("plugin.Install")
	if err != nil {
		return nil, err
	}
	installFunc := v.Interface().(func() error)

	v, err = interp.Eval("plugin.Check")
	if err != nil {
		return nil, err
	}
	checkFunc := v.Interface().(func() error)

	v, err = interp.Eval("plugin.Uninstall")
	if err != nil {
		return nil, err
	}
	uninstallFunc := v.Interface().(func() error)
	plg := customplugin.NewPlugin(modlue, plgId, installFunc, checkFunc, executeFunc, uninstallFunc, getNameFunc)
	return plg, nil
}
