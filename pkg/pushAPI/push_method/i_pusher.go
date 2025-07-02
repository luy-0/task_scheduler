package push_method

import "task_scheduler/pkg/pushAPI/base"

// Pusher 推送器接口
type IPusher interface {
	GetName() string                         // 推送器名称
	Push(msg base.Message) error             // 核心推送方法
	Validate(options base.PushOptions) error // 参数验证
	HealthCheck() bool                       // 健康检查
}
