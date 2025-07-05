package autobuy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"task_scheduler/internal/plugins"
	"task_scheduler/pkg/pushAPI"
)

// AutoBuyPlugin auto-buy插件实现
type AutoBuyPlugin struct{}

// AutoBuyTask auto-buy任务实现
type AutoBuyTask struct {
	name             string
	config           map[string]interface{}
	baseAmount       float64
	ahr999TimerTable Ahr999TimerTable
	pusher           pushAPI.PushAPI
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
	task := &AutoBuyTask{
		name:   "auto-buy",
		config: config,
	}

	// 解析基准金额配置
	if baseAmountRaw, exists := config["base_amount"]; exists {
		if baseAmount, ok := baseAmountRaw.(float64); ok {
			task.baseAmount = baseAmount
		} else {
			// 尝试从int转换
			if baseAmountInt, ok := baseAmountRaw.(int); ok {
				task.baseAmount = float64(baseAmountInt)
			} else {
				// 报错 yaml 中缺少参数
				return nil, fmt.Errorf("error, 配置中缺少 base_amount")

			}
		}
	} else {
		return nil, fmt.Errorf("error, 配置中缺少 base_amount")
	}

	// 解析AHR999倍数表配置
	if timerTableRaw, exists := config["ahr999_timer_table"]; exists {
		if timerTableStr, ok := timerTableRaw.(string); ok {
			// 解析JSON字符串格式的倍数表
			var timerTableMap map[string]interface{}
			if err := json.Unmarshal([]byte(timerTableStr), &timerTableMap); err != nil {
				return nil, fmt.Errorf("解析 ahr999_timer_table JSON 失败: %w", err)
			}

			task.ahr999TimerTable = make(Ahr999TimerTable)
			for rangeStr, multiplierRaw := range timerTableMap {
				var multiplier float64
				switch v := multiplierRaw.(type) {
				case float64:
					multiplier = v
				case int:
					multiplier = float64(v)
				default:
					return nil, fmt.Errorf("倍数配置格式错误: %v", multiplierRaw)
				}
				task.ahr999TimerTable[rangeStr] = multiplier
			}
		} else {
			return nil, fmt.Errorf("ahr999_timer_table 必须是字符串格式")
		}
	} else {
		return nil, fmt.Errorf("error, 配置中缺少 ahr999_timer_table")
	}

	task.pusher = pushAPI.NewPushAPI()
	task.pusher.Initialize(pushAPI.DefaultConfig(), pushAPI.WeChat)

	return task, nil
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
	investmentAmount, err := t.calculateInvestmentAmount(ahr999Value)
	if err != nil {
		return fmt.Errorf("计算定投金额失败: %w", err)
	}

	if debug {
		log.Printf("建议定投金额: $%.2f", investmentAmount)
	}

	// TODO: 这里可以添加实际的交易逻辑
	// 目前只是记录策略结果

	log.Printf("比特币定投策略 - 价格: $%.2f, AHR999: %.3f, 建议定投: $%.2f",
		currPrice, ahr999Value, investmentAmount)
	title := fmt.Sprintf("定投大饼 - $%.2f USDT", investmentAmount)
	content := fmt.Sprintf("价格: $%.2f\n\nAHR999: %.3f", currPrice, ahr999Value)
	// 推送消息
	t.pusher.PushNow(*pushAPI.NewNormalMessage("auto-buy", title, content), pushAPI.DefaultPushOptions())

	return nil
}

// calculateInvestmentAmount 根据AHR999指标计算定投金额
func (t *AutoBuyTask) calculateInvestmentAmount(ahr999 float64) (float64, error) {
	// 使用已解析的配置
	if len(t.ahr999TimerTable) == 0 {
		// 如果没有配置倍数表， 报错
		return 0.0, fmt.Errorf("没有配置倍数表")
	}

	// 使用新的计算逻辑
	return CalculateAmount(t.baseAmount, ahr999, t.ahr999TimerTable)
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
