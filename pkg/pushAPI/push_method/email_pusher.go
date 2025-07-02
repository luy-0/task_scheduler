package push_method

import (
	"log"
	"task_scheduler/pkg/pushAPI/base"
)

// EmailPusher 邮件推送器
type EmailPusher struct {
	BasePusher
}

// NewEmailPusher 创建邮件推送器
func NewEmailPusher() *EmailPusher {
	return &EmailPusher{
		BasePusher: BasePusher{Name: "email"},
	}
}

// Push 推送消息
func (ep *EmailPusher) Push(msg base.Message) error {
	log.Printf("邮件推送: %s - %s", msg.ID, msg.Content)
	// 这里应该实现真实的邮件推送逻辑
	// 例如使用SMTP发送邮件
	return nil
}
