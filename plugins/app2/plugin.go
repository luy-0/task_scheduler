package app2

import (
	"context"
	"fmt"
	"log"
	"time"

	"task_scheduler/internal/plugins"
)

// App2Plugin app2插件实现
type App2Plugin struct{}

// App2Task app2任务实现
type App2Task struct {
	name   string
	config map[string]interface{}
}

// NewPlugin 创建app2插件
func NewPlugin() plugins.Plugin {
	return &App2Plugin{}
}

// Name 返回插件名称
func (p *App2Plugin) Name() string {
	return "app2"
}

// CreateTask 创建任务实例
func (p *App2Plugin) CreateTask(config map[string]interface{}) (plugins.Task, error) {
	return &App2Task{
		name:   "app2",
		config: config,
	}, nil
}

// GetDefaultConfig 获取默认配置
func (p *App2Plugin) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"retry_count": 3,
		"data_path":   "/tmp/data",
	}
}

// Name 返回任务名称
func (t *App2Task) Name() string {
	return t.name
}

// Execute 执行任务
func (t *App2Task) Execute(ctx context.Context) error {
	log.Printf("开始执行 App2 任务")

	// 模拟任务执行
	time.Sleep(1 * time.Second)

	// 获取配置参数
	dataPath, _ := t.config["data_path"].(string)
	if dataPath == "" {
		dataPath = "/tmp/data"
	}

	retryCount, _ := t.config["retry_count"].(int)
	if retryCount == 0 {
		retryCount = 3
	}

	log.Printf("App2 任务执行完成: 数据路径=%s, 重试次数=%d", dataPath, retryCount)
	return nil
}

// ValidateConfig 验证配置
func (t *App2Task) ValidateConfig(config map[string]interface{}) error {
	// 检查重试次数配置
	if retryCount, exists := config["retry_count"]; exists {
		if retryVal, ok := retryCount.(int); ok {
			if retryVal < 0 || retryVal > 10 {
				return fmt.Errorf("retry_count 必须在0-10之间")
			}
		} else {
			return fmt.Errorf("retry_count 必须是整数类型")
		}
	}

	// 检查数据路径配置
	if dataPath, exists := config["data_path"]; exists {
		if path, ok := dataPath.(string); ok {
			if path == "" {
				return fmt.Errorf("data_path 不能为空")
			}
		} else {
			return fmt.Errorf("data_path 必须是字符串类型")
		}
	}

	return nil
}
