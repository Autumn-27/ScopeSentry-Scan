// main-------------------------------------
// @file      : testUtils.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/24 20:15
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
)

func ToBase62(num int64) string {
	chars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if num == 0 {
		return string(chars[0])
	}

	var result string
	for num > 0 {
		result = string(chars[num%62]) + result
		num /= 62
	}
	return result
}

func main() {
	config.Initialize()
	global.VERSION = "1.5"
	global.AppConfig.Debug = true
	// 初始化mongodb连接
	mongodb.Initialize()
	// 初始化redis连接
	redis.Initialize()
	// 初始化日志模块
	logger.NewLogger()
	utils.InitializeDnsTools()
	//_, b, _ := utils.Tools.GenerateIgnore("*.baidu.com\nrainy-sautyu.top")
	//for _, k := range b {
	//	flag := k.MatchString("baidu.com")
	//	fmt.Println(flag)
	//}
	similarity, err := utils.Tools.CompareContentSimilarity("adddw", "ddddddddddddddddddddddddwww")
	if err != nil {
		return
	}
	fmt.Println(similarity)
	//a := utils.DNS.QueryOne("dwas.dwadwasdwa")
	//fmt.Println(a)
	// 获取当前时间戳（秒）
	//now := time.Now().Unix()
	//fmt.Println(now)
	//// 定义2000年1月1日的时间
	//referenceTime := time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC)
	//
	//// 获取2000年1月1日的Unix时间戳（秒）
	//referenceTimestamp := referenceTime.Unix()
	//fmt.Println(referenceTimestamp)
	//// 计算从2000年到现在的时间差（秒）
	//secondsSince2000 := now - referenceTimestamp
	//fmt.Println(secondsSince2000)
	//base62Timestamp := ToBase62(secondsSince2000)
	//fmt.Println(base62Timestamp)
}
