package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
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
	config      *sandbox.Config
	containerID string
	mu          sync.Mutex
	cleaned     bool
}

// NewDockerSandbox
// receive the common SandboxConfig and convert it to a Docker-specific configuration
func NewDockerSandbox(ctx context.Context, config *sandbox.Config) (sandbox.Sandbox, error) {
	sandbox.InternalLogger.Ctx(ctx).Infof("Creating Docker client")
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer func(cli *client.Client) {
		err := cli.Close()
		if err != nil {
			sandbox.InternalLogger.Errorf("failed to close docker: %s", err.Error())
		}
	}(cli)
	sandbox.InternalLogger.Ctx(ctx).Infof("Docker client created successfully")

	if config.Language != "" && config.Version == "" {
		config.Version = config.BaseImage
	}
	// get runtime image
	config.Image, err = getRuntimeImage(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	// Construct docker
	return &DockerSandbox{
		client: cli,
		config: config,
	}, nil
}

// Execute execute code
func (ds *DockerSandbox) Execute(ctx context.Context, code string) (*sandbox.ExecutionResult, error) {
	defer func() {
		if err := recover(); err != nil {
			sandbox.InternalLogger.Errorf("failed to execute: %v", err)
		}
	}()
	start := time.Now()

	// Pull image.
	err := ds.ensureImage(ctx)
	if err != nil {
		return nil, err
	}

	fileManager, err := tempfile.NewTempFileManager("/var/tmp/")
	if err != nil {
		return nil, fmt.Errorf("failed to create file manager: %w", err)
	}
	fileName := "main." + ds.config.Suffix
	hostPath, err := fileManager.WriteFile(fileName, []byte(code), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	sandbox.InternalLogger.Infof("Write temp file successfully")
	defer func(fileManager *tempfile.TempFileManager) {
		err := fileManager.Cleanup()
		if err != nil {
			sandbox.InternalLogger.Errorf("failed to clean up: %v", err)
		}
	}(fileManager)

	id := uuid.New()
	containerName := fmt.Sprintf("mcp_%s_%s_%s", ds.config.Language, ds.config.Version, id.String())

	containerCfg := &container.Config{}
	hostCfg := &container.HostConfig{}
	resourcesCfg := &container.Resources{}

	// container config
	WithOptions(
		containerCfg,
		WithImage(ds.config.Image),
		WithCommand([]string{
			"tail", "-f", "/dev/null",
		}...),
	)

	// host config
	WithOptions(
		hostCfg,
		WithAutoRemove(false),
		WithBindMount(hostPath, hostPath),
		WithDiskMb(fileManager.GetDir(), ds.config.Resource.DiskMb),
	)

	// resource config
	WithOptions(
		resourcesCfg,
		WithMemory(ds.config.Resource.MemoryMb),
		WithCpuTimeout(ds.config.Resource.CpuTimeout),
	)
	sandbox.InternalLogger.Infof("the container configuration was successfully")

	resp, err := ds.client.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	sandbox.InternalLogger.Infof("Create container successfully")
	ds.containerID = resp.ID

	// Start the container
	err = ds.client.ContainerStart(ctx, ds.containerID, container.StartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	sandbox.InternalLogger.Infof("Build execution command successfully")
	// Dynamically construct the commands to be executed within the container based on the language.
	execCmd, err := buildExecutionCommand(ctx, ds.config, fileManager.GetDir(), hostPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution command: %w", err)
	}

	cmdCtx, cmdCancel := context.WithTimeout(ctx, ds.config.Resource.CpuTimeout)
	defer cmdCancel()

	// Execute commands within the already running container.
	execResp, err := ds.client.ContainerExecCreate(cmdCtx, ds.containerID, container.ExecOptions{
		Cmd:          execCmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to exec: %w", err)
	}

	// Attach to the exec instance to obtain the output stream.
	attachResp, err := ds.client.ContainerExecAttach(cmdCtx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to exec code: %w", err)
	}
	defer attachResp.Close()

	var stdoutBuf, stderrBuf strings.Builder
	_, err = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, attachResp.Reader)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return &sandbox.ExecutionResult{
				Stdout:   "",
				Stderr:   "command execution timeout",
				ExitCode: 124,
				Duration: ds.config.Resource.CpuTimeout,
			}, nil
		}
		return nil, fmt.Errorf("failed to get container stdout: %w", err)
	}

	// Check the exit status of the exec execution.
	inspectResp, err := ds.client.ContainerExecInspect(cmdCtx, execResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exec inspect: %w", err)
	}
	exitCode := inspectResp.ExitCode
	duration := time.Since(start)

	defer func() {
		err := ds.Cleanup(ctx)
		if err != nil {
			sandbox.InternalLogger.Errorf("failed to clean up: %s", err.Error())
		}
	}()

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

	cleanupOperation := func(ctx context.Context) error {
		return ds.client.ContainerRemove(ctx, ds.containerID, container.RemoveOptions{
			Force:         true, // force remove
			RemoveVolumes: true,
		})
	}

	retryDecorator := sandbox.WithRetry(3, 1)
	retryableCleanup := retryDecorator(cleanupOperation)
	err := retryableCleanup(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove container %s: %w", ds.containerID, err)
	}

	ds.containerID = ""
	ds.cleaned = true
	return nil
}

// ensureImage ensure image
func (ds *DockerSandbox) ensureImage(ctx context.Context) error {
	var buf bytes.Buffer
	_, err := ds.client.ImageInspect(ctx, ds.config.Image, client.ImageInspectWithRawResponse(&buf))
	if err == nil {
		sandbox.InternalLogger.Ctx(ctx).Infof("Image %s already exists, skip pulling", ds.config.Image)
		return nil
	}

	// Is the mirror image not found
	if !isImageNotFoundError(ctx, err) {
		sandbox.InternalLogger.Ctx(ctx).Errorf("failed to inspect image: %s", err.Error())
		return fmt.Errorf("failed to inspect image: %w", err)
	}

	// image pull
	pullResp, err := ds.client.ImagePull(ctx, ds.config.Image, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("faild to pull image: %w", err)
	}

	defer func(pullResp io.ReadCloser) {
		err := pullResp.Close()
		if err != nil {
			sandbox.InternalLogger.Errorf("failed to close pull image: %s", err.Error())
		}
	}(pullResp)

	_, err = io.Copy(io.Discard, pullResp)
	if err != nil {
		return fmt.Errorf("failed to read pull response: %w", err)
	}

	return nil
}
