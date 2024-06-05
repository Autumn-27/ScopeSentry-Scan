// Package main -----------------------------
// @file      : updateConfig.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/1/10 19:49
// -------------------------------------------
package main

import (
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/mongdbClient"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
)

func main() {
	system.SetUp()

	mongoClient, err := mongdbClient.Connect(system.AppConfig.Mongodb.Username, system.AppConfig.Mongodb.Password, system.AppConfig.Mongodb.IP, system.AppConfig.Mongodb.Port)
	if err != nil {
	}
	system.UpdateSubfinderApiConfig(mongoClient)
	system.UpdateDomainDicConfig(mongoClient)
}
