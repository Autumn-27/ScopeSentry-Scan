// Package portScanMode -----------------------------
// @file      : portScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/13 18:13
// -------------------------------------------
package portScanMode

import (
	"bufio"
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
	"strconv"
	"strings"
	"sync"
	"time"
)

type BannerInfo struct {
	Metadata json.RawMessage `json:"metadata"`
}

func PortScan(Domains string, Ports string, duplicates string) ([]types.AssetHttp, []types.AssetOther) {
	defer system.RecoverPanic("PortScan")
	var PortAlives []types.PortAlive
	CallBack := func(alive []types.PortAlive) {
		PortAlives = append(PortAlives, alive...)
	}
	if duplicates == "port" {
		scanPorts := ""
		p := scanResult.GetPortByHost(Domains)
		for _, tmpP := range strings.Split(Ports, ",") {
			found := false
			for _, port := range p {
				if port == tmpP {
					found = true
					break
				}
			}
			if !found {
				scanPorts += tmpP + ","
			}
		}
		if len(scanPorts) != 0 {
			scanPorts = scanPorts[:len(scanPorts)-1]
		}
		var err = NaabuScan([]string{Domains}, scanPorts, CallBack)
		if err != nil {
			system.SlogError(fmt.Sprintf("PortScan error: %v", err.Error()))
		}
	} else {
		var err = NaabuScan([]string{Domains}, Ports, CallBack)
		if err != nil {
			system.SlogError(fmt.Sprintf("PortScan error: %v", err.Error()))
		}
	}

	var targets []plugins.Target
	for _, value := range PortAlives {
		ip, _ := netip.ParseAddr(value.IP)
		p, _ := strconv.Atoi(value.Port)
		target := plugins.Target{
			Address: netip.AddrPortFrom(ip, uint16(p)),
			Host:    value.Host,
		}
		targets = append(targets, target)
	}

	// setup the scan system (mirrors command line options)
	fxConfig := scan.Config{
		DefaultTimeout: time.Duration(2) * time.Second,
		FastMode:       false,
		Verbose:        false,
		UDP:            false,
	}

	// run the scan
	results, err := scan.ScanTargets(targets, fxConfig)
	if err != nil {
		system.SlogError(fmt.Sprintf("PortScan error: %s\n", err))
	}

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
	// process the results
	NotificationMsg := "PortScan Result:\n"
	for _, result := range results {
		NotificationMsg += fmt.Sprintf("%v:%v - %v\n", result.Host, result.Port, result.Protocol)
		if result.Protocol == "http" || result.Protocol == "https" {
			portStr := strconv.Itoa(result.Port)
			url := result.Protocol + "://" + result.Host
			if portStr != "80" && portStr != "443" {
				url += ":" + portStr
			}
			urlList = append(urlList, url)
		} else {
			assetedOther := types.AssetOther{
				Host:      result.Host,
				IP:        result.IP,
				Port:      strconv.Itoa(result.Port),
				Protocol:  result.Protocol,
				TLS:       result.TLS,
				Transport: result.Transport,
				Version:   result.Version,
				Raw:       result.Raw,
				Type:      "other",
				Timestamp: system.GetTimeNow(),
				// Add additional fields specific to AssetOther if needed
			}
			assetOthers = append(assetOthers, assetedOther)
		}

	}
	if len(urlList) != 0 {
		httpxMode.HttpxScan(urlList, httpxResultsHandler)
	}
	assetOthersTemp := []types.AssetOther{}
	for _, portAlive := range PortAlives {
		found := false
		for _, rs := range results {
			if rs.Host == portAlive.Host && strconv.Itoa(rs.Port) == portAlive.Port {
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
	if system.NotificationConfig.PortScanNotification && len(results) != 0 {
		go system.SendNotification(NotificationMsg)
	}
	return httpxResults, assetOthers

}
func handleTcpConnection(host string, ip string, port string) (types.AssetOther, error) {
	assetedOther := types.AssetOther{
		Host:      host,
		IP:        ip,
		Port:      port,
		Protocol:  "",
		TLS:       false,
		Transport: "",
		Version:   "",
		Type:      "other",
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
