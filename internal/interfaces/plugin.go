// pool-------------------------------------
// @file      : interface.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:26
// -------------------------------------------

package interfaces

type Plugin interface {
	GetName() string
	SetName(name string)
	GetModule() string
	SetModule(name string)
	GetPluginId() string
	SetPluginId(id string)
	SetCustom(cu interface{})
	GetCustom() interface{}
	SetResult(ch chan interface{})
	SetParameter(args string)
	GetParameter() string
	SetTaskId(id string)
	GetTaskId() string
	SetTaskName(name string)
	GetTaskName() string
	Execute(input interface{}) (interface{}, error)
	Install() error
	Check() error
	Clone() Plugin
	Log(msg string, tp ...string)
}
