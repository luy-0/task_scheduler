package pushAPI

import (
	"fmt"
	"log"
	"time"
)

// PushAPIImpl PushAPI接口实现
type PushAPIImpl struct {
	controller *PushController
}

// NewPushAPI 创建PushAPI实例
func NewPushAPI() PushAPI {
	return &PushAPIImpl{}
}

// Initialize 初始化（选择内置推送方式）
func (api *PushAPIImpl) Initialize(cfg Config, method PushMethod) error {
	controller := NewPushController(cfg)

	if err := controller.Initialize(cfg, method); err != nil {
		return fmt.Errorf("初始化推送控制器失败: %w", err)
	}

	api.controller = controller
	log.Printf("推送API初始化成功，使用推送方式: %s", method.String())
	return nil
}

// InitializeWithPusher 高级初始化（自定义推送器）
func (api *PushAPIImpl) InitializeWithPusher(cfg Config, pusher Pusher) error {
	controller := NewPushController(cfg)

	if err := controller.InitializeWithPusher(cfg, pusher); err != nil {
		return fmt.Errorf("初始化推送控制器失败: %w", err)
	}

	api.controller = controller
	log.Printf("推送API初始化成功，使用自定义推送器: %s", pusher.Name())
	return nil
}

// PushNow 立即推送
func (api *PushAPIImpl) PushNow(message Message, options PushOptions) error {
	if api.controller == nil {
		return fmt.Errorf("推送API未初始化")
	}

	// 设置消息创建时间
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	return api.controller.PushNow(message, options)
}

// Enqueue 入队消息
func (api *PushAPIImpl) Enqueue(message Message, options PushOptions) error {
	if api.controller == nil {
		return fmt.Errorf("推送API未初始化")
	}

	// 设置消息创建时间
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	return api.controller.Enqueue(message, options)
}

// FlushQueue 刷新队列
func (api *PushAPIImpl) FlushQueue() error {
	if api.controller == nil {
		return fmt.Errorf("推送API未初始化")
	}

	return api.controller.FlushQueue()
}

// Stop 停止推送API
func (api *PushAPIImpl) Stop() {
	if api.controller != nil {
		api.controller.Stop()
	}
}

// GetQueueSize 获取队列大小
func (api *PushAPIImpl) GetQueueSize() int {
	if api.controller == nil {
		return 0
	}
	return api.controller.GetQueueSize()
}

// GetRegisteredPushers 获取已注册的推送器列表
func (api *PushAPIImpl) GetRegisteredPushers() []string {
	if api.controller == nil {
		return []string{}
	}
	return api.controller.GetRegisteredPushers()
}
