// Package subdomainMode -----------------------------
// @file      : runner.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/11 19:15
// -------------------------------------------
package subdomainMode

import (
	"bufio"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/system"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/util"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func SubDomainRunner(Host string) []types.SubdomainResult {
	defer system.RecoverPanic("SubDomainRunner")
	dicFilePath := path.Join(system.ConfigDir, "domainDic")
	system.SlogInfo(fmt.Sprintf("target %v ksubdomain start", Host))
	subdomainDict, err := util.ReadFileLiness(dicFilePath)
	if err != nil {
		system.SlogError(fmt.Sprintf("open file :%s error:%s", dicFilePath, err))
	}
	dicDomainList := []string{}
	dotIndex := strings.Index(Host, "*.")
	if dotIndex != -1 {
		for _, value := range subdomainDict {
			tmpDomain := strings.Replace(Host, "*", value, -1)
			dicDomainList = append(dicDomainList, tmpDomain)
		}
	} else {
		for _, value := range subdomainDict {
			dicDomainList = append(dicDomainList, value+"."+Host)
		}
	}
	subDoaminResult := Verify(dicDomainList)
	system.SlogInfo(fmt.Sprintf("target %v ksubdomain result %v ", Host, len(subDoaminResult)))
	system.SlogInfo(fmt.Sprintf("target %v ksubdomain end", Host))
	return subDoaminResult

}

func Verify(target []string) []types.SubdomainResult {
	randomString := util.GenerateRandomString(6)
	if len(target) == 0 {
		return []types.SubdomainResult{}
	}
	filename := util.CalculateMD5(target[0] + randomString)
	targetPath := filepath.Join(system.KsubdomainPath, "target", filename)
	resultPath := filepath.Join(system.KsubdomainPath, "result", filename)
	defer util.DeleteFile(targetPath)
	defer util.DeleteFile(resultPath)

	SubdomainWriteTarget(targetPath, target)
	args := []string{"v", "-f", targetPath, "-o", resultPath}
	cmd := exec.Command(system.KsubdomainExecPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		system.SlogError(fmt.Sprintf("ksubdomain verify 执行命令时出错：%s %s\n", err, output))
		return []types.SubdomainResult{}
	}
	result := GetSubdomainResult(resultPath)
	if len(result) == 0 {
		system.SlogInfo(fmt.Sprintf("verify target[0] %v get dns result 0", target[0]))
		return []types.SubdomainResult{}
	}
	return result
}
func Verify2(target []string) []types.SubdomainResult {
	randomString := util.GenerateRandomString(6)
	if len(target) == 0 {
		return []types.SubdomainResult{}
	}
	filename := util.CalculateMD5(target[0] + randomString)
	targetPath := filepath.Join(system.KsubdomainPath, "target", filename)
	resultPath := filepath.Join(system.KsubdomainPath, "result", filename)
	defer util.DeleteFile(targetPath)
	defer util.DeleteFile(resultPath)

	SubdomainWriteTarget(targetPath, target)
	args := []string{"v", "-f", targetPath, "-o", resultPath}
	cmd := exec.Command(system.KsubdomainExecPath, args...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		//system.SlogError(fmt.Sprintf("ksubdomain verify 执行命令时出错：%s %s\n", err, output))
		return []types.SubdomainResult{}
	}
	result := GetSubdomainResult(resultPath)
	if len(result) == 0 {
		//system.SlogInfo(fmt.Sprintf("verify target[0] %v get dns result 0", target[0]))
		return []types.SubdomainResult{}
	}
	return result
}

//func enum(target []string) []types.SubdomainResult {
//	randomString := util.GenerateRandomString(6)
//	if len(target) == 0 {
//		return nil
//	}
//	filename := util.CalculateMD5(target[0] + randomString)
//	targetPath := filepath.Join(system.KsubdomainPath, "target", filename)
//	resultPath := filepath.Join(system.KsubdomainPath, "result", filename)
//	SubdomainWriteTarget(targetPath, target)
//	args := []string{"e", "-f", targetPath, "-o", resultPath}
//	cmd := exec.Command(system.KsubdomainExecPath, args...)
//	output, err := cmd.CombinedOutput()
//	if err != nil {
//		system.SlogError(fmt.Sprintf("ksubdomain enum 执行命令时出错：%s %s\n", err, output))
//	}
//	result := GetSubdomainResult(resultPath)
//	if result == nil {
//		system.SlogInfo(fmt.Sprintf("enum target %v get dns result 0", target))
//		return nil
//	}
//	return result
//}

func SubdomainWriteTarget(targetPath string, data []string) {
	file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	defer file.Close()
	if err != nil {
		system.SlogError(fmt.Sprintf("SubdomainWriteTarget error: %v", err))
		return
	}
	for _, target := range data {
		if target == "" {
			continue
		}
		_, err = file.WriteString(target + "\n")
		if err != nil {
			system.SlogError(fmt.Sprintf("SubdomainWriteTarget write target error: %v", err))
			continue
		}
	}
}

func GetSubdomainResult(resultPath string) []types.SubdomainResult {
	file, err := os.Open(resultPath)
	if err != nil {
		return []types.SubdomainResult{}
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	subDomainResult := []types.SubdomainResult{}
	for scanner.Scan() {
		line := scanner.Text()
		Domains := strings.Split(line, "=>")
		system.SlogDebugLocal(fmt.Sprintf("Received DNS message in Ksubdoamin: %v", Domains))
		_domain := types.SubdomainResult{}
		_domain.Host = Domains[0]
		_domain.Type = "A"
		for i := 1; i < len(Domains); i++ {
			containsSpace := strings.Contains(Domains[i], " ")
			if containsSpace {
				result := strings.SplitN(Domains[i], " ", 2)
				_domain.Type = result[0]
				_domain.Value = append(_domain.Value, result[1])
			} else {
				_domain.IP = append(_domain.IP, Domains[i])
			}
		}
		time := system.GetTimeNow()
		_domain.Time = time
		subDomainResult = append(subDomainResult, _domain)
	}
	return subDomainResult
}
