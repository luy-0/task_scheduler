package pushAPI

import (
	"fmt"
	"log"
	"time"
)

// ExampleUsage 使用示例
func ExampleUsage() {
	// 创建推送API实例
	api := NewPushAPI()

	// 配置
	cfg := DefaultConfig()
	cfg.QueueSize = 100
	cfg.FlushInterval = 10 * time.Second

	// 初始化（使用内置推送方式）
	if err := api.Initialize(cfg, WeChat); err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer func() {
		if impl, ok := api.(*PushAPIImpl); ok {
			impl.Stop()
		}
	}()

	// 创建消息
	message := Message{
		ID:        "msg_001",
		Content:   "这是一条测试消息",
		Level:     "normal",
		Metadata:  map[string]interface{}{"source": "test"},
		CreatedAt: time.Now(),
	}

	// 推送选项
	options := PushOptions{
		Receivers: []string{"user1", "user2"},
		Priority:  5,
		Retry:     3,
	}

	// 立即推送
	if err := api.PushNow(message, options); err != nil {
		log.Printf("立即推送失败: %v", err)
	}

	// 入队推送
	message2 := Message{
		ID:      "msg_002",
		Content: "这是一条队列消息",
		Level:   "normal",
	}
	if err := api.Enqueue(message2, options); err != nil {
		log.Printf("入队失败: %v", err)
	}

	// 等待一段时间让队列刷新
	time.Sleep(15 * time.Second)

	// 手动刷新队列
	if err := api.FlushQueue(); err != nil {
		log.Printf("刷新队列失败: %v", err)
	}

	// 获取队列大小
	if impl, ok := api.(*PushAPIImpl); ok {
		fmt.Printf("队列大小: %d\n", impl.GetQueueSize())
		fmt.Printf("已注册的推送器: %v\n", impl.GetRegisteredPushers())
	}
}

// ExampleCustomPusher 自定义推送器示例
func ExampleCustomPusher() {
	// 创建自定义推送器
	customPusher := NewLogPusher()

	// 创建推送API实例
	api := NewPushAPI()

	// 配置
	cfg := DefaultConfig()

	// 使用自定义推送器初始化
	if err := api.InitializeWithPusher(cfg, customPusher); err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer func() {
		if impl, ok := api.(*PushAPIImpl); ok {
			impl.Stop()
		}
	}()

	// 创建消息
	message := Message{
		ID:      "custom_msg_001",
		Content: "这是一条自定义推送器消息",
		Level:   "emergency",
	}

	// 推送选项
	options := PushOptions{
		Receivers: []string{"admin"},
		Priority:  10,
		Retry:     0,
	}

	// 推送消息
	if err := api.PushNow(message, options); err != nil {
		log.Printf("推送失败: %v", err)
	}
}
