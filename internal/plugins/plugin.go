package plugins

import (
	"context"
	"time"
)

// Task 定义任务接口
type Task interface {
	// Name 返回任务名称
	Name() string

	// Execute 执行任务
	Execute(ctx context.Context) error

	// ValidateConfig 验证配置
	ValidateConfig(config map[string]interface{}) error
}

// Plugin 定义插件接口
type Plugin interface {
	// Name 返回插件名称
	Name() string

	// CreateTask 创建任务实例
	CreateTask(config map[string]interface{}) (Task, error)

	// GetDefaultConfig 获取默认配置
	GetDefaultConfig() map[string]interface{}
}

// TaskInfo 任务信息
type TaskInfo struct {
	Name     string                 `json:"name"`
	Schedule string                 `json:"schedule"`
	Config   map[string]interface{} `json:"config"`
	Enabled  bool                   `json:"enabled"`
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskName  string        `json:"task_name"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
}
