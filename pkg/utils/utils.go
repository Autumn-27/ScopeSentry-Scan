// utils-------------------------------------
// @file      : utils.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 22:34
// -------------------------------------------

package utils

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/projectdiscovery/gologger"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
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

func (t *UtilTools) GetSystemUsage() (int, float64) {
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

func (t *UtilTools) WriteContentFile(filPath string, fileContent string) error {
	// 将字符串写入文件
	return t.WriteByteContentFile(filPath, []byte(fileContent))
}

func (t *UtilTools) WriteByteContentFile(filPath string, fileContent []byte) error {
	// 将字符串写入文件
	if err := ioutil.WriteFile(filPath, fileContent, 0666); err != nil {
		fmt.Printf("Failed to create filPath: %s - %s", filPath, err)
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

func (t *UtilTools) DeleteFile(filePath string) {
	// 调用Remove函数删除文件
	err := os.Remove(filePath)
	if err != nil {
		gologger.Error().Msg(fmt.Sprintf("Failed to DeleteFile: %s - %s", filePath, err))
	}
}

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

	// 将文件内容写入本地文件
	err = ioutil.WriteFile(filePath, body, 0777)
	if err != nil {
		return false, fmt.Errorf("ioutil.WriteFile failed: %w", err)
	}

	return true, nil
}
