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
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"os"
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
				CrawlerTimeout = 30
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(CrawlerTimeout)*time.Minute)
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
			system.SlogInfoLocal(fmt.Sprintf("crawler begin %v", targetFileName))
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
				system.SlogError(fmt.Sprintf("%v CrawlerScan 获取命令标准输出管道错误：%s", targetFileName, err))
				return
			}
			stderrPipe, err := cmd.StderrPipe()
			if err != nil {
				system.SlogError(fmt.Sprintf("%v CrawlerScan 获取命令标准错误输出管道错误：%s", targetFileName, err))
				return
			}
			if err := cmd.Start(); err != nil {
				system.SlogError(fmt.Sprintf("%v CrawlerScan 启动命令错误：%s", targetFileName, err))
				return
			}
			if system.AppConfig.System.Debug {
				// 使用带缓冲的读取器来实时获取输出
				stdoutScanner := bufio.NewScanner(stdoutPipe)
				// 循环读取标准输出
				go func() {
					for stdoutScanner.Scan() {
						outputLine := stdoutScanner.Text()
						fmt.Println(outputLine)
					}
				}()
			}
			stderrScanner := bufio.NewScanner(stderrPipe)
			for stderrScanner.Scan() {
				// 处理标准错误输出
				errorLine := stderrScanner.Text()
				system.SlogErrorLocal(errorLine)
			}
			if err := cmd.Wait(); err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					system.SlogDebugLocal(fmt.Sprintf("%v CrawlerScan 超时退出：%s", targetFileName, err))
				} else {
					system.SlogError(fmt.Sprintf("%v CrawlerScan 执行命令时出错：%s", targetFileName, err))
				}
			}

			// Read the content of the file
			//resultContent, err := ioutil.ReadFile(resultPath)
			//if err != nil {
			//	system.SlogInfo(fmt.Sprintf("%v CrawlerScan read result 0: %s", targetFileName, err))
			//	return
			//}
			//var requests []Request
			//
			//// Unmarshal the JSON data into the slice
			//err = json.Unmarshal(resultContent, &requests)
			//if err != nil {
			//	system.SlogErrorLocal(fmt.Sprintf("%v CrawlerScan parse json err: %s", targetFileName, err))
			//	return
			//}
			defer util.DeleteFile(targetPath)
			defer util.DeleteFile(resultPath)

			file, err := os.Open(resultPath)
			if err != nil {
				if os.IsNotExist(err) {
					system.SlogDebugLocal(fmt.Sprintf("爬虫无结果：文件 %v 不存在: %v", resultPath, err))
				} else {
					system.SlogDebugLocal(fmt.Sprintf("无法打开文件 %v: %v", resultPath, err))
				}
				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if line == "[" || line == "]" {
					continue
				}
				line = strings.TrimRight(line, ",")
				var req Request
				err := json.Unmarshal([]byte(line), &req)
				if err != nil {
					system.SlogError(fmt.Sprintf("解析 JSON 错误: %s", err))
					continue
				}
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
						system.SlogDebugLocal(fmt.Sprintf("%v Get CrawlerScan Result Dedupli: %s %s ", targetFileName, req.Method, req.URL))
						continue
					}
				}
				system.SlogDebugLocal(fmt.Sprintf("%v Get CrawlerScan Result: %s %s ", targetFileName, req.Method, req.URL))
				crawlerResult := types.CrawlerResult{
					Url:    req.URL,
					Method: req.Method,
					Body:   body,
				}
				scanResult.CrawlerResult([]types.CrawlerResult{crawlerResult}, task.Id)
			}

			if err := scanner.Err(); err != nil {
				system.SlogError(fmt.Sprintf("读取文件错误: %v", err))
				return
			}
			system.SlogInfoLocal(fmt.Sprintf("crawler end %v", targetFileName))
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
