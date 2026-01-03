package auth

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
	"sync"
	"time"
)

type Session struct {
	Token      string
	DeviceName string
	ExpiresAt  time.Time
}

type DeviceCode struct {
	Code       string
	DeviceName string
	ExpiresAt  time.Time
}

type Manager struct {
	mu       sync.Mutex
	codes    map[string]DeviceCode
	sessions map[string]Session
	ttl      time.Duration
}

func NewManager(ttl time.Duration) *Manager {
	return &Manager{
		codes:    make(map[string]DeviceCode),
		sessions: make(map[string]Session),
		ttl:      ttl,
	}
}

func (m *Manager) GenerateDeviceCode(name string) DeviceCode {
	m.mu.Lock()
	defer m.mu.Unlock()
	code := randomCode(6)
	dc := DeviceCode{
		Code:       code,
		DeviceName: strings.TrimSpace(name),
		ExpiresAt:  time.Now().Add(5 * time.Minute),
	}
	m.codes[code] = dc
	return dc
}

func (m *Manager) VerifyDeviceCode(code string) (Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	dc, ok := m.codes[strings.ToUpper(strings.TrimSpace(code))]
	if !ok || time.Now().After(dc.ExpiresAt) {
		return Session{}, false
	}
	delete(m.codes, dc.Code)
	token := randomToken(32)
	sess := Session{
		Token:      token,
		DeviceName: dc.DeviceName,
		ExpiresAt:  time.Now().Add(m.ttl),
	}
	m.sessions[token] = sess
	return sess, true
}

func (m *Manager) ValidateToken(token string) (Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[token]
	if !ok || time.Now().After(s.ExpiresAt) {
		return Session{}, false
	}
	// 滚动续期
	s.ExpiresAt = time.Now().Add(m.ttl)
	m.sessions[token] = s
	return s, true
}

func randomCode(n int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[idx.Int64()]
	}
	return string(b)
}

func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

var defaultManager *Manager

func SetDefaultManager(m *Manager) {
	defaultManager = m
}

func DefaultManager() *Manager {
	return defaultManager
}
