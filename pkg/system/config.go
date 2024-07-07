// Package system -----------------------------
// @file      : system.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/9 21:57
// -------------------------------------------
package system

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/mongdbClient"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/redisClient"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type ScopeSentryConfig struct {
	System struct {
		NodeName       string `yaml:"NodeName"`
		TimeZoneName   string `yaml:"TimeZoneName"`
		MaxTaskNum     string `yaml:"MaxTaskNum"`
		DirscanThread  string `yaml:"DirscanThread"`
		PortscanThread string `yaml:"PortscanThread"`
		State          string `yaml:"State"`
		Running        int    `yaml:"Running"`
		Finished       int    `yaml:"Finished"`
		Debug          bool   `yaml:"Debug"`
		CrawlerThread  string `yaml:"CrawlerThread"`
		UrlThread      string `yaml:"UrlThread"`
		UrlMaxNum      string `yaml:"UrlMaxNum"`
		CrawlerTimeout string `json:"CrawlerTimeout"`
	} `yaml:"System"`
	Mongodb struct {
		IP       string `yaml:"IP"`
		Port     string `yaml:"Port"`
		Username string `yaml:"Username"`
		Password string `yaml:"Password"`
	} `yaml:"Mongodb"`
	Redis struct {
		IP       string `yaml:"IP"`
		Port     string `yaml:"Port"`
		Password string `yaml:"Password"`
	} `yaml:"Redis"`
}

var ConfigDir string
var ExtPath string
var CrawlerPath string
var CrawlerExecPath string
var KsubdomainPath string
var KsubdomainExecPath string
var AppConfig ScopeSentryConfig
var ConfigFileExists bool
var DebugFlag bool
var RedisClient *redisClient.RedisClient
var MongoClient *mongdbClient.MongoDBClient
var Projects []types.Project
var PortDict []types.PortDict
var SubdomainTakerFingers []types.SubdomainTakerFinger
var DirDict []string
var PocList map[string]types.PocData
var SensitiveRules []types.SensitiveRule
var WebFingers []types.WebFinger
var AppRFMutex sync.Mutex
var RunningNum = 0
var FinNum = 0
var NotificationConfig types.NotificationConfig
var NotificationApi []types.NotificationApi
var UpdateUrl string
var VERSION string
var UpdateSystemFlag chan bool
var CrawlerTarget chan types.CrawlerTask
var CrawlerThreadUpdateFlag chan bool
var CrawlerThreadNow int
var SensRegChan chan struct{}

func SetUp() bool {
	UpdateSystemFlag = make(chan bool)
	CrawlerTarget = make(chan types.CrawlerTask, 1)
	CrawlerThreadUpdateFlag = make(chan bool)
	SensRegChan = make(chan struct{}, 50)
	CrawlerThreadNow = 0
	VERSION = "1.3"
	UpdateUrl = "https://update.scope-sentry.top"
	PocList = make(map[string]types.PocData)
	dbFlag := InitDb()
	if !dbFlag {
		return dbFlag
	}
	LogInit(AppConfig.System.Debug)
	go UpdateInit()
	SlogInfoLocal("Start check crawler tool")
	flagCheck := CheckCrawler()
	if !flagCheck {
		return false
	}
	SlogInfoLocal("End check crawler tool")
	SlogInfoLocal("Start check ksubdomain tool")
	flagCheck = CheckKsubdomain()
	if !flagCheck {
		return false
	}
	SlogInfoLocal("End end ksubdomain tool")
	SlogInfoLocal("Start check Rustscan tool")
	flagCheck = CheckRustscan()
	if !flagCheck {
		return false
	}
	SlogInfoLocal("End end Rustscan tool")
	SlogInfoLocal("Start pulling data")
	UpdateSetUp()
	SlogInfoLocal("End pulling data")
	errj := json.Unmarshal(TakeOver_finger, &SubdomainTakerFingers)
	if errj != nil {
		fmt.Println("解析JSON失败:", errj)
	}
	SlogInfo(fmt.Sprintf("The current number of concurrent tasks: %s", AppConfig.System.MaxTaskNum))
	return true
}
func Test() {
	InitDb()
	LogInit(AppConfig.System.Debug)
}

func InitDb() bool {
	executableDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		SlogErrorLocal(fmt.Sprintf("Failed to retrieve the directory of the executable file: %s", err))
		return false
	}
	ConfigDir = filepath.Join(executableDir, "config")
	if err := os.MkdirAll(ConfigDir, os.ModePerm); err != nil {
		SlogErrorLocal(fmt.Sprintf("Failed to create system folder: %s", err))
		return false
	}
	ConfigFileExists = false
	scopeSentryConfigPath := filepath.Join(ConfigDir, "ScopeSentryConfig.yaml")
	if _, err := os.Stat(scopeSentryConfigPath); os.IsNotExist(err) {
		debugFlag := os.Getenv("Debug")
		if debugFlag == "1" {
			AppConfig.System.Debug = true
		} else {
			AppConfig.System.Debug = false
		}
		AppConfig.System.NodeName = os.Getenv("NodeName")
		AppConfig.System.TimeZoneName = os.Getenv("TimeZoneName")
		AppConfig.System.MaxTaskNum = "7"
		AppConfig.System.PortscanThread = "15"
		AppConfig.System.DirscanThread = "15"
		AppConfig.System.CrawlerThread = "2"
		AppConfig.System.UrlThread = "5"
		AppConfig.System.UrlMaxNum = "500"
		AppConfig.System.CrawlerTimeout = "1"
		AppConfig.Mongodb.IP = os.Getenv("Mongodb_IP")
		AppConfig.Mongodb.Port = os.Getenv("MONGODB_PORT")
		AppConfig.Mongodb.Username = os.Getenv("Mongodb_Username")
		AppConfig.Mongodb.Password = os.Getenv("Mongodb_Password")
		AppConfig.Redis.IP = os.Getenv("Redis_IP")
		AppConfig.Redis.Port = os.Getenv("Redis_PORT")
		AppConfig.Redis.Password = os.Getenv("Redis_Password")
		err := WriteYamlConfigToFile(filepath.Join(ConfigDir, "ScopeSentryConfig.yaml"), AppConfig)
		if err != nil {
			return false
		}
	} else {
		ConfigFileExists = true
		err = LoadYAMLFile(scopeSentryConfigPath, &AppConfig)
		if err != nil {
			return false
		}
	}
	if AppConfig.Mongodb.IP == "" || AppConfig.Redis.IP == "" {
		fmt.Println("Mongodb.IP 或 Redis.IP 为空，返回 false")
		return false
	}
	GetRedisClient()
	GetMongbClient()
	if RedisClient == nil {
		fmt.Println("RedisClient init err")
		return false
	}
	if MongoClient == nil {
		fmt.Println("MongoClient init err")
		return false
	}
	return true
}

func GetMongbClient() {
	fmt.Println("GetMongbClient begin")
	MongoClient, _ = mongdbClient.Connect(AppConfig.Mongodb.Username, AppConfig.Mongodb.Password, AppConfig.Mongodb.IP, AppConfig.Mongodb.Port)
	fmt.Println("GetMongbClient end")
}
func GetRedisClient() {
	fmt.Println("GetRedisClient begin")
	redisAddr := AppConfig.Redis.IP + ":" + AppConfig.Redis.Port
	redisPassword := AppConfig.Redis.Password
	RedisClient, _ = redisClient.NewRedisClient(redisAddr, redisPassword, 0)
	fmt.Println("GetRedisClient end")
}

func CheckRustscan() bool {
	executableDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		SlogError(fmt.Sprintf("Failed to retrieve the directory of the executable file:", err))
		return false
	}
	ExtPath = filepath.Join(executableDir, "ext")
	if err := os.MkdirAll(ExtPath, os.ModePerm); err != nil {
		SlogError(fmt.Sprintf("Failed to create ext folder:", err))
		return false
	}
	rustscanPath := filepath.Join(ExtPath, "rustscan")
	if err := os.MkdirAll(rustscanPath, os.ModePerm); err != nil {
		SlogError(fmt.Sprintf("Failed to create radPath folder:", err))
		return false
	}
	osType := runtime.GOOS
	// 判断操作系统类型
	var path string
	var dir string
	switch osType {
	case "windows":
		path = "rustscan.exe"
		dir = "win"
	case "linux":
		path = "rustscan"
		dir = "linux"
	default:
		dir = "darwin"
		path = "rustscan"
	}
	rustscanExecPath := filepath.Join(rustscanPath, path)
	if _, err := os.Stat(rustscanExecPath); os.IsNotExist(err) {
		resp, err := http.Get(fmt.Sprintf("%v/%v/%v", "https://raw.githubusercontent.com/Autumn-27/ScopeSentry-Scan/main/tools", dir, path))
		if err != nil {
			resp, err = http.Get(fmt.Sprintf("%v/%v/%v", "https://raw.githubusercontent.com/Autumn-27/ScopeSentry-Scan/main/tools", dir, path))
			if err != nil {
				SlogError(fmt.Sprintf("Error: %s", err))
				return false
			}
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			SlogError("Get rustscan Tool fail, go to https://github.com/boy-hack/ksubdomain/ Download the corresponding version and rename the executable program to ksubdomain/ksubdomain.exe and place it in the ext/rad file.")
			return false
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			SlogError(fmt.Sprintf("Read rustscan Tool file error: %s", err))
			return false
		}
		err = ioutil.WriteFile(KsubdomainExecPath, body, 0755)
		if err != nil {
			SlogError(fmt.Sprintf("Write Rad Tool Fail: %s", err))
			return false
		}
		if osType == "linux" {
			err = os.Chmod(KsubdomainExecPath, 0755)
			if err != nil {
				SlogError(fmt.Sprintf("Chmod rustscan Tool Fail: %s", err))
				return false
			}
		}
	}
	return true
}

func CheckKsubdomain() bool {
	executableDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		SlogError(fmt.Sprintf("Failed to retrieve the directory of the executable file:", err))
		return false
	}
	ExtPath = filepath.Join(executableDir, "ext")
	if err := os.MkdirAll(ExtPath, os.ModePerm); err != nil {
		SlogError(fmt.Sprintf("Failed to create ext folder:", err))
		return false
	}
	ksubdomainPath := filepath.Join(ExtPath, "ksubdomain")
	if err := os.MkdirAll(ksubdomainPath, os.ModePerm); err != nil {
		SlogError(fmt.Sprintf("Failed to create radPath folder:", err))
		return false
	}
	targetPath := filepath.Join(ksubdomainPath, "target")
	if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
		SlogError(fmt.Sprintf("Failed to create targetPath folder:", err))
		return false
	}
	resultPath := filepath.Join(ksubdomainPath, "result")
	if err := os.MkdirAll(resultPath, os.ModePerm); err != nil {
		SlogError(fmt.Sprintf("Failed to create resultPath folder:", err))
		return false
	}

	osType := runtime.GOOS
	// 判断操作系统类型
	var path string
	var dir string
	switch osType {
	case "windows":
		path = "ksubdomain.exe"
		dir = "win"
	case "linux":
		path = "ksubdomain"
		dir = "linux"
	default:
		dir = "darwin"
		path = "ksubdomain"
	}
	KsubdomainPath = filepath.Join(ExtPath, "ksubdomain")
	KsubdomainExecPath = filepath.Join(KsubdomainPath, path)
	if _, err := os.Stat(KsubdomainExecPath); os.IsNotExist(err) {
		resp, err := http.Get(fmt.Sprintf("%v/%v/%v", "https://raw.githubusercontent.com/Autumn-27/ScopeSentry-Scan/main/tools", dir, path))
		if err != nil {
			resp, err = http.Get(fmt.Sprintf("%v/%v/%v", "https://raw.githubusercontent.com/Autumn-27/ScopeSentry-Scan/main/tools", dir, path))
			if err != nil {
				SlogError(fmt.Sprintf("Error: %s", err))
				return false
			}
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			SlogError("Get ksubdomain Tool fail, go to https://github.com/boy-hack/ksubdomain/ Download the corresponding version and rename the executable program to ksubdomain/ksubdomain.exe and place it in the ext/rad file.")
			return false
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			SlogError(fmt.Sprintf("Read ksubdomain Tool file error: %s", err))
			return false
		}
		err = ioutil.WriteFile(KsubdomainExecPath, body, 0755)
		if err != nil {
			SlogError(fmt.Sprintf("Write Rad Tool Fail: %s", err))
			return false
		}
		if osType == "linux" {
			err = os.Chmod(KsubdomainExecPath, 0755)
			if err != nil {
				SlogError(fmt.Sprintf("Chmod ksubdomain Tool Fail: %s", err))
				return false
			}
		}
		// 检查ksudomain是否可以执行
		cmd := exec.Command(KsubdomainExecPath, "v", "-d", "scope-sentry.top")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			SlogError(fmt.Sprintf("ksubdomain Tool error: %s", err))
			return false
		}
		if err := cmd.Start(); err != nil {
			SlogError(fmt.Sprintf("ksubdomain Tool run start error: %s", err))
			return false
		}
		flag := 0
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), ".") {
				flag += 1
			}
			if flag == 20 {
				SlogError("ksubdomain get device error,Check whether the proxy is enabled.")
				return false
			}
		}
		if err := scanner.Err(); err != nil {
			SlogError(fmt.Sprintf("ksubdomain Tool run start f error: %s", err))
			return false
		}

		if err := cmd.Wait(); err != nil {
			SlogError(fmt.Sprintf("ksubdomain Tool run start f Command finished with error:: %s", err))
			return false
		}
	}
	return true
}

func CheckCrawler() bool {
	executableDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		SlogError(fmt.Sprintf("Failed to retrieve the directory of the executable file:", err))
		return false
	}
	ExtPath = filepath.Join(executableDir, "ext")
	if err := os.MkdirAll(ExtPath, os.ModePerm); err != nil {
		SlogError(fmt.Sprintf("Failed to create ext folder:", err))
		return false
	}
	radPath := filepath.Join(ExtPath, "rad")
	if err := os.MkdirAll(radPath, os.ModePerm); err != nil {
		SlogError(fmt.Sprintf("Failed to create radPath folder:", err))
		return false
	}
	targetPath := filepath.Join(radPath, "target")
	if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
		SlogError(fmt.Sprintf("Failed to create targetPath folder:", err))
		return false
	}
	resultPath := filepath.Join(radPath, "result")
	if err := os.MkdirAll(resultPath, os.ModePerm); err != nil {
		SlogError(fmt.Sprintf("Failed to create resultPath folder:", err))
		return false
	}

	osType := runtime.GOOS
	// 判断操作系统类型
	var path string
	var dir string
	switch osType {
	case "windows":
		path = "rad.exe"
		dir = "win"
	case "linux":
		path = "rad"
		dir = "linux"
	default:
		path = "rad"
		dir = "darwin"
	}
	CrawlerPath = filepath.Join(ExtPath, "rad")
	CrawlerExecPath = filepath.Join(radPath, path)
	if _, err := os.Stat(CrawlerExecPath); os.IsNotExist(err) {
		resp, err := http.Get(fmt.Sprintf("%v/%v/%v", "https://raw.githubusercontent.com/Autumn-27/ScopeSentry-Scan/main/tools", dir, path))
		if err != nil {
			resp, err = http.Get(fmt.Sprintf("%v/%v/%v", "https://raw.githubusercontent.com/Autumn-27/ScopeSentry-Scan/main/tools", dir, path))
			if err != nil {
				SlogError(fmt.Sprintf("Error: %s", err))
				return false
			}
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			SlogError("Get Rad Tool fail, go to https://github.com/chaitin/rad/ Download the corresponding version and rename the executable program to rad/rade.exe and place it in the ext/rad file.")
			return false
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			SlogError(fmt.Sprintf("Read Rad Tool file error: %s", err))
			return false
		}
		err = ioutil.WriteFile(CrawlerExecPath, body, 0755)
		if err != nil {
			SlogError(fmt.Sprintf("Write Rad Tool Fail: %s", err))
			return false
		}
		if osType == "linux" {
			err = os.Chmod(CrawlerExecPath, 0755)
			if err != nil {
				SlogError(fmt.Sprintf("Chmod Rad Tool Fail: %s", err))
				return false
			}
		}
	}
	return true
}

func UpdateInit() {
	defer RecoverPanic("UpdateInit")
	jsonData := map[string]string{"update": "init"}
	resp, _ := HTTPPostGetData(UpdateUrl+"/uptate/init", jsonData)
	if resp["code"] == nil {
		return
	}
	if resp["code"].(float64) != 200 {
		SlogDebugLocal(fmt.Sprintf("Update Init Error: %s", resp["message"]))
	}
}
func WriteYamlConfigToFile(path string, content interface{}) error {
	// 将配置内容转换为 YAML 格式
	yamlContent, err := yaml.Marshal(content)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to encode content to YAML: %v", err)
		// 处理错误，例如记录日志或返回错误
		myLog := CustomLog{
			Status: "Error",
			Msg:    errMsg,
		}
		PrintLog(myLog)
		return fmt.Errorf(errMsg)
	}

	// 将配置内容写入文件
	if err := ioutil.WriteFile(path, yamlContent, 0666); err != nil {
		errMsg := fmt.Sprintf("Failed to write config file %s: %v", path, err)
		// 处理错误，例如记录日志或返回错误
		myLog := CustomLog{
			Status: "Error",
			Msg:    errMsg,
		}
		PrintLog(myLog)
		return fmt.Errorf(errMsg)
	}
	return nil
}
func LoadYAMLFile(filePath string, target interface{}) error {
	// 读取 YAML 文件内容
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		errMsg := fmt.Sprintf("Error reading YAML file %s: %v", filePath, err)
		// 处理错误，例如记录日志或返回错误
		myLog := CustomLog{
			Status: "Error",
			Msg:    errMsg,
		}
		PrintLog(myLog)
		return fmt.Errorf(errMsg)
	}

	// 使用 yaml 库解析 YAML 内容到目标对象
	if err := yaml.Unmarshal(yamlFile, target); err != nil {
		errMsg := fmt.Sprintf("Error unmarshaling YAML content: %v", err)
		// 处理错误，例如记录日志或返回错误
		myLog := CustomLog{
			Status: "Error",
			Msg:    errMsg,
		}
		PrintLog(myLog)
		return fmt.Errorf(errMsg)
	}
	return nil
}

func GetDomainDic() []string {
	domainDidPath := filepath.Join(ConfigDir, "domainDic")
	// Open the file
	file, err := os.Open(domainDidPath)
	if err != nil {
		myLog := CustomLog{
			Status: "Error",
			Msg:    fmt.Sprintf("Failed to create domainDic file:", err),
		}
		PrintLog(myLog)
		return nil
	}
	defer file.Close()

	// Use bufio.Scanner to read the file content line by line
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	// Check for any errors that occurred during file reading
	if err := scanner.Err(); err != nil {
		myLog := CustomLog{
			Status: "Error",
			Msg:    fmt.Sprintf("Error reading the domainDic file:", err),
		}
		PrintLog(myLog)
		return nil
	}

	return lines
}

type Message struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func RefreshConfig() {
	ticker := time.Tick(3 * time.Second)
	for {
		<-ticker
		errorm := RedisClient.Ping(context.Background())
		if errorm != nil {
			GetRedisClient()
		}
		RefreshConfigNodeName := "refresh_config:" + AppConfig.System.NodeName
		exists, err := RedisClient.Exists(context.Background(), RefreshConfigNodeName)
		if err != nil {
			SlogError(fmt.Sprintf("RefreshConfig Error", err))
			continue
		}
		if exists {
			msg, err := RedisClient.PopFromListR(context.Background(), RefreshConfigNodeName)
			SlogInfo(fmt.Sprintf("recv RefreshConfig: %s", msg))
			if err != nil {
				myLog := CustomLog{
					Status: "Error",
					Msg:    fmt.Sprintf("RefreshConfig Error 2", err),
				}
				PrintLog(myLog)
				continue
			}
			jsonData := Message{}
			err2 := json.Unmarshal([]byte(msg), &jsonData)
			if err2 != nil {
				myLog := CustomLog{
					Status: "Error",
					Msg:    fmt.Sprintf("Task parse error", err),
				}
				PrintLog(myLog)
				continue
			}
			if jsonData.Name == "all" || jsonData.Name == AppConfig.System.NodeName {
				switch jsonData.Type {
				case "system":
					UpdateSystemConfig(true)
				case "subdomain":
					UpdateDomainDicConfig()
				case "dir":
					UpdateDirDicConfig()
				case "subfinder":
					UpdateSubfinderApiConfig()
				case "rad":
					UpdateRadConfig()
				case "sensitive":
					UpdateSensitive()
				case "nodeConfig":
					UpdateNode(true)
				case "project":
					UpdateProject()
				case "port":
					UpdatePort()
				case "poc":
					UpdatePoc(true)
				case "finger":
					UpdateWebFinger()
				case "notification":
					UpdateNotification()
				case "UpdateSystem":
					UpdateSystem()
				}
			}

		}
	}
}

func GetTimeNow() string {
	// 获取当前时间
	timeLocation, err := time.LoadLocation(AppConfig.System.TimeZoneName)
	if err != nil {
		ZapLog.Error(fmt.Sprintf("Error loading time zone:", err))
		return ""
	}
	currentTime := time.Now()
	var easternTime = currentTime.In(timeLocation)
	return easternTime.Format("2006-01-02 15:04:05")
}

func WriteToFile(path string, data []byte) error {
	err := ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func StartTask() {
	AppRFMutex.Lock()         // 锁定互斥锁
	defer AppRFMutex.Unlock() // 确保在函数结束时解锁
	RunningNum = RunningNum + 1
	SlogDebugLocal(fmt.Sprintf("Running start value: %d", RunningNum))
}

func EndTask() {
	AppRFMutex.Lock()         // 锁定互斥锁
	defer AppRFMutex.Unlock() // 确保在函数结束时解锁
	RunningNum = RunningNum - 1
	SlogDebugLocal(fmt.Sprintf("Running start value: %d", RunningNum))
	FinNum = FinNum + 1
	SlogDebugLocal(fmt.Sprintf("Running end value: %d", FinNum))
}

func GetRunFin() (int, int) {
	AppRFMutex.Lock()         // 锁定互斥锁
	defer AppRFMutex.Unlock() // 确保在函数结束时解锁
	return RunningNum, FinNum
}
