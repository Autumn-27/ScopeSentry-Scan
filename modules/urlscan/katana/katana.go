// katana-------------------------------------
// @file      : katana.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/11 21:33
// -------------------------------------------

package katana

import (
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/interfaces"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"github.com/projectdiscovery/katana/pkg/engine/standard"
	"github.com/projectdiscovery/katana/pkg/output"
	katanaTypes "github.com/projectdiscovery/katana/pkg/types"
	"math"
	"strconv"
	"sync"
)

type Plugin struct {
	Name      string
	Module    string
	Parameter string
	Id        string
	Result    chan interface{}
}

func NewPlugin() *Plugin {
	return &Plugin{
		Name:   "Katana",
		Module: "URLScan",
	}
}

func (p *Plugin) SetId(id string) {
	p.Id = id
}

func (p *Plugin) GetId() string {
	return p.Id
}

func (p *Plugin) SetResult(ch chan interface{}) {
	p.Result = ch
}

func (p *Plugin) SetName(name string) {
	p.Name = name
}

func (p *Plugin) GetName() string {
	return p.Name
}

func (p *Plugin) SetModule(module string) {
	p.Module = module
}

func (p *Plugin) GetModule() string {
	return p.Module
}

func (p *Plugin) Install() error {
	return nil
}

func (p *Plugin) Check() error {
	return nil
}

func (p *Plugin) SetParameter(args string) {
	p.Parameter = args
}

func (p *Plugin) GetParameter() string {
	return p.Parameter
}

func (p *Plugin) Execute(input interface{}) (interface{}, error) {
	data, ok := input.(types.AssetHttp)
	if !ok {
		logger.SlogError(fmt.Sprintf("%v error: %v input is not a string\n", p.Name, input))
		return nil, errors.New("input is not a string")
	}

	parameter := p.GetParameter()
	threads := 10
	timeout := 3
	maxDepth := 5
	if parameter != "" {
		args, err := utils.Tools.ParseArgs(parameter, "t", "timeout", "depth")
		if err != nil {
		} else {
			for key, value := range args {
				switch key {
				case "t":
					threads, _ = strconv.Atoi(value)
				case "timeout":
					timeout, _ = strconv.Atoi(value)
				case "depth":
					maxDepth, _ = strconv.Atoi(value)
				default:
					continue
				}
			}
		}
	}
	var urllist []string
	var mu sync.Mutex
	options := &katanaTypes.Options{
		MaxDepth:          maxDepth,    // Maximum depth to crawl
		FieldScope:        "rdn",       // Crawling Scope Field
		BodyReadSize:      math.MaxInt, // Maximum response size to read
		ScrapeJSResponses: true,
		ExtensionFilter:   []string{"png", "apng", "bmp", "gif", "ico", "cur", "jpg", "jpeg", "jfif", "pjp", "pjpeg", "svg", "tif", "tiff", "webp", "xbm", "3gp", "aac", "flac", "mpg", "mpeg", "mp3", "mp4", "m4a", "m4v", "m4p", "oga", "ogg", "ogv", "mov", "wav", "webm", "eot", "woff", "woff2", "ttf", "otf", "css"},
		KnownFiles:        "robotstxt,sitemapxml",
		Timeout:           timeout,       // Timeout is the time to wait for request in seconds
		Concurrency:       threads,       // Concurrency is the number of concurrent crawling goroutines
		Parallelism:       10,            // Parallelism is the number of urls processing goroutines
		Delay:             0,             // Delay is the delay between each crawl requests in seconds
		RateLimit:         150,           // Maximum requests to send per second
		Strategy:          "depth-first", // Visit strategy (depth-first, breadth-first)
		OnResult: func(result output.Result) { // Callback function to execute for result
			var r types.UrlResult
			r.Input = data.URL
			r.Source = result.Request.Source
			r.Output = result.Request.URL
			r.OutputType = result.Request.Attribute
			r.Status = result.Response.StatusCode
			r.Length = len(result.Response.Body)
			r.Body = result.Response.Body
			mu.Lock()
			urllist = append(urllist, result.Request.URL)
			mu.Unlock()
			p.Result <- r
		},
	}
	crawlerOptions, err := katanaTypes.NewCrawlerOptions(options)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("katana error %v", err.Error()))
	}
	defer crawlerOptions.Close()
	crawler, err := standard.New(crawlerOptions)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("katana standard.New error %v", err.Error()))
	}
	defer crawler.Close()
	err = crawler.Crawl(data.URL)
	if err != nil {
		logger.SlogErrorLocal(fmt.Sprintf("katana crawler.Crawl error %v: %v", input, err.Error()))
	}
	return urllist, nil
}

func (p *Plugin) Clone() interfaces.Plugin {
	return &Plugin{
		Name:   p.Name,
		Module: p.Module,
		Id:     p.Id,
	}
}
