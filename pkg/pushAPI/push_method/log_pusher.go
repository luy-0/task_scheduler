package push_method

import (
	"log"
	"task_scheduler/pkg/pushAPI/base"
	"time"
)

// LogPusher 日志推送器（用于测试）
type LogPusher struct {
	BasePusher
}

// NewLogPusher 创建日志推送器
func NewLogPusher() *LogPusher {
	return &LogPusher{
		BasePusher: BasePusher{Name: "log"},
	}
}

// Push 推送消息
func (lp *LogPusher) Push(msg base.Message) error {
	log.Printf("日志推送 [%s]: %s - %s", time.Now().Format("2006-01-02 15:04:05"), msg.ID, msg.Content)
	return nil
}
