package config

import (
	"testing"
)

func TestManagerDefaultsAndSubscribe(t *testing.T) {
	m, err := NewManager()
	if err != nil {
		t.Fatalf("new manager error: %v", err)
	}
	c := m.Current()
	if c.Server.Port == 0 {
		t.Fatalf("default not loaded")
	}
	calls := 0
	m.Subscribe(func(cfg *Config) {
		calls++
	})
	m.SetForTest(&Config{Server: ServerConfig{Port: 9090}})
	if calls == 0 {
		t.Fatalf("subscribe not triggered")
	}
	if m.Current().Server.Port != 9090 {
		t.Fatalf("set for test failed")
	}
}
