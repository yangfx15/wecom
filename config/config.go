package config

import (
	"encoding/json"
	"io/ioutil"
)

// Config 整个项目的配置
type Config struct {
	Mode       string `json:"mode"`
	Port       int    `json:"port"`
	*LogConfig `json:"log"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level"`
	Filename   string `json:"filename"`
	MaxSize    int    `json:"maxsize"`
	MaxAge     int    `json:"max_age"`
	MaxBackups int    `json:"max_backups"`
}

type DBConfig struct {
	Domain    string `json:"domain"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Database  string `json:"database"`
	DebugMode bool   `json:"debug_mode"`
}

// Conf 全局配置变量
var Conf = new(Config)

func InitConfig(level, name string) {
	Conf = &Config{
		Mode: "debug",
		Port: 80,
		LogConfig: &LogConfig{
			Level:      level,
			Filename:   name,
			MaxSize:    128,
			MaxAge:     7,
			MaxBackups: 30,
		},
	}
}

// Init 初始化配置；从指定文件加载配置文件
func Init(filePath string) error {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, Conf)
}
