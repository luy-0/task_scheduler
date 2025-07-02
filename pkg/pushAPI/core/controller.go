package core

import (
	"fmt"
	"log"
	"sync"
	"task_scheduler/pkg/pushAPI/base"
	"task_scheduler/pkg/pushAPI/push_method"
)

// PushController 核心推送控制器
type PushController struct {
	currentPusher push_method.IPusher // 当前激活的推送器
	delayHandler  DelayHandler        // 延迟处理模块
	// pusherRouter  PusherRouter   // 推送策略路由
	pushRegistry PusherRegistry // 推送器注册表
	config       base.PushConfig
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex
}

// NewPushController 创建推送控制器
func NewPushController(cfg base.PushConfig) *PushController {
	registry := NewPusherRegistry()
	// router := NewSimplePusherRouter(registry)

	return &PushController{
		// pusherRouter: router,
		pushRegistry: registry,
		config:       cfg,
		stopChan:     make(chan struct{}),
	}
}

// Initialize 初始化（选择内置推送方式）
func (pc *PushController) Initialize(cfg base.PushConfig, method base.PushMethod) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// 根据推送方式创建内置推送器
	var pusher push_method.IPusher
	switch method {
	case base.WeChat:
		pusher = push_method.NewWeChatPusher()
	case base.Email:
		pusher = push_method.NewEmailPusher()
	case base.SMS:
		pusher = push_method.NewSMSPusher()
	case base.Logger:
		pusher = push_method.NewLogPusher()
	default:
		return fmt.Errorf("不支持的推送方式: %s", method.String())
	}

	// 注册推送器
	if err := pc.pushRegistry.Register(method.String(), pusher); err != nil {
		return fmt.Errorf("注册推送器失败: %w", err)
	}

	pc.currentPusher = pusher
	pc.config = cfg

	// 创建延迟处理器
	pc.delayHandler = NewFileDelayHandler(cfg.DelayDir, cfg.ProcessedDir, pusher)

	// 启动延迟处理器
	if err := pc.delayHandler.Start(); err != nil {
		return fmt.Errorf("启动延迟处理器失败: %w", err)
	}

	return nil
}

// InitializeWithPusher 高级初始化（自定义推送器）
func (pc *PushController) InitializeWithPusher(cfg base.PushConfig, pusher push_method.IPusher) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pusher == nil {
		return fmt.Errorf("推送器不能为空")
	}

	// 注册推送器
	if err := pc.pushRegistry.Register(pusher.GetName(), pusher); err != nil {
		return fmt.Errorf("注册推送器失败: %w", err)
	}

	pc.currentPusher = pusher
	pc.config = cfg

	// 创建延迟处理器
	pc.delayHandler = NewFileDelayHandler(cfg.DelayDir, cfg.ProcessedDir, pusher)

	// 启动延迟处理器
	if err := pc.delayHandler.Start(); err != nil {
		return fmt.Errorf("启动延迟处理器失败: %w", err)
	}

	return nil
}

// PushNow 立即推送
func (pc *PushController) PushNow(message base.Message, options base.PushOptions) error {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	if pc.currentPusher == nil {
		return fmt.Errorf("推送器未初始化")
	}

	// 验证推送选项
	if err := pc.currentPusher.Validate(options); err != nil {
		return fmt.Errorf("推送选项验证失败: %w", err)
	}

	// 推送消息
	if err := pc.currentPusher.Push(message); err != nil {
		return fmt.Errorf("推送消息失败: %w", err)
	}

	log.Printf("消息推送成功: %s", message.ID)
	return nil
}

// Enqueue 入队消息（现在使用延迟文件处理）
func (pc *PushController) Enqueue(message base.Message, options base.PushOptions) error {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	if pc.delayHandler == nil {
		return fmt.Errorf("延迟处理器未初始化")
	}

	// 将消息写入延迟文件
	if err := pc.delayHandler.WriteDelayFile(message, options); err != nil {
		return fmt.Errorf("写入延迟文件失败: %w", err)
	}

	log.Printf("消息已写入延迟文件: %s", message.ID)
	return nil
}

// FlushQueue 刷新队列（现在处理延迟文件）
func (pc *PushController) FlushQueue() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.delayHandler == nil {
		return fmt.Errorf("延迟处理器未初始化")
	}

	// 处理延迟文件
	if err := pc.delayHandler.ProcessDelayFiles(); err != nil {
		return fmt.Errorf("处理延迟文件失败: %w", err)
	}

	return nil
}

// Stop 停止控制器
func (pc *PushController) Stop() {
	close(pc.stopChan)
	pc.wg.Wait()

	if pc.delayHandler != nil {
		pc.delayHandler.Stop()
	}
}

// GetRegisteredPushers 获取已注册的推送器列表
func (pc *PushController) GetRegisteredPushers() []string {
	if pc.pushRegistry == nil {
		return []string{}
	}
	return pc.pushRegistry.List()
}
