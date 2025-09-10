package docker

import (
	"context"
	"github.com/lemonlyue/code-sandbox-mcp/sandbox"
)

// NewDockerSandboxCreator 返回一个函数，该函数能创建一个新的 DockerSandbox 实例。
// 这个函数本身并不创建实例，而是返回一个创建实例的函数。
func NewDockerSandboxCreator() func(ctx context.Context, config *sandbox.Config) (sandbox.Sandbox, error) {
	return func(ctx context.Context, config *sandbox.Config) (sandbox.Sandbox, error) {
		// 这里是具体的创建逻辑
		ds, err := NewDockerSandbox(config)
		if err != nil {
			return nil, err
		}
		// 确保返回的 DockerSandbox 实现了 pkg/sandbox.Sandbox 接口
		return ds, nil
	}
}
