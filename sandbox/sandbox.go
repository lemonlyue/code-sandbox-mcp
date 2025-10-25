package sandbox

import (
	"context"
	"time"
)

// Config sandbox config
type Config struct {
	Suffix     string // file suffix
	Language   string // language
	Version    string // language version
	Image      string // container image
	WorkDir    string // work dir
	BaseImage  string // base image
	Entrypoint []string
	Timeout    time.Duration   // total timeout
	Resource   *ResourceConfig // resource config
	NetWork    *NetWorkConfig  // network config
}

type NetWorkConfig struct {
	Enabled bool
}

// ResourceConfig sandbox env resource limit
type ResourceConfig struct {
	CpuTimeout time.Duration //
	MemoryMb   int64
	DiskMb     int64
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
