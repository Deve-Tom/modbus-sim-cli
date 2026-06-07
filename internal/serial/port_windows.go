//go:build windows

// Package serial provides raw serial port operations with RS-485 support for Windows.
package serial

import (
	"fmt"
	"time"
)

// ErrTimeout is returned when a read operation times out.
var ErrTimeout = fmt.Errorf("serial: timeout")

// RS485Config configures RS-485 mode (simplified for Windows).
type RS485Config struct {
	Enabled            bool
	RtsHighDuringSend  bool
	RtsHighAfterSend   bool
	RxDuringTx         bool
	DelayRtsBeforeSend time.Duration
	DelayRtsAfterSend  time.Duration
}

// RawPort represents a raw serial port (placeholder for Windows).
type RawPort struct {
	// Placeholder - Windows serial port support would need golang.org/x/sys/windows
}

// OpenRawPort opens a serial port (placeholder for Windows).
// Windows serial port support requires additional implementation using Windows API.
func OpenRawPort(path string, baudRate, dataBits, stopBits int, parity string, rs485 *RS485Config) (*RawPort, error) {
	return nil, fmt.Errorf("raw serial port is not supported on Windows in this build")
}

// Close closes the serial port.
func (p *RawPort) Close() error {
	return nil
}

// ModemLines returns placeholder modem line status for Windows.
func (p *RawPort) ModemLines() (int, error) {
	return 0, fmt.Errorf("ModemLines not supported on Windows")
}

// SetRTS sets RTS (placeholder for Windows).
func (p *RawPort) SetRTS(on bool) error {
	return fmt.Errorf("SetRTS not supported on Windows")
}

// SetDTR sets DTR (placeholder for Windows).
func (p *RawPort) SetDTR(on bool) error {
	return fmt.Errorf("SetDTR not supported on Windows")
}

// Read reads from serial port (placeholder for Windows).
func (p *RawPort) Read(buf []byte, timeout time.Duration) (int, error) {
	return 0, fmt.Errorf("Read not supported on Windows")
}

// FormatModemLines returns a formatted string for modem lines (placeholder for Windows).
func FormatModemLines(status int) string {
	return "ModemLines not supported on Windows"
}
