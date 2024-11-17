// configupdater-------------------------------------
// @file      : type.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/8 12:15
// -------------------------------------------

package configupdater

type ConfigResult struct {
	Value string `bson:"value"`
}

type PluginInfo struct {
	Module string `bson:"module"`
	Hash   string `bson:"hash"`
	Source string `bson:"source"`
}
