package push_method

import (
	"fmt"
	"task_scheduler/pkg/pushAPI/base"
)

// BasePusher 基础推送器
type BasePusher struct {
	Name string
}

// GetName 返回推送器名称
func (bp *BasePusher) GetName() string {
	return bp.Name
}

// Validate 验证推送选项
func (bp *BasePusher) Validate(options base.PushOptions) error {
	if len(options.Receivers) == 0 {
		return fmt.Errorf("接收者列表不能为空")
	}

	if options.Priority < 0 || options.Priority > 10 {
		return fmt.Errorf("优先级必须在0-10之间")
	}

	if options.Retry < 0 || options.Retry > 5 {
		return fmt.Errorf("重试次数必须在0-5之间")
	}

	return nil
}

// HealthCheck 健康检查
func (bp *BasePusher) HealthCheck() bool {
	return true
}
