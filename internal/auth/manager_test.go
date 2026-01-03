package auth

import (
	"testing"
	"time"
)

func TestGenerateVerifyAndValidateToken(t *testing.T) {
	m := NewManager(1 * time.Minute)
	code := m.GenerateDeviceCode("phone")
	if code.Code == "" || code.DeviceName != "phone" {
		t.Fatalf("invalid device code")
	}
	sess, ok := m.VerifyDeviceCode(code.Code)
	if !ok {
		t.Fatalf("verify failed")
	}
	if sess.Token == "" || sess.DeviceName != "phone" {
		t.Fatalf("invalid session")
	}
	got, ok := m.ValidateToken(sess.Token)
	if !ok {
		t.Fatalf("validate failed")
	}
	if got.Token != sess.Token {
		t.Fatalf("token mismatch")
	}
}

func TestVerifyInvalidCode(t *testing.T) {
	m := NewManager(1 * time.Minute)
	_, ok := m.VerifyDeviceCode("ZZZZZZ")
	if ok {
		t.Fatalf("expected invalid code to fail")
	}
}
