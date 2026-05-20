# Modbus-Sim

A Modbus RTU/TCP Data Simulation CLI Tool written in Go.

## Features

- **Modbus RTU/TCP Server**: Simulates both TCP and RTU (serial) Modbus servers
- **Custom Register Configuration**: Define register ranges with initial values and labels via YAML
- **Multiple Byte Orders**: Supports ABCD (BigEndian), DCBA (LittleEndian), BADC, CDAB, BDAC (MidSwap)
- **Internationalization**: Supports English and Chinese languages
- **Structured Logging**: Configurable log output (console or JSON) with different log levels

## Installation

```bash
# Clone the repository
git clone &lt;repository-url&gt;
cd modbus-sim

# Build the project
go build -o modbus-sim .

# Or build with version info
go build -ldflags "-X main.Version=1.0.0 -X main.Commit=$(git rev-parse --short HEAD)" -o modbus-sim .
```

## Cross-Platform Build

### Using Make (Linux/macOS):

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build for specific platform
make build-windows    # Windows (amd64)
make build-linux      # Linux (amd64)
make build-linux-arm  # Linux ARM (32-bit, e.g., Raspberry Pi)
make build-linux-arm64 # Linux ARM64
make build-macos      # macOS (amd64)
make build-macos-arm64 # macOS ARM64 (Apple Silicon)
```

### Using Script (All platforms):

```bash
go run scripts/build-all.go
```

Build output will be in the `build/` directory.

## Usage

### Quick Start

Start a TCP server with default settings (100 registers, ABCD byte order, port 502):

```bash
./modbus-sim quick
```

Start with custom parameters:

```bash
./modbus-sim quick --mode tcp --addr :10502 --byte-order BDAC --registers 200
```

Start an RTU server:

```bash
./modbus-sim quick --mode rtu --addr /dev/ttyUSB0 --byte-order ABCD --registers 50
```

### Using Configuration Files

```bash
./modbus-sim run -c configs/example.yaml
```

### Commands

- `run` - Start the Modbus simulation server with a configuration file
- `quick` - Quick start a Modbus simulation server with command-line flags
- `version` - Print version information

### Flags

**run command:**
- `-c, --config` - Path to the configuration file (default: `configs/example.yaml`)

**quick command:**
- `-m, --mode` - Server mode: `tcp` or `rtu` (default: `tcp`)
- `-a, --addr` - Listen address (TCP) or serial port (RTU) (default: `:502`)
- `-b, --byte-order` - Byte order: `ABCD`, `DCBA`, `BADC`, `CDAB`, `BDAC` (default: `ABCD`)
- `-r, --registers` - Number of holding registers to initialize (default: `100`)

## Configuration

Example configuration file (`configs/example.yaml`):

```yaml
mode: tcp
listen_addr: ":502"
byte_order: ABCD

serial:
  baud_rate: 9600
  data_bits: 8
  stop_bits: 1
  parity: "none"

registers:
  - address: 0
    count: 10
    value: 0
    label: "temperature_sensor_1"
  - address: 100
    count: 10
    value: 0
    label: "pressure_sensor_1"

log_format: console
log_level: info
```

### Byte Order

The `byte_order` setting controls how multi-byte values are encoded:

| Order | Description |
|-------|-------------|
| `ABCD` | BigEndian (default) |
| `DCBA` | LittleEndian |
| `BADC` | BigEndian with word swap |
| `CDAB` | LittleEndian with word swap |
| `BDAC` | MidSwap (swap high bytes within each word) |

## Testing

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## Project Structure

```
modbus-sim/
├── main.go                     # Main entry point
├── Makefile                    # Build targets for cross-platform compilation
├── cmd/
│   └── root.go                 # CLI commands (cobra)
├── scripts/
│   └── build-all.go            # Cross-platform build script
├── configs/
│   └── example.yaml            # Example configuration
└── internal/
    ├── i18n/                   # Internationalization
    │   ├── i18n.go
    │   └── locales/
    │       ├── en.json
    │       └── zh.json
    ├── config/                 # Configuration loading
    │   ├── config.go
    │   └── config_test.go
    ├── byteorder/              # Byte order encoding
    │   ├── byteorder.go
    │   └── byteorder_test.go
    ├── register/               # Register management
    ├── simulator/              # Core simulation engine
    └── server/                 # TCP/RTU server
```

## License

MIT License
