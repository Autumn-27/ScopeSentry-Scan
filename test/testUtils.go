// main-------------------------------------
// @file      : testUtils.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/24 20:15
// -------------------------------------------

package main

import (
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

type Response struct {
	Error           bool       `json:"error"`
	ConsumedFPoint  int        `json:"consumed_fpoint"`
	RequiredFPoints int        `json:"required_fpoints"`
	Size            int        `json:"size"`
	Page            int        `json:"page"`
	Mode            string     `json:"mode"`
	Query           string     `json:"query"`
	Results         [][]string `json:"results"` // results 是一个包含数组的二维数组
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
	utils.InitializeRequests()
	//resultChan := make(chan string, 100)
	//go utils.Tools.ReadFileLineReader("C:\\Users\\autumn\\AppData\\Local\\JetBrains\\GoLand2023.3\\tmp\\GoLand\\ext\\katana\\result\\3MBJIanIesNNIFps", resultChan, context.Background())
	//time.Sleep(3 * time.Second)
	//for result := range resultChan {
	//	fmt.Println(result)
	//}
	//result := make(chan string)
	//// 设置超时时间和任务上下文管理
	//go utils.Tools.ExecuteCommandToChanWithTimeout("whoami", []string{}, result, 20*time.Minute, context.Background())
	////go utils.Tools.ExecuteCommandToChan("whoami", []string{}, result)
	//for i := range result {
	//	fmt.Println(i)
	//}
	//_, b, _ := utils.Tools.GenerateIgnore("*.baidu.com\nrainy-sautyu.top")
	//for _, k := range b {
	//	flag := k.MatchString("baidu.com")
	//	fmt.Println(flag)
	//}
	//similarity, err := utils.Tools.CompareContentSimilarity("adddw", "ddddddddddddddddddddddddwww")
	//if err != nil {
	//	return
	//}
	//fmt.Println(similarity)
	//a := utils.DNS.QueryOne("baidu.com")
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
	//encoded := base64.StdEncoding.EncodeToString([]byte("host=\"baidu.com\"&&port!=\"80\"&&port!=\"443\""))
	//url := fmt.Sprintf("https://fofa.info/api/v1/search/all?&key=%v&qbase64=%v&size=%v&fields=ip,port,domain,host,protocol,icp,title,", "", encoded, 10)
	//
	//res, _ := utils.Requests.HttpGetByte(url)
	//reader := bytes.NewReader(res)
	//decoder := json.NewDecoder(reader)
	//var result Response
	//if err := decoder.Decode(&result); err != nil {
	//}
	//if result.Error {
	//
	//}
	//fmt.Printf("%v", result.Results)
}
