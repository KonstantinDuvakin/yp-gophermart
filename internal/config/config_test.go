package config

import "testing"

func TestNewConfigDefaults(t *testing.T) {
	c := NewConfig()
	if c.RunAddress != "localhost:8080" {
		t.Errorf("RunAddress default = %q, want localhost:8080", c.RunAddress)
	}
	if c.DatabaseUri != "" || c.AccrualSystemAddress != "" {
		t.Errorf("expected empty DatabaseUri/AccrualSystemAddress by default")
	}
}

func TestApplyEnv_RandomJwtSecret(t *testing.T) {
	c := &Config{}
	c.ApplyEnv()
	// При отсутствии JWT_SECRET генерится случайный 32-байтный секрет (64 hex-символа).
	if len(c.JwtSecret) != 64 {
		t.Errorf("JwtSecret length = %d, want 64 (random hex secret)", len(c.JwtSecret))
	}
	if c.JwtSecret == "change_me" {
		t.Error("JwtSecret must not fall back to a public default")
	}
}

func TestApplyEnv_FromEnv(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "0.0.0.0:9000")
	t.Setenv("DATABASE_URI", "postgres://x")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual:8081")
	t.Setenv("JWT_SECRET", "super-secret")

	c := &Config{}
	c.ApplyEnv()

	if c.RunAddress != "0.0.0.0:9000" {
		t.Errorf("RunAddress = %q", c.RunAddress)
	}
	if c.DatabaseUri != "postgres://x" {
		t.Errorf("DatabaseUri = %q", c.DatabaseUri)
	}
	if c.AccrualSystemAddress != "http://accrual:8081" {
		t.Errorf("AccrualSystemAddress = %q", c.AccrualSystemAddress)
	}
	if c.JwtSecret != "super-secret" {
		t.Errorf("JwtSecret = %q", c.JwtSecret)
	}
}
