package pushAPI

import (
	"task_scheduler/pkg/pushAPI/base"
	"time"
)

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
	GetName() string                         // 推送器名称
	Push(msg base.Message) error             // 核心推送方法
	Validate(options base.PushOptions) error // 参数验证
	HealthCheck() bool                       // 健康检查
}

// PushMethod 推送方式枚举
type PushMethod int

const (
	WeChat PushMethod = iota // 微信推送
	Email                    // 邮件推送
	SMS                      // 短信推送
	Logger                   // 日志推送
)

// String 返回推送方式的字符串表示
func (pm PushMethod) String() string {
	switch pm {
	case WeChat:
		return "wechat"
	case Email:
		return "email"
	case SMS:
		return "sms"
	case Logger:
		return "logger"
	default:
		return "unknown"
	}
}

func (pm PushMethod) ToCore() base.PushMethod {
	return base.PushMethod(pm)
}

// Message 消息体定义
type Message struct {
	ID        string                 `json:"id"`         // 消息唯一标识
	Content   string                 `json:"content"`    // 消息内容
	Level     string                 `json:"level"`      // 紧急程度(emergency/normal)
	Metadata  map[string]interface{} `json:"metadata"`   // 扩展元数据
	CreatedAt time.Time              `json:"created_at"` // 创建时间
}

// PushOptions 推送选项
type PushOptions struct {
	Receivers []string `json:"receivers"` // 接收者列表
	Priority  int      `json:"priority"`  // 优先级
	Retry     int      `json:"retry"`     // 重试次数
}

// Config 推送配置
type Config struct {
	QueueSize     int           `json:"queue_size"`     // 队列大小
	FlushInterval time.Duration `json:"flush_interval"` // 刷新间隔
	DelayDir      string        `json:"delay_dir"`      // 延迟文件目录
	ProcessedDir  string        `json:"processed_dir"`  // 已处理文件目录
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		QueueSize:     1000,
		FlushInterval: 30 * time.Second,
		DelayDir:      "./tmp/delay",
		ProcessedDir:  "./tmp/processed",
	}
}
