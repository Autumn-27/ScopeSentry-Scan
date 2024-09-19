// Package util -----------------------------
// @file      : util.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/11 10:13
// -------------------------------------------
package util

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"io/ioutil"
	"math/rand"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ReadFileLiness(fileName string) ([]string, error) {
	result := []string{}
	f, err := os.Open(fileName)
	if err != nil {
		return result, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

func WriteStructArrayToFile(filename string, Result []types.AssetHttp) error {
	// 将结构体数组编码为 JSON 格式
	jsonData, err := json.MarshalIndent(Result, "", "  ")
	if err != nil {
		return err
	}

	// 将 JSON 数据写入文件
	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("结构体数组写入文件 %s 成功\n", filename)
	return nil
}
func GetFileExtension(url string) string {
	// 从URL中提取文件路径
	filePath := filepath.Base(url)

	// 获取文件的后缀名
	fileExtension := filepath.Ext(filePath)

	// 去除后缀名前的点号
	fileExtension = strings.TrimPrefix(fileExtension, ".")

	return fileExtension
}

func GenerateRandomString(length int) string {
	// 定义字符集
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 构建随机字符串
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func WriteContentFile(filPath string, fileContent string) bool {
	// 将字符串写入文件
	err := ioutil.WriteFile(filPath, []byte(fileContent), 0666)
	if err != nil {
		fmt.Printf("Failed to create filPath: %s - %s", filPath, err)
		return false
	}
	return true

}

func CalculateMD5(input string) string {
	// Convert the input string to bytes
	data := []byte(input)

	// Calculate the MD5 hash
	hash := md5.Sum(data)

	// Convert the hash to a hex string
	hashString := hex.EncodeToString(hash[:])

	return hashString
}

func DeleteFile(filePath string) {
	// 调用Remove函数删除文件
	err := os.Remove(filePath)
	if err != nil {
		fmt.Printf("Failed to DeleteFile: %s - %s", filePath, err)
	}
}
func GetSystemUsage() (int, float64) {
	// 获取CPU使用率
	percent, err := cpu.Percent(3*time.Second, false)
	if err != nil {
		fmt.Println("Failed to get CPU usage:", err)
		return 0, 0
	}
	cpuNum := 0
	if len(percent) > 0 {
		cpuNum = int(percent[0])
	}
	// 获取内存使用率
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Failed to get memory usage:", err)
		return 0, 0
	}
	return cpuNum, memInfo.UsedPercent
}

func GetRootDomain(input string) (string, error) {
	input = strings.TrimLeft(input, "http://")
	input = strings.TrimLeft(input, "https://")
	input = strings.TrimLeft(input, "//")
	input = strings.TrimLeft(input, "/")
	ip := net.ParseIP(input)
	if ip != nil {
		return input, nil
	}
	input = "https://" + input

	// 尝试解析为 URL
	u, err := url.Parse(input)
	if err == nil && u.Hostname() != "" {
		hostParts := strings.Split(u.Hostname(), ".")
		if len(hostParts) < 2 {
			return "", fmt.Errorf("域名格式不正确")
		}
		// 检查是否为复合域名
		if _, ok := compoundDomains[hostParts[len(hostParts)-2]+"."+hostParts[len(hostParts)-1]]; ok {
			return hostParts[len(hostParts)-3] + "." + hostParts[len(hostParts)-2] + "." + hostParts[len(hostParts)-1], nil
		}

		// 如果域名以 www 开头，特殊处理
		if hostParts[0] == "www" {
			return hostParts[len(hostParts)-2] + "." + hostParts[len(hostParts)-1], nil
		}

		return hostParts[len(hostParts)-2] + "." + hostParts[len(hostParts)-1], nil
	}
	return input, fmt.Errorf("输入既不是有效的 URL，也不是有效的 IP 地址")
}

var compoundDomains = map[string]bool{
	"ac.uk":  true,
	"co.uk":  true,
	"gov.uk": true,
	"ltd.uk": true,
	"me.uk":  true,
	"net.au": true,
	"org.au": true,
	"com.au": true,
	"edu.au": true,
	"gov.au": true,
	"asn.au": true,
	"id.au":  true,
	"com.cn": true,
	"net.cn": true,
	"org.cn": true,
	"gov.cn": true,
	"edu.cn": true,
	"mil.cn": true,
	"ac.cn":  true,
	"ah.cn":  true,
	"bj.cn":  true,
	"cq.cn":  true,
	"fj.cn":  true,
	"gd.cn":  true,
	"gs.cn":  true,
	"gx.cn":  true,
	"gz.cn":  true,
	"ha.cn":  true,
	"hb.cn":  true,
	"he.cn":  true,
	"hi.cn":  true,
	"hl.cn":  true,
	"hn.cn":  true,
	"jl.cn":  true,
	"js.cn":  true,
	"jx.cn":  true,
	"ln.cn":  true,
	"nm.cn":  true,
	"nx.cn":  true,
	"qh.cn":  true,
	"sc.cn":  true,
	"sd.cn":  true,
	"sh.cn":  true,
	"sn.cn":  true,
	"sx.cn":  true,
	"tj.cn":  true,
	"xj.cn":  true,
	"xz.cn":  true,
	"yn.cn":  true,
	"zj.cn":  true,
}
