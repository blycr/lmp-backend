package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port    int  `mapstructure:"port"`
	LANOnly bool `mapstructure:"lan_only"`
}

type FilesConfig struct {
	ShareDirs []string `mapstructure:"share_dirs"`
}

type AuthConfig struct {
	EnableDeviceAuth bool          `mapstructure:"enable_device_auth"`
	SessionTimeout   time.Duration `mapstructure:"session_timeout"`
}

type RateLimitConfig struct {
	Enabled           bool    `mapstructure:"enabled"`
	RequestsPerSecond float64 `mapstructure:"requests_per_second"`
	Burst             int     `mapstructure:"burst"`
}

type Config struct {
	Server ServerConfig    `mapstructure:"server"`
	Files  FilesConfig     `mapstructure:"files"`
	Auth   AuthConfig      `mapstructure:"auth"`
	Rate   RateLimitConfig `mapstructure:"rate_limit"`
}

func Load() (*Config, error) {
	v := viper.New()
	if p := os.Getenv("LMP_CONFIG"); p != "" {
		v.SetConfigFile(p)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("./backend")
		v.AddConfigPath("./backend/cmd/server")
		v.SetConfigType("json")
	}

	// defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.lan_only", true)
	v.SetDefault("auth.enable_device_auth", true)
	v.SetDefault("auth.session_timeout", time.Minute*30)
	v.SetDefault("files.share_dirs", []string{})
	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("rate_limit.requests_per_second", 10.0)
	v.SetDefault("rate_limit.burst", 20)

	_ = v.ReadInConfig() // optional

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
