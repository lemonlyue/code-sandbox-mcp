package sandbox

import (
	"context"
	"time"
)

// Config sandbox config
type Config struct {
	WorkDir  string
	Language string          // language
	Version  string          // language version
	Timeout  time.Duration   // total timeout
	Resource *ResourceConfig // resource config
	NetWork  *NetWorkConfig  // network config
	Engine   string          // sandbox engine types (such as "docker", "gvisor")
}

type NetWorkConfig struct {
	Enabled bool
}

// ResourceConfig sandbox env resource limit
type ResourceConfig struct {
	MaxCPUTime  time.Duration //
	MaxMemoryMB int64
	MaxDiskMB   int64
}

// ExecutionResult execution result
type ExecutionResult struct {
	Stdout   string        // standard output
	Stderr   string        // standard error
	ExitCode int           // exit code
	Duration time.Duration // duration
}

// Sandbox abstract interface
type Sandbox interface {
	// Execute Sandbox execution method
	//
	// ctx: context
	// code: execute code
	//
	// return:
	// *ExecutionResult: execution result
	// error:
	Execute(ctx context.Context, code string) (*ExecutionResult, error)

	// Cleanup clean and release all resources occupied by the sandbox(such as containers, networks, and temporary files.)
	Cleanup(ctx context.Context) error
}
