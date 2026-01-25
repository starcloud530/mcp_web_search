package utils

import (
	"gopkg.in/yaml.v3"
	"os"
)

// Config 应用配置结构
type Config struct {
	Server ServerConfig `yaml:"server"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// LoadConfig 加载配置文件
func LoadConfig(filePath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 解析 YAML 配置
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
