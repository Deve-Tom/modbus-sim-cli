// Package register provides register management for the Modbus simulator.
// It implements the mbserver.Register interface for custom register behavior.
package register

import (
	"fmt"
	"sync"

	"github.com/leijux/mbserver"
	"github.com/rs/zerolog"
)

// Manager manages Modbus registers with thread-safe access.
// It wraps mbserver.MemRegister to implement the mbserver.Register interface.
type Manager struct {
	mu        sync.RWMutex
	logger    zerolog.Logger
	memReg    *mbserver.MemRegister
}

// NewManager creates a new register manager with the given logger.
func NewManager(logger zerolog.Logger) *Manager {
	return &Manager{
		logger: logger,
		memReg: mbserver.NewMemRegister(),
	}
}

// RegisterConfig defines the configuration for a single register or register range.
type RegisterConfig struct {
	Address  uint16
	Count    uint16
	Type     string // "INT16", "UINT16", "INT32", "UINT32", "INT64", "UINT64", "FLOAT32", "FLOAT64"
	Values   []float64 // Initial values (one per data value, not per physical register)
	Label    string
}

// InitFromConfig initializes registers from configuration definitions.
func (m *Manager) InitFromConfig(regConfigs []RegisterConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, rc := range regConfigs {
		if rc.Count == 0 {
			continue
		}

		// Determine registers per value based on type
		regsPerValue := getRegistersPerValue(rc.Type)

		// Calculate number of data values
		numValues := int(rc.Count) / regsPerValue

		// Prepare values array
		values := rc.Values
		if len(values) == 0 {
			// Use default value 0 for all values
			values = make([]float64, numValues)
		} else if len(values) < numValues {
			// Pad with zeros if not enough values provided
			padded := make([]float64, numValues)
			copy(padded, values)
			values = padded
		}

		// Write values to registers
		for i := 0; i < numValues; i++ {
			value := values[i]
			startReg := int(rc.Address) + i*regsPerValue

			// Convert float64 to appropriate uint16 representation(s)
			uint16Values := convertToUint16s(value, rc.Type, regsPerValue)

			// Write each uint16 to consecutive registers
			for j, v := range uint16Values {
				addr := startReg + j
				if addr < len(m.memReg.HoldingRegisters) {
					m.memReg.HoldingRegisters[addr] = v
				}
			}
		}
	}

	m.logger.Info().Msg("registers initialized from config")
	return nil
}

// getRegistersPerValue returns the number of 16-bit registers per value for the given type.
func getRegistersPerValue(typeStr string) int {
	switch typeStr {
	case "INT16", "UINT16", "":
		return 1
	case "INT32", "UINT32", "FLOAT32":
		return 2
	case "INT64", "UINT64", "FLOAT64":
		return 4
	default:
		return 1
	}
}

// convertToUint16s converts a float64 value to uint16 representation(s) based on type.
func convertToUint16s(value float64, typeStr string, regsPerValue int) []uint16 {
	result := make([]uint16, regsPerValue)

	switch typeStr {
	case "INT16":
		// Clamp to int16 range and convert
		if value < -32768 {
			value = -32768
		} else if value > 32767 {
			value = 32767
		}
		result[0] = uint16(int16(value))

	case "UINT16", "", "FLOAT32", "FLOAT64":
		// For unspecified type or floating types, store as-is in uint16
		// This may lose precision for float types, but maintains basic compatibility
		if value < 0 {
			value = 0
		} else if value > 65535 {
			value = 65535
		}
		result[0] = uint16(value)

	case "INT32":
		// Convert to int32, split across 2 registers (big-endian style for ABCD)
		i32 := int32(value)
		result[0] = uint16((i32 >> 16) & 0xFFFF)
		result[1] = uint16(i32 & 0xFFFF)

	case "UINT32":
		// Convert to uint32, split across 2 registers
		u32 := uint32(value)
		result[0] = uint16((u32 >> 16) & 0xFFFF)
		result[1] = uint16(u32 & 0xFFFF)

	case "INT64":
		// Convert to int64, split across 4 registers
		i64 := int64(value)
		result[0] = uint16((i64 >> 48) & 0xFFFF)
		result[1] = uint16((i64 >> 32) & 0xFFFF)
		result[2] = uint16((i64 >> 16) & 0xFFFF)
		result[3] = uint16(i64 & 0xFFFF)

	case "UINT64":
		// Convert to uint64, split across 4 registers
		u64 := uint64(value)
		result[0] = uint16((u64 >> 48) & 0xFFFF)
		result[1] = uint16((u64 >> 32) & 0xFFFF)
		result[2] = uint16((u64 >> 16) & 0xFFFF)
		result[3] = uint16(u64 & 0xFFFF)

	default:
		// Default to UINT16 behavior
		if value < 0 {
			value = 0
		} else if value > 65535 {
			value = 65535
		}
		result[0] = uint16(value)
	}

	return result
}

// InitSimple initializes a contiguous range of holding registers with a default value.
func (m *Manager) InitSimple(startAddr uint16, count uint16, defaultValue uint16) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := uint16(0); i < count; i++ {
		addr := int(startAddr + i)
		if addr < len(m.memReg.HoldingRegisters) {
			m.memReg.HoldingRegisters[addr] = defaultValue
		}
	}

	m.logger.Info().
		Uint16("start", startAddr).
		Uint16("count", count).
		Msg("registers initialized (simple mode)")
	return nil
}

// GetMemRegister returns the underlying mbserver.MemRegister.
func (m *Manager) GetMemRegister() *mbserver.MemRegister {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.memReg
}

// Count returns the total number of registers.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.memReg.HoldingRegisters)
}

// SetRegisterValue sets the value of a register at the given address.
func (m *Manager) SetRegisterValue(address uint16, value uint16) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if int(address) >= len(m.memReg.HoldingRegisters) {
		return fmt.Errorf("register at address %d not found", address)
	}
	m.memReg.HoldingRegisters[address] = value
	m.logger.Debug().
		Uint16("address", address).
		Uint16("value", value).
		Msg("register value updated")
	return nil
}

// GetRegisterValue gets the value of a register at the given address.
func (m *Manager) GetRegisterValue(address uint16) (uint16, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if int(address) >= len(m.memReg.HoldingRegisters) {
		return 0, fmt.Errorf("register at address %d not found", address)
	}
	return m.memReg.HoldingRegisters[address], nil
}
