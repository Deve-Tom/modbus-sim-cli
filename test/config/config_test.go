package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"modbus-sim/internal/config"
)

// TestByteOrderConstants tests that all byte order constants are defined.
func TestByteOrderConstants(t *testing.T) {
	constants := map[config.ByteOrder]string{
		config.BigEndian:        "ABCD",
		config.LittleEndian:     "DCBA",
		config.BigEndianSwap:    "BADC",
		config.LittleEndianSwap: "CDAB",
		config.MidSwap:          "BDAC",
	}

	for bo, expected := range constants {
		if string(bo) != expected {
			t.Errorf("ByteOrder constant = %q, want %q", bo, expected)
		}
	}
}

// TestDefaultConfig tests that default configuration has sensible values.
func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.Mode != "tcp" {
		t.Errorf("DefaultConfig.Mode = %q, want %q", cfg.Mode, "tcp")
	}
	if cfg.ListenAddr != ":502" {
		t.Errorf("DefaultConfig.ListenAddr = %q, want %q", cfg.ListenAddr, ":502")
	}
	if cfg.ByteOrder != config.BigEndian {
		t.Errorf("DefaultConfig.ByteOrder = %q, want %q", cfg.ByteOrder, config.BigEndian)
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
    type: UINT16
    default_value: 42
    label: "test"
log_format: json
log_level: debug
`

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := config.LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if cfg.Mode != "tcp" {
		t.Errorf("cfg.Mode = %q, want %q", cfg.Mode, "tcp")
	}
	if cfg.ListenAddr != ":10502" {
		t.Errorf("cfg.ListenAddr = %q, want %q", cfg.ListenAddr, ":10502")
	}
	if cfg.ByteOrder != config.MidSwap {
		t.Errorf("cfg.ByteOrder = %q, want %q", cfg.ByteOrder, config.MidSwap)
	}
	if len(cfg.Registers) != 1 {
		t.Fatalf("cfg.Registers has %d entries, want 1", len(cfg.Registers))
	}
	if cfg.Registers[0].Address != 0 || cfg.Registers[0].Count != 10 {
		t.Errorf("cfg.Registers[0] = %+v, want {Address:0 Count:10}", cfg.Registers[0])
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

	_, err := config.LoadFromFile(cfgPath)
	if err == nil {
		t.Error("LoadFromFile should fail with invalid byte order")
	}
}

// TestRegisterTypeRegisterCount tests the RegisterCount method.
func TestRegisterTypeRegisterCount(t *testing.T) {
	tests := []struct {
		regType   config.RegisterType
		wantCount int
	}{
		{config.RegisterTypeInt16, 1},
		{config.RegisterTypeUint16, 1},
		{config.RegisterTypeInt32, 2},
		{config.RegisterTypeUint32, 2},
		{config.RegisterTypeFloat32, 2},
		{config.RegisterTypeInt64, 4},
		{config.RegisterTypeUint64, 4},
		{config.RegisterTypeFloat64, 4},
	}

	for _, tt := range tests {
		t.Run(string(tt.regType), func(t *testing.T) {
			if got := tt.regType.RegisterCount(); got != tt.wantCount {
				t.Errorf("RegisterCount() = %v, want %v", got, tt.wantCount)
			}
		})
	}
}

// TestLoadFromFileInvalidCountType tests that count is validated against type.
func TestLoadFromFileInvalidCountType(t *testing.T) {
	// count=10 with FLOAT32 (2 regs/value) = 5 values, should pass
	content := `
mode: tcp
registers:
  - address: 0
    count: 10
    type: FLOAT32
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	_, err := config.LoadFromFile(cfgPath)
	if err != nil {
		t.Errorf("Should pass with count=10 and FLOAT32: %v", err)
	}

	// count=10 with FLOAT64 (4 regs/value) = 2.5 values, should fail
	content = `
mode: tcp
registers:
  - address: 0
    count: 10
    type: FLOAT64
`
	cfgPath = filepath.Join(tmpDir, "test2.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	_, err = config.LoadFromFile(cfgPath)
	if err == nil {
		t.Error("Should fail with count=10 and FLOAT64 (not divisible)")
	}
}
