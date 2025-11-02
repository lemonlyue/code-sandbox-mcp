package docker

import (
	"context"
	"github.com/lemonlyue/code-sandbox-mcp/sandbox"
)

// NewDockerSandboxCreator Return a function that can create a new DockerSandbox instance.
// This function itself does not create an instance but returns a function that creates an instance.
func NewDockerSandboxCreator() func(ctx context.Context, config *sandbox.Config) (sandbox.Sandbox, error) {
	return func(ctx context.Context, config *sandbox.Config) (sandbox.Sandbox, error) {
		ds, err := NewDockerSandbox(ctx, config)
		if err != nil {
			return nil, err
		}
		// Make sure that the returned DockerSandbox implements the pkg/sandbox.Sandbox interface
		return ds, nil
	}
}
