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

// Config 结构体
type Config struct {
	NodeName     string        `yaml:"NodeName"`
	TimeZoneName string        `yaml:"TimeZoneName"`
	Debug        bool          `yaml:"debug"`
	MongoDB      MongoDBConfig `yaml:"mongodb"`
	Redis        RedisConfig   `yaml:"redis"`
}

type MongoDBConfig struct {
	IP       string `yaml:"ip"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type RedisConfig struct {
	IP       string `yaml:"ip"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
}

var (
	// AbsolutePath 全局变量
	AbsolutePath string
	ConfigPath   string
	ConfigDir    string
	// AppConfig Global variable to hold the loaded configuration
	AppConfig Config
	VERSION   string
)

// LoadConfig 读取配置文件并解析
func LoadConfig() error {
	// 尝试打开配置文件
	if _, err := os.Stat(ConfigPath); err == nil {
		// 配置文件存在，从文件读取
		if err := utils.ReadYAMLFile(ConfigPath, &AppConfig); err != nil {
			return err
		}
	} else {
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
	if err := utils.WriteYAMLFile(configFile, config); err != nil {
		return err
	}
	log.Printf("Configuration file created: %s", configFile)
	return nil
}

func Initialize() {
	AbsolutePath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	ConfigDir = filepath.Join(AbsolutePath, "config")
	ConfigPath = filepath.Join(ConfigDir, "config.yaml")
	err := LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
}
