# 定时任务调度器

这是一个可扩展的、基于配置的定时任务系统，使用 Go 语言开发。

## 功能特性

- 🕐 基于 cron 表达式的定时调度
- 🔌 插件化架构，支持动态扩展
- ⚙️ 基于 YAML 的配置管理
- 🛡️ 任务执行隔离和错误处理
- 📊 任务执行状态监控
- 🚀 优雅启动和停止

## 项目结构

```
task_scheduler/
├── configs/               # 配置文件目录
│   ├── config.yaml        # 主配置文件
│   └── tasks/             # 各任务配置
│       ├── app1.yaml
│       └── app2.yaml
├── internal/              # 内部模块
│   ├── core/              # 核心调度逻辑
│   │   ├── scheduler.go
│   │   └── task.go
│   ├── plugins/           # 插件接口定义
│   │   └── plugin.go
│   └── config/            # 配置加载与验证
│       └── loader.go
├── plugins/               # 插件实现
│   ├── app1/              # 任务1插件
│   │   └── plugin.go
│   └── app2/              # 任务2插件
│       └── plugin.go
├── main.go                # 入口文件
├── go.mod                 # 依赖管理
└── README.md              # 项目说明
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 运行程序

```bash
go run main.go
```

### 3. 查看日志

程序启动后会显示任务加载和执行日志：

```
2024/01/01 10:00:00 启动定时任务调度器...
2024/01/01 10:00:00 插件已注册: app1
2024/01/01 10:00:00 插件已注册: app2
2024/01/01 10:00:00 任务配置已加载: app1, 调度: */30 * * * * *
2024/01/01 10:00:00 任务配置已加载: app2, 调度: 0 */1 * * * *
2024/01/01 10:00:00 任务已添加: app1, 调度: */30 * * * * *
2024/01/01 10:00:00 任务已添加: app2, 调度: 0 */1 * * * *
2024/01/01 10:00:00 任务管理器已启动
2024/01/01 10:00:00 定时任务调度器已启动，按 Ctrl+C 停止...
```

## 配置说明

### 主配置文件 (configs/config.yaml)

```yaml
log_level: "info"
plugins_dir: "./plugins"
tasks:
  - name: "app1"
    config_file: "configs/tasks/app1.yaml"
    enabled: true
  - name: "app2"
    config_file: "configs/tasks/app2.yaml"
    enabled: true
```

### 任务配置文件 (configs/tasks/app1.yaml)

```yaml
schedule: "*/30 * * * * *"  # 每30秒执行一次
params:
  timeout: 30
  message: "Hello from App1 Task"
```

## 开发插件

### 1. 实现插件接口

```go
package myplugin

import (
    "context"
    "task_scheduler/internal/plugins"
)

type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "myplugin"
}

func (p *MyPlugin) CreateTask(config map[string]interface{}) (plugins.Task, error) {
    return &MyTask{config: config}, nil
}

func (p *MyPlugin) GetDefaultConfig() map[string]interface{} {
    return map[string]interface{}{
        "param1": "default_value",
    }
}

type MyTask struct {
    config map[string]interface{}
}

func (t *MyTask) Name() string {
    return "myplugin"
}

func (t *MyTask) Execute(ctx context.Context) error {
    // 实现任务逻辑
    return nil
}

func (t *MyTask) ValidateConfig(config map[string]interface{}) error {
    // 验证配置
    return nil
}
```

### 2. 注册插件

在 `main.go` 中添加插件注册：

```go
taskManager.RegisterPlugin(myplugin.NewPlugin())
```

### 3. 添加配置

在 `configs/config.yaml` 中添加任务配置：

```yaml
tasks:
  - name: "myplugin"
    config_file: "configs/tasks/myplugin.yaml"
    enabled: true
```

## 技术栈

- **调度引擎**: robfig/cron/v3
- **配置管理**: spf13/viper
- **语言**: Go 1.21+

## 许可证

MIT License 