package config

import (
	"fmt"
	"log"

	"task_scheduler/internal/plugins"

	"github.com/spf13/viper"
)

// Config 主配置结构
type Config struct {
	LogLevel   string       `mapstructure:"log_level"`
	PluginsDir string       `mapstructure:"plugins_dir"`
	Tasks      []TaskConfig `mapstructure:"tasks"`
}

// TaskConfig 任务配置结构
type TaskConfig struct {
	Name       string `mapstructure:"name"`
	ConfigFile string `mapstructure:"config_file"`
	Enabled    bool   `mapstructure:"enabled"`
}

// TaskScheduleConfig 任务调度配置
type TaskScheduleConfig struct {
	Schedule string                 `mapstructure:"schedule"`
	Params   map[string]interface{} `mapstructure:"params"`
}

// Loader 配置加载器
type Loader struct {
	configPath string
}

// NewLoader 创建配置加载器
func NewLoader(configPath string) *Loader {
	return &Loader{
		configPath: configPath,
	}
}

// LoadMainConfig 加载主配置
func (l *Loader) LoadMainConfig() (*Config, error) {
	viper.SetConfigFile(l.configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.PluginsDir == "" {
		config.PluginsDir = "./plugins"
	}

	return &config, nil
}

// LoadTaskConfig 加载任务配置
func (l *Loader) LoadTaskConfig(configFile string) (*TaskScheduleConfig, error) {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取任务配置文件失败: %w", err)
	}

	var taskConfig TaskScheduleConfig
	if err := viper.Unmarshal(&taskConfig); err != nil {
		return nil, fmt.Errorf("解析任务配置文件失败: %w", err)
	}

	return &taskConfig, nil
}

// LoadAllTasks 加载所有任务配置
func (l *Loader) LoadAllTasks(mainConfig *Config) ([]plugins.TaskInfo, error) {
	var tasks []plugins.TaskInfo

	for _, taskConfig := range mainConfig.Tasks {
		if !taskConfig.Enabled {
			log.Printf("任务已禁用: %s", taskConfig.Name)
			continue
		}

		// 加载任务调度配置
		scheduleConfig, err := l.LoadTaskConfig(taskConfig.ConfigFile)
		if err != nil {
			log.Printf("加载任务配置失败: %s, 错误: %v", taskConfig.Name, err)
			continue
		}

		// 创建任务信息
		taskInfo := plugins.TaskInfo{
			Name:     taskConfig.Name,
			Schedule: scheduleConfig.Schedule,
			Config:   scheduleConfig.Params,
			Enabled:  taskConfig.Enabled,
		}

		tasks = append(tasks, taskInfo)
		log.Printf("任务配置已加载: %s, 调度: %s", taskConfig.Name, scheduleConfig.Schedule)
	}

	return tasks, nil
}

// ValidateConfig 验证配置
func (l *Loader) ValidateConfig(config *Config) error {
	if config.LogLevel == "" {
		return fmt.Errorf("log_level 不能为空")
	}

	if config.PluginsDir == "" {
		return fmt.Errorf("plugins_dir 不能为空")
	}

	for _, task := range config.Tasks {
		if task.Name == "" {
			return fmt.Errorf("任务名称不能为空")
		}
		if task.ConfigFile == "" {
			return fmt.Errorf("任务配置文件路径不能为空: %s", task.Name)
		}
	}

	return nil
}
