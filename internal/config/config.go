package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// 服务配置（对标 Boot application.yml）
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Tracing    TracingConfig    `yaml:"tracing"`
	Channel    ChannelConfig    `yaml:"channel"`
	Deployment DeploymentConfig `yaml:"deployment"`
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

// 部署模式相关常量（对标 Java daxpay.platform.deployment.mode + DEPLOY_MODE 环境变量）
const (
	EnvDeployMode  = "DEPLOY_MODE"
	DeployModeProd = "PROD"
	DeployModeDev  = "DEV"
)

// DeploymentConfig：部署模式（PROD 生产 / DEV 联调）
//
// 对标 Java DeploymentModeEnforcer 读取的 daxpay.platform.deployment.mode。
// 与主应用联动：installer 通过共享 .env.biz 把同一个 DEPLOY_MODE 注入主应用与通道子应用容器，
// 切换主应用模式后重启本容器即自动跟随同值。Go 版无 Spring 的 fail-fast 机制，仅读取用于
// 启动展示与未来按需分支（不校验 actuator/health，Go 版无对应概念）。
type DeploymentConfig struct {
	Mode string `yaml:"mode"`
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
	// 环境变量覆盖部署模式（优先于配置文件，与主应用共享 DEPLOY_MODE）
	if envMode := os.Getenv(EnvDeployMode); envMode != "" {
		cfg.Deployment.Mode = strings.ToUpper(strings.TrimSpace(envMode))
	}
	// 默认 PROD（与主应用 application-prod.yml 的 ${DEPLOY_MODE:PROD} 一致）
	if cfg.Deployment.Mode == "" {
		cfg.Deployment.Mode = DeployModeProd
	}
	// 校验合法值（仅允许 PROD/DEV，防止误配）
	if cfg.Deployment.Mode != DeployModeProd && cfg.Deployment.Mode != DeployModeDev {
		return nil, fmt.Errorf("invalid deployment mode %q (env %s), only PROD/DEV allowed", cfg.Deployment.Mode, EnvDeployMode)
	}
	return cfg, nil
}
