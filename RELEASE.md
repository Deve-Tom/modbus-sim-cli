# Release Notes

## v0.0.2 (2025-05-26)

### Features

- **Random Update Interval**: Added configurable random value update interval (in seconds):
  - New global config: `random_interval` (default: 1.0)
  - New CLI flag: `--random-interval`

- **Improved Request/Response Logging**: Enhanced `show_data` feature to log meaningful data instead of raw hex:
  - Logs function names, addresses, counts, and values
  - Supports both TCP and RTU modes
  - TCP mode includes client connection/disconnection events

### Bug Fixes

- **Random Value Fluctuation**: Fixed critical issue where random value updates didn't work in TCP/RTU mode:
  - Implemented proper periodic random value updates with goroutine
  - Fixed configuration passing to register manager
  - RegisterManager now correctly handles both global and per-register random settings

- **Register Interface**: Implemented correct `mbserver.Register` interface in RegisterManager to enable proper data logging:
  - Added full implementation of ReadCoils, ReadDiscreteInputs, ReadHoldingRegisters, ReadInputRegisters
  - Implemented WriteSingleCoil, WriteSingleRegister, WriteMultipleCoils, WriteMultipleRegisters
  - All operations thread-safe with proper mutex locking

### Command-Line Flags

New/updated flags for the `quick` command:

| Flag | Description | Default |
|------|-------------|---------|
| `--random-interval` | Interval in seconds between random value updates | `1.0` |
| `--color` | Enable colored console output | `true` |
| `--random` | Enable random value fluctuation | `false` |
| `--random-min` | Minimum value for random fluctuation | `0` |
| `--random-max` | Maximum value for random fluctuation | `100` |
| `--show-data` | Show request and response data | `false` |

### Configuration Changes

Added global `random_interval` configuration:

```yaml
# Random value fluctuation settings
random_enable: false    # Globally enable random value fluctuation
random_min: 0          # Default minimum value for random fluctuation
random_max: 100        # Default maximum value for random fluctuation
random_interval: 1.0    # Interval in seconds between random value updates

registers:
  - address: 0
    count: 20
    type: FLOAT32
    default_value: 25.5
    random_enable: true
    random_min: 15.0
    random_max: 35.0
    label: "temperature"
```

### Documentation

- Updated README.md and README_zh.md with new features
- Updated example configuration file (configs/example.yaml)

---

## v0.0.1 (2025-05-21)

### Features

- **Multi-Type Register Support**: Added support for various data types beyond UINT16:
  - `INT16` / `UINT16` (1 register per value)
  - `INT32` / `UINT32` / `FLOAT32` (2 registers per value)
  - `INT64` / `UINT64` / `FLOAT64` (4 registers per value)

- **Flexible Value Configuration**: Register values can now be configured via:
  - `values` array: Explicit values for each data point
  - `default_value`: Single value applied to all data points

- **Random Value Fluctuation**: Added global and per-register random value fluctuation control:
  - Global settings: `random_enable`, `random_min`, `random_max`
  - Per-register override: `random_enable`, `random_min`, `random_max`

- **Colored Console Output**: Added `color_output` configuration option (defaults to `true`)

- **Request/Response Data Display**: Added `show_data` configuration option to enable detailed data logging

### Bug Fixes

- **FLOAT32/FLOAT64 Encoding**: Fixed IEEE 754 encoding implementation for FLOAT32 and FLOAT64 types. Uses `math.Float32bits()` and `math.Float64bits()` for proper binary representation.

- **ColorOutput Default Value**: Corrected default value to be `true` when not specified in configuration.

### Command-Line Flags

New flags for the `quick` command:

| Flag | Description | Default |
|------|-------------|---------|
| `--color` | Enable colored console output | `true` |
| `--random` | Enable random value fluctuation | `false` |
| `--random-min` | Minimum value for random fluctuation | `0` |
| `--random-max` | Maximum value for random fluctuation | `100` |
| `--show-data` | Show request and response data | `false` |

### Configuration Changes

The `registers` section in YAML config now supports:

```yaml
registers:
  - address: 0
    count: 20              # Physical 16-bit registers
    type: FLOAT32           # Data type
    default_value: 25.5     # Initial value for all
    random_enable: true     # Enable fluctuation
    random_min: 15.0
    random_max: 35.0
    label: "temperature"

  - address: 200
    count: 20
    type: UINT32
    values: [1000, 2000, 3000]  # Explicit values
    label: "flow_meter"
```

### Documentation

- Comprehensive README rewrite with configuration examples
- Chinese translation (README_zh.md) added
- Example configuration file (configs/example.yaml) updated
