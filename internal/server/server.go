// Package server provides the Modbus TCP and RTU server implementation.
// It wraps the mbserver library and integrates with the simulator.
package server

import (
	"encoding/hex"
	"fmt"
	"net"

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

	// Create server with custom register (use Manager directly)
	s.srv = mbserver.NewServer(
		mbserver.WithRegister(s.sim.RegisterManager()),
	)

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
	// Start TCP listener
	if err := s.srv.ListenTCP(s.sim.Config().ListenAddr); err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.sim.Config().ListenAddr, err)
	}

	// Start server
	go s.srv.Start()

	s.logger.Info().Str("addr", s.sim.Config().ListenAddr).Msg("TCP server started")
	return nil
}

// startRTU starts the Modbus RTU server over a serial port.
func (s *Server) startRTU() error {
	cfg := s.sim.Config()
	if cfg.Serial == nil {
		return fmt.Errorf("serial configuration is required for RTU mode")
	}

	// Open serial port
	serialConfig := &serial.Config{
		Address:  s.sim.Config().ListenAddr,
		BaudRate: cfg.Serial.BaudRate,
		DataBits: cfg.Serial.DataBits,
		StopBits: cfg.Serial.StopBits,
		Parity:   getParity(cfg.Serial.Parity),
	}

	if err := s.srv.ListenRTU(serialConfig); err != nil {
		return fmt.Errorf("failed to open serial port %s: %w", s.sim.Config().ListenAddr, err)
	}

	// Start server
	go s.srv.Start()

	s.logger.Info().
		Str("port", s.sim.Config().ListenAddr).
		Int("baud", cfg.Serial.BaudRate).
		Msg("RTU server started")
	return nil
}

// Stop gracefully stops the Modbus server.
func (s *Server) Stop() error {
	if s.srv != nil {
		s.srv.Shutdown()
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
