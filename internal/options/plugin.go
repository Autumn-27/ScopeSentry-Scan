// options-------------------------------------
// @file      : plugin.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/17 19:17
// -------------------------------------------

package options

type PluginOption struct {
	Name      string
	Module    string
	Parameter string
	PluginId  string
	Result    chan interface{}
	Custom    interface{}
	TaskId    string
	TaskName  string
	Log       func(msg string, tp ...string)
}
