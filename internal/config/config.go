package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// 服务配置（对标 Boot application.yml）
type Config struct {
	Server  ServerConfig    `yaml:"server"`
	Tracing TracingConfig   `yaml:"tracing"`
	Channel ChannelConfig   `yaml:"channel"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type TracingConfig struct {
	SampleRatio float64 `yaml:"sample_ratio"`
	OTLPEnabled bool    `yaml:"otlp_enabled"`
}

// ChannelConfig：通道子应用配置（传输加密等）
type ChannelConfig struct {
	TransportEncrypt TransportEncryptConfig `yaml:"transport-encrypt"`
}

// TransportEncryptConfig：主应用 ↔ 子应用 报文透明加密配置
//
// 对标 Java ChannelTransportEncryptProperties（prefix: daxpay.channel.transport-encrypt）。
// key 须恰好 32 字节 UTF-8 字符（AES-256），由 transport.NewEncryptor 校验。
type TransportEncryptConfig struct {
	Key string `yaml:"key"`
}

// 传输加密密钥环境变量名（生产环境强制注入，对标 Java CHANNEL_TRANSPORT_KEY）
const EnvTransportEncryptKey = "CHANNEL_TRANSPORT_KEY"

// 加载 YAML 配置；path 为空时使用默认值
func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{Port: 20100},
		Tracing: TracingConfig{
			SampleRatio: 1.0,
			OTLPEnabled: false,
		},
	}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 20100
	}
	// 环境变量覆盖传输加密密钥（优先于配置文件，生产部署用环境变量注入）
	if envKey := os.Getenv(EnvTransportEncryptKey); envKey != "" {
		cfg.Channel.TransportEncrypt.Key = envKey
	}
	// 传输加密强制常开：key 为空直接启动失败（对标 Java ChannelRestClientSupport 强制校验）
	if cfg.Channel.TransportEncrypt.Key == "" {
		return nil, fmt.Errorf("channel transport-encrypt key is empty (set 'channel.transport-encrypt.key' in config or env %s)", EnvTransportEncryptKey)
	}
	return cfg, nil
}
