//go:build linux

// Package serial provides raw serial port operations with RTS control
// and RS-485 kernel mode support for Linux.
package serial

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// Modem control line constants
const (
	TIOCM_DTR = 0x002
	TIOCM_RTS = 0x004
	TIOCM_CTS = 0x020
	TIOCM_CAR = 0x040
	TIOCM_RNG = 0x080
	TIOCM_DSR = 0x100

	TIOCMGET = 0x5415
	TIOCMSET = 0x5418

	// RS-485 ioctl and flags
	TIOCSRS485            = 0x542f
	SER_RS485_ENABLED     = 1 << 0
	SER_RS485_RTS_ON_SEND = 1 << 1
	SER_RS485_RTS_AFTER   = 1 << 2
	SER_RS485_RX_DURING   = 1 << 4
)

// ErrTimeout is returned when a read operation times out.
var ErrTimeout = fmt.Errorf("serial: timeout")

// RS485Config configures kernel-level RS-485 mode.
type RS485Config struct {
	Enabled           bool
	RtsHighDuringSend bool
	RtsHighAfterSend  bool
	RxDuringTx        bool
	DelayRtsBeforeSend time.Duration
	DelayRtsAfterSend  time.Duration
}

type rs485IoctlOpts struct {
	flags               uint32
	delayRtsBeforeSend  uint32
	delayRtsAfterSend   uint32
	padding             [5]uint32
}

// Baud rate to termios constant mapping.
var baudRates = map[int]uint32{
	50:      syscall.B50,
	75:      syscall.B75,
	110:     syscall.B110,
	134:     syscall.B134,
	150:     syscall.B150,
	200:     syscall.B200,
	300:     syscall.B300,
	600:     syscall.B600,
	1200:    syscall.B1200,
	1800:    syscall.B1800,
	2400:    syscall.B2400,
	4800:    syscall.B4800,
	9600:    syscall.B9600,
	19200:   syscall.B19200,
	38400:   syscall.B38400,
	57600:   syscall.B57600,
	115200:  syscall.B115200,
	230400:  syscall.B230400,
	460800:  syscall.B460800,
	500000:  syscall.B500000,
	576000:  syscall.B576000,
	921600:  syscall.B921600,
	1000000: syscall.B1000000,
	1152000: syscall.B1152000,
	1500000: syscall.B1500000,
	2000000: syscall.B2000000,
	2500000: syscall.B2500000,
	3000000: syscall.B3000000,
	3500000: syscall.B3500000,
	4000000: syscall.B4000000,
}

// Character size to termios constant mapping.
var charSizes = map[int]uint32{
	5: syscall.CS5,
	6: syscall.CS6,
	7: syscall.CS7,
	8: syscall.CS8,
}

// RawPort represents a raw serial port with full control over RTS/DTR and RS-485 mode.
type RawPort struct {
	f *os.File
}

// OpenRawPort opens a serial port with the specified configuration.
// If rs485 is non-nil and Enabled, kernel RS-485 mode is activated.
func OpenRawPort(path string, baudRate, dataBits, stopBits int, parity string, rs485 *RS485Config) (*RawPort, error) {
	f, err := os.OpenFile(path, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NDELAY|syscall.O_CLOEXEC, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}

	p := &RawPort{f: f}

	if err := p.configure(baudRate, dataBits, stopBits, parity); err != nil {
		f.Close()
		return nil, err
	}

	if rs485 != nil && rs485.Enabled {
		if err := p.EnableRS485(rs485); err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to enable RS-485 mode: %w", err)
		}
	}

	return p, nil
}

// Fd returns the underlying file descriptor.
func (p *RawPort) Fd() int {
	return int(p.f.Fd())
}

// Close closes the serial port.
func (p *RawPort) Close() error {
	return p.f.Close()
}

// Read reads from the serial port with a timeout.
// Returns ErrTimeout if no data is available within the timeout.
func (p *RawPort) Read(buf []byte, timeout time.Duration) (int, error) {
	fd := p.Fd()

	var rfds syscall.FdSet
	fdset(fd, &rfds)

	var tv *syscall.Timeval
	if timeout > 0 {
		t := syscall.NsecToTimeval(timeout.Nanoseconds())
		tv = &t
	}

	_, err := syscall.Select(fd+1, &rfds, nil, nil, tv)
	if err != nil {
		if err == syscall.EINTR {
			return 0, nil
		}
		return 0, fmt.Errorf("select error: %w", err)
	}

	if !fdisset(fd, &rfds) {
		return 0, ErrTimeout
	}

	n, err := syscall.Read(fd, buf)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Write writes data to the serial port.
func (p *RawPort) Write(data []byte) (int, error) {
	return p.f.Write(data)
}

// SetRTS sets the RTS (Request To Send) modem control line.
// For RS-485: RTS low = receive mode, RTS high = transmit mode.
func (p *RawPort) SetRTS(on bool) error {
	var status int
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(p.Fd()), uintptr(TIOCMGET), uintptr(unsafe.Pointer(&status)))
	if errno != 0 {
		return fmt.Errorf("TIOCMGET failed: %v", errno)
	}
	if on {
		status |= TIOCM_RTS
	} else {
		status &^= TIOCM_RTS
	}
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, uintptr(p.Fd()), uintptr(TIOCMSET), uintptr(unsafe.Pointer(&status)))
	if errno != 0 {
		return fmt.Errorf("TIOCMSET failed: %v", errno)
	}
	return nil
}

// SetDTR sets the DTR (Data Terminal Ready) modem control line.
func (p *RawPort) SetDTR(on bool) error {
	var status int
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(p.Fd()), uintptr(TIOCMGET), uintptr(unsafe.Pointer(&status)))
	if errno != 0 {
		return fmt.Errorf("TIOCMGET failed: %v", errno)
	}
	if on {
		status |= TIOCM_DTR
	} else {
		status &^= TIOCM_DTR
	}
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, uintptr(p.Fd()), uintptr(TIOCMSET), uintptr(unsafe.Pointer(&status)))
	if errno != 0 {
		return fmt.Errorf("TIOCMSET failed: %v", errno)
	}
	return nil
}

// ModemLines returns the current state of modem control lines.
func (p *RawPort) ModemLines() (int, error) {
	var status int
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(p.Fd()), uintptr(TIOCMGET), uintptr(unsafe.Pointer(&status)))
	if errno != 0 {
		return 0, fmt.Errorf("TIOCMGET failed: %v", errno)
	}
	return status, nil
}

// EnableRS485 enables kernel-level RS-485 mode via ioctl.
// This allows the kernel to automatically control RTS for RS-485 direction.
func (p *RawPort) EnableRS485(cfg *RS485Config) error {
	opts := rs485IoctlOpts{
		flags:             SER_RS485_ENABLED,
		delayRtsBeforeSend: uint32(cfg.DelayRtsBeforeSend / time.Millisecond),
		delayRtsAfterSend:  uint32(cfg.DelayRtsAfterSend / time.Millisecond),
	}

	if cfg.RtsHighDuringSend {
		opts.flags |= SER_RS485_RTS_ON_SEND
	}
	if cfg.RtsHighAfterSend {
		opts.flags |= SER_RS485_RTS_AFTER
	}
	if cfg.RxDuringTx {
		opts.flags |= SER_RS485_RX_DURING
	}

	r, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(p.Fd()),
		uintptr(TIOCSRS485),
		uintptr(unsafe.Pointer(&opts)))
	if errno != 0 {
		return fmt.Errorf("RS-485 ioctl failed: %v (kernel driver may not support RS-485 mode)", errno)
	}
	if r != 0 {
		return fmt.Errorf("RS-485 ioctl returned non-zero: %d", r)
	}

	return nil
}

// FormatModemLines returns a human-readable description of modem line states.
func FormatModemLines(status int) string {
	flags := []struct {
		bit  int
		name string
	}{
		{TIOCM_RTS, "RTS"},
		{TIOCM_CTS, "CTS"},
		{TIOCM_DTR, "DTR"},
		{TIOCM_DSR, "DSR"},
		{TIOCM_CAR, "CD"},
		{TIOCM_RNG, "RI"},
	}
	result := ""
	for _, f := range flags {
		if status&f.bit != 0 {
			result += f.name + "=high "
		} else {
			result += f.name + "=low "
		}
	}
	return result
}

func (p *RawPort) configure(baudRate, dataBits, stopBits int, parity string) error {
	termios := &syscall.Termios{}

	// Baud rate
	baudFlag, ok := baudRates[baudRate]
	if !ok {
		return fmt.Errorf("unsupported baud rate %d", baudRate)
	}
	termios.Cflag |= baudFlag
	termios.Ispeed = baudFlag
	termios.Ospeed = baudFlag

	// Data bits
	charSize, ok := charSizes[dataBits]
	if !ok {
		return fmt.Errorf("unsupported data bits %d", dataBits)
	}
	termios.Cflag |= charSize

	// Stop bits
	switch stopBits {
	case 1:
		// Default, no CSTOPB
	case 2:
		termios.Cflag |= syscall.CSTOPB
	default:
		return fmt.Errorf("unsupported stop bits %d", stopBits)
	}

	// Parity
	switch parity {
	case "N", "n", "none", "":
		// No parity
	case "E", "e", "even":
		termios.Cflag |= syscall.PARENB
		termios.Iflag |= syscall.INPCK
	case "O", "o", "odd":
		termios.Cflag |= syscall.PARENB | syscall.PARODD
		termios.Iflag |= syscall.INPCK
	default:
		return fmt.Errorf("unsupported parity %s", parity)
	}

	// Enable receiver, ignore modem control lines
	termios.Cflag |= syscall.CREAD | syscall.CLOCAL

	// Raw mode: zero-valued Termios struct has:
	// Iflag = 0 (no input processing)
	// Oflag = 0 (no output processing)
	// Lflag = 0 (no canonical mode, no echo, no signals)
	// This is the correct raw mode configuration.

	// VMIN/VTIME for non-blocking reads (we use select() for timeout)
	termios.Cc[syscall.VMIN] = 0
	termios.Cc[syscall.VTIME] = 0

	// Apply settings
	if _, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(p.Fd()),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(termios))); errno != 0 {
		return fmt.Errorf("failed to set termios: %v", errno)
	}

	return nil
}

// fdset implements FD_SET macro.
func fdset(fd int, fds *syscall.FdSet) {
	idx := fd / (syscall.FD_SETSIZE / len(fds.Bits)) % len(fds.Bits)
	pos := fd % (syscall.FD_SETSIZE / len(fds.Bits))
	fds.Bits[idx] = 1 << uint(pos)
}

// fdisset implements FD_ISSET macro.
func fdisset(fd int, fds *syscall.FdSet) bool {
	idx := fd / (syscall.FD_SETSIZE / len(fds.Bits)) % len(fds.Bits)
	pos := fd % (syscall.FD_SETSIZE / len(fds.Bits))
	return fds.Bits[idx]&(1<<uint(pos)) != 0
}
