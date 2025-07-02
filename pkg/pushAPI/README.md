# 推送组件 (PushAPI)

这是一个可扩展的推送组件，采用"核心控制器+推送策略插件"的架构模式，支持多种推送方式。

## 架构设计

### 三层架构
- **接口层**: 对外提供统一服务接口
- **控制层**: 管理消息生命周期和推送策略
- **实现层**: 具体推送方式的插件化实现

### 核心组件
- `PushController`: 核心推送控制器
- `MessageQueue`: 消息队列管理
- `DelayHandler`: 延迟文件处理
- `PusherRouter`: 推送策略路由
- `PusherRegistry`: 推送器注册表

## 功能特性

- 🔌 **插件化架构**: 支持自定义推送器
- 📨 **多种推送方式**: 微信、邮件、短信等
- 🕐 **延迟处理**: 支持延迟文件推送
- 📊 **消息队列**: 内存队列，支持批量推送
- 🛡️ **错误处理**: 完善的错误处理和重试机制
- 📈 **健康检查**: 推送器健康状态监控

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
    message := pushAPI.Message{
        ID:      "msg_001",
        Content: "这是一条测试消息",
        Level:   "normal",
    }

    // 推送选项
    options := pushAPI.PushOptions{
        Receivers: []string{"user1", "user2"},
        Priority:  5,
        Retry:     3,
    }

    // 立即推送
    if err := api.PushNow(message, options); err != nil {
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

func (mp *MyPusher) Push(msg pushAPI.Message) error {
    // 实现自定义推送逻辑
    log.Printf("自定义推送: %s", msg.Content)
    return nil
}

// 使用自定义推送器
api := pushAPI.NewPushAPI()
cfg := pushAPI.DefaultConfig()
customPusher := NewMyPusher()

if err := api.InitializeWithPusher(cfg, customPusher); err != nil {
    log.Fatalf("初始化失败: %v", err)
}
```

### 3. 队列推送

```go
// 入队消息
message := pushAPI.Message{
    ID:      "queue_msg_001",
    Content: "这是一条队列消息",
    Level:   "normal",
}

options := pushAPI.PushOptions{
    Receivers: []string{"user1"},
    Priority:  3,
    Retry:     2,
}

// 入队
if err := api.Enqueue(message, options); err != nil {
    log.Printf("入队失败: %v", err)
}

// 手动刷新队列
if err := api.FlushQueue(); err != nil {
    log.Printf("刷新队列失败: %v", err)
}
```

## 配置说明

### Config 配置结构

```go
type Config struct {
    QueueSize     int           // 队列大小
    FlushInterval time.Duration // 刷新间隔
    DelayDir      string        // 延迟文件目录
    ProcessedDir  string        // 已处理文件目录
}
```

### 默认配置

```go
func DefaultConfig() Config {
    return Config{
        QueueSize:     1000,
        FlushInterval: 30 * time.Second,
        DelayDir:      "./delay",
        ProcessedDir:  "./processed",
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
    ID        string                 // 消息唯一标识
    Content   string                 // 消息内容
    Level     string                 // 紧急程度(emergency/normal)
    Metadata  map[string]interface{} // 扩展元数据
    CreatedAt time.Time              // 创建时间
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
JSON Lines格式，每条消息一行

### 处理策略
- 每小时检查一次新文件
- 使用文件锁保证并发安全
- 成功推送后移动文件到processed目录
- 失败时保留原文件并记录错误日志

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

func (cp *CustomPusher) Push(msg pushAPI.Message) error {
    // 实现推送逻辑
    return nil
}
```

### 注册推送器

```go
registry := pushAPI.NewPusherRegistry()
customPusher := NewCustomPusher()
registry.Register("custom", customPusher)
```

## 注意事项

1. **线程安全**: 所有组件都是线程安全的
2. **资源管理**: 记得调用 `Stop()` 方法释放资源
3. **错误处理**: 推送失败不会影响其他消息
4. **配置验证**: 初始化时会验证配置参数
5. **健康检查**: 定期检查推送器健康状态 