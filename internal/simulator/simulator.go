// Package simulator orchestrates the Modbus simulation, managing registers,
// byte order, and providing the core simulation logic.
package simulator

import (
	"fmt"

	"github.com/leijux/mbserver"
	"github.com/rs/zerolog"

	"modbus-sim/internal/byteorder"
	"modbus-sim/internal/config"
	"modbus-sim/internal/register"
)

// Simulator is the core simulation engine that manages registers and byte ordering.
type Simulator struct {
	cfg           *config.Config
	logger        zerolog.Logger
	regManager    *register.Manager
	byteOrderImpl byteorder.Order
}

// New creates a new Simulator with the given configuration and logger.
func New(cfg *config.Config, logger zerolog.Logger) (*Simulator, error) {
	// Resolve byte order implementation
	bo, err := byteorder.Resolve(cfg.ByteOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve byte order: %w", err)
	}

	sim := &Simulator{
		cfg:           cfg,
		logger:        logger.With().Str("component", "simulator").Logger(),
		regManager:    register.NewManager(logger),
		byteOrderImpl: bo,
	}

	return sim, nil
}

// InitRegisters initializes registers from the configuration.
func (s *Simulator) InitRegisters() error {
	regDefs := make([]register.RegisterConfig, len(s.cfg.Registers))
	for i, rc := range s.cfg.Registers {
		// Determine type, defaulting to UINT16 if not specified
		regType := string(rc.Type)
		if regType == "" {
			regType = "UINT16"
		}

		// Determine registers per value based on type
		regsPerValue := getRegistersPerValue(regType)
		numValues := int(rc.Count) / regsPerValue

		// Prepare values - use Values array if specified, otherwise use DefaultValue
		var values []float64
		if len(rc.Values) > 0 {
			values = rc.Values
		} else if rc.DefaultValue != 0 {
			// Use DefaultValue for all values
			values = make([]float64, numValues)
			for j := range values {
				values[j] = rc.DefaultValue
			}
		}
		// If both are empty/nil, register.InitFromConfig will use zeros

		regDefs[i] = register.RegisterConfig{
			Address: rc.Address,
			Count:   rc.Count,
			Type:    regType,
			Values:  values,
			Label:   rc.Label,
		}
	}

	if len(regDefs) == 0 {
		s.logger.Warn().Msg("no register definitions in config, using default 100 registers")
		return s.regManager.InitSimple(0, 100, 0)
	}

	return s.regManager.InitFromConfig(regDefs)
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

// InitSimpleRegisters initializes a simple contiguous range of registers.
func (s *Simulator) InitSimpleRegisters(startAddr uint16, count uint16, defaultValue uint16) error {
	return s.regManager.InitSimple(startAddr, count, defaultValue)
}

// RegisterManager returns the underlying register manager for integration with mbserver.
func (s *Simulator) RegisterManager() *register.Manager {
	return s.regManager
}

// MemRegister returns the underlying mbserver.MemRegister for integration with mbserver.
func (s *Simulator) MemRegister() *mbserver.MemRegister {
	return s.regManager.GetMemRegister()
}

// ByteOrder returns the resolved byte order implementation.
func (s *Simulator) ByteOrder() byteorder.Order {
	return s.byteOrderImpl
}

// Config returns the simulator configuration.
func (s *Simulator) Config() *config.Config {
	return s.cfg
}

// Logger returns the simulator logger.
func (s *Simulator) Logger() zerolog.Logger {
	return s.logger
}

// EncodeUint32 encodes a uint32 value into a 4-byte slice using the configured byte order.
func (s *Simulator) EncodeUint32(v uint32) []byte {
	buf := make([]byte, 4)
	s.byteOrderImpl.PutUint32(buf, v)
	return buf
}

// DecodeUint32 decodes a uint32 value from a 4-byte slice using the configured byte order.
func (s *Simulator) DecodeUint32(buf []byte) uint32 {
	return s.byteOrderImpl.Uint32(buf)
}

// EncodeUint64 encodes a uint64 value into an 8-byte slice using the configured byte order.
func (s *Simulator) EncodeUint64(v uint64) []byte {
	buf := make([]byte, 8)
	s.byteOrderImpl.PutUint64(buf, v)
	return buf
}

// DecodeUint64 decodes a uint64 value from an 8-byte slice using the configured byte order.
func (s *Simulator) DecodeUint64(buf []byte) uint64 {
	return s.byteOrderImpl.Uint64(buf)
}

// Info returns a summary of the simulator state.
func (s *Simulator) Info() string {
	return fmt.Sprintf("mode=%s addr=%s byte_order=%s registers=%d",
		s.cfg.Mode, s.cfg.ListenAddr, s.cfg.ByteOrder, s.regManager.Count())
}
