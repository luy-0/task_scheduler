package push_method

import (
	"log"
	"task_scheduler/pkg/pushAPI/base"
)

// SMSPusher 短信推送器
type SMSPusher struct {
	BasePusher
}

// NewSMSPusher 创建短信推送器
func NewSMSPusher() *SMSPusher {
	return &SMSPusher{
		BasePusher: BasePusher{Name: "sms"},
	}
}

// Push 推送消息
func (sp *SMSPusher) Push(msg base.Message) error {
	log.Printf("短信推送: %s - %s", msg.ID, msg.Content)
	// 这里应该实现真实的短信推送逻辑
	// 例如调用短信服务商API
	return nil
}
