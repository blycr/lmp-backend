package config

import (
	"os"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"

	"github.com/spf13/viper"
)

type Manager struct {
	v    *viper.Viper
	curr atomic.Value
	mu   sync.Mutex
	subs []func(*Config)
}

func NewManager() (*Manager, error) {
	m := &Manager{v: viper.New()}
	if p := os.Getenv("LMP_CONFIG"); p != "" {
		m.v.SetConfigFile(p)
	} else {
		m.v.SetConfigName("config")
		m.v.AddConfigPath(".")
		m.v.AddConfigPath("./backend")
		m.v.AddConfigPath("./backend/cmd/server")
		m.v.SetConfigType("json")
	}
	m.v.SetDefault("server.port", 8080)
	m.v.SetDefault("server.lan_only", true)
	m.v.SetDefault("auth.enable_device_auth", true)
	m.v.SetDefault("auth.session_timeout", "30m")
	m.v.SetDefault("files.share_dirs", []string{})
	m.v.SetDefault("rate_limit.enabled", true)
	m.v.SetDefault("rate_limit.requests_per_second", 10.0)
	m.v.SetDefault("rate_limit.burst", 20)
	_ = m.v.ReadInConfig()
	var cfg Config
	if err := m.v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	m.curr.Store(cfg)
	m.v.WatchConfig()
	m.v.OnConfigChange(func(_ fsnotify.Event) {
		var next Config
		if err := m.v.Unmarshal(&next); err != nil {
			return
		}
		m.curr.Store(next)
		m.mu.Lock()
		for _, s := range m.subs {
			s(&next)
		}
		m.mu.Unlock()
	})
	return m, nil
}

func (m *Manager) Current() *Config {
	c := m.curr.Load().(Config)
	return &c
}

func (m *Manager) Subscribe(fn func(*Config)) {
	m.mu.Lock()
	m.subs = append(m.subs, fn)
	m.mu.Unlock()
}

func (m *Manager) SetForTest(cfg *Config) {
	m.curr.Store(*cfg)
	m.mu.Lock()
	for _, s := range m.subs {
		s(cfg)
	}
	m.mu.Unlock()
}
