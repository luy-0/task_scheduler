package core

import (
	"fmt"
	"sync"
	"task_scheduler/pkg/pushAPI/push_method"
)

// PusherRegistryImpl 推送器注册表实现
type PusherRegistryImpl struct {
	pushers map[string]push_method.IPusher
	mu      sync.RWMutex
}

// NewPusherRegistry 创建推送器注册表
func NewPusherRegistry() *PusherRegistryImpl {
	return &PusherRegistryImpl{
		pushers: make(map[string]push_method.IPusher),
	}
}

// Register 注册推送器
func (r *PusherRegistryImpl) Register(name string, pusher push_method.IPusher) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return fmt.Errorf("推送器名称不能为空")
	}

	if pusher == nil {
		return fmt.Errorf("推送器不能为空")
	}

	if _, exists := r.pushers[name]; exists {
		return fmt.Errorf("推送器已存在: %s", name)
	}

	r.pushers[name] = pusher
	return nil
}

// Get 获取推送器
func (r *PusherRegistryImpl) Get(name string) (push_method.IPusher, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pusher, exists := r.pushers[name]
	if !exists {
		return nil, fmt.Errorf("推送器不存在: %s", name)
	}

	return pusher, nil
}

// List 列出所有推送器名称
func (r *PusherRegistryImpl) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.pushers))
	for name := range r.pushers {
		names = append(names, name)
	}

	return names
}

// Unregister 注销推送器
func (r *PusherRegistryImpl) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.pushers[name]; !exists {
		return fmt.Errorf("推送器不存在: %s", name)
	}

	delete(r.pushers, name)
	return nil
}
