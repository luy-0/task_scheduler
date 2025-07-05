package push_method

import (
	"fmt"
	"task_scheduler/pkg/pushAPI/base"
	"time"

	serverchan "github.com/easychen/serverchan-sdk-golang"
)

// WeChatPusher 微信推送器
type WeChatPusher struct {
	sendKey string
}

// NewWeChatPusher 创建微信推送器
func NewWeChatPusher() *WeChatPusher {
	return &WeChatPusher{
		sendKey: "SCT7671TOKWWHhBntijf0DfzgF5luGPa", // 默认sendKey
	}
}

// NewWeChatPusherWithKey 使用指定sendKey创建微信推送器
func NewWeChatPusherWithKey(sendKey string) *WeChatPusher {
	return &WeChatPusher{
		sendKey: sendKey,
	}
}

// GetName 获取推送器名称
func (w *WeChatPusher) GetName() string {
	return "wechat"
}

// Push 推送消息
func (w *WeChatPusher) Push(msg base.Message) error {
	// 构建消息内容
	content := w.buildMessageContent(msg)

	// 发送消息
	resp, err := serverchan.ScSend(w.sendKey, msg.Title, content, nil)
	if err != nil {
		return fmt.Errorf("微信推送失败: %w", err)
	}

	// 检查响应
	if resp != nil && resp.Code != 0 {
		return fmt.Errorf("微信推送失败: %s", resp.Message)
	}

	return nil
}

// Validate 验证推送选项
func (w *WeChatPusher) Validate(options base.PushOptions) error {

	return nil
}

// HealthCheck 健康检查
func (w *WeChatPusher) HealthCheck() bool {
	// 发送测试消息进行健康检查
	testMsg := base.Message{
		ID:      "health_check",
		AppID:   "system",
		Title:   "健康检查",
		Content: "这是一条健康检查消息",
		Level:   base.Normal,
	}

	err := w.Push(testMsg)
	return err == nil
}

// buildMessageContent 构建消息内容
func (w *WeChatPusher) buildMessageContent(msg base.Message) string {
	content := msg.Content

	// 添加消息级别标识
	levelStr := "普通"
	switch msg.Level {
	case base.Emergency:
		levelStr = "紧急"
	case base.Normal:
		levelStr = "普通"
	}

	// 添加时间戳
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 构建完整内容
	fullContent := fmt.Sprintf("【%s】\n", levelStr)
	fullContent += fmt.Sprintf("时间: %s\n\n", timestamp)
	fullContent += fmt.Sprintf("来源: %s\n\n", msg.AppID)
	fullContent += fmt.Sprintf("消息ID: %s\n\n", msg.ID)
	fullContent += fmt.Sprintf("内容: \n\n\n%s\n\n", content)

	// 添加元数据
	if len(msg.Metadata) > 0 {
		fullContent += "\n【元数据】\n"
		for key, value := range msg.Metadata {
			fullContent += fmt.Sprintf("%s: %v\n", key, value)
		}
	}

	return fullContent
}

// SetSendKey 设置sendKey
func (w *WeChatPusher) SetSendKey(sendKey string) {
	w.sendKey = sendKey
}

// GetSendKey 获取sendKey
func (w *WeChatPusher) GetSendKey() string {
	return w.sendKey
}
