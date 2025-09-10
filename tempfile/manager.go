package tempfile

import (
	"io/fs"
	"os"
	"path/filepath"
)

// TempFileManager 管理临时文件和目录的生命周期
type TempFileManager struct {
	baseDir string // 临时目录的路径
}

// NewTempFileManager 创建一个新的临时文件管理器。
// baseDir 是临时目录的根路径，如果为空则使用系统默认临时目录
func NewTempFileManager(baseDir string) (*TempFileManager, error) {
	if baseDir == "" {
		baseDir = os.TempDir()
	} else {
		//
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return nil, err
		}
	}
	// 创建一个唯一的子目录，避免冲突
	dir, err := os.MkdirTemp(baseDir, "mcp-sandbox-*")
	if err != nil {
		return nil, err
	}
	return &TempFileManager{baseDir: dir}, nil
}

// WriteFile 将内容写入指定文件名的临时文件
func (m *TempFileManager) WriteFile(filename string, content []byte, perm fs.FileMode) (string, error) {
	filePath := filepath.Join(m.baseDir, filename)
	err := os.WriteFile(filePath, content, perm)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

// GetDir 返回临时目录的路径
func (m *TempFileManager) GetDir() string {
	return m.baseDir
}

// Cleanup 清理整个临时目录
func (m *TempFileManager) Cleanup() error {
	return os.RemoveAll(m.baseDir)
}
