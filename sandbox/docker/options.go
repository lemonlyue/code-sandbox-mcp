package docker

import "time"

type DockerSandboxConfig struct {
	Image       string
	MemoryMB    int64
	CPUShares   int64
	Timeout     time.Duration
	Language    string
	Version     string
	RetryPolicy RetryPolicy
}

type RetryPolicy struct {
	MaxRetries int
	RetryDelay time.Duration
}

type Option func(config *DockerSandboxConfig)

// WithImage set the container image
func WithImage(image string) Option {
	return func(config *DockerSandboxConfig) {
		config.Image = image
	}
}

// WithLanguage set the container language
func WithLanguage(language string) Option {
	return func(config *DockerSandboxConfig) {
		config.Language = language
	}
}

// WithVersion set the container language version
func WithVersion(version string) Option {
	return func(config *DockerSandboxConfig) {
		config.Version = version
	}
}

// WithMemoryMB set the container memory(MB) limit
func WithMemoryMB(memory int64) Option {
	return func(config *DockerSandboxConfig) {
		config.MemoryMB = memory
	}
}

// WithCPUShares set the container cpu shares
func WithCPUShares(shares int64) Option {
	return func(config *DockerSandboxConfig) {
		config.CPUShares = shares
	}
}

// WithTimeout set the container timeout
func WithTimeout(timeout time.Duration) Option {
	return func(config *DockerSandboxConfig) {
		config.Timeout = timeout
	}
}
