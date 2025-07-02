package app1

import (
	"context"
	"fmt"
	"log"

	"task_scheduler/internal/plugins"
)

// App1Plugin app1插件实现
type App1Plugin struct{}

// App1Task app1任务实现
type App1Task struct {
	name   string
	config map[string]interface{}
}

// NewPlugin 创建app1插件
func NewPlugin() plugins.Plugin {
	return &App1Plugin{}
}

// Name 返回插件名称
func (p *App1Plugin) Name() string {
	return "app1"
}

// CreateTask 创建任务实例
func (p *App1Plugin) CreateTask(config map[string]interface{}) (plugins.Task, error) {
	return &App1Task{
		name:   "app1",
		config: config,
	}, nil
}

// GetDefaultConfig 获取默认配置
func (p *App1Plugin) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"timeout": 30,
		"message": "Hello from App1",
	}
}

// Name 返回任务名称
func (t *App1Task) Name() string {
	return t.name
}

// Execute 执行任务
func (t *App1Task) Execute(ctx context.Context) error {
	log.Printf("开始执行 App1 任务")

	// 模拟任务执行
	// time.Sleep(2 * time.Second)

	// 获取配置参数
	message, _ := t.config["message"].(string)
	if message == "" {
		message = "Hello from App1"
	}

	log.Printf("App1 任务执行完成: %s", message)
	return nil
}

// ValidateConfig 验证配置
func (t *App1Task) ValidateConfig(config map[string]interface{}) error {
	// 检查超时配置
	if timeout, exists := config["timeout"]; exists {
		if timeoutVal, ok := timeout.(int); ok {
			if timeoutVal <= 0 {
				return fmt.Errorf("timeout 必须大于0")
			}
		} else {
			return fmt.Errorf("timeout 必须是整数类型")
		}
	}

	return nil
}
