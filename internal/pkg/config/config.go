package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Logger   LogConfig      `mapstructure:"logger"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type LogConfig struct {
	Mode       string
	Level      string
	FilePath   string // 日志文件路径
	MaxSize    int    // 单个文件最大尺寸 (MB)
	MaxBackups int    // 最大备份数量
	MaxAge     int    // 最大保留天数
	Compress   bool   // 是否压缩
}

// 全局配置变量
var Cfg *Config

// LoadConfig 从 configs/config.yaml 加载配置
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config") // 配置文件名 (不带后缀)
	viper.SetConfigType("yaml")   // 配置文件类型
	viper.AddConfigPath(".")      // 配置文件路径

	// 读取配置
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	// 将读取的配置反序列化到结构体中
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	Cfg = &cfg
	return &cfg, nil
}
