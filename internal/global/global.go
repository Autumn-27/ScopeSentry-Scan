// config-------------------------------------
// @file      : global.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/8 12:33
// -------------------------------------------

package global

import (
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
)

var (
	// AbsolutePath 全局变量
	AbsolutePath string
	ConfigPath   string
	ConfigDir    string
	// AppConfig Global variable to hold the loaded configuration
	AppConfig          Config
	VERSION            string
	FirstRun           bool
	DictPath           string
	ExtDir             string
	SensitiveRules     []types.SensitiveRule
	Projects           []types.Project
	WebFingers         []types.WebFinger
	NotificationApi    []types.NotificationApi
	NotificationConfig types.NotificationConfig
	PocDir             string
)
