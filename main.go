package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"task_scheduler/internal/config"
	"task_scheduler/internal/core"
	"task_scheduler/pkg/pushAPI"
	"task_scheduler/plugins/app1"
	"task_scheduler/plugins/app2"
	autobuy "task_scheduler/plugins/auto-buy"
)

func main() {
	main_pushAPI()
}

func main_pushAPI() {
	pushAPI.ExampleUsage()
}

func main_main() {
	log.Println("启动定时任务调度器...")

	// 创建配置加载器
	loader := config.NewLoader("configs/config.yaml")

	// 加载主配置
	mainConfig, err := loader.LoadMainConfig()
	if err != nil {
		log.Fatalf("加载主配置失败: %v", err)
	}

	// 验证配置
	if err := loader.ValidateConfig(mainConfig); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	// 创建任务管理器
	taskManager := core.NewTaskManager()

	// 注册插件
	taskManager.RegisterPlugin(app1.NewPlugin())
	taskManager.RegisterPlugin(app2.NewPlugin())
	taskManager.RegisterPlugin(autobuy.NewPlugin())

	// 加载所有任务配置
	tasks, err := loader.LoadAllTasks(mainConfig)
	if err != nil {
		log.Fatalf("加载任务配置失败: %v", err)
	}

	// 添加任务到调度器
	for _, task := range tasks {
		if err := taskManager.AddTask(task); err != nil {
			log.Printf("添加任务失败: %s, 错误: %v", task.Name, err)
			continue
		}
	}

	// 启动任务管理器
	taskManager.Start()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("定时任务调度器已启动，按 Ctrl+C 停止...")
	<-sigChan

	// 优雅停止
	log.Println("正在停止定时任务调度器...")
	taskManager.Stop()
	log.Println("定时任务调度器已停止")
}
