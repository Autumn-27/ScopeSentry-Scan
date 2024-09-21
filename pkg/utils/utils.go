// utils-------------------------------------
// @file      : utils.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 22:34
// -------------------------------------------

package utils

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type UtilTools struct{}

var Tools *UtilTools

func InitializeTools() {
	Tools = &UtilTools{}
}

// ReadYAMLFile 读取 YAML 文件并将其解析为目标结构体
func (t *UtilTools) ReadYAMLFile(filePath string, target interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(byteValue, target)
	if err != nil {
		return err
	}

	return nil
}

// WriteYAMLFile 将目标结构体序列化为 YAML 并写入到文件
func (t *UtilTools) WriteYAMLFile(filePath string, data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, yamlData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// GenerateRandomString 生产指定长度的随机字符串
func (t *UtilTools) GenerateRandomString(length int) string {
	// 定义字符集
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 构建随机字符串
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// GetSystemUsage 获取系统使用率
func (t *UtilTools) GetSystemUsage() (int, float64) {
	// 获取CPU使用率
	percent, err := cpu.Percent(3*time.Second, false)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("Failed to get CPU usage:", err))
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
		logger.SlogErrorLocal(fmt.Sprintf("Failed to get memory usage:", err))
		return 0, 0
	}
	return cpuNum, memInfo.UsedPercent
}

// WriteContentFile 将字符串写入指定文件
func (t *UtilTools) WriteContentFile(filPath string, fileContent string) error {
	// 将字符串写入文件
	return t.WriteByteContentFile(filPath, []byte(fileContent))
}

// WriteByteContentFile 将byte写入指定文件
func (t *UtilTools) WriteByteContentFile(filPath string, fileContent []byte) error {
	// 将字符串写入文件
	if err := ioutil.WriteFile(filPath, fileContent, 0666); err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("Failed to create filPath: %s - %s", filPath, err))
		return err
	}
	return nil
}

// MarshalYAMLToString 将目标结构体序列化为 YAML 字符串
func (t *UtilTools) MarshalYAMLToString(data interface{}) (string, error) {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}

// StructToJSON 将结构体序列化为json
func (t *UtilTools) StructToJSON(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// JSONToStruct 将 JSON 字符串反序列化为结构体
func (t *UtilTools) JSONToStruct(jsonStr []byte, result interface{}) error {
	return json.Unmarshal(jsonStr, result)
}

// ParseArgs 从字符串解析为参数
// 参数：
// - args: 输入的参数字符串，需要解析的参数，例如 "--name John --age 30"。
// - keys: 需要解析的键名列表，表示哪些键值对会被解析，比如传入 "name" 和 "age"。
//
// 返回值：
// - map[string]string: 返回一个包含键值对的 map，键是 keys 中指定的参数，值是解析出来的字符串。
// - error: 如果解析过程中遇到错误，则返回一个 error。
func (t *UtilTools) ParseArgs(args string, keys ...string) (map[string]string, error) {
	// 将参数字符串分割为切片
	argsSlice := strings.Fields(args)

	// 创建一个 FlagSet 对象来解析参数
	fs := flag.NewFlagSet("ParseArgs", flag.ContinueOnError)

	// 创建一个 map 用于存储 flag 的值
	values := make(map[string]*string)
	for _, key := range keys {
		// 初始化 map 并创建对应的 flag
		value := ""
		values[key] = &value
		fs.String(key, "", "a placeholder") // 创建一个 flag 用于存储 key 的值
	}

	// 解析参数
	err := fs.Parse(argsSlice)
	if err != nil {
		return nil, err
	}

	// 获取 key 对应的值并填充到结果 map 中
	result := make(map[string]string)
	for _, key := range keys {
		if valuePtr, ok := values[key]; ok {
			result[key] = *valuePtr
		} else {
			result[key] = ""
		}
	}

	return result, nil
}

// DeleteFile 删除指定文件
func (t *UtilTools) DeleteFile(filePath string) {
	// 调用Remove函数删除文件
	err := os.Remove(filePath)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("Failed to DeleteFile: %s - %s", filePath, err))
	}
}

// GetParameter 获取指定模块指定插件的参数
func (t *UtilTools) GetParameter(Parameters map[string]map[string]interface{}, module string, plugin string) (string, bool) {
	// 查找 module 是否存在
	if plugins, modOk := Parameters[module]; modOk {
		// 查找 plugin 是否存在
		if param, plugOk := plugins[plugin]; plugOk {
			return param.(string), true
		}
	}
	// 没有找到对应的参数，返回 false
	return "", false
}

// GetRootDomain 获取域名的根域名
func (t *UtilTools) GetRootDomain(input string) (string, error) {
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

// HttpGetDownloadFile 通过http下载文件到指定路径
func (t *UtilTools) HttpGetDownloadFile(url, filePath string) (bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("http.Get failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("ioutil.ReadAll failed: %w", err)
	}

	// 获取文件所在的目录路径
	dir := filepath.Dir(filePath)
	// 创建文件所在的目录（如果目录不存在）
	err = os.MkdirAll(dir, 0777) // MkdirAll 会递归创建目录
	if err != nil {
		return false, fmt.Errorf("os.MkdirAll failed: %w", err)
	}
	// 将文件内容写入本地文件
	err = ioutil.WriteFile(filePath, body, 0777)
	if err != nil {
		return false, fmt.Errorf("ioutil.WriteFile failed: %w", err)
	}

	return true, nil
}

func (t *UtilTools) ExecuteCommandWithTimeout(command string, args []string, timeout time.Duration) error {
	// 创建一个带有超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // 确保在函数结束后取消上下文，防止资源泄漏

	// 创建命令对象，使用带上下文的 exec.CommandContext
	cmd := exec.CommandContext(ctx, command, args...)

	// 执行命令，不获取输出
	err := cmd.Run()
	if err != nil {
		// 如果是超时错误，返回具体错误信息
		if ctx.Err() == context.DeadlineExceeded {
			return ctx.Err()
		}
		return err
	}

	// 如果没有错误，说明命令执行成功
	return nil
}

// ExecuteCommandToChan 执行指定命令，命令的输出每一行会发送到result的通道中。
func (t *UtilTools) ExecuteCommandToChan(cmdName string, args []string, result chan<- string) {
	// 创建命令
	cmd := exec.Command(cmdName, args...)

	// 获取命令输出管道（标准输出）
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		result <- fmt.Sprintf("Error getting stdout pipe: %v", err)
		close(result)
		return
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		result <- fmt.Sprintf("Error starting command: %v", err)
		close(result)
		return
	}

	// 使用 bufio 读取命令的标准输出
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		// 将每行输出发送到 result 通道
		result <- scanner.Text()
	}

	// 等待命令执行完毕
	if err := cmd.Wait(); err != nil {
		result <- fmt.Sprintf("Error waiting for command: %v", err)
	}

	// 关闭 result 通道，表示数据发送完毕
	close(result)
}

// CalculateMD5 计算字符串的md5值
func (t *UtilTools) CalculateMD5(input string) string {
	// Convert the input string to bytes
	data := []byte(input)

	// Calculate the MD5 hash
	hash := md5.Sum(data)

	// Convert the hash to a hex string
	hashString := hex.EncodeToString(hash[:])

	return hashString
}

// WriteLinesToFile 将字符串数组的每一行写入指定文件
func (t *UtilTools) WriteLinesToFile(filePath string, lines *[]string) error {

	// 获取文件所在的目录路径
	dir := filepath.Dir(filePath)
	// 创建文件所在的目录（如果目录不存在）
	err := os.MkdirAll(dir, 0777) // MkdirAll 会递归创建目录
	if err != nil {
		return fmt.Errorf("os.MkdirAll failed: %w", err)
	}

	// 打开文件，如果文件不存在则创建
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("无法打开文件: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	// 创建一个缓冲写入器
	writer := bufio.NewWriter(file)

	// 将数组中的每一行写入文件
	for _, line := range *lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("写入文件时出错: %v", err)
		}
	}

	// 刷新缓冲区，确保所有数据写入文件
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("刷新缓冲区时出错: %v", err)
	}

	return nil
}

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

func (t *UtilTools) GetTimeNow() string {
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

// EnsureDir 判断目录是否存在，不存在则创建
func (t *UtilTools) EnsureDir(dirPath string) error {
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
}

// ReadFileLineByLine 函数逐行读取文件，并将每一行发送到通道中
func (t *UtilTools) ReadFileLineByLine(filePath string, lineChan chan<- string) error {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close() // 函数结束时关闭文件

	// 使用 bufio.Scanner 逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineChan <- scanner.Text() // 将读取到的行发送到通道
	}

	// 检查读取过程中是否发生错误
	if err := scanner.Err(); err != nil {
		return err
	}

	close(lineChan) // 读取完毕后关闭通道
	return nil
}
