# code-sandbox-mcp

## Introduction

The MCP server is used to create a secure code sandbox environment, execute code in Docker containers, and provide code execution capabilities for AI applications.

For Chinese version, see [中文 README](https://github.com/lemonlyue/code-sandbox-mcp/blob/main/README-zh.md).

## Features

- Supports code execution in multiple programming languages (Python, PHP, Golang)
- Docker container-based isolated environment to ensure secure code execution
- Provides resource limitations (CPU timeout, memory limit, disk limit)
- Provides real-time interaction capabilities through SSE (Server-Sent Events)
- Easy-to-use tool interface for easy integration into AI applications

## Quick Start

### Prerequisites

- Go 1.25.1 or higher
- Docker environment

### Initialization
Pull the required programming language Docker images:

```bash
make init-images
```

### Build

Compile the project to the bin directory based on the target platform:

- Linux (amd64 architecture):
```bash
make build-linux
```

- macOS (Apple Silicon chips, arm64 architecture):

```bash
make build-darwin
```

- Windows (amd64 architecture):

```bash
make build-windows
```

### Run
Start the MCP server：
```bash
./bin/code-sandbox-mcp-server
```

The server will start on port 4000, providing the following endpoints:
- SSE endpoint: `/sse`
- Message endpoint: `/message`

## Code Execution Tool

The server registers a tool named `execute_code_in_sandbox` for executing code in a sandbox environment.

### Tool Parameters

| Parameter | Type | Required | Description |
|-------|-------|----------|-------|
| language | string | Yes      |  Programming language  |
| code | string | Yes      |  The code to be executed|
| version | string | No       |  Programming language version|

### Usage Example
Call the tool to execute Python code:
```json
{
    "tool": "execute_code_in_sandbox",
    "parameters": {
        "language": "python",
        "code": "print('Hello, World!')"
    }
}
```

## Project Structure

- `cmd/code-sandbox-mcp/main.go`: Server main entry point
- `sandbox/`: Sandbox core functionality implementation
- `sandbox/docker/`: Docker sandbox implementation
- `tempfile/`: Temporary file management (provides temporary file writing functionality, such as `WriteFile` method)
- `go.mod/go.sum`: Go dependency management
- `Makefile`: Build scripts

### Configuration Management

The project implements configuration management through `sandbox/config.go`, supporting:
- Loading YAML format configuration files (default paths include `./config.yaml` and `./config/config.yaml`)
- Monitoring configuration file changes and automatic reloading
- Configuration items include server information, runtime resource limits (CPU timeout, memory, disk), network settings, language-specific configurations (suffix, image, entrypoint, etc.)


### Cleanup
Clean up compiled files：
```bash
make clean
```

### Dependencies
Main dependencies include:
- Docker SDK for Go: For interacting with Docker
- trpc-mcp-go: Provides MCP server framework (version v0.0.7)
- viper: Configuration management
- fsnotify: File system notifications (for configuration file monitoring)
- golang.org/x/sys: System-related operation support
- golang.org/x/net: Network-related function support

See `go.mod` and `go.sum` files for the complete list of dependencies.
