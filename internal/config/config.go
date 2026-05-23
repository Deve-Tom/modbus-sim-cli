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

// RegisterType defines the data type for register values.
type RegisterType string

// Supported register types.
const (
	RegisterTypeInt16   RegisterType = "INT16"
	RegisterTypeUint16  RegisterType = "UINT16"
	RegisterTypeInt32   RegisterType = "INT32"
	RegisterTypeUint32  RegisterType = "UINT32"
	RegisterTypeInt64   RegisterType = "INT64"
	RegisterTypeUint64  RegisterType = "UINT64"
	RegisterTypeFloat32 RegisterType = "FLOAT32"
	RegisterTypeFloat64 RegisterType = "FLOAT64"
)

// ValidRegisterTypes lists all supported register types.
var ValidRegisterTypes = []RegisterType{
	RegisterTypeInt16,
	RegisterTypeUint16,
	RegisterTypeInt32,
	RegisterTypeUint32,
	RegisterTypeInt64,
	RegisterTypeUint64,
	RegisterTypeFloat32,
	RegisterTypeFloat64,
}

// IsValid checks whether the register type is a recognized value.
func (r RegisterType) IsValid() bool {
	for _, v := range ValidRegisterTypes {
		if r == v {
			return true
		}
	}
	return false
}

// RegisterCount returns the number of physical 16-bit registers per value for this type.
func (r RegisterType) RegisterCount() int {
	switch r {
	case RegisterTypeInt16, RegisterTypeUint16:
		return 1
	case RegisterTypeInt32, RegisterTypeUint32, RegisterTypeFloat32:
		return 2
	case RegisterTypeInt64, RegisterTypeUint64, RegisterTypeFloat64:
		return 4
	default:
		return 1
	}
}

// RegisterConfig defines the configuration for a single register or register range.
type RegisterConfig struct {
	// Address is the starting register address (0-based).
	Address uint16 `yaml:"address"`

	// Count is the number of physical 16-bit registers in this range.
	// Note: For multi-register types (INT32, UINT32, FLOAT32, INT64, UINT64, FLOAT64),
	// the actual number of data values is: Count / (registers_per_value_for_type)
	// Example: Count=20 with FLOAT32 (2 regs/value) = 10 FLOAT32 values
	Count uint16 `yaml:"count"`

	// Type is the data type for this register range (INT16, UINT16, INT32, UINT32, INT64, UINT64, FLOAT32, FLOAT64).
	Type RegisterType `yaml:"type"`

	// Label is an optional human-readable label for the register.
	Label string `yaml:"label,omitempty"`

	// Values is an optional array of initial values for the data values in this range.
	// The length should match the number of data values (Count / registers_per_type).
	// If not specified, DefaultValue is used for all values.
	// If DefaultValue is also not specified, all values default to 0.
	Values []float64 `yaml:"values,omitempty"`

	// DefaultValue is the initial value for all data values if Values is not specified.
	// Only applies to multi-value ranges; single values can use Values directly.
	DefaultValue float64 `yaml:"default_value,omitempty"`

	// RandomEnable enables random value fluctuation for this register range.
	RandomEnable bool `yaml:"random_enable"`

	// RandomMin is the minimum value for random fluctuation.
	RandomMin float64 `yaml:"random_min"`

	// RandomMax is the maximum value for random fluctuation.
	RandomMax float64 `yaml:"random_max"`
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

	// ColorOutput enables colored console output.
	ColorOutput *bool `yaml:"color_output,omitempty"`

	// ShowData enables logging of received requests and responses data.
	ShowData *bool `yaml:"show_data,omitempty"`

	// RandomEnable enables random value fluctuation globally.
	RandomEnable *bool `yaml:"random_enable,omitempty"`

	// RandomMin is the default minimum value for random fluctuation.
	RandomMin float64 `yaml:"random_min,omitempty"`

	// RandomMax is the default maximum value for random fluctuation.
	RandomMax float64 `yaml:"random_max,omitempty"`
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

	// Validate register types and count compatibility
	for i, reg := range cfg.Registers {
		if reg.Type != "" && !reg.Type.IsValid() {
			return nil, fmt.Errorf("invalid register type %q at index %d, must be one of: %v",
				reg.Type, i, ValidRegisterTypes)
		}

		// Validate count is compatible with type
		if reg.Type != "" && reg.Count > 0 {
			regCount := reg.Type.RegisterCount()
			if reg.Count%uint16(regCount) != 0 {
				return nil, fmt.Errorf(
					"invalid register count %d at index %d for type %s: count must be a multiple of %d (got %d registers, which is %d values × %d + %d remainder)",
					reg.Count, i, reg.Type, regCount, reg.Count, reg.Count/uint16(regCount), regCount, reg.Count%uint16(regCount))
			}
		}

		// Validate values array length if specified
		if len(reg.Values) > 0 && reg.Type != "" {
			expectedValues := int(reg.Count) / reg.Type.RegisterCount()
			if len(reg.Values) != expectedValues {
				return nil, fmt.Errorf(
					"invalid values length at index %d: expected %d values (count=%d / registers_per_%s=%d), got %d",
					i, expectedValues, reg.Count, reg.Type, reg.Type.RegisterCount(), len(reg.Values))
			}
		}
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
	// Default pointer fields to false if not set (except ColorOutput which defaults to true)
	if cfg.ColorOutput == nil {
		cfg.ColorOutput = newBool(true)
	}
	if cfg.ShowData == nil {
		cfg.ShowData = newBool(false)
	}
	if cfg.RandomEnable == nil {
		cfg.RandomEnable = newBool(false)
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
		LogFormat:    "console",
		LogLevel:     "info",
		ColorOutput:  newBool(true),
		ShowData:     newBool(false),
		RandomEnable: newBool(false),
	}
}

// newBool returns a pointer to a bool value.
func newBool(b bool) *bool {
	return &b
}
