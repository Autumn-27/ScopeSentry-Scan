// Package crawlerMode -----------------------------
// @file      : crawlerScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/17 22:48
// -------------------------------------------
package crawlerMode

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Request struct {
	Method  string `json:"Method"`
	URL     string `json:"URL"`
	B64Body string `json:"b64_body,omitempty"`
}

func CrawlerScan(tasks <-chan types.CrawlerTask) {
	defer system.RecoverPanic("CrawlerScan")
	//system.SetUp()

	for task := range tasks {
		if task.Id == "STOP Crawler" {
			system.SlogInfo("减少一个爬虫扫描线程")
			return
		}
		fileContent := ""
		for _, target := range task.Target {
			fileContent += target + "\n"
		}
		var wg sync.WaitGroup
		wg.Add(1)
		system.SlogInfo(fmt.Sprintf("crawler target %v", len(task.Target)))
		go func(fileContent string, task types.CrawlerTask) {
			defer task.Wg.Done()
			defer wg.Done()
			if fileContent == "" {
				return
			}
			CrawlerTimeout, err := strconv.Atoi(system.AppConfig.System.CrawlerTimeout)
			if err != nil {
				CrawlerTimeout = 1
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(CrawlerTimeout)*time.Hour)
			defer cancel()
			radModePath := system.CrawlerPath
			radPath := system.CrawlerExecPath
			command := radPath
			timeRandom := system.GetTimeNow()
			strRandom := util.GenerateRandomString(8)
			targetFileName := util.CalculateMD5(timeRandom + strRandom)
			targetPath := filepath.Join(radModePath, "target", targetFileName)
			resultPath := filepath.Join(radModePath, "result", targetFileName)
			radConfigPath := filepath.Join(radModePath, "rad_config.yml")
			flag := util.WriteContentFile(targetPath, fileContent)
			if !flag {
				system.SlogError(fmt.Sprintf("Write target file error"))
			}
			args := []string{"--url-file", targetPath, "--json", resultPath, "--config", radConfigPath}
			// 执行命令
			//cmd := exec.Command(command, args...)
			//output, err := cmd.CombinedOutput()
			//if err != nil {
			//	myLog := system.CustomLog{
			//		Status: "Error",
			//		Msg:    fmt.Sprintf("执行命令时出错：%s %s\n", err, output),
			//	}
			//	system.PrintLog(myLog)
			//}
			// 使用带超时的上下文执行命令
			cmd := exec.CommandContext(ctx, command, args...)

			// 创建管道以捕获命令输出
			stdoutPipe, err := cmd.StdoutPipe()
			if err != nil {
				system.SlogError(fmt.Sprintf("CrawlerScan 获取命令标准输出管道错误：%s", err))
				return
			}
			stderrPipe, err := cmd.StderrPipe()
			if err != nil {
				system.SlogError(fmt.Sprintf("CrawlerScan 获取命令标准错误输出管道错误：%s", err))
				return
			}
			if err := cmd.Start(); err != nil {
				system.SlogError(fmt.Sprintf("CrawlerScan 启动命令错误：%s", err))
				return
			}
			if system.AppConfig.System.Debug {
				// 使用带缓冲的读取器来实时获取输出
				stdoutScanner := bufio.NewScanner(stdoutPipe)
				stderrScanner := bufio.NewScanner(stderrPipe)

				// 循环读取标准输出
				go func() {
					for stdoutScanner.Scan() {
						outputLine := stdoutScanner.Text()
						fmt.Println(outputLine)
					}
				}()

				// 循环读取标准错误输出
				go func() {
					for stderrScanner.Scan() {
						// 处理标准错误输出
						errorLine := stderrScanner.Text()
						fmt.Println(errorLine)
					}
				}()
			}
			if err := cmd.Wait(); err != nil {
				system.SlogError(fmt.Sprintf("CrawlerScan 执行命令时出错：%s", err))
				return
			}
			// Read the content of the file
			resultContent, err := ioutil.ReadFile(resultPath)
			if err != nil {
				system.SlogInfo(fmt.Sprintf("CrawlerScan read result 0: %s", err))
				return
			}
			var requests []Request

			// Unmarshal the JSON data into the slice
			err = json.Unmarshal(resultContent, &requests)
			if err != nil {
				system.SlogErrorLocal(fmt.Sprintf("CrawlerScan parse json err: %s", err))
				return
			}
			defer util.DeleteFile(targetPath)
			defer util.DeleteFile(resultPath)
			var CrawlerResults []types.CrawlerResult
			// Print the parsed JSON data
			for _, req := range requests {
				body := ""
				if req.B64Body != "" {
					decodedBytes, err := base64.StdEncoding.DecodeString(req.B64Body)
					if err != nil {
						fmt.Println(err)
					}
					body = string(decodedBytes)
				}
				key := ""
				if req.Method == "GET" {
					key = req.URL
				} else {
					if body != "" {
						bodyKeyV := strings.Split(body, "&")
						for _, part := range bodyKeyV {
							bodyKey := strings.Split(part, "=")
							if len(bodyKey) > 1 {
								key += bodyKey[0]
							}
						}
					}
				}
				if key == "" {
					key = req.URL
				}
				if key != "" {
					dFlag := scanResult.CrawRedisDeduplication(key, task.Id)
					if dFlag {
						system.SlogDebugLocal(fmt.Sprintf("Get CrawlerScan Result Dedupli: %s %s ", req.Method, req.URL))
						continue
					}
				}
				system.SlogDebugLocal(fmt.Sprintf("Get CrawlerScan Result: %s %s ", req.Method, req.URL))
				crawlerResult := types.CrawlerResult{
					Url:    req.URL,
					Method: req.Method,
					Body:   body,
				}
				CrawlerResults = append(CrawlerResults, crawlerResult)
			}

			scanResult.CrawlerResult(CrawlerResults)

		}(fileContent, task)
		time.Sleep(5 * time.Second)
		wg.Wait()
		system.SlogInfo(fmt.Sprintf("target %s crawler scan completed", task.Host))
		scanResult.ProgressEnd("crawler", task.Host, task.Id)
	}
}

var mu sync.Mutex

func CrawlerThread(flag <-chan bool) {
	for f := range flag {
		system.SlogDebugLocal(fmt.Sprintf("接收到更新CrawlerThread: %v", f))
		oldCrawlerThread := system.CrawlerThreadNow
		newCrawlerThread, err := strconv.Atoi(system.AppConfig.System.CrawlerThread)
		if err != nil {
			fmt.Println("err:", err)
			return
		}
		if oldCrawlerThread != newCrawlerThread {
			mu.Lock()
			diff := newCrawlerThread - oldCrawlerThread
			if newCrawlerThread > oldCrawlerThread {
				for i := 0; i < diff; i++ {
					go CrawlerScan(system.CrawlerTarget)
					system.SlogInfo("创建一个新的爬虫扫描线程")
				}
			} else {
				for i := 0; i < -diff; i++ {
					system.CrawlerTarget <- types.CrawlerTask{Id: "STOP Crawler"} // 发送一个停止信号来停止一个goroutine
				}
			}
			system.CrawlerThreadNow = newCrawlerThread
			mu.Unlock()
		}
	}
}
