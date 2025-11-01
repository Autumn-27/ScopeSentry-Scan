// utils-------------------------------------
// @file      : utils.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 22:34
// -------------------------------------------

package utils

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/cespare/xxhash/v2"
	"github.com/hbollon/go-edlib"
	"github.com/ledongthuc/pdf"
	"github.com/nfnt/resize"
	"github.com/projectdiscovery/cdncheck"
	"github.com/projectdiscovery/httpx/runner"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
	"gopkg.in/yaml.v3"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type UtilTools struct {
	CdnCheckClient *cdncheck.Client
	fileLocks      map[string]*sync.Mutex
	fileMu         sync.Mutex
}

var Tools *UtilTools

func InitializeTools() {
	client := cdncheck.New()
	Tools = &UtilTools{
		CdnCheckClient: client,
		fileLocks:      make(map[string]*sync.Mutex),
	}
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

// 将整数转换为62进制字符串
func (t *UtilTools) ToBase62(num int64) string {
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

// GenerateHash 生成唯一hash
func (t *UtilTools) GenerateHash() string {
	timestamp := time.Now().Unix()
	secondsSince2000 := timestamp - 1747958400
	base62Timestamp := t.ToBase62(secondsSince2000)

	randoms := t.GenerateRandomString(4)
	hash := base62Timestamp + randoms
	return hash
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

// WriteByteContentFile 将byte写入指定文件, 覆盖写入
func (t *UtilTools) WriteByteContentFile(filPath string, fileContent []byte) error {
	// 将字符串写入文件
	err := t.EnsureFilePathExists(filPath)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filPath, fileContent, 0666); err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("Failed to create filPath: %s - %s", filPath, err))
		return err
	}
	return nil
}

// WriteContentFileAppend 将字符串写入指定文件
func (t *UtilTools) WriteContentFileAppend(filPath string, fileContent string) error {
	// 将字符串写入文件
	return t.AppendOrCreateFile(filPath, []byte(fileContent))
}

// getFileLock 获取文件锁
func (t *UtilTools) GetFileLock(filePath string) *sync.Mutex {
	t.fileMu.Lock()
	defer t.fileMu.Unlock()

	if _, exists := t.fileLocks[filePath]; !exists {
		t.fileLocks[filePath] = &sync.Mutex{}
	}
	return t.fileLocks[filePath]
}

// ClearAllLocks 清空所有文件锁
func (t *UtilTools) ClearAllLocks() {
	t.fileMu.Lock()
	defer t.fileMu.Unlock()

	t.fileLocks = make(map[string]*sync.Mutex) // 创建新的空 map，旧的会被垃圾回收
}

// AppendOrCreateFile 追加写入文件，如果文件不存在则创建
func (t *UtilTools) AppendOrCreateFile(filePath string, fileContent []byte) error {
	lock := t.GetFileLock(filePath)
	lock.Lock()
	defer lock.Unlock()

	// 确保文件路径存在
	err := t.EnsureFilePathExists(filePath)
	if err != nil {
		return err
	}

	// 打开文件，支持追加写入或创建新文件
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("Failed to open or create file: %s - %s", filePath, err))
		return err
	}
	defer file.Close()

	// 写入内容
	if _, err := file.Write(fileContent); err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("Failed to write to file: %s - %s", filePath, err))
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

	// 遍历 keys，为每个 key 添加一个 flag
	for _, key := range keys {
		value := ""
		values[key] = &value
		fs.StringVar(values[key], key, "", "a placeholder for "+key)
	}

	// 通过重定向标准输出，避免错误信息输出
	originalStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w
	// 解析参数
	err := fs.Parse(argsSlice)

	// 恢复标准输出
	w.Close()
	os.Stderr = originalStderr

	// 读取错误输出
	if err != nil {
		// 如果有错误，可以选择记录日志
		// fmt.Println("Ignored extra arguments:", err)
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
	// 检查文件是否存在
	if filePath == "" {
		return
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return
	} else if err != nil {
		return
	}

	// 文件存在，进行删除
	err := os.Remove(filePath)
	if err != nil {
		return
	}
}

func (t *UtilTools) DeleteFolder(folderPath string) error {
	err := os.RemoveAll(folderPath)
	if err != nil {
		return fmt.Errorf("删除文件夹失败: %v", err)
	}
	return nil
}

// GetParameter 获取指定模块指定插件的参数
func (t *UtilTools) GetParameter(Parameters map[string]map[string]string, module string, plugin string) (string, bool) {
	// 查找 module 是否存在
	if plugins, modOk := Parameters[module]; modOk {
		// 查找 plugin 是否存在
		if param, plugOk := plugins[plugin]; plugOk {
			return param, true
		}
	}
	// 没有找到对应的参数，返回 false
	return "", false
}

//// GetRootDomain 获取域名的根域名
//func (t *UtilTools) GetRootDomain(input string) (string, error) {
//	input = strings.TrimPrefix(input, "http://")
//	input = strings.TrimPrefix(input, "https://")
//	input = strings.Trim(input, "/")
//	ip := net.ParseIP(input)
//	if ip != nil {
//		return ip.String(), nil
//	}
//	input = "https://" + input
//
//	// 尝试解析为 URL
//	u, err := url.Parse(input)
//	if err == nil && u.Hostname() != "" {
//		ipHost := net.ParseIP(u.Hostname())
//		if ipHost != nil {
//			return ipHost.String(), nil
//		}
//		eTLDPlusOne, err := publicsuffix.EffectiveTLDPlusOne(u.Hostname())
//		if err != nil {
//			return input, fmt.Errorf("根域名解析错误")
//		}
//		return eTLDPlusOne, nil
//		//hostParts := strings.Split(u.Hostname(), ".")
//		//if len(hostParts) < 2 {
//		//	return "", fmt.Errorf("域名格式不正确")
//		//}
//		//if len(hostParts) == 2 {
//		//	return u.Hostname(), nil
//		//}
//		//// 检查是否为复合域名
//		//if _, ok := compoundDomains[hostParts[len(hostParts)-2]+"."+hostParts[len(hostParts)-1]]; ok {
//		//	return hostParts[len(hostParts)-3] + "." + hostParts[len(hostParts)-2] + "." + hostParts[len(hostParts)-1], nil
//		//}
//		//
//		//// 如果域名以 www 开头，特殊处理
//		//if hostParts[0] == "www" {
//		//	return hostParts[len(hostParts)-2] + "." + hostParts[len(hostParts)-1], nil
//		//}
//		//
//		//return hostParts[len(hostParts)-2] + "." + hostParts[len(hostParts)-1], nil
//	}
//	return input, fmt.Errorf("输入既不是有效的 URL，也不是有效的 IP 地址: %v", input)
//}

// GetRootDomain 提取输入中的根域名或 IP 地址
func (t *UtilTools) GetRootDomain(input string) (string, error) {
	u, err := t.SafeParseURL(input)
	if err != nil {
		return input, fmt.Errorf("URL 解析失败: %w", err)
	}

	hostname := u.Hostname()
	if hostname == "" {
		return input, fmt.Errorf("无法获取 Hostname")
	}

	// 是 IP 则直接返回
	if ip := net.ParseIP(hostname); ip != nil {
		return ip.String(), nil
	}

	// 提取有效根域名
	rootDomain, err := publicsuffix.EffectiveTLDPlusOne(hostname)
	if err != nil {
		return input, fmt.Errorf("根域名解析错误: %w", err)
	}

	return rootDomain, nil
}

// SafeParseURL 安全解析 URL，处理非法 %、中文域名、空格等问题
func (t *UtilTools) SafeParseURL(input string) (*url.URL, error) {
	if input == "" {
		return nil, fmt.Errorf("输入为空")
	}

	input = strings.TrimSpace(input)
	input = sanitizePercent(input)
	input = strings.ReplaceAll(input, " ", "%20")

	// 补充协议头
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		input = "https://" + input
	}

	u, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("URL 解析失败: %w", err)
	}

	// IDNA 处理中文域名
	asciiHost, err := idna.ToASCII(u.Hostname())
	if err != nil {
		return nil, fmt.Errorf("域名 IDNA 转换失败: %w", err)
	}

	// 补上端口（如有）
	u.Host = asciiHost + portSuffix(u.Host)

	return u, nil
}

// sanitizePercent 修复非法的 % 编码
func sanitizePercent(input string) string {
	var builder strings.Builder
	for i := 0; i < len(input); i++ {
		if input[i] == '%' {
			if i+2 >= len(input) || !isHexDigit(input[i+1]) || !isHexDigit(input[i+2]) {
				builder.WriteString("%25")
			} else {
				builder.WriteByte('%')
			}
		} else {
			builder.WriteByte(input[i])
		}
	}
	return builder.String()
}

// 判断是否为合法十六进制字符
func isHexDigit(b byte) bool {
	return ('0' <= b && b <= '9') || ('a' <= b && b <= 'f') || ('A' <= b && b <= 'F')
}

// 提取原 host 的端口号（如有）
func portSuffix(host string) string {
	if colon := strings.LastIndex(host, ":"); colon != -1 && colon < len(host)-1 {
		port := host[colon+1:]
		if _, err := fmt.Sscanf(port, "%d", new(int)); err == nil {
			return ":" + port
		}
	}
	return ""
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

func (t *UtilTools) ExecuteCommandWithTimeout(command string, args []string, timeout time.Duration, externalCtx context.Context) error {
	logger.SlogInfo(fmt.Sprintf("ExecuteCommandWithTimeout cmd: %v args %v", command, args))
	// 创建一个带有超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // 确保在函数结束后取消上下文，防止资源泄漏
	// 使用 select 来监控两个上下文：超时上下文和外部传入的上下文
	// 创建一个新的上下文，该上下文会在任意一个上下文被取消时取消
	mergedCtx, mergedCancel := context.WithCancel(ctx)
	go func() {
		// 监听外部上下文的取消
		select {
		case <-externalCtx.Done():
			mergedCancel() // 外部上下文取消时，取消合并的上下文
			return
		case <-ctx.Done():
			// 超时上下文取消时，合并上下文也取消
			mergedCancel()
			return
		}
	}()
	// 创建命令对象，使用带上下文的 exec.CommandContext
	cmd := exec.CommandContext(mergedCtx, command, args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	go io.Copy(io.Discard, stdout)
	go io.Copy(io.Discard, stderr)
	//stderrPipe, err := cmd.StderrPipe()
	//if err != nil {
	//	return fmt.Errorf("failed to get stderr: %w", err)
	//}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}
	//
	//go func() {
	//	scanner := bufio.NewScanner(stderrPipe)
	//	for scanner.Scan() {
	//		logger.SlogWarnLocal(fmt.Sprintf("%v stderr: %s", command, scanner.Text()))
	//	}
	//	if err := scanner.Err(); err != nil {
	//		logger.SlogWarnLocal(fmt.Sprintf("%v stderr scan error: %v", command, err))
	//	}
	//}()

	// 等待命令完成
	err := cmd.Wait()

	if err != nil {
		// 如果是上下文取消的错误
		if errors.Is(mergedCtx.Err(), context.Canceled) {
			// 上下文被取消
			logger.SlogWarnLocal(fmt.Sprintf("command execution canceled: %v", command))
			return nil
		}

		// 如果是超时错误
		if errors.Is(mergedCtx.Err(), context.DeadlineExceeded) {
			// 上下文超时
			logger.SlogWarnLocal(fmt.Sprintf("command execution timed out: %v", command))
			return nil
		}
		return err
	}
	// 如果没有错误，说明命令执行成功
	return nil
}

// ExecuteCommandToChanWithTimeout 执行指定命令，命令的输出每一行会发送到 result 的通道中，支持上下文管理和超时时间。
func (t *UtilTools) ExecuteCommandToChanWithTimeout(cmdName string, args []string, result chan<- string, timeout time.Duration, ctx context.Context) {
	logger.SlogInfo(fmt.Sprintf("ExecuteCommandToChanWithTimeout cmd: %v args %v", cmdName, args))
	// 使用超时时间包装上下文
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	defer close(result)
	// 创建命令
	cmd := exec.CommandContext(ctxWithTimeout, cmdName, args...)
	var wg sync.WaitGroup
	// 获取命令输出管道（标准输出）
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.SlogWarnLocal(fmt.Sprintf("Error getting stdout pipe: %v", err))
		return
	}

	// 获取命令错误输出管道（标准错误）
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.SlogWarnLocal(fmt.Sprintf("Error getting stderr pipe: %v", err))
		return
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		logger.SlogWarnLocal(fmt.Sprintf("Error getting starting command: %v", err))
		return
	}
	wg.Add(1)
	// 使用 goroutine 读取命令的标准输出
	go func() {
		defer wg.Done()
		reader := bufio.NewReaderSize(stdout, 7340032)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				logger.SlogWarnLocal(fmt.Sprintf("Error reading stdout: %v", err))
				break
			}
			select {
			case result <- strings.TrimRight(line, "\n"):
			case <-ctxWithTimeout.Done():
				return
			}
		}
	}()
	wg.Add(1)
	// 使用 goroutine 读取命令的错误输出
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			txt := scanner.Text()
			select {
			case <-ctxWithTimeout.Done():
				return
			default:
				result <- txt
				logger.SlogWarnLocal(fmt.Sprintf("Error getting stderr: %s", txt))
			}
		}
		if err := scanner.Err(); err != nil {
			logger.SlogWarnLocal(fmt.Sprintf("Error reading stderr: %v", err))
		}
	}()

	// 等待命令执行完毕
	err = cmd.Wait()
	if err != nil {
		select {
		case <-ctxWithTimeout.Done():
			return
		default:
			logger.SlogWarnLocal(fmt.Sprintf("Error waiting for command: %v", err))
		}
	}
	wg.Wait()
}

// ExecuteCommandToChan 执行指定命令，命令的输出每一行会发送到result的通道中。
func (t *UtilTools) ExecuteCommandToChan(cmdName string, args []string, result chan<- string) {
	logger.SlogInfo(fmt.Sprintf("ExecuteCommandToChan cmd: %v args %v", cmdName, args))
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

// EnsureFilePathExists 检查给定的文件路径，如果文件夹不存在则创建
func (t *UtilTools) EnsureFilePathExists(filePath string) error {
	// 获取文件的目录路径
	dir := filepath.Dir(filePath)

	err := t.EnsureDir(dir)
	if err != nil {
		return err
	}
	return nil
}

// ReadFileLineByLine 函数逐行读取文件，并将每一行发送到通道中
func (t *UtilTools) ReadFileLineByLine(filePath string, lineChan chan<- string, ctx context.Context) error {
	defer close(lineChan)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close() // 函数结束时关闭文件
	// 使用 bufio.Scanner 逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
			lineChan <- scanner.Text() // 将读取到的行发送到通道
		}
	}

	// 检查读取过程中是否发生错误
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// ReadFileLineReader 读取文件按行处理，每行是一个完整的JSON对象
func (t *UtilTools) ReadFileLineReader(filePath string, lineChan chan<- string, ctx context.Context) error {
	defer close(lineChan)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 使用大缓冲区来提高读取性能
	reader := bufio.NewReaderSize(file, 30*1024) // 30KB 缓冲区

	// 循环按行读取文件
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// 读取一行
			line, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				return err
			}
			if err == io.EOF && len(line) == 0 {
				return nil // 文件读取完毕
			}

			// 去除末尾的换行符和空格
			line = strings.TrimSpace(line)

			// 将每行发送到 lineChan
			select {
			case lineChan <- line:
			case <-ctx.Done():
				return nil
			}

			if err == io.EOF {
				return nil // 到达文件末尾，退出循环
			}
		}
	}
}

func (t *UtilTools) ReadFileLineReaderBytes(filePath string, lineChan chan<- []byte, ctx context.Context) error {
	defer close(lineChan)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 使用大缓冲区提高读取性能
	reader := bufio.NewReaderSize(file, 30*1024) // 30KB 缓冲区

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// 读取一行，ReadLine 返回 []byte
			lineBytes, isPrefix, err := reader.ReadLine()
			if err != nil && err != io.EOF {
				return err
			}
			if err == io.EOF && len(lineBytes) == 0 {
				return nil
			}

			// 如果行太长被分段，循环合并
			fullLine := append([]byte(nil), lineBytes...) // 可选择深拷贝，避免被复用覆盖
			for isPrefix {
				lineBytes, isPrefix, _ = reader.ReadLine()
				fullLine = append(fullLine, lineBytes...)
			}

			// 去除首尾空格
			fullLine = bytes.TrimSpace(fullLine)

			// 发送到 channel
			select {
			case lineChan <- fullLine:
			case <-ctx.Done():
				return nil
			}

			if err == io.EOF {
				return nil
			}
		}
	}
}

func (t *UtilTools) ReadFileToStringOptimized(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	// io.Copy 会自动按块读取，比手动循环略快
	_, err = io.Copy(&buf, file)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (t *UtilTools) CdnCheck(host string) (bool, string) {
	ip := net.ParseIP(host)
	matched, val, err := t.CdnCheckClient.CheckCDN(ip)
	if err != nil {
		return false, ""
	}
	if matched {
		return true, val
	} else {
		return false, ""
	}
}

func (t *UtilTools) WafCheck(ipStr string) (bool, string) {
	ip := net.ParseIP(ipStr)
	waf, s, err := t.CdnCheckClient.CheckWAF(ip)
	if err != nil {
		return false, ""
	}
	return waf, s
}

// GetDomain 提取URL中的域名
func (t *UtilTools) GetDomain(rawUrl string) string {
	// 解析 URL
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return rawUrl
	}

	// 提取域名
	domain := parsedUrl.Host
	if domain == "" {
		return rawUrl
	}
	// 去掉端口号（如果有）
	domain = strings.Split(domain, ":")[0]

	return domain
}

func removeDefaultPort(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// 检查端口并移除
	if (parsedURL.Scheme == "http" && parsedURL.Port() == "80") ||
		(parsedURL.Scheme == "https" && parsedURL.Port() == "443") {
		parsedURL.Host = parsedURL.Hostname() // 去掉端口
	}

	return parsedURL.String()
}

var SizeThreshold = 500 * 1024 // 500 KB

func (t *UtilTools) HttpxResultToAssetHttp(r runner.Result) types.AssetHttp {
	//defer Tools.DeleteFile(r.ScreenshotPath)
	Screenshot := ""
	if r.ScreenshotBytes != nil {
		if len(r.ScreenshotBytes) <= SizeThreshold {
			Screenshot = "data:image/png;base64," + base64.StdEncoding.EncodeToString(r.ScreenshotBytes)
		} else {
			Screenshot = "data:image/png;base64," + Tools.CompressAndEncodeScreenshot(r.ScreenshotBytes, 0.5)
		}
	}
	var ah = types.AssetHttp{
		Time:         Tools.GetTimeNow(),
		TLSData:      r.TLSData, // You may need to set an appropriate default value based on the actual type.
		Hashes:       r.Hashes,
		CDNName:      r.CDNName,
		Port:         r.Port,
		URL:          removeDefaultPort(r.URL), //去掉默认的80和443端口
		Title:        r.Title,
		Type:         "http",
		Error:        r.Error,
		ResponseBody: r.ResponseBody,
		Host:         t.GetDomain(r.URL),
		IP:           r.Host,
		FavIconMMH3:  r.FavIconMMH3,
		//FaviconPath:  r.FaviconPath,
		RawHeaders:   r.RawHeaders,
		Jarm:         r.JarmHash,
		Technologies: r.Technologies, // You may need to set an appropriate default value based on the actual type.
		StatusCode:   r.StatusCode,   // You may need to set an appropriate default value.
		Webcheck:     false,
		IconContent:  base64.StdEncoding.EncodeToString(r.FaviconData),
		CDN:          r.CDN,
		WebServer:    r.WebServer,
		Service:      r.Scheme,
		Screenshot:   Screenshot,
	}
	ah.LastScanTime = ah.Time
	return ah
}

func (t *UtilTools) IsMatchingFilter(fs []*regexp.Regexp, d []byte) bool {
	for _, r := range fs {
		if r.Match(d) {
			return true
		}
	}
	return false
}

//	func (t *UtilTools) AssetResultToAssetOther(r types.Asset) types.AssetOther {
//		var ao = types.AssetOther{
//
//		}
//		return ao
//	}
//
// generateIPRange 生成 IP 范围
func generateIPRange(start, end string) ([]string, error) {
	startIP := net.ParseIP(start)
	endIP := net.ParseIP(end)

	if startIP == nil || endIP == nil {
		return nil, fmt.Errorf("invalid IP addresses: %s, %s", start, end)
	}

	var ipList []string
	for ip := startIP; !ip.Equal(endIP); incrementIP(ip) {
		ipList = append(ipList, ip.String())
	}
	ipList = append(ipList, endIP.String()) // 添加结束 IP

	return ipList, nil
}

// incrementIP 增加 IP 地址
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] != 0 {
			break
		}
	}
}

// ipRangeToSlice 将 IP 网络转换为字符串切片
func ipRangeToSlice(ipNet *net.IPNet) []string {
	var ipList []string
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		ipList = append(ipList, ip.String())
	}
	return ipList
}

func (t *UtilTools) GenerateTarget(target string) ([]string, error) {
	if strings.Contains(target, "http://") || strings.Contains(target, "https://") {
		return []string{target}, nil
	}

	if strings.Contains(target, "-") {
		// 处理 IP 范围
		parts := strings.Split(target, "-")
		if len(parts) != 2 {
			return []string{target}, nil
		}
		startIP := strings.TrimSpace(parts[0])
		endIP := strings.TrimSpace(parts[1])
		ipRange, err := generateIPRange(startIP, endIP)
		if err != nil {
			return []string{target}, nil
		}
		return ipRange, err
	}

	if strings.Contains(target, "/") {
		// 处理网络
		_, ipNet, err := net.ParseCIDR(target)
		if err != nil {
			return []string{target}, nil
		}
		return ipRangeToSlice(ipNet), nil
	}

	// 返回单个目标
	return []string{target}, nil
}

func (t *UtilTools) GenerateIgnore(ignore string) ([]string, []*regexp.Regexp, error) {
	var ignoreList []string
	var regexList []*regexp.Regexp
	for _, ta := range strings.Split(ignore, "\n") {
		ta = strings.ReplaceAll(ta, "http://", "")
		ta = strings.ReplaceAll(ta, "https://", "")
		ta = strings.TrimSpace(ta)
		if !strings.Contains(ta, "*") {
			result, err := t.GenerateTarget(ta)
			if err != nil {
				return nil, nil, err
			}
			ignoreList = append(ignoreList, result...)
		} else {
			// 转义并替换为正则表达式
			tEscaped := regexp.QuoteMeta(ta)
			tEscaped = strings.ReplaceAll(tEscaped, `\*`, `.*`)
			regex, err := regexp.Compile(tEscaped)
			if err != nil {
				return nil, nil, err
			}
			regexList = append(regexList, regex)
		}
	}
	return ignoreList, regexList, nil
}

func (t *UtilTools) IsSuffixURL(rawURL string, suffix string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("URL 解析错误:", err)
		return false
	}

	// 获取路径部分，去掉查询参数
	path := parsedURL.Path

	// 判断路径是否以 ".js" 结尾
	return strings.HasSuffix(path, suffix)
}

// CompareContentSimilarity 计算编辑距离相似度
func (t *UtilTools) CompareContentSimilarity(content1, content2 string) (float64, error) {
	// 使用编辑距离计算相似度
	similarity, err := edlib.StringsSimilarity(content1, content2, edlib.Levenshtein)
	if err != nil {
		return 0, fmt.Errorf("error calculating similarity: %v", err)
	}
	percentage := similarity * 100
	result := math.Round(float64(percentage*100)) / 100
	return result, nil
}

// UnzipSrcToDest 将文件src解压到dest
func (t *UtilTools) UnzipSrcToDest(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func (t *UtilTools) UntarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
				return err
			}

			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

func (t *UtilTools) UnzipFile(src, dest string) error {
	if strings.HasSuffix(src, ".zip") {
		return t.UnzipSrcToDest(src, dest)
	} else if strings.HasSuffix(src, ".tar.gz") {
		return t.UntarGz(src, dest)
	}
	return fmt.Errorf("unsupported file format: %s", src)
}

// RemoveStringDuplicates 数组去重
func (t *UtilTools) RemoveStringDuplicates(arr []string) []string {
	unique := make(map[string]bool)
	result := []string{}

	for _, value := range arr {
		if !unique[value] {
			unique[value] = true
			result = append(result, value)
		}
	}

	return result
}

func (t *UtilTools) HandleLinuxTemp() {
	tempDir := "/tmp"

	// 定义 Linux 下的 find 命令
	findCmd := fmt.Sprintf(
		`find %s -type d \( -regex '.*/[0-9]\{9\}' -o -name 'nuclei*' -o -name '.org.chromium.Chromium*' -o -name '*_badger' -o -name 'rod' \) -exec rm -rf {} +`,
		tempDir,
	)

	// 执行命令
	cmd := exec.Command("bash", "-c", findCmd)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("执行 Linux find 命令时出错: %v\n", err)
		logger.SlogWarn("清空临时文件出错，请手动清空/tmp目录，防止磁盘占用过大")
		return
	}
}

// CompressAndEncodeScreenshot 压缩图片并返回 Base64 编码的字符串。
// 它接收原始图片字节流 r.ScreenshotBytes，返回 Base64 编码的图片字符串。
// 该函数会无损压缩 PNG 图像并缩小为原尺寸的 scaleFactor。 0.5 = 50%
func (t *UtilTools) CompressAndEncodeScreenshot(screenshotBytes []byte, scaleFactor float64) string {
	if screenshotBytes == nil {
		logger.SlogWarn("No screenshot data provided.")
		return ""
	}

	// 创建字节缓冲区直接处理图像数据
	var buf bytes.Buffer

	// 使用 image.DecodeConfig 获取图像类型和尺寸，不加载整个图片
	config, _, err := image.DecodeConfig(bytes.NewReader(screenshotBytes))
	if err != nil {
		logger.SlogWarn(fmt.Sprintf("Error getting image config:", err))
		return base64.StdEncoding.EncodeToString(screenshotBytes) // 返回原始 Base64 编码
	}

	// 计算新的图片尺寸（通过缩放比例）
	newWidth := uint(float64(config.Width) * scaleFactor)
	newHeight := uint(float64(config.Height) * scaleFactor)

	// 解码并处理图片的流式操作
	img, imgType, err := image.Decode(bytes.NewReader(screenshotBytes))
	if err != nil {
		logger.SlogWarn(fmt.Sprintf("Error decoding image:", err))
		return base64.StdEncoding.EncodeToString(screenshotBytes) // 返回原始 Base64 编码
	}

	// 压缩图片（无损或有损方式，调整尺寸）
	compressedImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

	// 根据图片格式进行不同的编码
	switch imgType {
	case "jpeg":
		// 使用有损压缩
		err = jpeg.Encode(&buf, compressedImg, nil)
		if err != nil {
			logger.SlogWarn(fmt.Sprintf("Error encoding JPEG image:", err))
			return base64.StdEncoding.EncodeToString(buf.Bytes()) // 错误时返回 Base64 编码
		}
	case "png":
		// 使用无损压缩
		err = png.Encode(&buf, compressedImg)
		if err != nil {
			logger.SlogWarn(fmt.Sprintf("Error encoding PNG image:", err))
			return base64.StdEncoding.EncodeToString(buf.Bytes()) // 错误时返回 Base64 编码
		}
	default:
		logger.SlogWarn(fmt.Sprintf("Unsupported image format:", imgType))
		return base64.StdEncoding.EncodeToString(buf.Bytes()) // 默认返回 Base64 编码
	}

	// 返回压缩并编码后的 Base64 字符串
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// Command 简介调用exec.Command
func (t *UtilTools) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

func (t *UtilTools) GetPdfContent(filePath string) string {
	f, r, err := pdf.Open(filePath)
	// remember close file
	defer f.Close()
	if err != nil {
		logger.SlogWarn(fmt.Sprintf("GetPdfContent error: %v", err))
		return ""
	}
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		logger.SlogWarn(fmt.Sprintf("GetPdfContent GetPlainText error: %v", err))
		return ""
	}
	buf.ReadFrom(b)

	return buf.String()
}

func (t *UtilTools) MoveContents(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 忽略根目录
		if path == srcDir {
			return nil
		}

		// 计算目标路径
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// 是文件就移动
			return t.MoveFile(path, dstPath)
		}
	})
}

func (t *UtilTools) MoveFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return err
	}
	return os.Rename(src, dst) // 也可以先 copy 再删原文件，防止跨盘失败
}

func (t *UtilTools) EqualStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (t *UtilTools) CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (t *UtilTools) IsJson(str string) bool {
	// JSON 字符串通常以 `{` 开始，以 `}` 结束
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return false
	}

	// 简单检查是否以 `{` 开始，以 `}` 结束
	if str[0] == '{' && str[len(str)-1] == '}' {
		var js map[string]interface{}
		// 尝试将字符串解析为 JSON 对象
		if err := json.Unmarshal([]byte(str), &js); err == nil {
			return true
		}
	}
	return false
}

func (t *UtilTools) ModifyJSONValues(jsonStr string, value string) (string, error) {
	// 解析 JSON 为 interface{}
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	// 替换所有值为 "test"
	updatedData := t.ReplaceValuesWithTest(data, value)

	// 将修改后的数据重新编码为 JSON 字符串
	updatedJSON, err := json.Marshal(updatedData)
	if err != nil {
		return "", err
	}

	return string(updatedJSON), nil
}

func (t *UtilTools) ReplaceValuesWithTest(data interface{}, value string) interface{} {
	// 如果是 map 类型（JSON 对象），直接遍历并替换每个值
	if m, ok := data.(map[string]interface{}); ok {
		for key := range m {
			m[key] = value // 替换为 "test"
		}
		return m
	}

	// 如果是 slice 类型（JSON 数组），直接遍历并替换每个值
	if arr, ok := data.([]interface{}); ok {
		for i := range arr {
			arr[i] = value // 替换为 "test"
		}
		return arr
	}

	// 如果是其他类型（比如 string），直接返回 "test"
	return value
}

func (t *UtilTools) HashXX64String(input string) string {
	hash := xxhash.Sum64String(input)
	return fmt.Sprintf("%x", hash)
}
