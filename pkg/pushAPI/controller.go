package pushAPI

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// PushController 核心推送控制器
type PushController struct {
	currentPusher Pusher         // 当前激活的推送器
	queue         MessageQueue   // 消息队列
	delayHandler  DelayHandler   // 延迟处理模块
	pusherRouter  PusherRouter   // 推送策略路由
	pushRegistry  PusherRegistry // 推送器注册表
	config        Config
	stopChan      chan struct{}
	wg            sync.WaitGroup
	mu            sync.RWMutex
}

// NewPushController 创建推送控制器
func NewPushController(cfg Config) *PushController {
	registry := NewPusherRegistry()
	router := NewSimplePusherRouter(registry)
	queue := NewMemoryQueue(cfg.QueueSize)

	return &PushController{
		queue:        queue,
		pusherRouter: router,
		pushRegistry: registry,
		config:       cfg,
		stopChan:     make(chan struct{}),
	}
}

// Initialize 初始化（选择内置推送方式）
func (pc *PushController) Initialize(cfg Config, method PushMethod) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// 根据推送方式创建内置推送器
	var pusher Pusher
	switch method {
	case WeChat:
		pusher = NewWeChatPusher()
	case Email:
		pusher = NewEmailPusher()
	case SMS:
		pusher = NewSMSPusher()
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

	// 启动队列刷新
	pc.startQueueFlusher()

	return nil
}

// InitializeWithPusher 高级初始化（自定义推送器）
func (pc *PushController) InitializeWithPusher(cfg Config, pusher Pusher) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pusher == nil {
		return fmt.Errorf("推送器不能为空")
	}

	// 注册推送器
	if err := pc.pushRegistry.Register(pusher.Name(), pusher); err != nil {
		return fmt.Errorf("注册推送器失败: %w", err)
	}

	pc.currentPusher = pusher
	pc.config = cfg

	// 创建延迟处理器
	pc.delayHandler = NewFileDelayHandler(cfg.DelayDir, cfg.ProcessedDir, pusher)

	// 启动队列刷新
	pc.startQueueFlusher()

	return nil
}

// PushNow 立即推送
func (pc *PushController) PushNow(message Message, options PushOptions) error {
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

// Enqueue 入队消息
func (pc *PushController) Enqueue(message Message, options PushOptions) error {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	if pc.queue == nil {
		return fmt.Errorf("消息队列未初始化")
	}

	// 将选项信息添加到消息元数据中
	if message.Metadata == nil {
		message.Metadata = make(map[string]interface{})
	}
	message.Metadata["push_options"] = options

	// 入队
	if err := pc.queue.Enqueue(message); err != nil {
		return fmt.Errorf("消息入队失败: %w", err)
	}

	log.Printf("消息已入队: %s", message.ID)
	return nil
}

// FlushQueue 刷新队列
func (pc *PushController) FlushQueue() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.queue == nil || pc.currentPusher == nil {
		return fmt.Errorf("队列或推送器未初始化")
	}

	// 获取所有消息
	messages, err := pc.queue.DequeueAll()
	if err != nil {
		return fmt.Errorf("获取队列消息失败: %w", err)
	}

	if len(messages) == 0 {
		return nil
	}

	// 批量推送
	for _, msg := range messages {
		if err := pc.currentPusher.Push(msg); err != nil {
			log.Printf("推送消息失败: %s, 错误: %v", msg.ID, err)
			continue
		}
		log.Printf("队列消息推送成功: %s", msg.ID)
	}

	return nil
}

// startQueueFlusher 启动队列刷新器
func (pc *PushController) startQueueFlusher() {
	pc.wg.Add(1)
	go func() {
		defer pc.wg.Done()

		ticker := time.NewTicker(pc.config.FlushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-pc.stopChan:
				return
			case <-ticker.C:
				if err := pc.FlushQueue(); err != nil {
					log.Printf("刷新队列失败: %v", err)
				}
			}
		}
	}()
}

// Stop 停止控制器
func (pc *PushController) Stop() {
	close(pc.stopChan)
	pc.wg.Wait()

	if pc.delayHandler != nil {
		pc.delayHandler.Stop()
	}
}

// GetQueueSize 获取队列大小
func (pc *PushController) GetQueueSize() int {
	if pc.queue == nil {
		return 0
	}
	return pc.queue.Size()
}

// GetRegisteredPushers 获取已注册的推送器列表
func (pc *PushController) GetRegisteredPushers() []string {
	if pc.pushRegistry == nil {
		return []string{}
	}
	return pc.pushRegistry.List()
}
