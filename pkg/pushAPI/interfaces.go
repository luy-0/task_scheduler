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

	// 定时推送方法
	PushAt(message Message, options PushOptions, scheduledAt time.Time) error
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

// MessageLevel 消息级别枚举
type MessageLevel int

const (
	Normal    MessageLevel = iota // 普通消息
	Emergency                     // 紧急消息
)

// String 返回消息级别的字符串表示
func (ml MessageLevel) String() string {
	switch ml {
	case Normal:
		return "normal"
	case Emergency:
		return "emergency"
	default:
		return "normal"
	}
}

// ToCore 转换为内部MessageLevel
func (ml MessageLevel) ToCore() base.MessageLevel {
	return base.MessageLevel(ml)
}

// SendStatus 发送状态枚举
type SendStatus int

const (
	StatusInitialized SendStatus = iota // 初始化
	StatusPending                       // 等待发送
	StatusSuccess                       // 成功
	StatusFailed                        // 失败
)

// String 返回发送状态的字符串表示
func (ss SendStatus) String() string {
	switch ss {
	case StatusInitialized:
		return "initialized"
	case StatusPending:
		return "pending"
	case StatusSuccess:
		return "success"
	case StatusFailed:
		return "failed"
	default:
		return "initialized"
	}
}

// ToCore 转换为内部SendStatus
func (ss SendStatus) ToCore() base.SendStatus {
	return base.SendStatus(ss)
}

// Message 消息体定义
type Message struct {
	ID         string                 `json:"id"`          // 消息唯一标识，自动生成格式：{app_id}_YYMMDD_{gen_id}
	AppID      string                 `json:"app_id"`      // 发送方ID，标志消息来源
	Title      string                 `json:"title"`       // 消息标题
	Content    string                 `json:"content"`     // 消息内容
	Level      MessageLevel           `json:"level"`       // 紧急程度
	Metadata   map[string]interface{} `json:"metadata"`    // 扩展元数据
	CreatedAt  time.Time              `json:"created_at"`  // 创建时间
	SentAt     time.Time              `json:"sent_at"`     // 最终成功发送时间
	SendStatus SendStatus             `json:"send_status"` // 发送状态
}

// NewMessage 创建新消息
func NewMessage(appID, title, content string, level MessageLevel) *Message {
	coreMsg := base.NewMessage(appID, title, content, level.ToCore())
	return &Message{
		ID:         coreMsg.ID,
		AppID:      coreMsg.AppID,
		Title:      coreMsg.Title,
		Content:    coreMsg.Content,
		Level:      level,
		Metadata:   coreMsg.Metadata,
		CreatedAt:  coreMsg.CreatedAt,
		SentAt:     coreMsg.SentAt,
		SendStatus: SendStatus(coreMsg.SendStatus),
	}
}

// NewMessageWithDefaultLevel 创建新消息（使用默认级别Normal）
func NewNormalMessage(appID, title, content string) *Message {
	return NewMessage(appID, title, content, Normal)
}

// SetMetadata 设置元数据
func (m *Message) SetMetadata(key string, value interface{}) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}
	m.Metadata[key] = value
}

// GetMetadata 获取元数据
func (m *Message) GetMetadata(key string) (interface{}, bool) {
	if m.Metadata == nil {
		return nil, false
	}
	value, exists := m.Metadata[key]
	return value, exists
}

// SetSentAt 设置发送时间
func (m *Message) SetSentAt(sentAt time.Time) {
	m.SentAt = sentAt
}

// SetSendStatus 设置发送状态
func (m *Message) SetSendStatus(status SendStatus) {
	m.SendStatus = status
}

// ToCore 转换为内部Message
func (m *Message) ToCore() base.Message {
	return base.Message{
		ID:         m.ID,
		AppID:      m.AppID,
		Title:      m.Title,
		Content:    m.Content,
		Level:      m.Level.ToCore(),
		Metadata:   m.Metadata,
		CreatedAt:  m.CreatedAt,
		SentAt:     m.SentAt,
		SendStatus: m.SendStatus.ToCore(),
	}
}

// FromCore 从内部Message创建
func FromCore(coreMsg base.Message) *Message {
	return &Message{
		ID:         coreMsg.ID,
		AppID:      coreMsg.AppID,
		Title:      coreMsg.Title,
		Content:    coreMsg.Content,
		Level:      MessageLevel(coreMsg.Level),
		Metadata:   coreMsg.Metadata,
		CreatedAt:  coreMsg.CreatedAt,
		SentAt:     coreMsg.SentAt,
		SendStatus: SendStatus(coreMsg.SendStatus),
	}
}

// PushOptions 推送选项
type PushOptions struct {
	Receivers []string `json:"receivers"` // 接收者列表
	Priority  int      `json:"priority"`  // 优先级
	Retry     int      `json:"retry"`     // 重试次数
}

// ToCore 转换为内部PushOptions
func (po *PushOptions) ToCore() base.PushOptions {
	return base.PushOptions{
		Receivers: po.Receivers,
		Priority:  po.Priority,
		Retry:     po.Retry,
	}
}

// Config 推送配置
type Config struct {
	QueueSize     int           `json:"queue_size"`     // 队列大小
	FlushInterval time.Duration `json:"flush_interval"` // 刷新间隔
	WorkingDir    string        `json:"working_dir"`    // 工作目录（存放延迟和定时消息）
	HistoryDir    string        `json:"history_dir"`    // 历史消息记录目录
	WeChatConfig  WeChatConfig  `json:"wechat_config"`  // 微信推送配置
}

// WeChatConfig 微信推送配置
type WeChatConfig struct {
	SendKey string `json:"send_key"` // 方糖气球sendKey
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		QueueSize:     1000,
		FlushInterval: 30 * time.Second,
		WorkingDir:    "./tmp/working",
		HistoryDir:    "./tmp/history",
		WeChatConfig: WeChatConfig{
			SendKey: "SCT7671TOKWWHhBntijf0DfzgF5luGPa", // 默认sendKey
		},
	}
}
