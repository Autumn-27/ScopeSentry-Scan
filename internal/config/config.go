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
)

// Config 结构体用来存储读取的配置
type Config struct {
	NodeName     string        `yaml:"NodeName"`
	TimeZoneName string        `yaml:"TimeZoneName"`
	Debug        bool          `yaml:"Debug"`
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

// AppConfig Global variable to hold the loaded configuration
var AppConfig Config

// LoadConfig 读取配置文件并解析
func LoadConfig(configFile string) error {
	// 尝试打开配置文件
	if _, err := os.Stat(configFile); err == nil {
		// 配置文件存在，从文件读取
		if err := utils.ReadYAMLFile(configFile, &AppConfig); err != nil {
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
		if err := createConfigFile(configFile, AppConfig); err != nil {
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
	log.Printf("Configuration file created from environment variables: %s", configFile)
	return nil
}

func Initialize(configFile string) {
	err := LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
}
