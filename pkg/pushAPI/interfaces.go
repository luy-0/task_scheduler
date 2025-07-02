package pushAPI

// PushAPI 模块接口定义
type PushAPI interface {
	// 初始化（选择内置推送方式）
	Initialize(cfg Config, method PushMethod) error

	// 高级初始化（自定义推送器）
	InitializeWithPusher(cfg Config, pusher Pusher) error

	// 推送方法
	PushNow(message Message, options PushOptions) error
	Enqueue(message Message, options PushOptions) error
	FlushQueue() error
}

// Pusher 推送器接口
type Pusher interface {
	Name() string                       // 推送器名称
	Push(msg Message) error             // 核心推送方法
	Validate(options PushOptions) error // 参数验证
	HealthCheck() bool                  // 健康检查
}

// MessageQueue 消息队列接口
type MessageQueue interface {
	Enqueue(msg Message) error
	DequeueAll() ([]Message, error)
	Clear() error
	Size() int
}

// DelayHandler 延迟处理模块接口
type DelayHandler interface {
	ProcessDelayFiles() error
	WriteDelayFile(msg Message, options PushOptions) error
	Start() error
	Stop() error
}

// PusherRouter 推送策略路由接口
type PusherRouter interface {
	Route(msg Message) ([]Pusher, error)
}

// PusherRegistry 推送器注册表接口
type PusherRegistry interface {
	Register(name string, pusher Pusher) error
	Get(name string) (Pusher, error)
	List() []string
	Unregister(name string) error
}
