package docker

import (
	"bytes"
	"context"
	"errors"
	"github.com/lemonlyue/code-sandbox-mcp/sandbox"
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

	execCommand := config.Entrypoint[2]
	tmpl := template.Must(template.New("command").Parse(execCommand))
	data := EntrypointTmpl{
		ExecFile: filePath,
		Path:     path,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return []string{}, err
	}

	config.Entrypoint[2] = buf.String()
	return config.Entrypoint, nil
}
