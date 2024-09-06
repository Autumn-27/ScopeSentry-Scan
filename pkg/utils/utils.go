// utils-------------------------------------
// @file      : utils.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 22:34
// -------------------------------------------

package utils

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

// ReadYAMLFile 读取 YAML 文件并将其解析为目标结构体
func ReadYAMLFile(filePath string, target interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(byteValue, target)
	if err != nil {
		return err
	}

	return nil
}

// WriteYAMLFile 将目标结构体序列化为 YAML 并写入到文件
func WriteYAMLFile(filePath string, data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, yamlData, 0644)
	if err != nil {
		return err
	}

	return nil
}
