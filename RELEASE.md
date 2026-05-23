# Release Notes

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
