package core

// import (
// 	"fmt"
// 	"strings"
// 	"task_scheduler/pkg/pushAPI/base"
// 	"task_scheduler/pkg/pushAPI/push_method"
// )

// // SimplePusherRouter 简单推送策略路由
// type SimplePusherRouter struct {
// 	registry PusherRegistry
// }

// // NewSimplePusherRouter 创建简单推送策略路由
// func NewSimplePusherRouter(registry PusherRegistry) *SimplePusherRouter {
// 	return &SimplePusherRouter{
// 		registry: registry,
// 	}
// }

// // Route 路由消息到推送器
// func (r *SimplePusherRouter) Route(msg base.Message) ([]push_method.IPusher, error) {
// 	var pushers []push_method.IPusher

// 	// 根据消息级别选择推送器
// 	switch strings.ToLower(msg.Level) {
// 	case "emergency":
// 		// 紧急消息：尝试所有可用的推送器
// 		names := r.registry.List()
// 		for _, name := range names {
// 			if pusher, err := r.registry.Get(name); err == nil {
// 				pushers = append(pushers, pusher)
// 			}
// 		}
// 	case "normal":
// 		// 普通消息：使用默认推送器（第一个可用的）
// 		names := r.registry.List()
// 		if len(names) > 0 {
// 			if pusher, err := r.registry.Get(names[0]); err == nil {
// 				pushers = append(pushers, pusher)
// 			}
// 		}
// 	default:
// 		// 默认使用第一个可用的推送器
// 		names := r.registry.List()
// 		if len(names) > 0 {
// 			if pusher, err := r.registry.Get(names[0]); err == nil {
// 				pushers = append(pushers, pusher)
// 			}
// 		}
// 	}

// 	if len(pushers) == 0 {
// 		return nil, fmt.Errorf("没有可用的推送器")
// 	}

// 	return pushers, nil
// }
