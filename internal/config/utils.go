// config-------------------------------------
// @file      : utils.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/7 19:30
// -------------------------------------------

package config

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"os"
	"time"
)

var timeZoneOffsets = map[string]int{
	"UTC":                 0,
	"Asia/Shanghai":       8 * 60 * 60,
	"Asia/Tokyo":          9 * 60 * 60,
	"Asia/Kolkata":        5*60*60 + 30*60,
	"Europe/London":       0,
	"Europe/Berlin":       1 * 60 * 60,
	"Europe/Paris":        1 * 60 * 60,
	"America/New_York":    -5 * 60 * 60,
	"America/Chicago":     -6 * 60 * 60,
	"America/Denver":      -7 * 60 * 60,
	"America/Los_Angeles": -8 * 60 * 60,
	"Australia/Sydney":    10 * 60 * 60,
	"Australia/Perth":     8 * 60 * 60,
	"Asia/Singapore":      8 * 60 * 60,
	"Asia/Hong_Kong":      8 * 60 * 60,
	"Europe/Moscow":       3 * 60 * 60,
	"America/Sao_Paulo":   -3 * 60 * 60,
	"Africa/Johannesburg": 2 * 60 * 60,
	"Asia/Dubai":          4 * 60 * 60,
	"Pacific/Auckland":    12 * 60 * 60,
}

func GetTimeNow() string {
	// 获取当前时间
	timeZoneName := global.AppConfig.TimeZoneName

	var location *time.Location
	var err error

	// 查找时区名称对应的偏移量
	offset, exists := timeZoneOffsets[timeZoneName]
	if exists {
		// 如果存在映射，使用固定时区
		location = time.FixedZone(timeZoneName, offset)
	} else {
		// 如果映射不存在，尝试直接加载时区名称
		location, err = time.LoadLocation(timeZoneName)
		if err != nil {
			// 如果加载失败，使用系统默认时区
			fmt.Printf("Time zone not found: %s, using system default time zone\n", timeZoneName)
			location = time.Local
		}
	}
	currentTime := time.Now()
	var easternTime = currentTime.In(location)
	return easternTime.Format("2006-01-02 15:04:05")
}

func EnsureDir(dirPath string) error {
	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// 如果目录不存在，则创建目录
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
		return nil
	} else {
		return nil
	}
	return nil
}
