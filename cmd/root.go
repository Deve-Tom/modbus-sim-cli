// Package cmd provides the CLI commands for the Modbus simulator.
// It uses cobra for command-line parsing and includes: run, quick, version.
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"modbus-sim/internal/config"
	"modbus-sim/internal/i18n"
	"modbus-sim/internal/server"
	"modbus-sim/internal/simulator"
)

// Version information (can be set via ldflags at build time)
var (
	Version = "dev"
	Commit  = "none"
)

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
	quickMode      string
	quickAddr      string
	quickByteOrder string
	quickRegisters int
)

var quickCmd = &cobra.Command{
	Use:   "quick",
	Short: "Quick start Modbus simulator",
	Long:  "Quick start a Modbus simulation server with command-line flags",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.DefaultConfig()
		cfg.Mode = quickMode
		cfg.ListenAddr = quickAddr
		cfg.ByteOrder = config.ByteOrder(quickByteOrder)

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

// init registers all CLI commands and flags.
func init() {
	// run command flags
	runCmd.Flags().StringVarP(&runCfgFile, "config", "c", "configs/example.yaml",
		"Path to the configuration file")

	// quick command flags
	quickCmd.Flags().StringVarP(&quickMode, "mode", "m", "tcp", "Server mode: tcp or rtu")
	quickCmd.Flags().StringVarP(&quickAddr, "addr", "a", ":502", "Listen address (TCP) or serial port (RTU)")
	quickCmd.Flags().StringVarP(&quickByteOrder, "byte-order", "b", "ABCD", "Byte order: ABCD, DCBA, BADC, CDAB, BDAC")
	quickCmd.Flags().IntVarP(&quickRegisters, "registers", "r", 100, "Number of holding registers to initialize")

	// Register subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(quickCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute runs the root command. This is the main entry point for the CLI.
func Execute() {
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
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
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
