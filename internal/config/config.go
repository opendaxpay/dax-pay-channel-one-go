package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// 服务配置（对标 Boot application.yml）
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Tracing TracingConfig `yaml:"tracing"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type TracingConfig struct {
	SampleRatio float64 `yaml:"sample_ratio"`
	OTLPEnabled bool    `yaml:"otlp_enabled"`
}

// 加载 YAML 配置；path 为空时使用默认值
func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{Port: 20100},
		Tracing: TracingConfig{
			SampleRatio: 1.0,
			OTLPEnabled: false,
		},
	}
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 20100
	}
	return cfg, nil
}
