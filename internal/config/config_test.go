package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestByteOrderConstants tests that all byte order constants are defined.
func TestByteOrderConstants(t *testing.T) {
	constants := map[ByteOrder]string{
		BigEndian:        "ABCD",
		LittleEndian:    "DCBA",
		BigEndianSwap:   "BADC",
		LittleEndianSwap: "CDAB",
		MidSwap:         "BDAC",
	}

	for bo, expected := range constants {
		if string(bo) != expected {
			t.Errorf("ByteOrder constant = %q, want %q", bo, expected)
		}
	}
}

// TestDefaultConfig tests that default configuration has sensible values.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Mode != "tcp" {
		t.Errorf("DefaultConfig.Mode = %q, want %q", cfg.Mode, "tcp")
	}
	if cfg.ListenAddr != ":502" {
		t.Errorf("DefaultConfig.ListenAddr = %q, want %q", cfg.ListenAddr, ":502")
	}
	if cfg.ByteOrder != BigEndian {
		t.Errorf("DefaultConfig.ByteOrder = %q, want %q", cfg.ByteOrder, BigEndian)
	}
	if cfg.Serial == nil {
		t.Error("DefaultConfig.Serial should not be nil")
	}
}

// TestLoadFromFile tests loading a YAML config file.
func TestLoadFromFile(t *testing.T) {
	content := `
mode: tcp
listen_addr: ":10502"
byte_order: BDAC
registers:
  - address: 0
    count: 10
    value: 42
    label: "test"
log_format: json
log_level: debug
`

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if cfg.Mode != "tcp" {
		t.Errorf("cfg.Mode = %q, want %q", cfg.Mode, "tcp")
	}
	if cfg.ListenAddr != ":10502" {
		t.Errorf("cfg.ListenAddr = %q, want %q", cfg.ListenAddr, ":10502")
	}
	if cfg.ByteOrder != MidSwap {
		t.Errorf("cfg.ByteOrder = %q, want %q", cfg.ByteOrder, MidSwap)
	}
	if len(cfg.Registers) != 1 {
		t.Fatalf("cfg.Registers has %d entries, want 1", len(cfg.Registers))
	}
	if cfg.Registers[0].Address != 0 || cfg.Registers[0].Count != 10 || cfg.Registers[0].Value != 42 {
		t.Errorf("cfg.Registers[0] = %+v, want {Address:0 Count:10 Value:42}", cfg.Registers[0])
	}
}

// TestLoadFromFileInvalidByteOrder tests that loading a config with invalid byte order fails.
func TestLoadFromFileInvalidByteOrder(t *testing.T) {
	content := `
mode: tcp
listen_addr: ":502"
byte_order: INVALID
`

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadFromFile(cfgPath)
	if err == nil {
		t.Error("LoadFromFile should fail with invalid byte order")
	}
}