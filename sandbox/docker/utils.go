package docker

// getRuntimeImage Return the Docker image based on language and version.
func getRuntimeImage(language string, version string) (string, error) {
	return "golang:1.25.1-alpine", nil
}

func buildExecutionCommand(language string, filePath string) ([]string, error) {
	return []string{
		"go", "run", filePath,
	}, nil
}
