# code-sandbox-mcp

## 简介

MCP 服务器用于创建安全的代码沙箱环境，在 Docker 容器中执行代码，并为 AI 应用提供代码执行能力。

 [英文 README](https://github.com/lemonlyue/code-sandbox-mcp/blob/main/README.md)

## 功能特点

- 支持多种编程语言的代码执行（Python、PHP、Golang）
- 基于 Docker 容器的隔离环境，确保代码执行安全
- 提供资源限制（CPU 超时、内存限制、磁盘限制）
- 通过 SSE（服务器发送事件）提供实时交互能力
- 简单易用的工具接口，方便集成到 AI 应用中

## 快速开始

### 前置要求
- Go 1.25.1 或更高版本
- Docker 环境

### 初始化
拉取所需的编程语言 Docker 镜像：
```bash
make init-images
```

### 构建
根据目标平台编译项目到 `bin` 目录：

- Linux (amd64 架构)：
```bash
make build-linux
```

- macOS (Apple Silicon 芯片，arm64 架构)：
```bash
make build-darwin
```

- Windows (amd64 架构)：
```bash
make build-windows
```

### 运行
启动 MCP 服务器：
```bash
./bin/code-sandbox-mcp-server
```

服务器将在 4000 端口启动，提供以下端点：
- SSE 端点: `/sse`
- 消息端点: `/message`

## 代码执行工具

服务器注册了一个名为`execute_code_in_sandbox`的工具，用于在沙箱环境中执行代码。

### 工具参数

| 参数名 | 类型 | 是否必需 | 描述 |
|-------|-------|---|-------|
| language | string | 是 |  编程语言  |
| code | string | 是 |  需要执行的代码  |
| version | string | 是 |  编程语言版本  |

### 使用示例
调用工具执行 Python 代码：
```json
{
    "tool": "execute_code_in_sandbox",
    "parameters": {
        "language": "python",
        "code": "print('Hello, World!')"
    }
}
```

## 项目结构
- `cmd/code-sandbox-mcp/main.go`: 服务器主入口
- `sandbox/`: 沙箱核心功能实现
- `sandbox/docker/`: Docker 沙箱实现
- `tempfile/`: 临时文件管理（提供临时文件写入功能，如WriteFile方法）
- `go.mod/go.sum`: Go 依赖管理
- `Makefile`: 构建脚本

### 配置管理
项目通过`sandbox/config.go`实现配置管理功能，支持：
- 加载 YAML 格式的配置文件（默认路径包括`./config.yaml`和`./config/config.yaml`）
- 监控配置文件变化并自动重载
- 配置项包括服务器信息、运行时资源限制（CPU 超时、内存、磁盘）、网络设置、语言特定配置（后缀、镜像、入口点等）

### 清理
清理编译生成的文件：
```bash
make clean
```

### 依赖项
主要依赖项包括：
- Docker SDK for Go: 用于与 Docker 交互
- trpc-mcp-go: 提供 MCP 服务器框架（版本 v0.0.7）
- viper: 配置管理
- fsnotify: 文件系统通知（用于配置文件监控）
- golang.org/x/sys: 系统相关操作支持
- golang.org/x/net: 网络相关功能支持

完整依赖列表参见`go.mod`和`go.sum`文件。