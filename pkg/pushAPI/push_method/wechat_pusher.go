package push_method

import (
	"log"
	"task_scheduler/pkg/pushAPI/base"
)

// WeChatPusher 微信推送器
type WeChatPusher struct {
	BasePusher
}

// NewWeChatPusher 创建微信推送器
func NewWeChatPusher() *WeChatPusher {
	return &WeChatPusher{
		BasePusher: BasePusher{Name: "wechat"},
	}
}

// Push 推送消息
func (wp *WeChatPusher) Push(msg base.Message) error {
	log.Printf("微信推送: %s - %s", msg.ID, msg.Content)
	// 这里应该实现真实的微信推送逻辑
	// 例如调用微信企业号API或微信公众号API
	return nil
}
