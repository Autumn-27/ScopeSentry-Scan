// config-------------------------------------
// @file      : global.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/8 12:33
// -------------------------------------------

package config

var (
	// AbsolutePath 全局变量
	AbsolutePath string
	ConfigPath   string
	ConfigDir    string
	// AppConfig Global variable to hold the loaded configuration
	AppConfig Config
	VERSION   string
	FirstRun  bool
)
