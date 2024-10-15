// main-------------------------------------
// @file      : testUtils.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/24 20:15
// -------------------------------------------

package main

import (
	"fmt"
	"time"
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
	//utils.InitializeDnsTools()
	//a := utils.DNS.QueryOne("dwas.dwadwasdwa")
	//fmt.Println(a)
	// 获取当前时间戳（秒）
	now := time.Now().Unix()
	fmt.Println(now)
	// 定义2000年1月1日的时间
	referenceTime := time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC)

	// 获取2000年1月1日的Unix时间戳（秒）
	referenceTimestamp := referenceTime.Unix()
	fmt.Println(referenceTimestamp)
	// 计算从2000年到现在的时间差（秒）
	secondsSince2000 := now - referenceTimestamp
	fmt.Println(secondsSince2000)
	base62Timestamp := ToBase62(secondsSince2000)
	fmt.Println(base62Timestamp)
}
