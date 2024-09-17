// config-------------------------------------
// @file      : config.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 22:01
// -------------------------------------------

package config

import (
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"log"
	"os"
	"path/filepath"
)

// LoadConfig 读取配置文件并解析
func LoadConfig() error {
	// 尝试打开配置文件
	if _, err := os.Stat(ConfigPath); err == nil {
		FirstRun = false
		// 配置文件存在，从文件读取
		if err := utils.Tools.ReadYAMLFile(ConfigPath, &AppConfig); err != nil {
			return err
		}
	} else {
		FirstRun = true
		// 配置文件不存在，从环境变量读取
		AppConfig = Config{
			NodeName:     getEnv("NodeName", ""),
			TimeZoneName: getEnv("TimeZoneName", "Asia/Shanghai"),
			MongoDB: MongoDBConfig{
				IP:       getEnv("MONGODB_IP", ""),
				Port:     getEnv("MONGODB_PORT", "27017"),
				User:     getEnv("MONGODB_USER", ""),
				Password: getEnv("MONGODB_PASSWORD", ""),
				Database: getEnv("MONGODB_DATABASE", ""),
			},
			Redis: RedisConfig{
				IP:       getEnv("REDIS_IP", ""),
				Port:     getEnv("REDIS_PORT", "6379"),
				Password: getEnv("REDIS_PASSWORD", ""),
			},
		}
		// 创建配置文件
		if err := createConfigFile(ConfigPath, AppConfig); err != nil {
			return err
		}
		if AppConfig.MongoDB.IP == "" {
			return fmt.Errorf("missing required MongoDB IP configuration")
		}

	}
	return nil
}

// getEnv 从环境变量中读取配置值，如果环境变量未设置，则返回默认值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func createConfigFile(configFile string, config Config) error {
	if err := utils.Tools.WriteYAMLFile(configFile, config); err != nil {
		return err
	}
	log.Printf("Configuration file created: %s", configFile)
	return nil
}

func CreateDir() {
	dirs := []string{
		ConfigDir,
		filepath.Join(ConfigDir, "dir"),
		filepath.Join(ConfigDir, "subdomain"),
		DictPath,
		ExtDir,
		filepath.Join(ExtDir, "rad"),
		PocDir,
		filepath.Join(AbsolutePath, "data"),
	}

	for _, dir := range dirs {
		err := EnsureDir(dir)
		if err != nil {
			log.Fatalf("%s create error: %v", dir, err)
		}
	}

}

func Initialize() {
	AbsolutePath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	ConfigDir = filepath.Join(AbsolutePath, "config")
	ConfigPath = filepath.Join(ConfigDir, "config.yaml")
	ModulesConfigPath = filepath.Join(ConfigDir, "modules.yaml")
	DictPath = filepath.Join(AbsolutePath, "dictionaries")
	ExtDir = filepath.Join(AbsolutePath, "ext")
	PocDir = filepath.Join(AbsolutePath, "poc")
	CreateDir()
	err := LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
}
