// Package cmd provides the CLI commands for the Modbus simulator.
// It uses cobra for command-line parsing and includes: run, quick, version.
package cmd

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"modbus-sim/internal/config"
	"modbus-sim/internal/i18n"
	ser "modbus-sim/internal/serial"
	"modbus-sim/internal/server"
	"modbus-sim/internal/simulator"
)

// Version information (can be set via ldflags at build time)
var (
	Version = "dev"
	Commit  = "none"
)

var langFlag string

var rootCmd = &cobra.Command{
	Use:   "modbus-sim",
	Short: "Modbus RTU/TCP Data Simulation CLI Tool",
	Long:  "Modbus RTU/TCP Data Simulation CLI Tool",
}

// --- run command ---

var runCfgFile string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start Modbus simulator with config file",
	Long:  "Start the Modbus simulation server with a configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration from file
		cfg, err := config.LoadFromFile(runCfgFile)
		if err != nil {
			log.Error().Err(err).Msg(i18n.T("ConfigLoadFailed", map[string]interface{}{
				"Error": err.Error(),
			}))
			return err
		}

		log.Info().Str("file", runCfgFile).Msg(
			i18n.T("ConfigLoaded", map[string]interface{}{"File": runCfgFile}))

		return startSimulator(cfg)
	},
}

// --- quick command ---

var (
	quickMode           string
	quickAddr           string
	quickSlaveID        uint8
	quickByteOrder      string
	quickRegisters      int
	quickRandom         bool
	quickRandomMin      float64
	quickRandomMax      float64
	quickRandomInterval float64
	quickColor          bool
	quickShowData       bool
)

var quickCmd = &cobra.Command{
	Use:   "quick",
	Short: "Quick start Modbus simulator",
	Long:  "Quick start a Modbus simulation server with command-line flags",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.DefaultConfig()
		cfg.Mode = quickMode
		cfg.ListenAddr = quickAddr
		cfg.SlaveID = quickSlaveID
		cfg.ByteOrder = config.ByteOrder(quickByteOrder)
		cfg.RandomEnable = &quickRandom
		cfg.RandomMin = quickRandomMin
		cfg.RandomMax = quickRandomMax
		cfg.RandomInterval = quickRandomInterval
		cfg.ColorOutput = &quickColor
		cfg.ShowData = &quickShowData

		log.Info().Msg(
			i18n.T("QuickStartInfo", map[string]interface{}{
				"Mode":      quickMode,
				"Addr":      quickAddr,
				"Registers": quickRegisters,
			}))

		// Create simulator and initialize simple registers
		sim, err := simulator.New(cfg, log.Logger)
		if err != nil {
			return fmt.Errorf("failed to create simulator: %w", err)
		}

		if err := sim.InitSimpleRegisters(0, uint16(quickRegisters), 0); err != nil {
			return fmt.Errorf("failed to initialize registers: %w", err)
		}

		// Create and start server
		srv := server.New(sim)
		if err := srv.Start(); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}

		// Wait for shutdown signal
		waitForShutdown(srv)

		return nil
	},
}

// --- version command ---

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(
			i18n.T("VersionInfo", map[string]interface{}{
				"Version": Version,
				"Commit":  Commit,
			}))
	},
}

// --- serial-dump command ---

var (
	dumpPort      string
	dumpBaud      int
	dumpDataBits  int
	dumpStopBits  int
	dumpParity    string
	dumpRTSLow    bool
	dumpRS485     bool
	dumpDTRLow    bool
)

var serialDumpCmd = &cobra.Command{
	Use:   "serial-dump",
	Short: "Dump raw serial port data for debugging",
	Long:  "Open a serial port and print all received data in hex format. Useful for diagnosing RS-485/serial connectivity issues.\n\nThis tool provides RTS/DTR control and RS-485 kernel mode support to help diagnose\nwhy a serial port may not be receiving data.\n\nCommon RS-485 issue: the transceiver is stuck in transmit mode because RTS is high.\nUse --rts-low to force RTS low and enable receive mode.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Setup console logger
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

		parity := dumpParity

		// Build RS-485 config if requested
		var rs485Cfg *ser.RS485Config
		if dumpRS485 {
			rs485Cfg = &ser.RS485Config{
				Enabled:           true,
				RtsHighDuringSend: true,
				RtsHighAfterSend:  false,
				RxDuringTx:        false,
				DelayRtsAfterSend: 1 * time.Millisecond,
			}
		}

		log.Info().
			Str("port", dumpPort).
			Int("baud", dumpBaud).
			Int("data_bits", dumpDataBits).
			Int("stop_bits", dumpStopBits).
			Str("parity", parity).
			Bool("rts_low", dumpRTSLow).
			Bool("dtr_low", dumpDTRLow).
			Bool("rs485_kernel", dumpRS485).
			Msg("Opening serial port for raw data dump...")

		port, err := ser.OpenRawPort(dumpPort, dumpBaud, dumpDataBits, dumpStopBits, parity, rs485Cfg)
		if err != nil {
			return fmt.Errorf("failed to open serial port %s: %w", dumpPort, err)
		}
		defer port.Close()

		// Show modem line status after opening
		if lines, err := port.ModemLines(); err == nil {
			log.Info().Str("modem_lines", ser.FormatModemLines(lines)).Msg("Modem line status after open")
		}

		// Set RTS low if requested (for RS-485 receive mode)
		if dumpRTSLow {
			if err := port.SetRTS(false); err != nil {
				log.Warn().Err(err).Msg("Failed to set RTS low")
			} else {
				log.Info().Msg("RTS set LOW (RS-485 receive mode enabled)")
			}
		}

		// Set DTR low if requested
		if dumpDTRLow {
			if err := port.SetDTR(false); err != nil {
				log.Warn().Err(err).Msg("Failed to set DTR low")
			} else {
				log.Info().Msg("DTR set LOW")
			}
		}

		// Show modem line status after RTS/DTR control
		if lines, err := port.ModemLines(); err == nil {
			log.Info().Str("modem_lines", ser.FormatModemLines(lines)).Msg("Modem line status after configuration")
		}

		log.Info().Msg("Serial port opened. Waiting for data... (Ctrl+C to stop)")

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		totalBytes := 0
		buf := make([]byte, 512)

		for {
			select {
			case <-sigCh:
				log.Info().Int("total_bytes", totalBytes).Msg("Stopped")
				return nil
			default:
			}

			n, err := port.Read(buf, 1*time.Second)
			if err != nil {
				if err == ser.ErrTimeout {
					// Timeout is normal, just continue
					continue
				}
				log.Error().Err(err).Msg("Read error")
				continue
			}
			if n > 0 {
				totalBytes += n
				data := buf[:n]
				log.Info().
					Int("bytes", n).
					Int("total", totalBytes).
					Str("hex", hex.EncodeToString(data)).
					Msg("received")

				// Also print a formatted byte-by-byte view for easier analysis
				fmt.Printf("  Raw bytes: ")
				for i, b := range data {
					if i > 0 {
						fmt.Printf(" ")
					}
					fmt.Printf("0x%02X", b)
				}
				fmt.Println()
			}
		}
	},
}

// init registers all CLI commands and flags.
func init() {
	// root command flags
	rootCmd.PersistentFlags().StringVarP(&langFlag, "lang", "l", "en", "Language for output (en, zh)")

	// run command flags
	runCmd.Flags().StringVarP(&runCfgFile, "config", "c", "configs/example.yaml",
		"Path to the configuration file")

	// quick command flags
	quickCmd.Flags().StringVarP(&quickMode, "mode", "m", "tcp", "Server mode: tcp or rtu")
	quickCmd.Flags().StringVarP(&quickAddr, "addr", "a", ":502", "Listen address (TCP) or serial port (RTU)")
	quickCmd.Flags().Uint8VarP(&quickSlaveID, "slave-id", "s", 1, "Modbus slave/device address (1-247)")
	quickCmd.Flags().StringVarP(&quickByteOrder, "byte-order", "b", "ABCD", "Byte order: ABCD, DCBA, BADC, CDAB, BDAC")
	quickCmd.Flags().IntVarP(&quickRegisters, "registers", "r", 100, "Number of holding registers to initialize")
	quickCmd.Flags().BoolVarP(&quickRandom, "random", "", false, "Enable random value fluctuation")
	quickCmd.Flags().Float64VarP(&quickRandomMin, "random-min", "", 0, "Minimum value for random fluctuation")
	quickCmd.Flags().Float64VarP(&quickRandomMax, "random-max", "", 100, "Maximum value for random fluctuation")
	quickCmd.Flags().Float64VarP(&quickRandomInterval, "random-interval", "", 1.0, "Interval in seconds between random value updates")
	quickCmd.Flags().BoolVarP(&quickColor, "color", "", true, "Enable colored console output")
	quickCmd.Flags().BoolVarP(&quickShowData, "show-data", "", false, "Show request and response data")

	// Register subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(quickCmd)
	rootCmd.AddCommand(versionCmd)

	// serial-dump command flags
	serialDumpCmd.Flags().StringVarP(&dumpPort, "port", "p", "/dev/ttyAMA3", "Serial port device path")
	serialDumpCmd.Flags().IntVarP(&dumpBaud, "baud", "b", 9600, "Baud rate")
	serialDumpCmd.Flags().IntVarP(&dumpDataBits, "data-bits", "d", 8, "Data bits (7 or 8)")
	serialDumpCmd.Flags().IntVarP(&dumpStopBits, "stop-bits", "", 1, "Stop bits (1 or 2)")
	serialDumpCmd.Flags().StringVarP(&dumpParity, "parity", "", "none", "Parity: none, even, odd")
	serialDumpCmd.Flags().BoolVar(&dumpRTSLow, "rts-low", true, "Force RTS low (enable RS-485 receive mode)")
	serialDumpCmd.Flags().BoolVar(&dumpDTRLow, "dtr-low", false, "Force DTR low")
	serialDumpCmd.Flags().BoolVar(&dumpRS485, "rs485", false, "Enable kernel RS-485 mode (auto RTS control)")
	rootCmd.AddCommand(serialDumpCmd)
}

// updateCommandDescriptions updates command descriptions with i18n translations.
// This must be called after i18n.MustInit() in main().
func updateCommandDescriptions() {
	rootCmd.Short = i18n.T("AppDescription", nil)
	rootCmd.Long = i18n.T("AppDescription", nil)

	runCmd.Short = i18n.T("RunShort", nil)
	runCmd.Long = i18n.T("RunLong", nil)

	quickCmd.Short = i18n.T("QuickShort", nil)
	quickCmd.Long = i18n.T("QuickLong", nil)

	// Update flag usages
	runCmd.Flags().Lookup("config").Usage = i18n.T("FlagConfigUsage", nil)
	quickCmd.Flags().Lookup("mode").Usage = i18n.T("FlagModeUsage", nil)
	quickCmd.Flags().Lookup("addr").Usage = i18n.T("FlagAddrUsage", nil)
	quickCmd.Flags().Lookup("byte-order").Usage = i18n.T("FlagByteOrderUsage", nil)
	quickCmd.Flags().Lookup("registers").Usage = i18n.T("FlagRegistersUsage", nil)
}

// Execute runs the root command. This is the main entry point for the CLI.
func Execute() {
	rootCmd.ParseFlags(os.Args[1:])
	if langFlag != "" {
		i18n.SetLanguage(langFlag)
	}
	updateCommandDescriptions()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// startSimulator creates a simulator from config, starts the server, and waits for shutdown.
func startSimulator(cfg *config.Config) error {
	// Setup logger
	setupLogger(cfg)

	log.Info().Msg(i18n.T("ServerStarting", nil))

	// Create simulator
	sim, err := simulator.New(cfg, log.Logger)
	if err != nil {
		return fmt.Errorf("failed to create simulator: %w", err)
	}

	// Initialize registers
	if err := sim.InitRegisters(); err != nil {
		return fmt.Errorf("failed to initialize registers: %w", err)
	}

	log.Info().Msg(
		i18n.T("RegisterInitialized", map[string]interface{}{
			"Count": sim.RegisterManager().Count(),
		}))

	// Create and start server
	srv := server.New(sim)
	if err := srv.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	log.Info().Msg(i18n.T("ServerStarted", nil))

	// Wait for shutdown signal
	waitForShutdown(srv)

	return nil
}

// setupLogger configures zerolog based on the config settings.
func setupLogger(cfg *config.Config) {
	switch cfg.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if cfg.LogFormat == "json" {
		// JSON format for production
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	} else {
		// Console format for development
		colorOutput := cfg.ColorOutput != nil && *cfg.ColorOutput
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: !colorOutput})
	}
}

// waitForShutdown blocks until an interrupt signal is received, then stops the server.
func waitForShutdown(srv *server.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh

	log.Info().Str("signal", sig.String()).Msg("received shutdown signal")

	if err := srv.Stop(); err != nil {
		log.Error().Err(err).Msg("error stopping server")
	}

	log.Info().Msg(i18n.T("ServerStopped", nil))
}
