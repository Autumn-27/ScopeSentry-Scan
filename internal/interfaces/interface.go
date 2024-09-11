// modules-------------------------------------
// @file      : interface.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/10 19:38
// -------------------------------------------

package interfaces

type ModuleRunner interface {
	ModuleRun() error
	SetInput(chan interface{})
	GetName() string
}
