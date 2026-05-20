// Package config defines the configuration structures and loading logic
// for the Modbus simulator, including byte order constants.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ByteOrder represents the byte ordering for multi-byte register values.
type ByteOrder string

// Supported byte order constants.
const (
	// BigEndian - ABCD: bytes are in big-endian order [A][B][C][D]
	BigEndian ByteOrder = "ABCD"

	// LittleEndian - DCBA: bytes are in little-endian order [D][C][B][A]
	LittleEndian ByteOrder = "DCBA"

	// BigEndianSwap - BADC: big-endian with word swap [B][A][D][C]
	BigEndianSwap ByteOrder = "BADC"

	// LittleEndianSwap - CDAB: little-endian with word swap [C][D][A][B]
	LittleEndianSwap ByteOrder = "CDAB"

	// MidSwap - BDAC: swap high bytes within each word [B][D][A][C]
	// For a 32-bit value with bytes [A][B][C][D]:
	//   After BDAC encoding: [B][D][A][C]
	//   This swaps A<->B and C<->D within their word positions.
	MidSwap ByteOrder = "BDAC"
)

// ValidByteOrders lists all supported byte order values.
var ValidByteOrders = []ByteOrder{
	BigEndian,
	LittleEndian,
	BigEndianSwap,
	LittleEndianSwap,
	MidSwap,
}

// IsValid checks whether the byte order string is a recognized value.
func (b ByteOrder) IsValid() bool {
	for _, v := range ValidByteOrders {
		if b == v {
			return true
		}
	}
	return false
}

// RegisterConfig defines the configuration for a single register or register range.
type RegisterConfig struct {
	// Address is the starting register address (0-based).
	Address uint16 `yaml:"address"`

	// Count is the number of registers in this range.
	Count uint16 `yaml:"count"`

	// Value is the initial value for the first register.
	Value uint16 `yaml:"value"`

	// Label is an optional human-readable label for the register.
	Label string `yaml:"label,omitempty"`
}

// SerialConfig holds serial port configuration for RTU mode.
type SerialConfig struct {
	// BaudRate is the serial port baud rate (e.g., 9600, 19200, 115200).
	BaudRate int `yaml:"baud_rate"`

	// DataBits is the number of data bits (typically 8).
	DataBits int `yaml:"data_bits"`

	// StopBits is the number of stop bits (1 or 2).
	StopBits int `yaml:"stop_bits"`

	// Parity is the parity mode: "none", "even", "odd".
	Parity string `yaml:"parity"`
}

// Config is the top-level configuration structure loaded from YAML.
type Config struct {
	// Mode is the server mode: "tcp" or "rtu".
	Mode string `yaml:"mode"`

	// ListenAddr is the TCP listen address (e.g., ":502").
	ListenAddr string `yaml:"listen_addr"`

	// ByteOrder is the byte order for multi-register values.
	ByteOrder ByteOrder `yaml:"byte_order"`

	// Serial holds serial port configuration (required for RTU mode).
	Serial *SerialConfig `yaml:"serial,omitempty"`

	// Registers is the list of register configurations.
	Registers []RegisterConfig `yaml:"registers"`

	// LogFormat is the log output format: "console" or "json".
	LogFormat string `yaml:"log_format"`

	// LogLevel is the minimum log level: "debug", "info", "warn", "error".
	LogLevel string `yaml:"log_level"`
}

// LoadFromFile reads and parses a YAML configuration file.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	// Validate byte order
	if cfg.ByteOrder != "" && !cfg.ByteOrder.IsValid() {
		return nil, fmt.Errorf("invalid byte order %q, must be one of: %v",
			cfg.ByteOrder, ValidByteOrders)
	}

	// Set defaults
	if cfg.Mode == "" {
		cfg.Mode = "tcp"
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":502"
	}
	if cfg.ByteOrder == "" {
		cfg.ByteOrder = BigEndian
	}
	if cfg.LogFormat == "" {
		cfg.LogFormat = "console"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return &cfg, nil
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Mode:       "tcp",
		ListenAddr: ":502",
		ByteOrder:  BigEndian,
		Serial: &SerialConfig{
			BaudRate: 9600,
			DataBits: 8,
			StopBits: 1,
			Parity:   "none",
		},
		LogFormat: "console",
		LogLevel:  "info",
	}
}