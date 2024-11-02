// config-------------------------------------
// @file      : type.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/8 12:32
// -------------------------------------------

package global

// Config 结构体
type Config struct {
	NodeName     string        `yaml:"NodeName"`
	State        int           `yaml:"state"`
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
