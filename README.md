# Modbus-Sim CLI

<div align="center">

**A powerful and flexible Modbus RTU/TCP simulation tool written in Go**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)]()

**English** | [简体中文](README_zh.md)

</div>

## 📖 Overview

Modbus-Sim is a command-line tool that simulates Modbus devices for testing, development, and debugging purposes. It supports both **Modbus TCP** (Ethernet) and **Modbus RTU** (Serial) protocols, making it ideal for:

- 🧪 Testing Modbus client applications
- 🔧 Debugging industrial communication systems
- 📚 Learning Modbus protocol fundamentals
- 🏭 Simulating field devices in development environments
- 🎓 Educational purposes for industrial automation

## ✨ Features

- **Dual Protocol Support**: Full support for both Modbus TCP and Modbus RTU protocols
- **Flexible Configuration**: YAML-based configuration for complex register layouts
- **Quick Start Mode**: One-command startup for rapid testing
- **Multiple Byte Orders**: Support for ABCD, DCBA, BADC, CDAB, and BDAC byte ordering
- **Custom Register Ranges**: Define multiple register blocks with custom initial values
- **Multi-Type Registers**: Support for INT16, UINT16, INT32, UINT32, INT64, UINT64, FLOAT32, FLOAT64 data types
- **Random Value Fluctuation**: Configurable random value changes for simulation
- **Internationalization**: Built-in support for English and Chinese languages
- **Structured Logging**: Configurable log formats (console/JSON) with multiple log levels
- **Colored Output**: Optional colored console output for better readability
- **Request/Response Logging**: Optional detailed data logging for debugging
- **Cross-Platform**: Pre-built binaries for Windows, Linux, macOS (including ARM)
- **Zero Dependencies**: Single binary deployment, no runtime dependencies required

## 🚀 Quick Start

### Option 1: Download Pre-built Binary

Visit the [Releases page](https://github.com/Deve-Tom/modbus-sim-cli/releases) to download pre-built binaries for your platform.

### Option 2: Build from Source

**Prerequisites:**
- Go 1.21 or later
- Git

```bash
# Clone the repository
git clone https://github.com/Deve-Tom/modbus-sim-cli.git
cd modbus-sim-cli

# Build the binary
go build -o modbus-sim .

# Build with version information
go build -ldflags "-X 'modbus-sim/cmd.Version=1.0.0' -X 'modbus-sim/cmd.Commit=$(git rev-parse --short HEAD)'" -o modbus-sim .
```

### Option 3: Install via Go

```bash
go install github.com/Deve-Tom/modbus-sim-cli@latest
```

## 🛠️ Building

### Cross-Platform Compilation

#### Using Make (Linux/macOS)

```bash
# Build for current platform
make build

# Build for all supported platforms
make build-all

# Build for specific platforms
make build-windows      # Windows amd64
make build-linux        # Linux amd64
make build-linux-arm    # Linux ARM (Raspberry Pi)
make build-linux-arm64  # Linux ARM64
make build-macos        # macOS Intel
make build-macos-arm64  # macOS Apple Silicon
```

#### Using Build Script

```bash
# Build for all platforms
go run scripts/build-all.go
```

Build artifacts are stored in the `build/` directory organized by platform:

```
build/
├── windows_amd64/
│   └── modbus-sim.exe
├── linux_amd64/
│   └── modbus-sim
├── linux_arm/
│   └── modbus-sim
├── linux_arm64/
│   └── modbus-sim
├── darwin_amd64/
│   └── modbus-sim
└── darwin_arm64/
    └── modbus-sim
```

## 💻 Usage

### Quick Start Mode

The quickest way to get started is using the `quick` command with sensible defaults.

#### Modbus TCP Server

```bash
# Start TCP server with defaults (port 502, 100 registers, ABCD byte order)
./modbus-sim quick

# Custom TCP server on port 10502 with 200 registers
./modbus-sim quick --mode tcp --addr :10502 --registers 200

# Use BDAC byte order (MidSwap)
./modbus-sim quick --mode tcp --addr :502 --byte-order BDAC --registers 50
```

#### Modbus RTU Server

**Linux:**
```bash
# Start RTU server on /dev/ttyUSB0 with default serial settings
sudo ./modbus-sim quick --mode rtu --addr /dev/ttyUSB0 --registers 50

# Custom baud rate and register count
sudo ./modbus-sim quick --mode rtu --addr /dev/ttyUSB0 --byte-order ABCD --registers 100
```

**macOS:**
```bash
# Common macOS serial port names
./modbus-sim quick --mode rtu --addr /dev/cu.usbserial-XXXX --registers 50
./modbus-sim quick --mode rtu --addr /dev/cu.usbmodem-XXXX --registers 50

# List available serial ports
ls /dev/cu.*
# or
ls /dev/tty.*
```

**Windows:**
```powershell
# Use COM port numbers
.\modbus-sim.exe quick --mode rtu --addr COM3 --registers 50
.\modbus-sim quick --mode rtu --addr COM4 --byte-order ABCD --registers 100

# List available COM ports (PowerShell)
Get-CimInstance Win32_SerialPort | Select-Object DeviceID, Description

# Or use Command Prompt
mode
```

> **Note:** RTU mode requires appropriate permissions to access serial ports.
> - **Linux:** Use `sudo` or add your user to the `dialout` group
> - **macOS:** Add your user to the `dialout` or `uucp` group
> - **Windows:** Run as Administrator if encountering permission issues

### Configuration File Mode

For more complex setups, use a YAML configuration file:

```bash
# Start with default configuration
./modbus-sim run

# Specify custom configuration file
./modbus-sim run --config configs/example.yaml
./modbus-sim run -c my-config.yaml
```

### Command Reference

| Command | Description |
|---------|-------------|
| `quick` | Quick start with command-line flags |
| `run` | Start server with configuration file |
| `version` | Display version information |

### Flags

#### Global Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--lang` | `-l` | Language (en, zh) | `en` |

#### Quick Command Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--mode` | `-m` | Server mode: `tcp` or `rtu` | `tcp` |
| `--addr` | `-a` | Listen address (TCP) or serial port (RTU) | `:502` |
| `--byte-order` | `-b` | Byte order: `ABCD`, `DCBA`, `BADC`, `CDAB`, `BDAC` | `ABCD` |
| `--registers` | `-r` | Number of holding registers to initialize | `100` |
| `--random` | | Enable random value fluctuation | `false` |
| `--random-min` | | Minimum value for random fluctuation | `0` |
| `--random-max` | | Maximum value for random fluctuation | `100` |
| `--random-interval` | | Interval in seconds between random value updates | `1.0` |
| `--color` | | Enable colored console output | `true` |
| `--show-data` | | Show request and response data | `false` |

#### Run Command Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--config` | `-c` | Path to configuration file | `configs/example.yaml` |

## ⚙️ Configuration

### Configuration File Structure

Create a YAML configuration file to define your Modbus device simulation:

```yaml
# Server mode: "tcp" or "rtu"
mode: tcp

# TCP listen address (e.g., ":502" for standard Modbus TCP port)
# For RTU mode, use serial.address instead
listen_addr: ":502"

# Byte order for multi-register values
# Supported values:
#   ABCD - BigEndian (default, most common)
#   DCBA - LittleEndian
#   BADC - BigEndian with word swap
#   CDAB - LittleEndian with word swap
#   BDAC - MidSwap (swap high bytes within each word)
byte_order: ABCD

# Serial port configuration (required for RTU mode, ignored for TCP)
serial:
  address: "/dev/ttyUSB0"  # Serial port device path (required for RTU mode)
  baud_rate: 9600      # Common values: 9600, 19200, 38400, 57600, 115200
  data_bits: 8         # Typically 8
  stop_bits: 1         # 1 or 2
  parity: "none"       # "none", "odd", or "even"

# Display settings
color_output: true      # Enable colored console output
show_data: false        # Enable logging of requests and responses data

# Random value fluctuation settings
random_enable: false    # Globally enable random value fluctuation
random_min: 0           # Default minimum value for random fluctuation
random_max: 100         # Default maximum value for random fluctuation
random_interval: 1.0    # Interval in seconds between random value updates

# Register definitions
# Each entry defines a contiguous range of holding registers.
registers:
  # Temperature sensor registers (addresses 0-19, 20 physical registers)
  # With FLOAT32 type (2 regs per value), this gives 10 FLOAT32 values
  - address: 0
    count: 20
    type: FLOAT32
    label: "temperature_sensor_1"
    default_value: 0     # Used if values array is not specified
    random_enable: true
    random_min: 15.0
    random_max: 35.0

  # Pressure sensor registers (addresses 100-119, 20 physical registers)
  # With FLOAT32 type, this gives 10 FLOAT32 values
  - address: 100
    count: 20
    type: FLOAT32
    label: "pressure_sensor_1"
    default_value: 1013.25

  # Flow meter registers (addresses 200-219, 20 physical registers)
  # With UINT32 type (2 regs per value), this gives 10 UINT32 values
  # Using values array to specify each value explicitly
  - address: 200
    count: 20
    type: UINT32
    label: "flow_meter_1"
    values: [1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000]

  # Device status registers (addresses 300-309, 10 physical registers)
  # With UINT16 type (1 reg per value), this gives 10 UINT16 values
  - address: 300
    count: 10
    type: UINT16
    label: "device_status"
    default_value: 1

# Logging configuration
log_format: console   # "console" for human-readable, "json" for production
log_level: info       # "debug", "info", "warn", or "error"
```

### Register Configuration Fields

Each register entry supports the following fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `address` | uint16 | Yes | Starting register address (0-based) |
| `count` | uint16 | Yes | Number of **physical 16-bit registers** |
| `type` | string | No | Data type (INT16, UINT16, INT32, UINT32, INT64, UINT64, FLOAT32, FLOAT64). Defaults to UINT16 if not specified. |
| `label` | string | No | Human-readable label for this register range |
| `values` | float64[] | No | Array of initial values (one per data value, not per physical register) |
| `default_value` | float64 | No | Initial value for all data values if `values` is not specified |
| `random_enable` | bool | No | Enable random fluctuation for this range |
| `random_min` | float64 | No | Minimum value for random fluctuation |
| `random_max` | float64 | No | Maximum value for random fluctuation |

### How Count and Type Work Together

The `count` field specifies the **number of physical 16-bit registers**, while `type` determines how many registers each data value occupies:

| Type | Registers per Value | Example |
|------|-------------------|---------|
| INT16, UINT16 | 1 | count=10 → 10 values |
| INT32, UINT32, FLOAT32 | 2 | count=20 → 10 values |
| INT64, UINT64, FLOAT64 | 4 | count=20 → 5 values |

**Example:** `count: 20` with `type: FLOAT32` means:
- 20 physical registers
- 10 FLOAT32 values (20 / 2 registers per value)
- Each FLOAT32 value occupies 2 consecutive physical registers

### Initial Values

You can specify initial values using either:

1. **`default_value`**: Single value applied to all data values in the range
   ```yaml
   - address: 0
     count: 20
     type: FLOAT32
     default_value: 25.5  # All 10 FLOAT32 values start at 25.5
   ```

2. **`values`**: Array of values, one per data value
   ```yaml
   - address: 0
     count: 20
     type: FLOAT32
     values: [25.5, 26.0, 26.5, 27.0, 27.5, 28.0, 28.5, 29.0, 29.5, 30.0]
   ```

If neither is specified, all values default to 0.

### Register Configuration Examples

#### Example 1: Simple Device with Sequential Registers

```yaml
registers:
  - address: 0
    count: 50
    value: 0
    label: "all_registers"
```

#### Example 2: Multiple Sensor Groups

```yaml
registers:
  # Temperature sensors (0-19)
  - address: 0
    count: 20
    value: 250  # 25.0°C (assuming scaling factor of 10)
    label: "temp_sensors"

  # Humidity sensors (100-109)
  - address: 100
    count: 10
    value: 500  # 50.0% RH
    label: "humidity_sensors"

  # Pressure sensors (200-209)
  - address: 200
    count: 10
    value: 1013  # 1013 hPa
    label: "pressure_sensors"
```

#### Example 3: Mixed Initial Values

```yaml
registers:
  # Control registers (writable)
  - address: 0
    count: 5
    value: 0
    label: "control_regs"

  # Status registers (read-only simulation)
  - address: 100
    count: 10
    value: 1
    label: "status_regs"

  # Setpoint registers
  - address: 200
    count: 5
    value: 500
    label: "setpoints"
```

### Byte Order Explained

Byte order determines how multi-byte values (like 32-bit integers or floats) are stored across consecutive 16-bit registers.

Consider a 32-bit value `0x12345678` stored in two consecutive registers:

| Byte Order | Register 0 | Register 1 | Description |
|------------|------------|------------|-------------|
| **ABCD** | `0x1234` | `0x5678` | BigEndian (network order, most common) |
| **DCBA** | `0x5678` | `0x1234` | LittleEndian (Intel order) |
| **BADC** | `0x3412` | `0x7856` | BigEndian with byte swap |
| **CDAB** | `0x7856` | `0x3412` | LittleEndian with byte swap |
| **BDAC** | `0x1256` | `0x3478` | MidSwap (word byte swap) |

**Recommendation:** Use `ABCD` unless your target device specifically requires a different byte order. Check your device documentation for the correct setting.

### Random Value Fluctuation

The simulator supports random value fluctuation for simulating dynamic sensor data. This feature can be enabled globally or per-register-range.

#### CLI Usage

```bash
# Enable random fluctuation with default range (0-100)
./modbus-sim quick --random

# Custom value range
./modbus-sim quick --random --random-min 10 --random-max 50

# Custom update interval (update every 2 seconds)
./modbus-sim quick --random --random-interval 2.0
```

#### Configuration File Usage

```yaml
# Global settings (apply to all registers)
random_enable: true
random_min: 0
random_max: 100
random_interval: 1.0  # Update every 1 second by default

# Per-register override
registers:
  - address: 0
    count: 10
    value: 0
    type: UINT16
    random_enable: true   # Override global setting
    random_min: 15.0     # Override global min
    random_max: 35.0     # Override global max
```

### Colored Output

The simulator supports colored console output for better readability. This feature is enabled by default but can be disabled.

#### CLI Usage

```bash
# Disable colored output
./modbus-sim quick --color=false

# Enable colored output (default)
./modbus-sim quick --color=true
```

#### Configuration File Usage

```yaml
color_output: false  # Disable colored output
```

### Request/Response Data Logging

The simulator can log the actual data sent and received, which is useful for debugging Modbus communication.

#### CLI Usage

```bash
# Enable data logging
./modbus-sim quick --show-data
```

#### Configuration File Usage

```yaml
show_data: true  # Enable data logging
```

#### Output Format

When enabled, the simulator logs:
- **TCP Mode**: Client connection/disconnection events and data frames
  ```
  {"level":"info","component":"server","client":"192.168.1.100:54321","message":"client connected"}
  {"level":"debug","component":"server","start":0,"count":10,"values":[0,1,2,3,4,5,6,7,8,9],"message":"ReadHoldingRegisters"}
  {"level":"debug","component":"server","start":0,"value":1234,"message":"WriteSingleRegister"}
  ```
- **RTU Mode**: Data frames (without client info)
  ```
  {"level":"debug","component":"server","start":0,"count":10,"values":[0,1,2,3,4,5,6,7,8,9],"message":"ReadHoldingRegisters"}
  ```

## 🧪 Testing & Integration

### Testing with Modbus Client Tools

#### Python (pymodbus)

```python
from pymodbus.client import ModbusTcpClient

# Connect to the simulator
client = ModbusTcpClient('localhost', port=502)

if client.connect():
    # Read 10 holding registers starting from address 0
    result = client.read_holding_registers(address=0, count=10, slave=1)
    if not result.isError():
        print(f'Register values: {result.registers}')
    
    # Write a single register
    client.write_register(address=0, value=1234, slave=1)
    
    # Write multiple registers
    values = [100, 200, 300, 400, 500]
    client.write_registers(address=0, values=values, slave=1)
    
    client.close()
```

Install: `pip install pymodbus`

#### Python (modbus-tk)

```python
import modbus_tk.modbus_tcp as modbus_tcp

master = modbus_tcp.TcpMaster(host='localhost', port=502)
master.set_timeout(5.0)

# Read registers
result = master.execute(1, modbus_tk.defines.READ_HOLDING_REGISTERS, 0, 10)
print(f'Values: {result}')

# Write register
master.execute(1, modbus_tk.defines.WRITE_SINGLE_REGISTER, 0, output_value=1234)

master.close()
```

Install: `pip install modbus-tk`

#### Node.js (modbus-serial)

```javascript
const ModbusRTU = require("modbus-serial");
const client = new ModbusRTU();

client.connectTCP("localhost", { port: 502 })
  .then(() => {
    client.setID(1);
    
    // Read holding registers
    client.readHoldingRegisters(0, 10, (err, data) => {
      console.log("Values:", data.data);
    });
    
    // Write single register
    client.writeRegister(0, 1234);
  });
```

Install: `npm install modbus-serial`

#### Command Line (Raw TCP)

```bash
# Read holding registers (Function Code 03)
# Read 10 registers starting from address 0
echo -ne '\x00\x01\x00\x00\x00\x06\x01\x03\x00\x00\x00\x0a' | nc -w 2 localhost 502

# Write single register (Function Code 06)
# Write value 1234 to address 0
echo -ne '\x00\x01\x00\x00\x00\x06\x01\x06\x00\x00\x04\xd2' | nc -w 2 localhost 502
```

### Graphical Tools

- **QModBus**: Cross-platform Modbus tester ([Download](https://sourceforge.net/projects/qmodbus/))
- **Modbus Poll**: Professional Modbus master simulator (Windows)
- **CAS Modbus Scanner**: Free Modbus scanning tool
- **Simply Modbus**: Easy-to-use Modbus client

### Testing RTU Mode

For RTU testing, you can use virtual serial ports:

**Linux/macOS:**
```bash
# Install socat for virtual serial ports
sudo apt install socat        # Debian/Ubuntu/Kali
brew install socat            # macOS

# Create a pair of connected virtual serial ports
socat -d -d pty,raw,echo=0 pty,raw,echo=0
# Output: PTY is /dev/pts/5 and PTY is /dev/pts/6

# Terminal 1: Start Modbus RTU simulator
sudo ./modbus-sim quick --mode rtu --addr /dev/pts/5 --registers 50

# Terminal 2: Use a Modbus RTU client to connect to /dev/pts/6
python3 rtu_client.py  # Your client code connecting to /dev/pts/6
```

**Windows:**
```powershell
# Option 1: Use com0com (null modem emulator)
# Download from: https://sourceforge.net/projects/com0com/
# Install and create virtual COM port pairs (e.g., COM10 <-> COM11)

# Terminal 1: Start Modbus RTU simulator
.\modbus-sim.exe quick --mode rtu --addr COM10 --registers 50

# Terminal 2: Use a Modbus RTU client to connect to COM11

# Option 2: Use physical USB-to-Serial adapters
# Connect two USB-to-Serial adapters with a null modem cable
# Or use a single adapter with loopback (TX to RX)
```

### Serial Port Names by Platform

Different operating systems use different naming conventions for serial ports:

| Platform | Device Pattern | Example | Description |
|----------|---------------|---------|-------------|
| **Linux** | `/dev/ttyUSB*` | `/dev/ttyUSB0` | USB-to-Serial adapters (FTDI, CH340, etc.) |
| **Linux** | `/dev/ttyACM*` | `/dev/ttyACM0` | CDC-ACM devices (Arduino, etc.) |
| **Linux** | `/dev/ttyS*` | `/dev/ttyS0` | Built-in serial ports |
| **macOS** | `/dev/cu.usbserial-*` | `/dev/cu.usbserial-1420` | USB-to-Serial adapters |
| **macOS** | `/dev/cu.usbmodem-*` | `/dev/cu.usbmodem14201` | CDC-ACM devices |
| **macOS** | `/dev/cu.*` | `/dev/cu.Bluetooth-Incoming-Port` | All serial devices |
| **Windows** | `COM*` | `COM3`, `COM4` | All serial ports |

**Tips:**
- **Linux:** Use `dmesg | grep tty` to find newly connected devices
- **macOS:** Use `ls /dev/cu.*` to list all available serial ports
- **Windows:** Use Device Manager or `mode` command to find COM ports
- Prefer `/dev/cu.*` over `/dev/tty.*` on macOS for better compatibility

### Supported Modbus Function Codes

| Code | Name | Description | Support |
|------|------|-------------|---------|
| 01 | Read Coils | Read discrete outputs | ✅ |
| 02 | Read Discrete Inputs | Read discrete inputs | ✅ |
| 03 | Read Holding Registers | Read analog outputs | ✅ |
| 04 | Read Input Registers | Read analog inputs | ✅ |
| 05 | Write Single Coil | Write discrete output | ✅ |
| 06 | Write Single Register | Write analog output | ✅ |
| 15 | Write Multiple Coils | Write multiple discrete outputs | ✅ |
| 16 | Write Multiple Registers | Write multiple analog outputs | ✅ |

## 📁 Project Structure

```
modbus-sim-cli/
├── main.go                     # Application entry point
├── Makefile                    # Build automation
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
├── LICENSE                     # MIT License
├── README.md                   # This file (English)
├── README_zh.md                # Chinese version
├── .gitignore                  # Git ignore rules
│
├── cmd/
│   └── root.go                 # CLI commands (cobra framework)
│
├── configs/
│   └── example.yaml            # Example configuration file
│
├── scripts/
│   └── build-all.go            # Cross-platform build script
│
├── build/                      # Compiled binaries (generated)
│   ├── linux_amd64/
│   ├── linux_arm/
│   ├── linux_arm64/
│   ├── darwin_amd64/
│   ├── darwin_arm64/
│   └── windows_amd64/
│
└── internal/
    ├── i18n/                   # Internationalization
    │   ├── i18n.go             # i18n implementation
    │   └── locales/
    │       ├── en.json         # English translations
    │       └── zh.json         # Chinese translations
    │
    ├── config/                 # Configuration management
    │   ├── config.go           # Config loading and validation
    │   └── config_test.go      # Configuration tests
    │
    ├── byteorder/              # Byte order handling
    │   ├── byteorder.go        # Byte order implementations
    │   └── byteorder_test.go   # Byte order tests
    │
    ├── register/               # Register management
    │   └── register.go         # Register storage and operations
    │
    ├── simulator/              # Core simulation engine
    │   └── simulator.go        # Simulator orchestration
    │
    └── server/                 # Network server
        └── server.go           # TCP/RTU server implementation
```

## 🔍 Troubleshooting

### Common Issues

#### 1. Port 502 Permission Denied

**Problem:** Cannot bind to port 502 (privileged port)

**Solution:**
```bash
# Option 1: Use sudo (Linux/macOS)
sudo ./modbus-sim quick

# Option 2: Use a non-privileged port (>1024)
./modbus-sim quick --addr :10502

# Option 3: Set CAP_NET_BIND_SERVICE capability (Linux)
sudo setcap 'cap_net_bind_service=+ep' ./modbus-sim

# Windows: Run as Administrator or use port >1024
.\modbus-sim.exe quick --addr :10502
```

#### 2. Serial Port Access Denied (RTU Mode)

**Problem:** Cannot open serial port

**Linux Solution:**
```bash
# Option 1: Use sudo
sudo ./modbus-sim quick --mode rtu --addr /dev/ttyUSB0

# Option 2: Add user to dialout group
sudo usermod -a -G dialout $USER
# Logout and login again

# Option 3: Change device permissions (temporary)
sudo chmod 666 /dev/ttyUSB0
```

**macOS Solution:**
```bash
# Option 1: Add user to dialout or uucp group
sudo dscl . append /Groups/dialout GroupMembership $(whoami)
# or
sudo dscl . append /Groups/uucp GroupMembership $(whoami)

# Option 2: Use cu.* devices instead of tty.*
./modbus-sim quick --mode rtu --addr /dev/cu.usbserial-XXXX

# List available ports
ls -l /dev/cu.*
ls -l /dev/tty.*
```

**Windows Solution:**
```powershell
# Option 1: Run as Administrator
# Right-click on PowerShell or Command Prompt -> "Run as Administrator"
.\modbus-sim.exe quick --mode rtu --addr COM3

# Option 2: Check if COM port is in use
# PowerShell
Get-CimInstance Win32_SerialPort | Select-Object DeviceID, Description

# Command Prompt
mode

# Option 3: Try different COM port number
.\modbus-sim.exe quick --mode rtu --addr COM4
```

#### 3. No Response to Modbus Requests

**Checklist:**
- Verify server is running and listening on the correct port
- Check firewall rules allow connections
- Ensure correct slave/unit ID (default is 1)
- Verify byte order matches your client expectations
- Check logs for error messages

```bash
# Check if port is listening
netstat -tlnp | grep 502
ss -tlnp | grep 502

# Check running processes
ps aux | grep modbus-sim

# Test connectivity
telnet localhost 502
nc -zv localhost 502
```

#### 4. Incorrect Register Values

**Possible causes:**
- Wrong byte order setting
- Wrong register address (Modbus uses 0-based addressing)
- Data type mismatch (16-bit vs 32-bit values)

**Debug tip:** Enable debug logging to see raw requests and responses:
```yaml
log_level: debug
log_format: console
```

### Getting Help

- 📖 Check this README for usage examples
- 🐛 Report issues on [GitHub Issues](https://github.com/Deve-Tom/modbus-sim-cli/issues)
- 💬 Discuss features and ask questions in [Discussions](https://github.com/Deve-Tom/modbus-sim-cli/discussions)

## 📊 Use Cases

### Industrial Automation Testing

Simulate PLCs, HMIs, sensors, and actuators during development:

```bash
# Simulate a temperature controller
./modbus-sim quick --mode tcp --addr :502 --registers 100

# Client application can now read/write temperature setpoints
```

### SCADA System Development

Test SCADA applications without physical hardware:

```yaml
# scada-test.yaml
mode: tcp
listen_addr: ":502"
registers:
  - address: 0
    count: 50
    value: 0
    label: "analog_inputs"
  - address: 100
    count: 50
    value: 0
    label: "digital_inputs"
  - address: 200
    count: 50
    value: 0
    label: "control_outputs"
```

### IoT Gateway Testing

Validate IoT gateway Modbus integration:

```bash
# Simulate multiple field devices on different ports
./modbus-sim quick --mode tcp --addr :5020 --registers 50 &
./modbus-sim quick --mode tcp --addr :5021 --registers 50 &
./modbus-sim quick --mode tcp --addr :5022 --registers 50 &
```

### Education and Training

Learn Modbus protocol hands-on:

```bash
# Start simulator
./modbus-sim quick --mode tcp --addr :502

# Practice with different tools
# Try reading/writing registers using Python, Node.js, or command line
```

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/modbus-sim-cli.git
cd modbus-sim-cli

# Install dependencies
go mod download

# Run tests
go test ./... -v

# Build and test locally
go build -o modbus-sim .
./modbus-sim quick
```

### Code Standards

- Follow Go best practices and idioms
- Write tests for new functionality
- Update documentation as needed
- Use meaningful commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [mbserver](https://github.com/leijux/mbserver) - Modbus server library
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [zerolog](https://github.com/rs/zerolog) - Structured logging
- [goburrow/serial](https://github.com/goburrow/serial) - Serial port communication

## 📞 Support

- 📧 **Issues**: [GitHub Issues](https://github.com/Deve-Tom/modbus-sim-cli/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/Deve-Tom/modbus-sim-cli/discussions)
- 📖 **Documentation**: This README and code comments

---

<div align="center">

**Made with ❤️ using Go**

[⭐ Star this repo](https://github.com/Deve-Tom/modbus-sim-cli) if you find it useful!

</div>
