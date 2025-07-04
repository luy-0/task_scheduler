package pushAPI

import (
	"fmt"
	"log"
	"task_scheduler/pkg/pushAPI/base"
	"task_scheduler/pkg/pushAPI/core"
	"task_scheduler/pkg/pushAPI/push_method"
	"time"
)

// PushAPIImpl PushAPI接口实现
type PushAPIImpl struct {
	controller *core.PushController
}

// NewPushAPI 创建PushAPI实例
func NewPushAPI() PushAPI {
	return &PushAPIImpl{}
}

// Initialize 初始化（选择内置推送方式）
func (api *PushAPIImpl) Initialize(cfg Config, method PushMethod) error {
	// 转换配置
	coreConfig := base.PushConfig{
		QueueSize:     cfg.QueueSize,
		FlushInterval: cfg.FlushInterval,
		WorkingDir:    cfg.WorkingDir,
		HistoryDir:    cfg.HistoryDir,
	}
	coreMethod := method.ToCore()

	controller := core.NewPushController(coreConfig)

	if err := controller.Initialize(coreConfig, coreMethod); err != nil {
		return fmt.Errorf("初始化推送控制器失败: %w", err)
	}

	api.controller = controller
	log.Printf("推送API初始化成功，使用推送方式: %s", method.String())
	return nil
}

// InitializeWithPusher 高级初始化（自定义推送器）
func (api *PushAPIImpl) InitializeWithPusher(cfg Config, pusher Pusher) error {
	// 转换配置
	coreConfig := base.PushConfig{
		QueueSize:     cfg.QueueSize,
		FlushInterval: cfg.FlushInterval,
		WorkingDir:    cfg.WorkingDir,
		HistoryDir:    cfg.HistoryDir,
	}

	controller := core.NewPushController(coreConfig)

	// 转换推送器
	corePusher := &corePusherAdapter{pusher: pusher}
	if err := controller.InitializeWithPusher(coreConfig, corePusher); err != nil {
		return fmt.Errorf("初始化推送控制器失败: %w", err)
	}

	api.controller = controller
	log.Printf("推送API初始化成功，使用自定义推送器: %s", pusher.GetName())
	return nil
}

// PushNow 立即推送
func (api *PushAPIImpl) PushNow(message Message, options PushOptions) error {
	if api.controller == nil {
		return fmt.Errorf("推送API未初始化")
	}

	// 如果消息ID为空，自动生成
	if message.ID == "" {
		coreMsg := base.NewMessage(message.AppID, message.Title, message.Content, message.Level.ToCore())
		message.ID = coreMsg.ID
	}

	// 设置消息创建时间
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	// 设置发送状态为等待发送
	message.SetSendStatus(StatusPending)

	// 转换消息和选项
	coreMessage := message.ToCore()
	coreOptions := options.ToCore()

	// 推送消息
	if err := api.controller.PushNow(coreMessage, coreOptions); err != nil {
		// 推送失败，设置状态为失败
		message.SetSendStatus(StatusFailed)
		return err
	}

	// 推送成功，设置发送时间和状态
	message.SetSentAt(time.Now())
	message.SetSendStatus(StatusSuccess)

	return nil
}

// Enqueue 入队消息
func (api *PushAPIImpl) Enqueue(message Message, options PushOptions) error {
	if api.controller == nil {
		return fmt.Errorf("推送API未初始化")
	}

	// 如果消息ID为空，自动生成
	if message.ID == "" {
		coreMsg := base.NewMessage(message.AppID, message.Title, message.Content, message.Level.ToCore())
		message.ID = coreMsg.ID
	}

	// 设置消息创建时间
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	// 设置发送状态为等待发送
	message.SetSendStatus(StatusPending)

	// 转换消息和选项
	coreMessage := message.ToCore()
	coreOptions := options.ToCore()

	return api.controller.Enqueue(coreMessage, coreOptions)
}

// FlushQueue 刷新队列
func (api *PushAPIImpl) FlushQueue() error {
	if api.controller == nil {
		return fmt.Errorf("推送API未初始化")
	}

	return api.controller.FlushQueue()
}

// PushAt 定时推送
func (api *PushAPIImpl) PushAt(message Message, options PushOptions, scheduledAt time.Time) error {
	if api.controller == nil {
		return fmt.Errorf("推送API未初始化")
	}

	// 如果消息ID为空，自动生成
	if message.ID == "" {
		coreMsg := base.NewMessage(message.AppID, message.Title, message.Content, message.Level.ToCore())
		message.ID = coreMsg.ID
	}

	// 设置消息创建时间
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	// 设置发送状态为等待发送
	message.SetSendStatus(StatusPending)

	// 转换消息和选项
	coreMessage := message.ToCore()
	coreOptions := options.ToCore()

	// 安排定时推送
	return api.controller.PushAt(coreMessage, coreOptions, scheduledAt)
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
	return 0 // 现在使用文件存储，队列大小为0
}

// GetRegisteredPushers 获取已注册的推送器列表
func (api *PushAPIImpl) GetRegisteredPushers() []string {
	if api.controller == nil {
		return []string{}
	}
	return api.controller.GetRegisteredPushers()
}

// corePusherAdapter 适配器，将外部推送器转换为内部推送器
type corePusherAdapter struct {
	pusher push_method.IPusher
}

func (cpa *corePusherAdapter) GetName() string {
	return cpa.pusher.GetName()
}

func (cpa *corePusherAdapter) Push(msg base.Message) error {
	return cpa.pusher.Push(msg)
}

func (cpa *corePusherAdapter) Validate(options base.PushOptions) error {
	return cpa.pusher.Validate(options)
}

func (cpa *corePusherAdapter) HealthCheck() bool {
	return cpa.pusher.HealthCheck()
}
