package core

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"task_scheduler/internal/plugins"

	"github.com/robfig/cron/v3"
)

// TaskManager 任务管理器
type TaskManager struct {
	cron    *cron.Cron
	tasks   map[string]*ManagedTask
	plugins map[string]plugins.Plugin
	results []plugins.TaskResult
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// ManagedTask 管理的任务
type ManagedTask struct {
	Info    plugins.TaskInfo
	Plugin  plugins.Plugin
	Task    plugins.Task
	EntryID cron.EntryID
}

// NewTaskManager 创建任务管理器
func NewTaskManager() *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskManager{
		cron:    cron.New(cron.WithSeconds()),
		tasks:   make(map[string]*ManagedTask),
		plugins: make(map[string]plugins.Plugin),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// RegisterPlugin 注册插件
func (tm *TaskManager) RegisterPlugin(plugin plugins.Plugin) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.plugins[plugin.Name()] = plugin
	log.Printf("插件已注册: %s", plugin.Name())
}

// AddTask 添加任务
func (tm *TaskManager) AddTask(info plugins.TaskInfo) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 检查插件是否存在
	plugin, exists := tm.plugins[info.Name]
	if !exists {
		return fmt.Errorf("插件不存在: %s", info.Name)
	}

	// 创建任务实例
	task, err := plugin.CreateTask(info.Config)
	if err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}

	// 验证配置
	if err := task.ValidateConfig(info.Config); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 添加定时任务
	entryID, err := tm.cron.AddFunc(info.Schedule, func() {
		tm.executeTask(task, info)
	})
	if err != nil {
		return fmt.Errorf("添加定时任务失败: %w", err)
	}

	// 保存任务信息
	managedTask := &ManagedTask{
		Info:    info,
		Plugin:  plugin,
		Task:    task,
		EntryID: entryID,
	}
	tm.tasks[info.Name] = managedTask

	log.Printf("任务已添加: %s, 调度: %s", info.Name, info.Schedule)
	return nil
}

// executeTask 执行任务
func (tm *TaskManager) executeTask(task plugins.Task, info plugins.TaskInfo) {
	startTime := time.Now()
	result := plugins.TaskResult{
		TaskName:  info.Name,
		StartTime: startTime,
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(tm.ctx, 30*time.Second)
	defer cancel()

	// 执行任务
	err := task.Execute(ctx)

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		log.Printf("任务执行失败: %s, 错误: %v", info.Name, err)
	} else {
		result.Success = true
		log.Printf("任务执行成功: %s, 耗时: %v", info.Name, result.Duration)
	}

	// 保存执行结果
	tm.mu.Lock()
	tm.results = append(tm.results, result)
	// 只保留最近100条记录
	if len(tm.results) > 100 {
		tm.results = tm.results[len(tm.results)-100:]
	}
	tm.mu.Unlock()
}

// Start 启动任务管理器
func (tm *TaskManager) Start() {
	tm.cron.Start()
	log.Println("任务管理器已启动")
}

// Stop 停止任务管理器
func (tm *TaskManager) Stop() {
	tm.cancel()
	tm.cron.Stop()
	log.Println("任务管理器已停止")
}

// GetTasks 获取所有任务
func (tm *TaskManager) GetTasks() map[string]*ManagedTask {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make(map[string]*ManagedTask)
	for k, v := range tm.tasks {
		result[k] = v
	}
	return result
}

// GetResults 获取执行结果
func (tm *TaskManager) GetResults() []plugins.TaskResult {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make([]plugins.TaskResult, len(tm.results))
	copy(result, tm.results)
	return result
}
