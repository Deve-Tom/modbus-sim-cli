// Package server provides custom Modbus function handlers with detailed logging.
// These handlers replace mbserver's default handlers to add request/response logging
// and slave address verification for RTU mode.
package server

import (
	"encoding/binary"
	"fmt"

	"github.com/leijux/mbserver"
)

// funcCodeName returns a human-readable name for a Modbus function code.
func funcCodeName(code uint8) string {
	switch code {
	case 1:
		return "ReadCoils"
	case 2:
		return "ReadDiscreteInputs"
	case 3:
		return "ReadHoldingRegisters"
	case 4:
		return "ReadInputRegisters"
	case 5:
		return "WriteSingleCoil"
	case 6:
		return "WriteSingleRegister"
	case 15:
		return "WriteMultipleCoils"
	case 16:
		return "WriteMultipleRegisters"
	default:
		return fmt.Sprintf("Unknown(%d)", code)
	}
}

// getSlaveAddress extracts the slave address from a Modbus frame.
// For RTU frames, this returns the slave device address.
// For TCP frames, this returns 0 (Unit ID is in the MBAP header, not in the PDU).
func getSlaveAddress(frame mbserver.Framer) uint8 {
	if rtu, ok := frame.(*mbserver.RTUFrame); ok {
		return rtu.Address
	}
	return 0
}

// isRTUFrame checks if the frame is an RTU frame.
func isRTUFrame(frame mbserver.Framer) bool {
	_, ok := frame.(*mbserver.RTUFrame)
	return ok
}

// registerAddressAndNumber extracts register address and count from frame data.
func registerAddressAndNumber(frame mbserver.Framer) (register, numRegs int) {
	data := frame.GetData()
	return int(binary.BigEndian.Uint16(data[0:2])), int(binary.BigEndian.Uint16(data[2:4]))
}

// registerAddressAndValue extracts register address and value from frame data.
func registerAddressAndValue(frame mbserver.Framer) (int, uint16) {
	data := frame.GetData()
	register := int(binary.BigEndian.Uint16(data[0:2]))
	value := binary.BigEndian.Uint16(data[2:4])
	return register, value
}

// logRequest logs a Modbus request with full context.
func (s *Server) logRequest(frame mbserver.Framer) {
	funcCode := frame.GetFunction()
	slaveAddr := getSlaveAddress(frame)

	evt := s.logger.Info()
	if isRTUFrame(frame) {
		evt = evt.Uint8("slave_addr", slaveAddr)
	}

	switch funcCode {
	case 1, 2, 3, 4:
		reg, numRegs := registerAddressAndNumber(frame)
		evt = evt.Uint8("func", funcCode).
			Str("func_name", funcCodeName(funcCode)).
			Int("register", reg).
			Int("count", numRegs)
	case 5, 6:
		reg, value := registerAddressAndValue(frame)
		evt = evt.Uint8("func", funcCode).
			Str("func_name", funcCodeName(funcCode)).
			Int("register", reg).
			Uint16("value", value)
	case 15, 16:
		reg, numRegs := registerAddressAndNumber(frame)
		evt = evt.Uint8("func", funcCode).
			Str("func_name", funcCodeName(funcCode)).
			Int("register", reg).
			Int("count", numRegs)
	default:
		evt = evt.Uint8("func", funcCode).
			Str("func_name", funcCodeName(funcCode))
	}

	evt.Msg("modbus request")
}

// logResponse logs a Modbus response.
func (s *Server) logResponse(frame mbserver.Framer, data []byte, exception mbserver.Exception) {
	funcCode := frame.GetFunction()
	isException := exception != mbserver.Success

	evt := s.logger.Info()
	if isRTUFrame(frame) {
		evt = evt.Uint8("slave_addr", getSlaveAddress(frame))
	}

	evt = evt.Uint8("func", funcCode).
		Str("func_name", funcCodeName(funcCode))

	if isException {
		evt = evt.Str("exception", exception.Error()).
			Bool("error", true)
	} else {
		// For read operations, log the returned values
		switch funcCode {
		case 3, 4:
			if len(data) > 0 {
				byteCount := int(data[0])
				if byteCount > 0 && len(data) >= 1+byteCount {
					values := mbserver.BytesToUint16(data[1 : 1+byteCount])
					evt = evt.Int("byte_count", byteCount).
						Uints16("values", values)
				}
			}
		case 1, 2:
			if len(data) > 0 {
				evt = evt.Int("byte_count", int(data[0]))
			}
		case 5, 6:
			evt.Msg("write ok")
			return
		case 15, 16:
			evt.Msg("write ok")
			return
		}
	}

	evt.Msg("modbus response")
}

// checkSlaveID checks if the request's slave address matches the configured SlaveID.
// For RTU mode, this is critical as only the matching slave should respond.
// Returns true if the request should be processed.
func (s *Server) checkSlaveID(frame mbserver.Framer) bool {
	if !isRTUFrame(frame) {
		return true
	}

	requestAddr := getSlaveAddress(frame)
	expectedAddr := s.sim.Config().SlaveID

	if requestAddr == 0 {
		// Broadcast address (slave 0) - process but don't respond
		// In Modbus RTU, address 0 is broadcast; slaves process but don't respond
		s.logger.Debug().
			Uint8("broadcast_addr", requestAddr).
			Msg("received broadcast request")
		return true
	}

	if requestAddr != expectedAddr {
		s.logger.Warn().
			Uint8("request_addr", requestAddr).
			Uint8("expected_addr", expectedAddr).
			Msg("slave address mismatch, request ignored")
		return false
	}

	return true
}

// --- Custom function handlers with logging ---

// loggingReadCoils handles function code 1 with logging.
func loggingReadCoils(r mbserver.Register, frame mbserver.Framer) ([]byte, mbserver.Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	if register > 65535 || register+numRegs > 65536 {
		return []byte{}, mbserver.IllegalDataAddress
	}

	dataSize := numRegs / 8
	if (numRegs % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)

	coils, exception := r.ReadCoils(register, numRegs)
	if exception != mbserver.Success {
		return []byte{}, exception
	}

	for i, value := range coils {
		if value {
			shift := uint(i) % 8
			data[1+i/8] |= byte(1 << shift)
		}
	}
	return data, mbserver.Success
}

// loggingReadDiscreteInputs handles function code 2 with logging.
func loggingReadDiscreteInputs(r mbserver.Register, frame mbserver.Framer) ([]byte, mbserver.Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	if register > 65535 || register+numRegs > 65536 {
		return []byte{}, mbserver.IllegalDataAddress
	}

	dataSize := numRegs / 8
	if (numRegs % 8) != 0 {
		dataSize++
	}
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)

	discreteInputs, exception := r.ReadDiscreteInputs(register, numRegs)
	if exception != mbserver.Success {
		return []byte{}, exception
	}

	for i, value := range discreteInputs {
		if value {
			shift := uint(i) % 8
			data[1+i/8] |= byte(1 << shift)
		}
	}
	return data, mbserver.Success
}

// loggingReadHoldingRegisters handles function code 3 with logging.
func loggingReadHoldingRegisters(r mbserver.Register, frame mbserver.Framer) ([]byte, mbserver.Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	if register > 65535 || register+numRegs > 65536 {
		return []byte{}, mbserver.IllegalDataAddress
	}

	hRegisters, exception := r.ReadHoldingRegisters(register, numRegs)
	if exception != mbserver.Success {
		return []byte{}, exception
	}

	data := make([]byte, 1, 1+numRegs*2)
	data[0] = byte(numRegs * 2)
	data = append(data, mbserver.Uint16ToBytes(hRegisters)...)

	return data, mbserver.Success
}

// loggingReadInputRegisters handles function code 4 with logging.
func loggingReadInputRegisters(r mbserver.Register, frame mbserver.Framer) ([]byte, mbserver.Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	if register > 65535 || register+numRegs > 65536 {
		return []byte{}, mbserver.IllegalDataAddress
	}

	iRegisters, exception := r.ReadInputRegisters(register, numRegs)
	if exception != mbserver.Success {
		return []byte{}, exception
	}

	data := make([]byte, 1, 1+numRegs*2)
	data[0] = byte(numRegs * 2)
	data = append(data, mbserver.Uint16ToBytes(iRegisters)...)

	return data, mbserver.Success
}

// loggingWriteSingleCoil handles function code 5 with logging.
func loggingWriteSingleCoil(r mbserver.Register, frame mbserver.Framer) ([]byte, mbserver.Exception) {
	register, value := registerAddressAndValue(frame)
	if value != 0 {
		value = 1
	}

	if exception := r.WriteSingleCoil(register, value != 0); exception != mbserver.Success {
		return []byte{}, exception
	}

	return frame.GetData()[0:4], mbserver.Success
}

// loggingWriteSingleRegister handles function code 6 with logging.
func loggingWriteSingleRegister(r mbserver.Register, frame mbserver.Framer) ([]byte, mbserver.Exception) {
	register, value := registerAddressAndValue(frame)

	if exception := r.WriteSingleRegister(register, value); exception != mbserver.Success {
		return []byte{}, exception
	}

	return frame.GetData()[0:4], mbserver.Success
}

// loggingWriteMultipleCoils handles function code 15 with logging.
func loggingWriteMultipleCoils(r mbserver.Register, frame mbserver.Framer) ([]byte, mbserver.Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	if register > 65535 || numRegs > 65535-register {
		return []byte{}, mbserver.IllegalDataAddress
	}

	expectedBytes := (numRegs + 7) / 8
	if len(valueBytes) < expectedBytes {
		return []byte{}, mbserver.IllegalDataValue
	}

	bitCount := 0
	bitValue := make([]bool, numRegs)

	for i, value := range valueBytes {
		for bitPos := uint(0); bitPos < 8; bitPos++ {
			bitValue[(i*8)+int(bitPos)] = bitAtPosition(value, bitPos) != 0
			bitCount++
			if bitCount >= numRegs {
				break
			}
		}
		if bitCount >= numRegs {
			break
		}
	}

	if exception := r.WriteMultipleCoils(register, bitValue); exception != mbserver.Success {
		return []byte{}, exception
	}

	return frame.GetData()[0:4], mbserver.Success
}

// loggingWriteMultipleRegisters handles function code 16 with logging.
func loggingWriteMultipleRegisters(r mbserver.Register, frame mbserver.Framer) ([]byte, mbserver.Exception) {
	register, numRegs := registerAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]

	if register > 65535 || numRegs > 65535-register {
		return []byte{}, mbserver.IllegalDataAddress
	}

	if len(valueBytes)/2 != numRegs {
		return []byte{}, mbserver.IllegalDataValue
	}

	values := mbserver.BytesToUint16(valueBytes)
	if exception := r.WriteMultipleRegisters(register, values); exception != mbserver.Success {
		return []byte{}, exception
	}

	return frame.GetData()[0:4], mbserver.Success
}

// bitAtPosition returns the bit value at the given position.
func bitAtPosition(value uint8, pos uint) uint8 {
	return (value >> pos) & 0x01
}

// makeLoggingHandler wraps a Modbus function handler with request/response logging
// and slave address verification for RTU mode.
func (s *Server) makeLoggingHandler(fn mbserver.Function) mbserver.Function {
	return func(r mbserver.Register, frame mbserver.Framer) ([]byte, mbserver.Exception) {
		// Log the incoming request
		s.logRequest(frame)

		// Check slave ID for RTU mode
		if !s.checkSlaveID(frame) {
			// For RTU, when the slave address doesn't match,
			// we return success with empty data. The mbserver will still
			// send a response frame, but the master should ignore it
			// since the address won't match.
			// Note: Ideally we'd not respond at all, but mbserver's handler
			// always writes back. This is a known limitation.
			s.logger.Debug().Msg("request for different slave, returning empty response")
			return []byte{}, mbserver.IllegalFunction
		}

		// Execute the actual handler
		data, exception := fn(r, frame)

		// Log the response
		s.logResponse(frame, data, exception)

		return data, exception
	}
}

// registerLoggingHandlers creates logging-wrapped function handlers for all
// standard Modbus function codes.
func (s *Server) registerLoggingHandlers() []mbserver.OptionFunc {
	handlers := map[uint8]mbserver.Function{
		1:  loggingReadCoils,
		2:  loggingReadDiscreteInputs,
		3:  loggingReadHoldingRegisters,
		4:  loggingReadInputRegisters,
		5:  loggingWriteSingleCoil,
		6:  loggingWriteSingleRegister,
		15: loggingWriteMultipleCoils,
		16: loggingWriteMultipleRegisters,
	}

	opts := make([]mbserver.OptionFunc, 0, len(handlers))
	for funcCode, handler := range handlers {
		wrapped := s.makeLoggingHandler(handler)
		opts = append(opts, mbserver.WithRegisterFunction(funcCode, wrapped))
	}

	return opts
}
