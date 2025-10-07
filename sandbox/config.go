package sandbox

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"os"
)

type SandboxConfig struct {
	Server    ServerConfig              `yaml:"server" mapstructure:"server"`
	Runtimes  runtimeConfig             `yaml:"runtimes" mapstructure:"runtimes"`
	Languages map[string]languageConfig `yaml:"languages" mapstructure:"languages"` // 修改这里
}

type languageConfig struct {
	Suffix       string          `yaml:"suffix" mapstructure:"suffix"`
	DefaultImage string          `yaml:"default_image" mapstructure:"default_image"`
	BaseImage    string          `yaml:"base_image" mapstructure:"base_image"`
	Entrypoint   []string        `yaml:"entrypoint" mapstructure:"entrypoint"`
	Resources    resourcesConfig `yaml:"resources" mapstructure:"resources"`
}

type ServerConfig struct {
	Name    string `yaml:"name" mapstructure:"name"`
	Version string `yaml:"version" mapstructure:"version"`
}

type runtimeConfig struct {
	Resources     resourcesConfig `yaml:"resources" mapstructure:"resources"`
	Network       networkConfig   `yaml:"network" mapstructure:"network"`
	Engine        string          `yaml:"engine" mapstructure:"engine"`
	CleanupOnExit bool            `yaml:"cleanup_on_exit" mapstructure:"cleanup_on_exit"`
	WorkDir       string          `yaml:"work_dir" mapstructure:"work_dir"`
	Timeout       string          `yaml:"timeout" mapstructure:"timeout"`
}

type resourcesConfig struct {
	CpuTimeout string `yaml:"cpu_timeout" mapstructure:"cpu_timeout"`
	MemoryMb   int64  `yaml:"memory_mb" mapstructure:"memory_mb"`
	DiskMb     int64  `yaml:"disk_mb" mapstructure:"disk_mb"`
}

type networkConfig struct {
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
}

type ConfigManager struct {
	config *SandboxConfig
}

func NewConfigManager(configPath ...string) (*ConfigManager, error) {
	manager := &ConfigManager{}

	finalConfigPath := manager.determineConfigPath(configPath...)

	if err := manager.loadConfig(finalConfigPath); err != nil {
		return nil, err
	}

	// watch config
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		if err := manager.loadConfig(finalConfigPath); err != nil {

		}
	})

	return manager, nil
}

// determineConfigPath
func (cm *ConfigManager) determineConfigPath(configPaths ...string) string {
	for _, path := range configPaths {
		if path != "" {
			if _, err := os.Stat(path); err == nil {
				return path // 文件存在，直接返回
			}
			return path
		}
	}

	defaultPaths := []string{
		"./config.yaml",
		"./config/config.yaml",
	}

	for _, path := range defaultPaths {
		expandedPath := os.ExpandEnv(path)
		if _, err := os.Stat(expandedPath); err == nil {
			return expandedPath
		}
	}

	return "./config.yaml"
}

func (cm *ConfigManager) loadConfig(configPath string) error {
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	var config SandboxConfig
	if err := viper.Unmarshal(&config); err != nil {
		return err
	}

	cm.config = &config
	return nil
}

func (cm *ConfigManager) GetConfig() *SandboxConfig {
	return cm.config
}

func (cm *ConfigManager) GetRuntimeEngine() string {
	return cm.config.Runtimes.Engine
}

func (cm *ConfigManager) GetLanguageConfig(language string) languageConfig {
	return cm.config.Languages[language]
}
