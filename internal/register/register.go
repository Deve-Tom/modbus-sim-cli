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
	Address uint16
	Count   uint16
	Value   uint16
	Label   string
}

// InitFromConfig initializes registers from configuration definitions.
func (m *Manager) InitFromConfig(regConfigs []RegisterConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, rc := range regConfigs {
		if rc.Count == 0 {
			rc.Count = 1
		}
		for i := uint16(0); i < rc.Count; i++ {
			addr := int(rc.Address + i)
			if addr < len(m.memReg.HoldingRegisters) {
				m.memReg.HoldingRegisters[addr] = rc.Value
			}
		}
	}

	m.logger.Info().Msg("registers initialized from config")
	return nil
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
