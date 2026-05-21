# Modbus-Sim CLI

<div align="center">

**一个用 Go 语言编写的强大且灵活的 Modbus RTU/TCP 仿真工具**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)]()

[English](README.md) | **简体中文**

</div>

## 📖 概述

Modbus-Sim 是一个命令行工具，用于仿真 Modbus 设备，适用于测试、开发和调试。它同时支持 **Modbus TCP**（以太网）和 **Modbus RTU**（串行）协议，非常适合：

- 🧪 测试 Modbus 客户端应用程序
- 🔧 调试工业通信系统
- 📚 学习 Modbus 协议基础知识
- 🏭 在开发环境中仿真现场设备
- 🎓 工业自动化教学用途

## ✨ 特性

- **双协议支持**：完整支持 Modbus TCP 和 Modbus RTU 协议
- **灵活配置**：基于 YAML 的配置，支持复杂的寄存器布局
- **快速启动模式**：一键启动，快速测试
- **多种字节序**：支持 ABCD、DCBA、BADC、CDAB 和 BDAC 字节序
- **自定义寄存器范围**：定义多个寄存器块，设置自定义初始值
- **国际化**：内置支持英文和中文
- **结构化日志**：可配置的日志格式（控制台/JSON）和多个日志级别
- **跨平台**：为 Windows、Linux、macOS（包括 ARM）提供预构建二进制文件
- **零依赖**：单二进制文件部署，无需运行时依赖

## 🚀 快速开始

### 选项 1：下载预构建二进制文件

访问 [Releases 页面](https://github.com/Deve-Tom/modbus-sim-cli/releases) 下载适合您平台的预构建二进制文件。

### 选项 2：从源码构建

**前置要求：**
- Go 1.21 或更高版本
- Git

```bash
# 克隆仓库
git clone https://github.com/Deve-Tom/modbus-sim-cli.git
cd modbus-sim-cli

# 构建二进制文件
go build -o modbus-sim .

# 带版本信息构建
go build -ldflags "-X 'modbus-sim/cmd.Version=1.0.0' -X 'modbus-sim/cmd.Commit=$(git rev-parse --short HEAD)'" -o modbus-sim .
```

### 选项 3：通过 Go 安装

```bash
go install github.com/Deve-Tom/modbus-sim-cli@latest
```

## 🛠️ 构建

### 跨平台编译

#### 使用 Make（Linux/macOS）

```bash
# 为当前平台构建
make build

# 为所有支持的平台构建
make build-all

# 为特定平台构建
make build-windows      # Windows amd64
make build-linux        # Linux amd64
make build-linux-arm    # Linux ARM（树莓派）
make build-linux-arm64  # Linux ARM64
make build-macos        # macOS Intel
make build-macos-arm64  # macOS Apple Silicon
```

#### 使用构建脚本

```bash
# 为所有平台构建
go run scripts/build-all.go
```

构建产物存储在按平台组织的 `build/` 目录中：

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

## 💻 使用方法

### 快速启动模式

使用 `quick` 命令和合理的默认值是最快的启动方式。

#### Modbus TCP 服务器

```bash
# 使用默认值启动 TCP 服务器（端口 502，100 个寄存器，ABCD 字节序）
./modbus-sim quick

# 在端口 10502 上启动自定义 TCP 服务器，200 个寄存器
./modbus-sim quick --mode tcp --addr :10502 --registers 200

# 使用 BDAC 字节序（MidSwap）
./modbus-sim quick --mode tcp --addr :502 --byte-order BDAC --registers 50
```

#### Modbus RTU 服务器

**Linux：**
```bash
# 在 /dev/ttyUSB0 上启动 RTU 服务器，使用默认串口设置
sudo ./modbus-sim quick --mode rtu --addr /dev/ttyUSB0 --registers 50

# 自定义波特率和寄存器数量
sudo ./modbus-sim quick --mode rtu --addr /dev/ttyUSB0 --byte-order ABCD --registers 100
```

**macOS：**
```bash
# 常见的 macOS 串口名称
./modbus-sim quick --mode rtu --addr /dev/cu.usbserial-XXXX --registers 50
./modbus-sim quick --mode rtu --addr /dev/cu.usbmodem-XXXX --registers 50

# 列出可用的串口
ls /dev/cu.*
# 或
ls /dev/tty.*
```

**Windows：**
```powershell
# 使用 COM 端口号
.\modbus-sim.exe quick --mode rtu --addr COM3 --registers 50
.\modbus-sim quick --mode rtu --addr COM4 --byte-order ABCD --registers 100

# 列出可用的 COM 端口（PowerShell）
Get-CimInstance Win32_SerialPort | Select-Object DeviceID, Description

# 或使用命令提示符
mode
```

> **注意：** RTU 模式需要适当的权限来访问串口。
> - **Linux：** 使用 `sudo` 或将用户添加到 `dialout` 组
> - **macOS：** 将用户添加到 `dialout` 或 `uucp` 组
> - **Windows：** 如果遇到权限问题，请以管理员身份运行

### 配置文件模式

对于更复杂的设置，使用 YAML 配置文件：

```bash
# 使用默认配置启动
./modbus-sim run

# 指定自定义配置文件
./modbus-sim run --config configs/example.yaml
./modbus-sim run -c my-config.yaml
```

### 命令参考

| 命令 | 描述 |
|------|------|
| `quick` | 使用命令行标志快速启动 |
| `run` | 使用配置文件启动服务器 |
| `version` | 显示版本信息 |

### 命令行标志

#### 全局标志

| 标志 | 简写 | 描述 | 默认值 |
|------|------|------|--------|
| `--lang` | `-l` | 语言（en, zh） | `en` |

#### Quick 命令标志

| 标志 | 简写 | 描述 | 默认值 |
|------|------|------|--------|
| `--mode` | `-m` | 服务器模式：`tcp` 或 `rtu` | `tcp` |
| `--addr` | `-a` | 监听地址（TCP）或串口（RTU） | `:502` |
| `--byte-order` | `-b` | 字节序：`ABCD`、`DCBA`、`BADC`、`CDAB`、`BDAC` | `ABCD` |
| `--registers` | `-r` | 要初始化的保持寄存器数量 | `100` |

#### Run 命令标志

| 标志 | 简写 | 描述 | 默认值 |
|------|------|------|--------|
| `--config` | `-c` | 配置文件路径 | `configs/example.yaml` |

## ⚙️ 配置

### 配置文件结构

创建 YAML 配置文件来定义您的 Modbus 设备仿真：

```yaml
# 服务器模式："tcp" 或 "rtu"
mode: tcp

# TCP 监听地址（例如 ":502" 为标准 Modbus TCP 端口）
# 对于 RTU 模式，这指定串口（例如 "/dev/ttyUSB0"）
listen_addr: ":502"

# 多寄存器值的字节序
# 支持的值：
#   ABCD - 大端序（默认，最常见）
#   DCBA - 小端序
#   BADC - 大端序带字交换
#   CDAB - 小端序带字交换
#   BDAC - 中字交换（交换每个字内的高字节）
byte_order: ABCD

# 串口配置（RTU 模式必需，TCP 模式忽略）
serial:
  baud_rate: 9600      # 常用值：9600, 19200, 38400, 57600, 115200
  data_bits: 8         # 通常为 8
  stop_bits: 1         # 1 或 2
  parity: "none"       # "none"、"odd" 或 "even"

# 寄存器定义
# 每个条目定义一个连续的保持寄存器范围
registers:
  # 温度传感器寄存器（地址 0-9）
  - address: 0
    count: 10
    value: 0
    label: "temperature_sensor_1"

  # 压力传感器寄存器（地址 100-109）
  - address: 100
    count: 10
    value: 0
    label: "pressure_sensor_1"

  # 流量计寄存器（地址 200-204）
  - address: 200
    count: 5
    value: 1000
    label: "flow_meter_1"

  # 设备状态寄存器（地址 300-309）
  - address: 300
    count: 10
    value: 0
    label: "device_status"

# 日志配置
log_format: console   # "console" 用于人类可读，"json" 用于生产环境
log_level: info       # "debug"、"info"、"warn" 或 "error"
```

### 寄存器配置示例

#### 示例 1：具有连续寄存器的简单设备

```yaml
registers:
  - address: 0
    count: 50
    value: 0
    label: "all_registers"
```

#### 示例 2：多个传感器组

```yaml
registers:
  # 温度传感器（0-19）
  - address: 0
    count: 20
    value: 250  # 25.0°C（假设缩放因子为 10）
    label: "temp_sensors"

  # 湿度传感器（100-109）
  - address: 100
    count: 10
    value: 500  # 50.0% RH
    label: "humidity_sensors"

  # 压力传感器（200-209）
  - address: 200
    count: 10
    value: 1013  # 1013 hPa
    label: "pressure_sensors"
```

#### 示例 3：混合初始值

```yaml
registers:
  # 控制寄存器（可写）
  - address: 0
    count: 5
    value: 0
    label: "control_regs"

  # 状态寄存器（只读仿真）
  - address: 100
    count: 10
    value: 1
    label: "status_regs"

  # 设定点寄存器
  - address: 200
    count: 5
    value: 500
    label: "setpoints"
```

### 字节序详解

字节序决定多字节值（如 32 位整数或浮点数）如何存储在连续的 16 位寄存器中。

考虑一个 32 位值 `0x12345678` 存储在两个连续寄存器中：

| 字节序 | 寄存器 0 | 寄存器 1 | 描述 |
|--------|----------|----------|------|
| **ABCD** | `0x1234` | `0x5678` | 大端序（网络序，最常见） |
| **DCBA** | `0x5678` | `0x1234` | 小端序（Intel 序） |
| **BADC** | `0x3412` | `0x7856` | 大端序带字节交换 |
| **CDAB** | `0x7856` | `0x3412` | 小端序带字节交换 |
| **BDAC** | `0x1256` | `0x3478` | 中字交换（字内字节交换） |

**建议：** 除非目标设备特别要求不同的字节序，否则使用 `ABCD`。请查看设备文档以获取正确的设置。

## 🧪 测试与集成

### 使用 Modbus 客户端工具测试

#### Python (pymodbus)

```python
from pymodbus.client import ModbusTcpClient

# 连接到仿真器
client = ModbusTcpClient('localhost', port=502)

if client.connect():
    # 从地址 0 开始读取 10 个保持寄存器
    result = client.read_holding_registers(address=0, count=10, slave=1)
    if not result.isError():
        print(f'寄存器值: {result.registers}')
    
    # 写入单个寄存器
    client.write_register(address=0, value=1234, slave=1)
    
    # 写入多个寄存器
    values = [100, 200, 300, 400, 500]
    client.write_registers(address=0, values=values, slave=1)
    
    client.close()
```

安装：`pip install pymodbus`

#### Python (modbus-tk)

```python
import modbus_tk.modbus_tcp as modbus_tcp

master = modbus_tcp.TcpMaster(host='localhost', port=502)
master.set_timeout(5.0)

# 读取寄存器
result = master.execute(1, modbus_tk.defines.READ_HOLDING_REGISTERS, 0, 10)
print(f'值: {result}')

# 写入寄存器
master.execute(1, modbus_tk.defines.WRITE_SINGLE_REGISTER, 0, output_value=1234)

master.close()
```

安装：`pip install modbus-tk`

#### Node.js (modbus-serial)

```javascript
const ModbusRTU = require("modbus-serial");
const client = new ModbusRTU();

client.connectTCP("localhost", { port: 502 })
  .then(() => {
    client.setID(1);
    
    // 读取保持寄存器
    client.readHoldingRegisters(0, 10, (err, data) => {
      console.log("值:", data.data);
    });
    
    // 写入单个寄存器
    client.writeRegister(0, 1234);
  });
```

安装：`npm install modbus-serial`

#### 命令行（原始 TCP）

```bash
# 读取保持寄存器（功能码 03）
# 从地址 0 开始读取 10 个寄存器
echo -ne '\x00\x01\x00\x00\x00\x06\x01\x03\x00\x00\x00\x0a' | nc -w 2 localhost 502

# 写入单个寄存器（功能码 06）
# 向地址 0 写入值 1234
echo -ne '\x00\x01\x00\x00\x00\x06\x01\x06\x00\x00\x04\xd2' | nc -w 2 localhost 502
```

### 图形化工具

- **QModBus**：跨平台 Modbus 测试工具（[下载](https://sourceforge.net/projects/qmodbus/)）
- **Modbus Poll**：专业 Modbus 主站仿真器（Windows）
- **CAS Modbus Scanner**：免费 Modbus 扫描工具
- **Simply Modbus**：易于使用的 Modbus 客户端

### 测试 RTU 模式

对于 RTU 测试，您可以使用虚拟串口：

**Linux/macOS：**
```bash
# 安装 socat 以创建虚拟串口
sudo apt install socat        # Debian/Ubuntu/Kali
brew install socat            # macOS

# 创建一对连接的虚拟串口
socat -d -d pty,raw,echo=0 pty,raw,echo=0
# 输出：PTY is /dev/pts/5 and PTY is /dev/pts/6

# 终端 1：启动 Modbus RTU 仿真器
sudo ./modbus-sim quick --mode rtu --addr /dev/pts/5 --registers 50

# 终端 2：使用 Modbus RTU 客户端连接到 /dev/pts/6
python3 rtu_client.py  # 您的客户端代码连接到 /dev/pts/6
```

**Windows：**
```powershell
# 选项 1：使用 com0com（虚拟串口模拟器）
# 下载地址：https://sourceforge.net/projects/com0com/
# 安装并创建虚拟 COM 端口对（例如 COM10 <-> COM11）

# 终端 1：启动 Modbus RTU 仿真器
.\modbus-sim.exe quick --mode rtu --addr COM10 --registers 50

# 终端 2：使用 Modbus RTU 客户端连接到 COM11

# 选项 2：使用物理 USB 转串口适配器
# 用零调制解调器电缆连接两个 USB 转串口适配器
# 或使用单个适配器进行环回（TX 连接到 RX）
```

### 各平台串口名称

不同的操作系统使用不同的串口命名约定：

| 平台 | 设备模式 | 示例 | 描述 |
|------|----------|------|------|
| **Linux** | `/dev/ttyUSB*` | `/dev/ttyUSB0` | USB 转串口适配器（FTDI、CH340 等） |
| **Linux** | `/dev/ttyACM*` | `/dev/ttyACM0` | CDC-ACM 设备（Arduino 等） |
| **Linux** | `/dev/ttyS*` | `/dev/ttyS0` | 内置串口 |
| **macOS** | `/dev/cu.usbserial-*` | `/dev/cu.usbserial-1420` | USB 转串口适配器 |
| **macOS** | `/dev/cu.usbmodem-*` | `/dev/cu.usbmodem14201` | CDC-ACM 设备 |
| **macOS** | `/dev/cu.*` | `/dev/cu.Bluetooth-Incoming-Port` | 所有串口设备 |
| **Windows** | `COM*` | `COM3`, `COM4` | 所有串口 |

**提示：**
- **Linux：** 使用 `dmesg \| grep tty` 查找新连接的设备
- **macOS：** 使用 `ls /dev/cu.*` 列出所有可用串口
- **Windows：** 使用设备管理器或 `mode` 命令查找 COM 端口
- 在 macOS 上优先使用 `/dev/cu.*` 而不是 `/dev/tty.*` 以获得更好的兼容性

### 支持的 Modbus 功能码

| 功能码 | 名称 | 描述 | 支持 |
|--------|------|------|------|
| 01 | 读线圈 | 读取离散输出 | ✅ |
| 02 | 读离散输入 | 读取离散输入 | ✅ |
| 03 | 读保持寄存器 | 读取模拟输出 | ✅ |
| 04 | 读输入寄存器 | 读取模拟输入 | ✅ |
| 05 | 写单个线圈 | 写入离散输出 | ✅ |
| 06 | 写单个寄存器 | 写入模拟输出 | ✅ |
| 15 | 写多个线圈 | 写入多个离散输出 | ✅ |
| 16 | 写多个寄存器 | 写入多个模拟输出 | ✅ |

## 📁 项目结构

```
modbus-sim-cli/
├── main.go                     # 应用程序入口点
├── Makefile                    # 构建自动化
├── go.mod                      # Go 模块定义
├── go.sum                      # 依赖校验和
├── LICENSE                     # MIT 许可证
├── README.md                   # 本文档（英文）
├── README_zh.md                # 本文档（中文）
├── .gitignore                  # Git 忽略规则
│
├── cmd/
│   └── root.go                 # CLI 命令（cobra 框架）
│
├── configs/
│   └── example.yaml            # 示例配置文件
│
├── scripts/
│   └── build-all.go            # 跨平台构建脚本
│
├── build/                      # 编译后的二进制文件（生成）
│   ├── linux_amd64/
│   ├── linux_arm/
│   ├── linux_arm64/
│   ├── darwin_amd64/
│   ├── darwin_arm64/
│   └── windows_amd64/
│
└── internal/
    ├── i18n/                   # 国际化
    │   ├── i18n.go             # i18n 实现
    │   └── locales/
    │       ├── en.json         # 英文翻译
    │       └── zh.json         # 中文翻译
    │
    ├── config/                 # 配置管理
    │   ├── config.go           # 配置加载和验证
    │   └── config_test.go      # 配置测试
    │
    ├── byteorder/              # 字节序处理
    │   ├── byteorder.go        # 字节序实现
    │   └── byteorder_test.go   # 字节序测试
    │
    ├── register/               # 寄存器管理
    │   └── register.go         # 寄存器存储和操作
    │
    ├── simulator/              # 核心仿真引擎
    │   └── simulator.go        # 仿真器编排
    │
    └── server/                 # 网络服务器
        └── server.go           # TCP/RTU 服务器实现
```

## 🔍 故障排除

### 常见问题

#### 1. 端口 502 权限被拒绝

**问题：** 无法绑定到端口 502（特权端口）

**解决方案：**
```bash
# 选项 1：使用 sudo（Linux/macOS）
sudo ./modbus-sim quick

# 选项 2：使用非特权端口（>1024）
./modbus-sim quick --addr :10502

# 选项 3：设置 CAP_NET_BIND_SERVICE 能力（Linux）
sudo setcap 'cap_net_bind_service=+ep' ./modbus-sim

# Windows：以管理员身份运行或使用端口 >1024
.\modbus-sim.exe quick --addr :10502
```

#### 2. 串口访问被拒绝（RTU 模式）

**问题：** 无法打开串口

**Linux 解决方案：**
```bash
# 选项 1：使用 sudo
sudo ./modbus-sim quick --mode rtu --addr /dev/ttyUSB0

# 选项 2：将用户添加到 dialout 组
sudo usermod -a -G dialout $USER
# 注销并重新登录

# 选项 3：更改设备权限（临时）
sudo chmod 666 /dev/ttyUSB0
```

**macOS 解决方案：**
```bash
# 选项 1：将用户添加到 dialout 或 uucp 组
sudo dscl . append /Groups/dialout GroupMembership $(whoami)
# 或
sudo dscl . append /Groups/uucp GroupMembership $(whoami)

# 选项 2：使用 cu.* 设备而不是 tty.*
./modbus-sim quick --mode rtu --addr /dev/cu.usbserial-XXXX

# 列出可用端口
ls -l /dev/cu.*
ls -l /dev/tty.*
```

**Windows 解决方案：**
```powershell
# 选项 1：以管理员身份运行
# 右键单击 PowerShell 或命令提示符 -> “以管理员身份运行”
.\modbus-sim.exe quick --mode rtu --addr COM3

# 选项 2：检查 COM 端口是否被占用
# PowerShell
Get-CimInstance Win32_SerialPort | Select-Object DeviceID, Description

# 命令提示符
mode

# 选项 3：尝试不同的 COM 端口号
.\modbus-sim.exe quick --mode rtu --addr COM4
```

#### 3. 对 Modbus 请求无响应

**检查清单：**
- 验证服务器是否正在运行并在正确的端口上监听
- 检查防火墙规则是否允许连接
- 确保使用正确的从站/单元 ID（默认为 1）
- 验证字节序是否与客户端期望匹配
- 检查日志中的错误消息

```bash
# 检查端口是否在监听
netstat -tlnp | grep 502
ss -tlnp | grep 502

# 检查运行的进程
ps aux | grep modbus-sim

# 测试连通性
telnet localhost 502
nc -zv localhost 502
```

#### 4. 寄存器值不正确

**可能原因：**
- 错误的字节序设置
- 错误的寄存器地址（Modbus 使用基于 0 的地址）
- 数据类型不匹配（16 位 vs 32 位值）

**调试技巧：** 启用调试日志以查看原始请求和响应：
```yaml
log_level: debug
log_format: console
```

### 获取帮助

- 📖 查看本 README 以获取使用示例
- 🐛 在 [GitHub Issues](https://github.com/Deve-Tom/modbus-sim-cli/issues) 上报告问题
- 💬 在 [Discussions](https://github.com/Deve-Tom/modbus-sim-cli/discussions) 中讨论功能和提问

## 📊 使用场景

### 工业自动化测试

在开发过程中仿真 PLC、HMI、传感器和执行器：

```bash
# 仿真温度控制器
./modbus-sim quick --mode tcp --addr :502 --registers 100

# 客户端应用程序现在可以读取/写入温度设定点
```

### SCADA 系统开发

在没有物理硬件的情况下测试 SCADA 应用程序：

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

### IoT 网关测试

验证 IoT 网关的 Modbus 集成：

```bash
# 在不同端口上仿真多个现场设备
./modbus-sim quick --mode tcp --addr :5020 --registers 50 &
./modbus-sim quick --mode tcp --addr :5021 --registers 50 &
./modbus-sim quick --mode tcp --addr :5022 --registers 50 &
```

### 教育和培训

动手学习 Modbus 协议：

```bash
# 启动仿真器
./modbus-sim quick --mode tcp --addr :502

# 使用不同工具练习
# 尝试使用 Python、Node.js 或命令行读取/写入寄存器
```

## 🤝 贡献

欢迎贡献！以下是您可以帮助的方式：

1. **Fork** 仓库
2. **创建** 功能分支（`git checkout -b feature/amazing-feature`）
3. **提交** 您的更改（`git commit -m 'Add amazing feature'`）
4. **推送** 到分支（`git push origin feature/amazing-feature`）
5. **开启** Pull Request

### 开发设置

```bash
# 克隆您的 fork
git clone https://github.com/YOUR_USERNAME/modbus-sim-cli.git
cd modbus-sim-cli

# 安装依赖
go mod download

# 运行测试
go test ./... -v

# 本地构建和测试
go build -o modbus-sim .
./modbus-sim quick
```

### 代码规范

- 遵循 Go 最佳实践和习惯用法
- 为新功能编写测试
- 根据需要更新文档
- 使用有意义的提交消息

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [mbserver](https://github.com/leijux/mbserver) - Modbus 服务器库
- [cobra](https://github.com/spf13/cobra) - CLI 框架
- [zerolog](https://github.com/rs/zerolog) - 结构化日志
- [goburrow/serial](https://github.com/goburrow/serial) - 串口通信

## 📞 支持

- 📧 **问题反馈**：[GitHub Issues](https://github.com/Deve-Tom/modbus-sim-cli/issues)
- 💬 **讨论**：[GitHub Discussions](https://github.com/Deve-Tom/modbus-sim-cli/discussions)
- 📖 **文档**：本 README 和代码注释

---

<div align="center">

**用 Go ❤️ 打造**

[⭐ 给仓库加星](https://github.com/Deve-Tom/modbus-sim-cli) 如果您觉得它有用！

</div>
