package pushAPI

import (
	"fmt"
	"log"
	"task_scheduler/pkg/pushAPI/push_method"
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

	// 初始化（使用微信推送）
	if err := api.Initialize(cfg, Logger); err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer func() {
		if impl, ok := api.(*PushAPIImpl); ok {
			impl.Stop()
		}
	}()

	// 创建消息（使用新的构造函数）
	message := NewNormalMessage("app1", "测试消息", "这是一条测试消息")
	message.SetMetadata("source", "test")
	message.SetMetadata("user_id", "12345")

	// 推送选项
	options := PushOptions{
		Receivers: []string{"user1", "user2"},
		Priority:  5,
		Retry:     3,
	}

	// 立即推送
	if err := api.PushNow(*message, options); err != nil {
		log.Printf("立即推送失败: %v", err)
	}

	// 创建紧急消息
	emergencyMessage := NewMessage("app1", "紧急通知", "这是一条紧急消息", Emergency)
	emergencyMessage.SetMetadata("alert_type", "system_error")

	// 推送紧急消息
	if err := api.PushNow(*emergencyMessage, options); err != nil {
		log.Printf("紧急消息推送失败: %v", err)
	}

	// 入队推送（现在使用文件存储）
	delayMessage := NewNormalMessage("app2", "延迟消息", "这是一条延迟消息")
	delayMessage.SetMetadata("delay_reason", "scheduled")

	if err := api.Enqueue(*delayMessage, options); err != nil {
		log.Printf("入队失败: %v", err)
	}

	// 等待一段时间让延迟处理器处理
	time.Sleep(15 * time.Second)

	// 手动刷新队列（处理延迟文件）
	if err := api.FlushQueue(); err != nil {
		log.Printf("刷新队列失败: %v", err)
	}

	// 获取队列大小（现在总是0，因为使用文件存储）
	if impl, ok := api.(*PushAPIImpl); ok {
		fmt.Printf("队列大小: %d\n", impl.GetQueueSize())
		fmt.Printf("已注册的推送器: %v\n", impl.GetRegisteredPushers())
	}

	// 演示消息状态变化
	fmt.Printf("消息ID: %s\n", message.ID)
	fmt.Printf("消息状态: %s\n", message.SendStatus.String())
	fmt.Printf("发送时间: %v\n", message.SentAt)
}

// ExampleCustomPusher 自定义推送器示例
func ExampleCustomPusher() {
	// 创建自定义推送器
	customPusher := push_method.NewLogPusher()

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
	message := NewMessage("custom_app", "自定义推送器消息", "这是一条自定义推送器消息", Emergency)

	// 推送选项
	options := PushOptions{
		Receivers: []string{"admin"},
		Priority:  10,
		Retry:     0,
	}

	// 推送消息
	if err := api.PushNow(*message, options); err != nil {
		log.Printf("推送失败: %v", err)
	}
}
