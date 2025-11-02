package tempfile

import (
	"io/fs"
	"os"
	"path/filepath"
)

// TempFileManager Manage the lifecycle of temporary files and directories.
type TempFileManager struct {
	baseDir string // temporary directories
}

// NewTempFileManager Create a temporary file manager.
// baseDir It is root path of the temporary. If it is empty, the system's default temporary directory will be used
func NewTempFileManager(baseDir string) (*TempFileManager, error) {
	if baseDir == "" {
		baseDir = os.TempDir()
	} else {
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return nil, err
		}
	}
	dir, err := os.MkdirTemp(baseDir, "mcp-sandbox-*")
	if err != nil {
		return nil, err
	}
	return &TempFileManager{baseDir: dir}, nil
}

// WriteFile Write the content to a temporary file with the specified file name.
func (m *TempFileManager) WriteFile(filename string, content []byte, perm fs.FileMode) (string, error) {
	filePath := filepath.Join(m.baseDir, filename)
	err := os.WriteFile(filePath, content, perm)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

// GetDir Return temporary directories.
func (m *TempFileManager) GetDir() string {
	return m.baseDir
}

// Cleanup Clean up the temporary directories.
func (m *TempFileManager) Cleanup() error {
	return os.RemoveAll(m.baseDir)
}
