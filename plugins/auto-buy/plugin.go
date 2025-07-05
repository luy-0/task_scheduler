package autobuy

import (
	"context"
	"fmt"
	"log"
	"task_scheduler/internal/plugins"
)

// AutoBuyPlugin auto-buy插件实现
type AutoBuyPlugin struct{}

// AutoBuyTask auto-buy任务实现
type AutoBuyTask struct {
	name   string
	config map[string]interface{}
}

// NewPlugin 创建auto-buy插件
func NewPlugin() plugins.Plugin {
	return &AutoBuyPlugin{}
}

// Name 返回插件名称
func (p *AutoBuyPlugin) Name() string {
	return "auto-buy"
}

// CreateTask 创建任务实例
func (p *AutoBuyPlugin) CreateTask(config map[string]interface{}) (plugins.Task, error) {
	return &AutoBuyTask{
		name:   "auto-buy",
		config: config,
	}, nil
}

// GetDefaultConfig 获取默认配置
func (p *AutoBuyPlugin) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": true,
		"debug":   false,
	}
}

// Name 返回任务名称
func (t *AutoBuyTask) Name() string {
	return t.name
}

// Execute 执行任务
func (t *AutoBuyTask) Execute(ctx context.Context) error {
	log.Printf("开始执行 Auto-Buy 任务")

	// 获取配置参数
	enabled, _ := t.config["enabled"].(bool)
	if !enabled {
		log.Printf("Auto-Buy 任务已禁用")
		return nil
	}

	debug, _ := t.config["debug"].(bool)

	// 执行比特币定投逻辑
	if err := t.executeBitcoinStrategy(debug); err != nil {
		return fmt.Errorf("执行比特币定投策略失败: %w", err)
	}

	log.Printf("Auto-Buy 任务执行完成")
	return nil
}

// executeBitcoinStrategy 执行比特币定投策略
func (t *AutoBuyTask) executeBitcoinStrategy(debug bool) error {
	// 获取当前ahr999指标
	currPrice, ahr999Value, err := GetAhr999()
	if err != nil {
		return fmt.Errorf("获取AHR999指标失败: %w", err)
	}

	if debug {
		log.Printf("当前比特币价格: $%.2f", currPrice)
		log.Printf("当前AHR999指标: %.3f", ahr999Value)
	}

	// 根据AHR999指标决定定投策略
	investmentAmount := t.calculateInvestmentAmount(ahr999Value)

	if debug {
		log.Printf("建议定投金额: $%.2f", investmentAmount)
	}

	// TODO: 这里可以添加实际的交易逻辑
	// 目前只是记录策略结果
	log.Printf("比特币定投策略执行完成 - 价格: $%.2f, AHR999: %.3f, 建议定投: $%.2f",
		currPrice, ahr999Value, investmentAmount)

	return nil
}

// calculateInvestmentAmount 根据AHR999指标计算定投金额
func (t *AutoBuyTask) calculateInvestmentAmount(ahr999 float64) float64 {
	// 基础定投金额
	baseAmount := 100.0

	// 根据AHR999指标调整定投金额
	// AHR999 < 0.5: 大幅增加定投 (熊市)
	// AHR999 0.5-1.0: 正常定投
	// AHR999 > 1.0: 减少定投 (牛市)

	if ahr999 < 0.5 {
		return baseAmount * 2.0 // 熊市加倍定投
	} else if ahr999 > 1.0 {
		return baseAmount * 0.5 // 牛市减半定投
	} else {
		return baseAmount // 正常定投
	}
}

// ValidateConfig 验证配置
func (t *AutoBuyTask) ValidateConfig(config map[string]interface{}) error {
	// 检查enabled配置
	if enabled, exists := config["enabled"]; exists {
		if _, ok := enabled.(bool); !ok {
			return fmt.Errorf("enabled 必须是布尔类型")
		}
	}

	// 检查debug配置
	if debug, exists := config["debug"]; exists {
		if _, ok := debug.(bool); !ok {
			return fmt.Errorf("debug 必须是布尔类型")
		}
	}

	return nil
}
