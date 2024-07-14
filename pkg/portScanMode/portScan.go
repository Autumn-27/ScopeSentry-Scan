// Package portScanMode -----------------------------
// @file      : portScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/13 18:13
// -------------------------------------------
package portScanMode

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/httpxMode"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/scanResult"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"github.com/praetorian-inc/fingerprintx/pkg/plugins"
	"github.com/praetorian-inc/fingerprintx/pkg/scan"
	"net"
	"net/netip"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BannerInfo struct {
	Metadata json.RawMessage `json:"metadata"`
}

//	func PortScan(Domains string, Ports string, duplicates string) ([]types.AssetHttp, []types.AssetOther) {
//		defer system.RecoverPanic("PortScan")
//		var PortAlives []types.PortAlive
//		CallBack := func(alive []types.PortAlive) {
//			PortAlives = append(PortAlives, alive...)
//		}
//		if duplicates == "port" {
//			scanPorts := ""
//			//p := scanResult.GetPortByHost(Domains)
//			//for _, tmpP := range strings.Split(Ports, ",") {
//			//	found := false
//			//	for _, port := range p {
//			//		if port == tmpP {
//			//			found = true
//			//			break
//			//		}
//			//	}
//			//	if !found {
//			//		scanPorts += tmpP + ","
//			//	}
//			//}
//			//if len(scanPorts) != 0 {
//			//	scanPorts = scanPorts[:len(scanPorts)-1]
//			//}
//			var err = NaabuScan([]string{Domains}, scanPorts, CallBack)
//			if err != nil {
//				system.SlogError(fmt.Sprintf("PortScan error: %v", err.Error()))
//			}
//		} else {
//			var err = NaabuScan([]string{Domains}, Ports, CallBack)
//			if err != nil {
//				system.SlogError(fmt.Sprintf("PortScan error: %v", err.Error()))
//			}
//		}
//
//		var targets []plugins.Target
//		for _, value := range PortAlives {
//			ip, _ := netip.ParseAddr(value.IP)
//			p, _ := strconv.Atoi(value.Port)
//			target := plugins.Target{
//				Address: netip.AddrPortFrom(ip, uint16(p)),
//				Host:    value.Host,
//			}
//			targets = append(targets, target)
//		}
//
//		// setup the scan system (mirrors command line options)
//		fxConfig := scan.Config{
//			DefaultTimeout: time.Duration(2) * time.Second,
//			FastMode:       false,
//			Verbose:        false,
//			UDP:            false,
//		}
//
//		// run the scan
//		results, err := scan.ScanTargets(targets, fxConfig)
//		if err != nil {
//			system.SlogError(fmt.Sprintf("PortScan error: %s\n", err))
//		}
//
//		var httpxResults []types.AssetHttp
//		var httpxResultsMutex sync.Mutex
//		httpxResultsHandler := func(r types.AssetHttp) {
//			//fmt.Printf("Result in process: %s %s %d\n", r.Host, r.StatusCode)
//			httpxResultsMutex.Lock()
//			httpxResults = append(httpxResults, r)
//			httpxResultsMutex.Unlock()
//		}
//		urlList := []string{}
//		assetOthers := []types.AssetOther{}
//		// process the results
//		NotificationMsg := "PortScan Result:\n"
//		for _, result := range results {
//			NotificationMsg += fmt.Sprintf("%v:%v - %v\n", result.Host, result.Port, result.Protocol)
//			if result.Protocol == "http" || result.Protocol == "https" {
//				portStr := strconv.Itoa(result.Port)
//				url := result.Protocol + "://" + result.Host
//				if portStr != "80" && portStr != "443" {
//					url += ":" + portStr
//				}
//				urlList = append(urlList, url)
//			} else {
//				assetedOther := types.AssetOther{
//					Host:      result.Host,
//					IP:        result.IP,
//					Port:      strconv.Itoa(result.Port),
//					Protocol:  result.Protocol,
//					TLS:       result.TLS,
//					Transport: result.Transport,
//					Version:   result.Version,
//					Raw:       result.Raw,
//					Type:      "other",
//					Timestamp: system.GetTimeNow(),
//					// Add additional fields specific to AssetOther if needed
//				}
//				assetOthers = append(assetOthers, assetedOther)
//			}
//
//		}
//		if len(urlList) != 0 {
//			httpxMode.HttpxScan(urlList, httpxResultsHandler)
//		}
//		assetOthersTemp := []types.AssetOther{}
//		for _, portAlive := range PortAlives {
//			found := false
//			for _, rs := range results {
//				if rs.Host == portAlive.Host && strconv.Itoa(rs.Port) == portAlive.Port {
//					found = true
//					break
//				}
//			}
//			if !found {
//				assetedOther, err := handleTcpConnection(portAlive.Host, portAlive.IP, portAlive.Port)
//				if err != nil {
//					continue
//				}
//				NotificationMsg += fmt.Sprintf("%v:%v\n", assetedOther.Host, assetedOther.Port)
//				assetOthersTemp = append(assetOthersTemp, assetedOther)
//			}
//		}
//		assetOthers = append(assetOthers, assetOthersTemp...)
//		if system.NotificationConfig.PortScanNotification && len(results) != 0 {
//			go system.SendNotification(NotificationMsg)
//		}
//		return httpxResults, assetOthers
//
// }
func handleTcpConnection(host string, ip string, port string) (types.AssetOther, error) {
	assetedOther := types.AssetOther{
		Host:      host,
		IP:        ip,
		Port:      port,
		Protocol:  "unknown",
		TLS:       false,
		Transport: "",
		Version:   "",
		Type:      "other",
		Timestamp: system.GetTimeNow(),
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", ip, port), 5*time.Second)
	if err != nil {
		return assetedOther, err
	}
	defer conn.Close()
	err = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		return assetedOther, err
	}
	reader := bufio.NewReader(conn)
	bannerBytes, err := reader.ReadBytes('\n')
	if err != nil {
		return assetedOther, err
	}
	assetedOther.Raw = bannerBytes
	return assetedOther, nil
}

func PortScan2(domain string, ports string, duplicates string) ([]types.AssetHttp, []types.AssetOther) {
	var excludePorts string
	excludePorts = ""
	if duplicates == "port" {
		excludePorts = scanResult.GetPortByHost(domain)
	}
	portResultChan := make(chan string, 100)
	var wg sync.WaitGroup
	wg.Add(1)
	go RustScan(domain, ports, excludePorts, portResultChan, &wg)
	var httpxResults []types.AssetHttp
	var httpxResultsMutex sync.Mutex
	httpxResultsHandler := func(r types.AssetHttp) {
		//fmt.Printf("Result in process: %s %s %d\n", r.Host, r.StatusCode)
		httpxResultsMutex.Lock()
		httpxResults = append(httpxResults, r)
		httpxResultsMutex.Unlock()
	}
	urlList := []string{}
	assetOthers := []types.AssetOther{}
	fxConfig := scan.Config{
		DefaultTimeout: time.Duration(3) * time.Second,
		FastMode:       false,
		Verbose:        false,
		UDP:            false,
	}
	var PortAlives []types.PortAlive
	NotificationMsg := "PortScan Result:\n"
	fingerPortFound := []string{}
	for result := range portResultChan {
		portParts := strings.SplitN(result, ":", 2)
		if len(portParts) == 2 {
			ip, _ := netip.ParseAddr(portParts[0])
			p, _ := strconv.Atoi(portParts[1])
			PortAlives = append(PortAlives, types.PortAlive{
				Port: portParts[1],
				IP:   portParts[0],
				Host: domain,
			})
			target := plugins.Target{
				Address: netip.AddrPortFrom(ip, uint16(p)),
				Host:    domain,
			}
			targets := make([]plugins.Target, 1)
			targets = append(targets, target)
			wg.Add(1)
			go func(targets []plugins.Target) {
				defer wg.Done()
				fingerResults, err := scan.ScanTargets(targets, fxConfig)
				if err != nil {
					system.SlogError(fmt.Sprintf("PortScan error: %s\n", err))
				}
				for _, fingerResult := range fingerResults {
					NotificationMsg += fmt.Sprintf("%v:%v - %v\n", fingerResult.Host, fingerResult.Port, fingerResult.Protocol)
					fingerPort := strconv.Itoa(fingerResult.Port)
					fingerPortFound = append(fingerPortFound, fingerPort)
					if fingerResult.Protocol == "http" || fingerResult.Protocol == "https" {
						portStr := fingerPort
						url := fingerResult.Protocol + "://" + fingerResult.Host
						if portStr != "80" && portStr != "443" {
							url += ":" + portStr
						}
						urlList = append(urlList, url)
					} else {
						assetedOther := types.AssetOther{
							Host:      fingerResult.Host,
							IP:        fingerResult.IP,
							Port:      fingerPort,
							Protocol:  fingerResult.Protocol,
							TLS:       fingerResult.TLS,
							Transport: fingerResult.Transport,
							Version:   fingerResult.Version,
							Raw:       fingerResult.Raw,
							Type:      "other",
							Timestamp: system.GetTimeNow(),
							// Add additional fields specific to AssetOther if needed
						}
						assetOthers = append(assetOthers, assetedOther)
					}

				}
			}(targets)
		} else {
			system.SlogError(fmt.Sprintf("%v portscan parse error:%v", domain, result))
		}
	}
	if len(urlList) != 0 {
		httpxMode.HttpxScan(urlList, httpxResultsHandler)
	}
	assetOthersTemp := []types.AssetOther{}
	for _, portAlive := range PortAlives {
		found := false
		for _, rs := range fingerPortFound {
			if rs == portAlive.Port {
				found = true
				break
			}
		}
		if !found {
			assetedOther, err := handleTcpConnection(portAlive.Host, portAlive.IP, portAlive.Port)
			if err != nil {
				continue
			}
			NotificationMsg += fmt.Sprintf("%v:%v\n", assetedOther.Host, assetedOther.Port)
			assetOthersTemp = append(assetOthersTemp, assetedOther)
		}
	}
	assetOthers = append(assetOthers, assetOthersTemp...)
	if system.NotificationConfig.PortScanNotification && len(fingerPortFound) != 0 {
		go system.SendNotification(NotificationMsg)
	}
	wg.Wait()
	return httpxResults, assetOthers
}

func RustScan(domain string, ports string, exclude string, portResultChan chan<- string, wg *sync.WaitGroup) {
	system.SlogDebugLocal("RustScan start")
	defer wg.Done()
	defer close(portResultChan)
	PortBatchSize := system.AppConfig.System.PortBatchSize
	if PortBatchSize == "" {
		PortBatchSize = "600"
	}
	PortTimeOut := system.AppConfig.System.PortTimeOut
	if PortTimeOut == "" {
		PortTimeOut = "3000"
	}
	var cmd *exec.Cmd
	args := []string{"-b", PortBatchSize, "-t", PortTimeOut, "-a", domain, "-r", ports, "--accessible", "--scripts", "None"}
	if exclude != "" {
		args = append(args, "-e")
		args = append(args, exclude)
	}
	rustScanExecPath := system.RustScanExecPath
	cmd = exec.Command(rustScanExecPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		system.SlogError(fmt.Sprintf("RustScan StdoutPipe error： %v", err))
		return
	}
	if err := cmd.Start(); err != nil {
		system.SlogError(fmt.Sprintf("RustScan cmd.Start error： %v", err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1800*time.Second)
	defer cancel()
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			system.SlogError("RustScan Command execution timed out")
			return
		default:
			r := scanner.Text()
			system.SlogDebugLocal(r)
			if strings.Contains(r, "Open") {
				openParts := strings.SplitN(r, " ", 2)
				portResultChan <- openParts[1]
				continue
			}
			if strings.Contains(r, "->") {
				system.SlogInfo(fmt.Sprintf("%v Port alive: %v", domain, r))
				continue
			}
			system.SlogError(fmt.Sprintf("%v PortScan error: %v", domain, r))
		}
	}
	if err := scanner.Err(); err != nil {
		system.SlogError(fmt.Sprintf("%v RustScan scanner.Err error： %v", domain, err))
	}
	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		system.SlogError(fmt.Sprintf("%v RustScan cmd.Wait error： %v", domain, err))
	}
	system.SlogDebugLocal("RustScan end")
}
