// Package server provides the Modbus TCP and RTU server implementation.
// It wraps the mbserver library and integrates with the simulator.
package server

import (
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/goburrow/serial"
	"github.com/leijux/mbserver"
	"github.com/rs/zerolog"

	"modbus-sim/internal/simulator"
)

// Server wraps the Modbus server and manages its lifecycle.
type Server struct {
	logger   zerolog.Logger
	sim      *simulator.Simulator
	srv      *mbserver.Server
	showData bool
}

// New creates a new Server with the given simulator.
func New(sim *simulator.Simulator) *Server {
	showData := sim.Config().ShowData != nil && *sim.Config().ShowData
	return &Server{
		logger:   sim.Logger().With().Str("component", "server").Logger(),
		sim:      sim,
		showData: showData,
	}
}

// Start starts the Modbus server based on the configured mode (TCP or RTU).
func (s *Server) Start() error {
	cfg := s.sim.Config()

	// Enable/disable data logging in register manager
	s.sim.RegisterManager().SetShowData(s.showData)

	// Create server with custom register and logging function handlers
	opts := []mbserver.OptionFunc{
		mbserver.WithRegister(s.sim.RegisterManager()),
	}
	opts = append(opts, s.registerLoggingHandlers()...)
	s.srv = mbserver.NewServer(opts...)

	// Start random value updates
	s.sim.StartRandomUpdates()

	switch cfg.Mode {
	case "tcp":
		return s.startTCP()
	case "rtu":
		return s.startRTU()
	default:
		return fmt.Errorf("unsupported server mode: %s (must be 'tcp' or 'rtu')", cfg.Mode)
	}
}

// loggingListener wraps a net.Listener to log connections.
type loggingListener struct {
	net.Listener
	logger zerolog.Logger
}

// Accept implements net.Listener and logs connection info.
func (l *loggingListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	addr := conn.RemoteAddr().String()
	l.logger.Info().Str("client", addr).Msg("client connected")
	return &loggingConn{Conn: conn, logger: l.logger, remoteAddr: addr}, nil
}

// loggingConn wraps a net.Conn to log read/write operations.
type loggingConn struct {
	net.Conn
	logger     zerolog.Logger
	remoteAddr string
}

// Read logs the data read from the connection.
func (c *loggingConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if n > 0 {
		c.logger.Debug().
			Str("client", c.remoteAddr).
			Str("data", hex.EncodeToString(b[:n])).
			Msg("received")
	}
	return
}

// Write logs the data written to the connection.
func (c *loggingConn) Write(b []byte) (n int, err error) {
	if len(b) > 0 {
		c.logger.Debug().
			Str("client", c.remoteAddr).
			Str("data", hex.EncodeToString(b)).
			Msg("sent")
	}
	n, err = c.Conn.Write(b)
	return
}

// startTCP starts the Modbus TCP server.
func (s *Server) startTCP() error {
	// Start TCP listener - mbserver handles the actual TCP server internally
	if err := s.srv.ListenTCP(s.sim.Config().ListenAddr); err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.sim.Config().ListenAddr, err)
	}

	// Start server in background
	go s.srv.Start()

	s.logger.Info().
		Str("addr", s.sim.Config().ListenAddr).
		Uint8("slave_id", s.sim.Config().SlaveID).
		Msg("TCP server started")
	return nil
}

// startRTU starts the Modbus RTU server over a serial port.
func (s *Server) startRTU() error {
	cfg := s.sim.Config()
	if cfg.Serial == nil {
		return fmt.Errorf("serial configuration is required for RTU mode")
	}

	// Determine serial port address: prefer Serial.Address, fallback to ListenAddr for backward compatibility
	serialAddr := cfg.Serial.Address
	if serialAddr == "" {
		serialAddr = cfg.ListenAddr
	}
	if serialAddr == "" {
		return fmt.Errorf("serial port address is required for RTU mode (set serial.address in config)")
	}

	// Open serial port
	serialConfig := &serial.Config{
		Address:  serialAddr,
		BaudRate: cfg.Serial.BaudRate,
		DataBits: cfg.Serial.DataBits,
		StopBits: cfg.Serial.StopBits,
		Parity:   getParity(cfg.Serial.Parity),
		Timeout:  0, // No timeout - mbserver's acceptSerialRequests exits on ErrTimeout, so we must not timeout
	}

	// Configure RS-485 kernel mode if enabled
	if cfg.Serial.RS485 != nil && cfg.Serial.RS485.Enabled {
		serialConfig.RS485 = serial.RS485Config{
			Enabled:            true,
			RtsHighDuringSend:  cfg.Serial.RS485.RtsHighDuringSend,
			RtsHighAfterSend:   cfg.Serial.RS485.RtsHighAfterSend,
			RxDuringTx:         cfg.Serial.RS485.RxDuringTx,
			DelayRtsBeforeSend: time.Duration(cfg.Serial.RS485.DelayRtsBeforeSend) * time.Millisecond,
			DelayRtsAfterSend:  time.Duration(cfg.Serial.RS485.DelayRtsAfterSend) * time.Millisecond,
		}
		s.logger.Info().Msg("RS-485 kernel mode enabled")
	}

	if err := s.srv.ListenRTU(serialConfig); err != nil {
		return fmt.Errorf("failed to open serial port %s: %w", serialAddr, err)
	}

	// Start server
	go s.srv.Start()

	// Log complete serial port configuration for debugging
	logEvent := s.logger.Info().
		Str("port", cfg.Serial.Address).
		Int("baud_rate", cfg.Serial.BaudRate).
		Int("data_bits", cfg.Serial.DataBits).
		Int("stop_bits", cfg.Serial.StopBits).
		Str("parity", cfg.Serial.Parity).
		Uint8("slave_id", cfg.SlaveID)
	if cfg.Serial.RS485 != nil && cfg.Serial.RS485.Enabled {
		logEvent = logEvent.
			Bool("rs485_enabled", true).
			Bool("rs485_rts_high_during_send", cfg.Serial.RS485.RtsHighDuringSend).
			Bool("rs485_rts_high_after_send", cfg.Serial.RS485.RtsHighAfterSend).
			Int("rs485_delay_rts_after_send_ms", cfg.Serial.RS485.DelayRtsAfterSend)
	}
	logEvent.Msg("RTU server started")

	// If showData is enabled, start RTU logging goroutine
	if s.showData {
		go s.startRTULogging(serialAddr)
	}

	return nil
}

// startRTULogging logs RTU serial data. Since RTU doesn't have connection concept,
// we log device information and all data frames.
func (s *Server) startRTULogging(serialPort string) {
	s.logger.Info().
		Str("port", serialPort).
		Msg("RTU data logging started (device connected)")

	// Note: RTU doesn't have connection/disconnection events like TCP.
	// The mbserver library handles serial communication internally.
	// For detailed RTU frame logging, we would need to modify the mbserver library
	// or intercept at a lower level. The current implementation logs through
	// the Register interface methods (ReadCoils, ReadHoldingRegisters, etc.)

	// For now, we'll log a periodic heartbeat to show the logger is active
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.logger.Debug().
				Str("port", serialPort).
				Msg("RTU still listening")
		}
	}
}

// Stop gracefully stops the Modbus server.
// For RTU mode, it uses a timeout since the serial port Read may block
// and prevent mbserver.Shutdown() from completing.
func (s *Server) Stop() error {
	if s.srv != nil {
		done := make(chan struct{})
		go func() {
			s.srv.Shutdown()
			close(done)
		}()

		select {
		case <-done:
			// Shutdown completed normally
		case <-time.After(3 * time.Second):
			s.logger.Warn().Msg("server shutdown timed out, forcing exit")
		}
	}
	// Stop random value updates
	s.sim.StopRandomUpdates()
	s.logger.Info().Msg("server stopped")
	return nil
}

// IsRunning returns whether the server is currently running.
func (s *Server) IsRunning() bool {
	return s.srv != nil
}

// getParity converts a parity string to the goburrow/serial format.
func getParity(p string) string {
	switch p {
	case "odd":
		return "O"
	case "even":
		return "E"
	case "none", "":
		return "N"
	default:
		return "N"
	}
}
