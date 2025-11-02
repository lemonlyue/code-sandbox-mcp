package docker

import (
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"time"
)

type ImageTmpl struct {
	Version  string `json:"version"`
	Language string `json:"language"`
}

type EntrypointTmpl struct {
	ExecFile string `json:"exec_file"`
	Path     string `json:"path"`
}

type ConfigOption func(*container.Config)

type HostConfigOption func(*container.HostConfig)

type ResourceConfigOption func(*container.Resources)

func WithOptions[T any](cfg *T, opts ...func(*T)) {
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
}

func WithImage(image string) ConfigOption {
	return func(cfg *container.Config) {
		cfg.Image = image
	}
}

func WithCommand(cmd ...string) ConfigOption {
	return func(cfg *container.Config) {
		cfg.Cmd = cmd
	}
}

func WithBindMount(source string, target string) HostConfigOption {
	return func(cfg *container.HostConfig) {
		cfg.Mounts = append(cfg.Mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: source,
			Target: target,
		})
	}
}

func WithAutoRemove(autoRemove bool) HostConfigOption {
	return func(cfg *container.HostConfig) {
		cfg.AutoRemove = autoRemove
	}
}

func WithMemory(memoryMb int64) ResourceConfigOption {
	return func(cfg *container.Resources) {
		limitMemory := memoryMb * 1024 * 1024
		cfg.Memory = limitMemory
		cfg.MemorySwap = limitMemory
	}
}

func WithDiskMb(baseDir string, diskMb int64) HostConfigOption {
	return func(cfg *container.HostConfig) {
		cfg.Tmpfs = map[string]string{
			baseDir: fmt.Sprintf("size=%vm", diskMb),
		}
	}
}

func WithCpuTimeout(cpuLimit time.Duration) ResourceConfigOption {
	return func(cfg *container.Resources) {
		cfg.CPUPeriod = 100000
		cfg.CPUQuota = int64(float64(cfg.CPUPeriod) * 100000)
	}
}
