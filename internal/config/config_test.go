package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	content := `
frontends:
- port: 8080
- port: 8081

backends:
- "http://backend1:9001"
- "http://backend2:9002"

healthcheck:
  interval: "10s"
  timeout: "2s"
  path: "/health"

logging:
  level: "info"
  format: "json"

metrics:
  enabled: true
  port: 9090
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test loading config
	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify frontend ports
	if len(cfg.Frontends) != 2 {
		t.Errorf("Expected 2 frontends, got %d", len(cfg.Frontends))
	}
	if cfg.Frontends[0].Port != 8080 {
		t.Errorf("Expected frontend port 8080, got %d", cfg.Frontends[0].Port)
	}
	if cfg.Frontends[1].Port != 8081 {
		t.Errorf("Expected frontend port 8081, got %d", cfg.Frontends[1].Port)
	}

	// Verify backends
	if len(cfg.Backends) != 2 {
		t.Errorf("Expected 2 backends, got %d", len(cfg.Backends))
	}
	if cfg.Backends[0] != "http://backend1:9001" {
		t.Errorf("Expected backend1 URL, got %s", cfg.Backends[0])
	}
	if cfg.Backends[1] != "http://backend2:9002" {
		t.Errorf("Expected backend2 URL, got %s", cfg.Backends[1])
	}

	// Verify healthcheck
	if cfg.HealthCheck.Interval != 10*time.Second {
		t.Errorf("Expected 10s interval, got %v", cfg.HealthCheck.Interval)
	}
	if cfg.HealthCheck.Timeout != 2*time.Second {
		t.Errorf("Expected 2s timeout, got %v", cfg.HealthCheck.Timeout)
	}
	if cfg.HealthCheck.Path != "/health" {
		t.Errorf("Expected /health path, got %s", cfg.HealthCheck.Path)
	}

	// Verify logging
	if cfg.Logging.Level != "info" {
		t.Errorf("Expected info log level, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Expected json log format, got %s", cfg.Logging.Format)
	}

	// Verify metrics
	if !cfg.Metrics.Enabled {
		t.Error("Expected metrics to be enabled")
	}
	if cfg.Metrics.Port != 9090 {
		t.Errorf("Expected metrics port 9090, got %d", cfg.Metrics.Port)
	}
}

func TestLoadDefaults(t *testing.T) {
	// Create a minimal config file
	content := `
frontends:
- port: 8080

backends:
- "http://backend1:9001"
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test loading config with defaults
	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify default values
	if cfg.HealthCheck.Interval != 10*time.Second {
		t.Errorf("Expected default 10s interval, got %v", cfg.HealthCheck.Interval)
	}
	if cfg.HealthCheck.Timeout != 2*time.Second {
		t.Errorf("Expected default 2s timeout, got %v", cfg.HealthCheck.Timeout)
	}
	if cfg.HealthCheck.Path != "/health" {
		t.Errorf("Expected default /health path, got %s", cfg.HealthCheck.Path)
	}
	if cfg.Metrics.Port != 9090 {
		t.Errorf("Expected default metrics port 9090, got %d", cfg.Metrics.Port)
	}
}

func TestLoadInvalidFile(t *testing.T) {
	// Test non-existent file
	_, err := Load("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error loading non-existent file")
	}

	// Test invalid YAML
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("invalid: yaml: content:")); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	_, err = Load(tmpfile.Name())
	if err == nil {
		t.Error("Expected error loading invalid YAML")
	}
}
