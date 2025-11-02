package docker

import (
	"bytes"
	"context"
	"errors"
	"github.com/lemonlyue/code-sandbox-mcp/sandbox"
	"strings"
	"text/template"
)

// getRuntimeImage Return the Docker image based on language and version.
func getRuntimeImage(ctx context.Context, config *sandbox.Config) (string, error) {
	tmpl := template.Must(template.New("docker").Parse(config.Image))

	data := ImageTmpl{
		Version:  config.Version,
		Language: config.Language,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// buildExecutionCommand build execution command
func buildExecutionCommand(ctx context.Context, config *sandbox.Config, path string, filePath string) ([]string, error) {
	if len(config.Entrypoint) != 3 {
		return []string{}, errors.New("failed to build execution command")
	}

	entrypoint := make([]string, len(config.Entrypoint))
	copy(entrypoint, config.Entrypoint)

	execCommand := entrypoint[2]
	tmpl := template.Must(template.New("command").Parse(execCommand))
	data := EntrypointTmpl{
		ExecFile: filePath,
		Path:     path,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return []string{}, err
	}

	entrypoint[2] = buf.String()
	return entrypoint, nil
}

// isImageNotFoundError
func isImageNotFoundError(ctx context.Context, err error) bool {
	return strings.Contains(err.Error(), "No such image")
}
