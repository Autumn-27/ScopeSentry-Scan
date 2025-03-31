// main-------------------------------------
// @file      : testSens.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/7/28 18:17
// -------------------------------------------

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/bigcache"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/configupdater"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/contextmanager"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/handler"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/notification"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pebbledb"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/plugins"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/results"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/modules/urlsecurity/trufflehog"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"net/url"
	"path"
	"path/filepath"
	"time"
)

func main() {
	DebugInit()
	plg := trufflehog.NewPlugin()
	plg.Install()
	//data, _ := os.ReadFile("C:\\Users\\autumn\\Desktop\\PushSdk-8.3.68.0 (1).zip")
	//input := types.UrlResult{
	//	Output: "httpx://test.com",
	//	Body:   string(data),
	//	Status: 200,
	//}
	//plg.Execute(input)

	resultChan := make(chan string, 100)
	go func() {
		err := utils.Tools.ReadFileLineReader("C:\\Users\\autumn\\AppData\\Local\\JetBrains\\GoLand2023.3\\tmp\\GoLand\\ext\\katana\\NlxJtbZYo3wwIrW9", resultChan, context.Background())
		if err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("ReadFileLineReader %v", err))
		}
	}()
	time.Sleep(2 * time.Second)
	var katanaResult types.KatanaResult
	for result := range resultChan {
		err := json.Unmarshal([]byte(result), &katanaResult)
		if err != nil {
			continue
		}
		var r types.UrlResult
		parsedURL, err := url.Parse(katanaResult.Request.URL)
		urlPath := ""
		if err != nil {
			urlPath = katanaResult.Request.URL
		} else {
			urlPath = parsedURL.Path
		}
		r.Ext = path.Ext(urlPath)
		r.Ext = path.Ext(parsedURL.Path)
		r.Input = "http://dwasdwadwa"
		r.Source = katanaResult.Request.Source
		r.Output = katanaResult.Request.URL
		r.OutputType = katanaResult.Request.Attribute
		r.Status = katanaResult.Response.StatusCode
		r.Length = len(katanaResult.Response.Body)
		r.Body = katanaResult.Response.Body
		r.Time = utils.Tools.GetTimeNow()
		if err != nil {
		}
		plg.Execute(r)
	}

}

func DebugInit() {
	// 初始化系统信息
	config.Initialize()
	var err error
	// 初始化mongodb连接
	mongodb.Initialize()
	// 初始化redis连接
	redis.Initialize()
	// 初始化日志模块
	err = logger.NewLogger()
	if err != nil {
	}
	// 初始化任务计数器
	handler.InitHandle()
	// 更新配置、加载字典
	configupdater.Initialize()
	// 初始化模块配置
	err = config.ModulesInitialize()
	if err != nil {
		return
	}
	// 初始化上下文管理器
	contextmanager.NewContextManager()
	// 初始化tools
	utils.InitializeTools()
	utils.InitializeDnsTools()
	utils.InitializeRequests()
	utils.InitializeResults()
	utils.InitializeProxyRequestsPool()
	// 初始化通知模块
	notification.InitializeNotification()
	// 初始化协程池
	pool.Initialize()
	// 初始化个模块的协程池
	pool.PoolManage.InitializeModulesPools(config.ModulesConfig)
	go pool.StartMonitoring()
	// 初始化内存缓存
	err = bigcache.Initialize()
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("bigcache Initialize error: %v", err))
		return
	}
	// 初始化本地持久化缓存
	pebbledbSetting := pebbledb.Settings{
		DBPath:       filepath.Join(global.AbsolutePath, "data", "pebbledb"),
		CacheSize:    64 << 20,
		MaxOpenFiles: 500,
	}
	pebbledbOption := pebbledb.GetPebbleOptions(&pebbledbSetting)
	if !global.AppConfig.Debug {
		pebbledbOption.Logger = nil
	}
	pedb, err := pebbledb.NewPebbleDB(pebbledbOption, pebbledbSetting.DBPath)
	if err != nil {
		return
	}
	pebbledb.PebbleStore = pedb
	defer func(PebbleStore *pebbledb.PebbleDB) {
		_ = PebbleStore.Close()
	}(pebbledb.PebbleStore)

	// 初始化结果处理队列，(正常的时候将该初始化放入任务开始时，任务执行完毕关闭结果队列)
	results.InitializeResultQueue()
	defer results.Close()

	// 初始化全局插件管理器
	plugins.GlobalPluginManager = plugins.NewPluginManager()
	err = plugins.GlobalPluginManager.InitializePlugins()
	if err != nil {
		return
	}
}
