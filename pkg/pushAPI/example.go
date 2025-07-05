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

	// 入队推送（延迟消息）
	delayMessage := NewNormalMessage("app2", "延迟消息", "这是一条延迟消息")
	delayMessage.SetMetadata("delay_reason", "scheduled")

	if err := api.Enqueue(*delayMessage, options); err != nil {
		log.Printf("入队失败: %v", err)
	}

	// 再入队一条延迟消息
	delayMessage2 := NewNormalMessage("app2", "延迟消息2", "这是一条第二条延迟消息")
	delayMessage2.SetMetadata("delay_reason", "scheduled2")
	if err := api.Enqueue(*delayMessage2, options); err != nil {
		log.Printf("入队失败: %v", err)
	}

	// 手动刷新队列（会自动合并并发送所有延迟消息）
	if err := api.FlushQueue(); err != nil {
		log.Printf("刷新队列失败: %v", err)
	}

	// 获取队列大小（现在总是0，因为使用文件存储）
	if impl, ok := api.(*PushAPIImpl); ok {
		fmt.Printf("已注册的推送器: %v\n", impl.GetRegisteredPushers())
	}

	// 演示消息状态变化
	fmt.Printf("消息ID: %s\n", message.ID)
	fmt.Printf("消息状态: %s\n", message.SendStatus.String())
	fmt.Printf("发送时间: %v\n", message.SentAt)

	// 演示定时推送功能
	fmt.Println("\n6. 演示定时推送功能")

	// 定时推送（会自动合并并发送所有延迟消息）
	scheduledTime := time.Now().Add(1 * time.Minute)
	scheduledMessage := NewNormalMessage("app3", "定时通知", "这是一条定时发送的消息")
	if err := api.PushAt(*scheduledMessage, options, scheduledTime); err != nil {
		log.Printf("安排定时推送失败: %v", err)
	} else {
		fmt.Printf("定时消息已安排: %s -> %s\n", scheduledMessage.ID, scheduledTime.Format("15:04:05"))
	}

	// 创建另一个定时消息（20秒后发送）
	scheduledTime2 := time.Now().Add(20 * time.Second)
	scheduledMessage2 := NewMessage("app3", "紧急定时通知", "这是一条紧急定时消息", Emergency)
	scheduledMessage2.SetMetadata("scheduled_reason", "urgent_reminder")

	if err := api.PushAt(*scheduledMessage2, options, scheduledTime2); err != nil {
		log.Printf("安排定时推送失败: %v", err)
	} else {
		fmt.Printf("定时消息已安排: %s -> %s\n", scheduledMessage2.ID, scheduledTime2.Format("15:04:05"))
	}

	// 等待一段时间让定时推送处理器处理
	fmt.Println("等待30秒让定时推送处理器处理...")
	time.Sleep(30 * time.Second)

	fmt.Println("定时推送演示完成")
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
