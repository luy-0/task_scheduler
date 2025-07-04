# 推送组件 (PushAPI)

这是一个可扩展的推送组件，采用"核心控制器+推送策略插件"的架构模式，支持多种推送方式。

## 架构设计

### 三层架构
- **接口层**: 对外提供统一服务接口 (`api.go`, `types.go`)
- **控制层**: 管理消息生命周期和推送策略 (`core/`)
- **实现层**: 具体推送方式的插件化实现 (`push_method/`)

### 核心组件
- `core.PushController`: 核心推送控制器
- `core.DelayHandler`: 延迟文件处理
- `core.PusherRouter`: 推送策略路由
- `core.PusherRegistry`: 推送器注册表
- `push_method/`: 各种推送器实现

## 功能特性

- 🔌 **插件化架构**: 支持自定义推送器
- 📨 **多种推送方式**: 微信、邮件、短信等
- 🕐 **延迟处理**: 使用文件存储延迟消息
- 📊 **文件队列**: 基于文件的延迟消息处理
- 🛡️ **错误处理**: 完善的错误处理和重试机制
- 📈 **健康检查**: 推送器健康状态监控

## 目录结构

```
pkg/pushAPI/
├── api.go              # 对外API接口实现
├── types.go            # 对外类型定义
├── example_test.go     # 使用示例
├── README.md           # 文档说明
├── core/               # 核心实现
│   ├── types.go        # 内部类型定义
│   ├── interfaces.go   # 内部接口定义
│   ├── base_pusher.go  # 基础推送器
│   ├── controller.go   # 推送控制器
│   ├── delay_handler.go # 延迟文件处理
│   ├── registry.go     # 推送器注册表
│   ├── router.go       # 推送策略路由
│   └── queue.go        # 内存队列（已废弃）
└── push_method/        # 推送器实现
    ├── wechat_pusher.go # 微信推送器
    ├── email_pusher.go  # 邮件推送器
    ├── sms_pusher.go    # 短信推送器
    └── log_pusher.go    # 日志推送器
```

## 快速开始

### 1. 基本使用

```go
package main

import (
    "log"
    "time"
    "task_scheduler/pkg/pushAPI"
)

func main() {
    // 创建推送API实例
    api := pushAPI.NewPushAPI()

    // 配置
    cfg := pushAPI.DefaultConfig()
    cfg.QueueSize = 100
    cfg.FlushInterval = 10 * time.Second

    // 初始化（使用微信推送）
    if err := api.Initialize(cfg, pushAPI.WeChat); err != nil {
        log.Fatalf("初始化失败: %v", err)
    }

    // 创建消息
    message := pushAPI.NewMessageWithDefaultLevel("app1", "测试消息", "这是一条测试消息")
    message.SetMetadata("source", "test")

    // 推送选项
    options := pushAPI.PushOptions{
        Receivers: []string{"user1", "user2"},
        Priority:  5,
        Retry:     3,
    }

    // 立即推送
    if err := api.PushNow(*message, options); err != nil {
        log.Printf("推送失败: %v", err)
    }
}
```

### 2. 自定义推送器

```go
// 创建自定义推送器
type MyPusher struct {
    pushAPI.BasePusher
}

func NewMyPusher() *MyPusher {
    return &MyPusher{
        BasePusher: pushAPI.BasePusher{Name: "my_pusher"},
    }
}

func (mp *MyPusher) Name() string {
    return mp.BasePusher.Name
}

func (mp *MyPusher) Push(msg pushAPI.Message) error {
    // 实现自定义推送逻辑
    log.Printf("自定义推送: %s", msg.Content)
    return nil
}

func (mp *MyPusher) Validate(options pushAPI.PushOptions) error {
    // 验证配置
    return nil
}

func (mp *MyPusher) HealthCheck() bool {
    return true
}

// 使用自定义推送器
api := pushAPI.NewPushAPI()
cfg := pushAPI.DefaultConfig()
customPusher := NewMyPusher()

if err := api.InitializeWithPusher(cfg, customPusher); err != nil {
    log.Fatalf("初始化失败: %v", err)
}
```

### 3. 延迟推送

```go
// 入队消息（使用文件存储）
message := pushAPI.Message{
    ID:      "delay_msg_001",
    Content: "这是一条延迟消息",
    Level:   "normal",
}

options := pushAPI.PushOptions{
    Receivers: []string{"user1"},
    Priority:  3,
    Retry:     2,
}

// 入队（写入延迟文件）
if err := api.Enqueue(message, options); err != nil {
    log.Printf("入队失败: %v", err)
}

// 手动处理延迟文件
if err := api.FlushQueue(); err != nil {
    log.Printf("处理延迟文件失败: %v", err)
}
```

## 配置说明

### Config 配置结构

```go
type Config struct {
    QueueSize     int           // 队列大小（已废弃，使用文件存储）
    FlushInterval time.Duration // 刷新间隔
    DelayDir      string        // 延迟文件目录
    ProcessedDir  string        // 已处理文件目录
    HistoryDir    string        // 历史消息记录目录
9    WorkingDir    string        // 定时推送工作目录
}
```

### 默认配置

```go
func DefaultConfig() Config {
    return Config{
        QueueSize:     1000,
        FlushInterval: 30 * time.Second,
        DelayDir:      "./tmp/delay",
        ProcessedDir:  "./tmp/processed",
        HistoryDir:    "./tmp/history",
        WorkingDir:    "./tmp/working",
    }
}
```

## 推送方式

### 内置推送器

1. **WeChatPusher**: 微信推送
2. **EmailPusher**: 邮件推送
3. **SMSPusher**: 短信推送
4. **LogPusher**: 日志推送（用于测试）

### 推送方式枚举

```go
type PushMethod int

const (
    WeChat PushMethod = iota // 微信推送
    Email                    // 邮件推送
    SMS                      // 短信推送
)
```

## 消息结构

### Message 消息体

```go
type Message struct {
    ID         string                 // 消息唯一标识，自动生成格式：{app_id}_YYMMDD_{gen_id}
    AppID      string                 // 发送方ID，标志消息来源
    Title      string                 // 消息标题
    Content    string                 // 消息内容
    Level      MessageLevel           // 紧急程度（枚举）
    Metadata   map[string]interface{} // 扩展元数据
    CreatedAt  time.Time              // 创建时间
    SentAt     time.Time              // 最终成功发送时间
    SendStatus SendStatus             // 发送状态（枚举）
}
```

### 消息级别枚举

```go
type MessageLevel int

const (
    Normal MessageLevel = iota // 普通消息
    Emergency                  // 紧急消息
)
```

### 发送状态枚举

```go
type SendStatus int

const (
    StatusInitialized SendStatus = iota // 初始化
    StatusPending                       // 等待发送
    StatusSuccess                       // 成功
    StatusFailed                        // 失败
)
```

### 创建消息

```go
// 使用默认级别（Normal）创建消息
message := NewMessageWithDefaultLevel("app1", "消息标题", "消息内容")

// 指定级别创建消息
message := NewMessage("app1", "紧急通知", "紧急消息内容", Emergency)

// 设置元数据
message.SetMetadata("user_id", "12345")
message.SetMetadata("source", "system")

// 获取元数据
if value, exists := message.GetMetadata("user_id"); exists {
    fmt.Printf("用户ID: %v\n", value)
}
```

### PushOptions 推送选项

```go
type PushOptions struct {
    Receivers []string // 接收者列表
    Priority  int      // 优先级 (0-10)
    Retry     int      // 重试次数 (0-5)
}
```

## 延迟处理

### 文件命名规则
- 格式: `delay_{timestamp}.msg`
- 示例: `delay_20230701_1200.msg`

### 存储格式
JSON格式，包含消息和推送选项

### 处理策略
- 每小时检查一次新文件
- 使用文件锁保证并发安全
- 成功推送后移动文件到processed目录
- 失败时保留原文件并记录错误日志

## 定时推送

### 功能概述
推送组件支持定时推送功能，可以指定消息在特定时间发送。

### 时间精度
- 时间精度为分钟，自动舍弃秒和毫秒
- 每分钟检查一次是否有需要发送的消息

### 文件组织
- 按4小时时间段组织文件：`scheduled_20240101_08.json`、`scheduled_20240101_12.json`
- 存储格式：JSON数组，每条记录包含消息、推送选项和计划发送时间

### 定时推送API

```go
// 安排定时推送（1小时后发送）
scheduledTime := time.Now().Add(1 * time.Hour)
message := pushAPI.NewNormalMessage("app1", "定时通知", "这是一条定时消息")
options := pushAPI.PushOptions{
    Receivers: []string{"user1"},
    Priority:  5,
    Retry:     2,
}

if err := api.PushAt(*message, options, scheduledTime); err != nil {
    log.Printf("安排定时推送失败: %v", err)
} else {
    fmt.Printf("定时消息已安排: %s -> %s\n", message.ID, scheduledTime.Format("15:04:05"))
}
```

### 文件命名规则
- 格式: `scheduled_YYYYMMDD_HH.json`
- 示例: `scheduled_20240101_08.json` (1月1日8-12点时间段)
- 时间段划分: 0-4点、4-8点、8-12点、12-16点、16-20点、20-24点

### 处理机制
- 每分钟触发一次检查
- 扫描当前4小时时间段的文件
- 发送到期的消息并从文件中移除
- 自动记录发送历史（成功/失败）

### 定时消息结构

```go
type ScheduledMessage struct {
    Message     Message     // 消息内容
    Options     PushOptions // 推送选项
    ScheduledAt time.Time   // 计划发送时间
}
```

## 历史记录

### 功能概述
推送组件会自动记录所有消息的发送历史，包括成功和失败的记录。

### 文件组织
- 按月份组织文件：`success_send_202401.json`、`failed_send_202401.json`
- 存储格式：JSON数组，每条记录包含完整的消息和发送信息

### 记录内容
成功记录包含：
- 时间戳
- 发送方ID
- 发送途径（推送器名称）
- 消息标题
- 消息内容
- 消息ID
- 消息级别
- 接收者列表
- 优先级
- 重试次数

失败记录额外包含：
- 失败原因

### 历史记录结构

```go
type HistoryRecord struct {
    Timestamp   time.Time // 时间
    AppID       string    // 发送方
    PusherName  string    // 发送途径
    Title       string    // 标题
    Content     string    // 发送内容
    MessageID   string    // 消息ID
    Level       string    // 消息级别
    Receivers   []string  // 接收者
    Priority    int       // 优先级
    RetryCount  int       // 重试次数
    ErrorReason string    // 失败原因（仅失败记录）
}
```

## 错误处理

### 验证规则
- 接收者列表不能为空
- 优先级必须在0-10之间
- 重试次数必须在0-5之间

### 重试机制
- 支持配置重试次数
- 失败时记录详细错误日志
- 不会影响其他消息的推送

## 扩展开发

### 实现自定义推送器

1. 实现 `Pusher` 接口
2. 继承 `BasePusher` 获取基础功能
3. 实现具体的推送逻辑

```go
type CustomPusher struct {
    pushAPI.BasePusher
    // 自定义字段
}

func (cp *CustomPusher) Name() string {
    return cp.BasePusher.Name
}

func (cp *CustomPusher) Push(msg pushAPI.Message) error {
    // 实现推送逻辑
    return nil
}

func (cp *CustomPusher) Validate(options pushAPI.PushOptions) error {
    // 验证配置
    return nil
}

func (cp *CustomPusher) HealthCheck() bool {
    return true
}
```

## 架构变更说明

### v2.0 主要变更

1. **延迟处理重构**: 从内存队列改为文件存储
2. **代码结构优化**: 内部实现移至 `core/` 目录
3. **推送器分离**: 推送器实现移至 `push_method/` 目录
4. **接口简化**: 外部接口保持稳定，内部实现重构

### 向后兼容性

- 外部API接口保持不变
- 配置结构保持不变
- 消息和选项结构保持不变

## 注意事项

1. **线程安全**: 所有组件都是线程安全的
2. **资源管理**: 记得调用 `Stop()` 方法释放资源
3. **错误处理**: 推送失败不会影响其他消息
4. **配置验证**: 初始化时会验证配置参数
5. **健康检查**: 定期检查推送器健康状态
6. **文件存储**: 延迟消息现在使用文件存储，确保目录权限正确 