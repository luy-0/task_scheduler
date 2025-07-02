package core

import (
	"task_scheduler/pkg/pushAPI/base"
	"task_scheduler/pkg/pushAPI/push_method"
)

// DelayHandler 延迟处理模块接口
type DelayHandler interface {
	ProcessDelayFiles() error
	WriteDelayFile(msg base.Message, options base.PushOptions) error
	Start() error
	Stop() error
}

// PusherRouter 推送策略路由接口
type PusherRouter interface {
	Route(msg base.Message) ([]push_method.IPusher, error)
}

// PusherRegistry 推送器注册表接口
type PusherRegistry interface {
	Register(name string, pusher push_method.IPusher) error
	Get(name string) (push_method.IPusher, error)
	List() []string
	Unregister(name string) error
}
