// Package portScanMode -----------------------------
// @file      : nabbuScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/13 20:32
// -------------------------------------------
package portScanMode

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/types"
	"github.com/projectdiscovery/naabu/v2/pkg/result"
	"github.com/projectdiscovery/naabu/v2/pkg/runner"
	"strconv"
)

func NaabuScan(Domain []string, Ports string, Callbak func(alive []types.PortAlive)) error {
	defer system.RecoverPanic("NaabuScan")
	PortscanThread, err := strconv.Atoi(system.AppConfig.System.PortscanThread)
	if err != nil {
		PortscanThread = 10
	}
	options := runner.Options{
		Host:       Domain,
		Ports:      Ports,
		Passive:    true,
		ExcludeCDN: true,
		Retries:    1,
		Threads:    PortscanThread,
		OnResult: func(hr *result.HostResult) {
			fmt.Println(hr.Host, hr.IP, hr.Ports)
			PortAlives := []types.PortAlive{}
			for _, p := range hr.Ports {
				portAlive := types.PortAlive{}
				portAlive.Host = hr.Host
				portAlive.Port = strconv.Itoa(p.Port)
				portAlive.IP = hr.IP
				PortAlives = append(PortAlives, portAlive)
			}
			Callbak(PortAlives)

		},
	}

	naabuRunner, err := runner.NewRunner(&options)
	if err != nil {
		return err
	}
	defer naabuRunner.Close()

	return naabuRunner.RunEnumeration()
}
