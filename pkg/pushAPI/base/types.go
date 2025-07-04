package base

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

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

// ParseMessageLevel 解析消息级别字符串
func ParseMessageLevel(level string) MessageLevel {
	switch strings.ToLower(level) {
	case "emergency":
		return Emergency
	case "normal":
		return Normal
	default:
		return Normal
	}
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

// ParseSendStatus 解析发送状态字符串
func ParseSendStatus(status string) SendStatus {
	switch strings.ToLower(status) {
	case "pending":
		return StatusPending
	case "success":
		return StatusSuccess
	case "failed":
		return StatusFailed
	case "initialized":
		return StatusInitialized
	default:
		return StatusInitialized
	}
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
	mu         sync.RWMutex           `json:"-"`           // 用于metadata操作的互斥锁
}

// NewMessage 创建新消息
func NewMessage(appID, title, content string, level MessageLevel) *Message {
	return &Message{
		ID:         generateMessageID(appID),
		AppID:      appID,
		Title:      title,
		Content:    content,
		Level:      level,
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
		SendStatus: StatusInitialized,
	}
}

// NewMessageWithDefaultLevel 创建新消息（使用默认级别Normal）
func NewMessageWithDefaultLevel(appID, title, content string) *Message {
	return NewMessage(appID, title, content, Normal)
}

// SetMetadata 设置元数据
func (m *Message) SetMetadata(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}
	m.Metadata[key] = value
}

// GetMetadata 获取元数据
func (m *Message) GetMetadata(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
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

// generateMessageID 生成消息ID
func generateMessageID(appID string) string {
	now := time.Now()
	dateStr := now.Format("060102_150405")                // YYMMDD格式
	nanoStr := fmt.Sprintf("%06d", now.Nanosecond()/1000) // 微秒部分

	return fmt.Sprintf("%s_%s_%s", appID, dateStr, nanoStr)
}

// PushOptions 推送选项
type PushOptions struct {
	Receivers []string `json:"receivers"` // 接收者列表
	Priority  int      `json:"priority"`  // 优先级
	Retry     int      `json:"retry"`     // 重试次数
}

// PushConfig 推送配置
type PushConfig struct {
	QueueSize     int           `json:"queue_size"`     // 队列大小
	FlushInterval time.Duration `json:"flush_interval"` // 刷新间隔
	DelayDir      string        `json:"delay_dir"`      // 延迟文件目录
	ProcessedDir  string        `json:"processed_dir"`  // 已处理文件目录
	HistoryDir    string        `json:"history_dir"`    // 历史消息记录目录
	WorkingDir    string        `json:"working_dir"`    // 定时推送工作目录
}

// DefaultConfig 返回默认配置
func DefaultConfig() PushConfig {
	return PushConfig{
		QueueSize:     1000,
		FlushInterval: 30 * time.Second,
		DelayDir:      "./delay",
		ProcessedDir:  "./processed",
		HistoryDir:    "./history",
		WorkingDir:    "./working",
	}
}

// DelayMessage 延迟消息结构
type DelayMessage struct {
	Message Message     `json:"message"`
	Options PushOptions `json:"options"`
}

// ScheduledMessage 定时消息结构
type ScheduledMessage struct {
	Message     Message     `json:"message"`
	Options     PushOptions `json:"options"`
	ScheduledAt time.Time   `json:"scheduled_at"` // 计划发送时间
}

// HistoryRecord 历史记录结构
type HistoryRecord struct {
	Timestamp   time.Time `json:"timestamp"`    // 时间
	AppID       string    `json:"app_id"`       // 发送方
	PusherName  string    `json:"pusher_name"`  // 发送途径
	Title       string    `json:"title"`        // 标题
	Content     string    `json:"content"`      // 发送内容
	MessageID   string    `json:"message_id"`   // 消息ID
	Level       string    `json:"level"`        // 消息级别
	Receivers   []string  `json:"receivers"`    // 接收者
	Priority    int       `json:"priority"`     // 优先级
	RetryCount  int       `json:"retry_count"`  // 重试次数
	ErrorReason string    `json:"error_reason"` // 失败原因（仅失败记录）
}

// NewSuccessHistoryRecord 创建成功发送历史记录
func NewSuccessHistoryRecord(msg Message, pusherName string, options PushOptions) *HistoryRecord {
	return &HistoryRecord{
		Timestamp:  time.Now(),
		AppID:      msg.AppID,
		PusherName: pusherName,
		Title:      msg.Title,
		Content:    msg.Content,
		MessageID:  msg.ID,
		Level:      msg.Level.String(),
		Receivers:  options.Receivers,
		Priority:   options.Priority,
		RetryCount: options.Retry,
	}
}

// NewFailedHistoryRecord 创建失败发送历史记录
func NewFailedHistoryRecord(msg Message, pusherName string, options PushOptions, errorReason string) *HistoryRecord {
	return &HistoryRecord{
		Timestamp:   time.Now(),
		AppID:       msg.AppID,
		PusherName:  pusherName,
		Title:       msg.Title,
		Content:     msg.Content,
		MessageID:   msg.ID,
		Level:       msg.Level.String(),
		Receivers:   options.Receivers,
		Priority:    options.Priority,
		RetryCount:  options.Retry,
		ErrorReason: errorReason,
	}
}
