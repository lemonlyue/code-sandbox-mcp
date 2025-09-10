package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
	"github.com/lemonlyue/code-sandbox-mcp/sandbox"
	"github.com/lemonlyue/code-sandbox-mcp/tempfile"
	"io"
	"strings"
	"sync"
	"time"
)

// DockerSandbox It is the Docker implementation of the Sandbox interface.
type DockerSandbox struct {
	client      *client.Client
	config      *DockerSandboxConfig
	containerID string
	mu          sync.Mutex
	cleaned     bool
}

//var retryPolicy config.RetryConfig
//
//func SetRetryPolicy(policy config.RetryConfig) {
//	retryPolicy = policy
//}

// NewDockerSandbox
// receive the common SandboxConfig and convert it to a Docker-specific configuration
func NewDockerSandbox(config *sandbox.Config) (sandbox.Sandbox, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker Client: %w", err)
	}

	var opts []Option
	// convert it to a Docker-specific configuration

	//
	if config.Language != "" {
		// If the version is empty, the default is the latest version
		if config.Version == "" {
			config.Version = "latest"
		}
		runtimeImage, err := getRuntimeImage(config.Language, config.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to get image: %w", err)
		}
		opts = append(opts, WithImage(runtimeImage))
		opts = append(opts, WithLanguage(config.Language))
		opts = append(opts, WithVersion(config.Version))
	}

	if config.Resource != nil {

	}

	// Construct docker
	return NewSandbox(cli, opts...)
}

// NewSandbox construct docker
func NewSandbox(cli *client.Client, opts ...Option) (sandbox.Sandbox, error) {
	cfg := &DockerSandboxConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return &DockerSandbox{
		client: cli,
		config: cfg,
	}, nil
}

// Execute execute code
func (ds *DockerSandbox) Execute(ctx context.Context, code string) (*sandbox.ExecutionResult, error) {
	start := time.Now()

	// Pull image.
	pullResp, err := ds.client.ImagePull(ctx, ds.config.Image, image.PullOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}
	defer func(pullResp io.ReadCloser) {
		err := pullResp.Close()
		if err != nil {
			// todo
		}
	}(pullResp)

	//// 将拉取日志输出到标准输出，确认拉取过程完全完成
	//io.Copy(os.Stdout, pullResp) // 加入这行以读取拉取响应流

	fileManager, err := tempfile.NewTempFileManager("/var/tmp/")
	if err != nil {
		return nil, fmt.Errorf("failed to create file manager: %w", err)
	}
	hostPath, err := fileManager.WriteFile("main.go", []byte(code), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	defer func(fileManager *tempfile.TempFileManager) {
		err := fileManager.Cleanup()
		if err != nil {
			// todo
		}
	}(fileManager)

	id := uuid.New()
	containerName := fmt.Sprintf("mcp_%s_%s_%s", ds.config.Language, ds.config.Version, id.String())
	resp, err := ds.client.ContainerCreate(ctx, &container.Config{
		Image: ds.config.Image,
		Cmd: []string{
			"tail", "-f", "/dev/null", // Keep the container running without exiting
		},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: hostPath,
				Target: hostPath,
			},
		},
		AutoRemove: false, // Manual control in Cleanup
	}, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	ds.containerID = resp.ID

	// Start the container
	err = ds.client.ContainerStart(ctx, ds.containerID, container.StartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Dynamically construct the commands to be executed within the container based on the language.
	execCmd, err := buildExecutionCommand(ds.config.Language, hostPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution command: %w", err)
	}

	// Execute commands within the already running container.
	execResp, err := ds.client.ContainerExecCreate(ctx, ds.containerID, container.ExecOptions{
		Cmd:          execCmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to exec: %w", err)
	}

	// Attach to the exec instance to obtain the output stream.
	attachResp, err := ds.client.ContainerExecAttach(ctx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to exec code: %w", err)
	}
	defer attachResp.Close()

	var stdoutBuf, stderrBuf strings.Builder
	_, err = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, attachResp.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stdout: %w", err)
	}

	// Check the exit status of the exec execution.
	inspectResp, err := ds.client.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exec inspect: %w", err)
	}
	exitCode := inspectResp.ExitCode
	duration := time.Since(start)

	return &sandbox.ExecutionResult{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}

// Cleanup clean container resources
func (ds *DockerSandbox) Cleanup(ctx context.Context) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// idempotence check
	if ds.cleaned || ds.containerID == "" {
		return nil
	}

	var err error
	// retry
	for i := 0; i < ds.config.RetryPolicy.MaxRetries; i++ {
		// remove container
		err = ds.client.ContainerRemove(ctx, ds.containerID, container.RemoveOptions{
			Force:         true, // force remove
			RemoveVolumes: true,
		})
		if err == nil {
			ds.containerID = "" // reset container id
			ds.cleaned = true
			return nil
		}

		if i < ds.config.RetryPolicy.MaxRetries-1 {
			select {
			case <-time.After(ds.config.RetryPolicy.RetryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("failed to remove container %s: %w", ds.containerID, err)
}
