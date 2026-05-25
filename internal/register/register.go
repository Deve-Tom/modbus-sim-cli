// Package register provides register management for the Modbus simulator.
// It implements the mbserver.Register interface for custom register behavior.
package register

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/leijux/mbserver"
	"github.com/rs/zerolog"
)

// RandomRange defines the configuration for random value fluctuation.
type RandomRange struct {
	Enable bool
	Min    float64
	Max    float64
}

// RegisterDefinition stores information about a configured register range,
// including its type and random fluctuation configuration.
type RegisterDefinition struct {
	Address     uint16
	Count       uint16
	Type        string
	Label       string
	RandomRange RandomRange
}

// Manager manages Modbus registers with thread-safe access.
// It implements the mbserver.Register interface for custom register behavior.
type Manager struct {
	mu             sync.RWMutex
	logger         zerolog.Logger
	memReg         *mbserver.MemRegister
	definitions    []RegisterDefinition
	randomEnabled  bool
	randomTicker   *time.Ticker
	randomStopChan chan struct{}
	randomInterval time.Duration
	showData       bool
}

// NewManager creates a new register manager with the given logger.
func NewManager(logger zerolog.Logger) *Manager {
	rand.Seed(time.Now().UnixNano())
	return &Manager{
		logger:         logger,
		memReg:         mbserver.NewMemRegister(),
		randomInterval: 1 * time.Second, // Default update every 1 second
	}
}

// RegisterConfig defines the configuration for a single register or register range.
type RegisterConfig struct {
	Address      uint16
	Count        uint16
	Type         string    // "INT16", "UINT16", "INT32", "UINT32", "INT64", "UINT64", "FLOAT32", "FLOAT64"
	Values       []float64 // Initial values (one per data value, not per physical register)
	Label        string
	RandomEnable bool
	RandomMin    float64
	RandomMax    float64
}

// InitFromConfig initializes registers from configuration definitions.
func (m *Manager) InitFromConfig(regConfigs []RegisterConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Initialize definitions array
	m.definitions = make([]RegisterDefinition, len(regConfigs))

	for i, rc := range regConfigs {
		if rc.Count == 0 {
			continue
		}

		// Save definition for later use in random updates
		m.definitions[i] = RegisterDefinition{
			Address: rc.Address,
			Count:   rc.Count,
			Type:    rc.Type,
			Label:   rc.Label,
			RandomRange: RandomRange{
				Enable: rc.RandomEnable,
				Min:    rc.RandomMin,
				Max:    rc.RandomMax,
			},
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
		for j := 0; j < numValues; j++ {
			value := values[j]
			startReg := int(rc.Address) + j*regsPerValue

			// Convert float64 to appropriate uint16 representation(s)
			uint16Values := convertToUint16s(value, rc.Type, regsPerValue)

			// Write each uint16 to consecutive registers
			for k, v := range uint16Values {
				addr := startReg + k
				if addr < len(m.memReg.HoldingRegisters) {
					m.memReg.HoldingRegisters[addr] = v
				}
			}
		}
	}

	m.logger.Info().
		Int("definitions", len(m.definitions)).
		Msg("registers initialized from config")
	return nil
}

// GetRegistersPerValue returns the number of 16-bit registers per value for the given type.
// This is the exported version for use by external packages like simulator.
func GetRegistersPerValue(typeStr string) int {
	return getRegistersPerValue(typeStr)
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

	case "UINT16", "":
		// For unspecified type, store as-is in uint16
		if value < 0 {
			value = 0
		} else if value > 65535 {
			value = 65535
		}
		result[0] = uint16(value)

	case "FLOAT32":
		// Encode IEEE 754 float32 into 2 registers (big-endian for ABCD byte order)
		f32 := float32(value)
		bits := math.Float32bits(f32)
		result[0] = uint16((bits >> 16) & 0xFFFF)
		result[1] = uint16(bits & 0xFFFF)

	case "FLOAT64":
		// Encode IEEE 754 float64 into 4 registers (big-endian for ABCD byte order)
		bits := math.Float64bits(value)
		result[0] = uint16((bits >> 48) & 0xFFFF)
		result[1] = uint16((bits >> 32) & 0xFFFF)
		result[2] = uint16((bits >> 16) & 0xFFFF)
		result[3] = uint16(bits & 0xFFFF)

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

// SetShowData enables or disables detailed request/response logging.
func (m *Manager) SetShowData(enable bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.showData = enable
}

// SetRandomInterval sets the interval between random value updates.
func (m *Manager) SetRandomInterval(interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.randomInterval = interval
}

// ==================== mbserver.Register interface implementation ====================

// ReadCoils implements the mbserver.Register interface.
func (m *Manager) ReadCoils(start, count int) ([]bool, mbserver.Exception) {
	m.mu.RLock()
	values, ex := m.memReg.ReadCoils(start, count)
	m.mu.RUnlock()

	if m.showData {
		m.logger.Debug().
			Int("start", start).
			Int("count", count).
			Bools("values", values).
			Msg("ReadCoils")
	}

	return values, ex
}

// ReadDiscreteInputs implements the mbserver.Register interface.
func (m *Manager) ReadDiscreteInputs(start, count int) ([]bool, mbserver.Exception) {
	m.mu.RLock()
	values, ex := m.memReg.ReadDiscreteInputs(start, count)
	m.mu.RUnlock()

	if m.showData {
		m.logger.Debug().
			Int("start", start).
			Int("count", count).
			Bools("values", values).
			Msg("ReadDiscreteInputs")
	}

	return values, ex
}

// ReadHoldingRegisters implements the mbserver.Register interface.
func (m *Manager) ReadHoldingRegisters(start, count int) ([]uint16, mbserver.Exception) {
	m.mu.RLock()
	values, ex := m.memReg.ReadHoldingRegisters(start, count)
	m.mu.RUnlock()

	if m.showData {
		m.logger.Debug().
			Int("start", start).
			Int("count", count).
			Uints16("values", values).
			Msg("ReadHoldingRegisters")
	}

	return values, ex
}

// ReadInputRegisters implements the mbserver.Register interface.
func (m *Manager) ReadInputRegisters(start, count int) ([]uint16, mbserver.Exception) {
	m.mu.RLock()
	values, ex := m.memReg.ReadInputRegisters(start, count)
	m.mu.RUnlock()

	if m.showData {
		m.logger.Debug().
			Int("start", start).
			Int("count", count).
			Uints16("values", values).
			Msg("ReadInputRegisters")
	}

	return values, ex
}

// WriteSingleCoil implements the mbserver.Register interface.
func (m *Manager) WriteSingleCoil(start int, value bool) mbserver.Exception {
	m.mu.Lock()
	ex := m.memReg.WriteSingleCoil(start, value)
	m.mu.Unlock()

	if m.showData {
		m.logger.Debug().
			Int("start", start).
			Bool("value", value).
			Msg("WriteSingleCoil")
	}

	return ex
}

// WriteSingleRegister implements the mbserver.Register interface.
func (m *Manager) WriteSingleRegister(start int, value uint16) mbserver.Exception {
	m.mu.Lock()
	ex := m.memReg.WriteSingleRegister(start, value)
	m.mu.Unlock()

	if m.showData {
		m.logger.Debug().
			Int("start", start).
			Uint16("value", value).
			Msg("WriteSingleRegister")
	}

	return ex
}

// WriteMultipleCoils implements the mbserver.Register interface.
func (m *Manager) WriteMultipleCoils(start int, values []bool) mbserver.Exception {
	m.mu.Lock()
	ex := m.memReg.WriteMultipleCoils(start, values)
	m.mu.Unlock()

	if m.showData {
		m.logger.Debug().
			Int("start", start).
			Int("count", len(values)).
			Bools("values", values).
			Msg("WriteMultipleCoils")
	}

	return ex
}

// WriteMultipleRegisters implements the mbserver.Register interface.
func (m *Manager) WriteMultipleRegisters(start int, values []uint16) mbserver.Exception {
	m.mu.Lock()
	ex := m.memReg.WriteMultipleRegisters(start, values)
	m.mu.Unlock()

	if m.showData {
		m.logger.Debug().
			Int("start", start).
			Int("count", len(values)).
			Uints16("values", values).
			Msg("WriteMultipleRegisters")
	}

	return ex
}

// ==================== End mbserver.Register interface ====================

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

// StartRandomUpdates starts periodic random value updates for registers
// that have RandomEnable set to true.
func (m *Manager) StartRandomUpdates(globalEnable bool, globalMin, globalMax float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If already running, stop first
	if m.randomStopChan != nil {
		close(m.randomStopChan)
		m.randomTicker.Stop()
	}

	m.randomEnabled = globalEnable

	// Check if any register has random enabled (overriding global)
	anyRegisterEnabled := false
	for _, def := range m.definitions {
		if def.RandomRange.Enable {
			anyRegisterEnabled = true
			break
		}
	}

	// Only start if global is enabled OR any individual register is enabled
	if !globalEnable && !anyRegisterEnabled {
		m.logger.Debug().Msg("random updates not enabled, skipping")
		return
	}

	m.logger.Info().
		Bool("global", globalEnable).
		Float64("global_min", globalMin).
		Float64("global_max", globalMax).
		Dur("interval", m.randomInterval).
		Msg("starting random value updates")

	m.randomStopChan = make(chan struct{})
	m.randomTicker = time.NewTicker(m.randomInterval)

	// Start the update goroutine
	go m.runRandomUpdates(globalMin, globalMax)
}

// StopRandomUpdates stops the periodic random value updates.
func (m *Manager) StopRandomUpdates() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.randomStopChan != nil {
		close(m.randomStopChan)
		m.randomTicker.Stop()
		m.randomStopChan = nil
		m.randomTicker = nil
		m.logger.Info().Msg("stopped random value updates")
	}
}

// runRandomUpdates is the goroutine that periodically updates registers with random values.
func (m *Manager) runRandomUpdates(globalMin, globalMax float64) {
	for {
		select {
		case <-m.randomTicker.C:
			m.updateRandomValues(globalMin, globalMax)
		case <-m.randomStopChan:
			return
		}
	}
}

// updateRandomValues updates all enabled registers with random values.
func (m *Manager) updateRandomValues(globalMin, globalMax float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, def := range m.definitions {
		// Determine if this range should be updated
		shouldUpdate := m.randomEnabled || def.RandomRange.Enable
		if !shouldUpdate {
			continue
		}

		// Determine min/max values for this range
		minVal := globalMin
		maxVal := globalMax
		if def.RandomRange.Min != 0 || def.RandomRange.Max != 0 {
			// Use register-specific values if they are set
			if def.RandomRange.Min != 0 {
				minVal = def.RandomRange.Min
			}
			if def.RandomRange.Max != 0 {
				maxVal = def.RandomRange.Max
			}
		}

		// Ensure min < max
		if minVal >= maxVal {
			continue
		}

		// Determine registers per value
		regsPerValue := getRegistersPerValue(def.Type)
		numValues := int(def.Count) / regsPerValue

		// Update each value in the range
		for i := 0; i < numValues; i++ {
			// Generate random value in range [minVal, maxVal)
			randomVal := minVal + rand.Float64()*(maxVal-minVal)

			// Convert to appropriate uint16 representation(s)
			uint16Values := convertToUint16s(randomVal, def.Type, regsPerValue)

			// Write to registers
			startReg := int(def.Address) + i*regsPerValue
			for j, v := range uint16Values {
				addr := startReg + j
				if addr < len(m.memReg.HoldingRegisters) {
					m.memReg.HoldingRegisters[addr] = v
				}
			}
		}
	}
}
