package docker

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

type ConfigOption func(*container.Config)

type HostConfigOption func(*container.HostConfig)

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
