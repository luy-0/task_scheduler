package autobuy

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Ahr999TimerTable AHR999指标对应的倍数表
type Ahr999TimerTable map[string]float64

// CalculateAmount 根据AHR999指标计算定投金额
// baseAmount: 基准金额
// ahr999Value: AHR999指标值
// timerTable: AHR999区间对应的倍数表
func CalculateAmount(baseAmount float64, ahr999Value float64, timerTable Ahr999TimerTable) (float64, error) {
	amount, multiplier, rangeStr, err := GetRecommendedAmount(baseAmount, ahr999Value, timerTable)
	if err != nil {
		return 0, err
	}
	log.Printf("定投金额: %f, 倍数: %f, 区间: %s", amount, multiplier, rangeStr)
	return amount, nil
}

// findMultiplier 根据AHR999值查找对应的倍数
func findMultiplier(ahr999Value float64, timerTable Ahr999TimerTable) (float64, error) {
	// 按区间查找，支持以下格式：
	// "<0.4": 4
	// "0.4-0.6": 3
	// "0.6-0.8": 2
	// "0.8-1.2": 1
	// "1.2-1.4": 0.6
	// "1.4-1.6": 0.3
	// "1.6-1.8": 0.15

	for rangeStr, multiplier := range timerTable {
		if isInRange(ahr999Value, rangeStr) {
			return multiplier, nil
		}
	}

	// 如果没有找到匹配的区间，返回默认倍数1.0
	return 0, fmt.Errorf("没有找到匹配的区间，拒接执行")
}

// isInRange 判断AHR999值是否在指定区间内
func isInRange(ahr999Value float64, rangeStr string) bool {
	rangeStr = strings.TrimSpace(rangeStr)

	// 处理 "<0.4" 格式
	if strings.HasPrefix(rangeStr, "<") {
		valueStr := strings.TrimPrefix(rangeStr, "<")
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return false
		}
		return ahr999Value < value
	}

	// 处理 ">1.8" 格式
	if strings.HasPrefix(rangeStr, ">") {
		valueStr := strings.TrimPrefix(rangeStr, ">")
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return false
		}
		return ahr999Value > value
	}

	// 处理 "0.4-0.6" 格式
	if strings.Contains(rangeStr, "-") {
		parts := strings.Split(rangeStr, "-")
		if len(parts) != 2 {
			return false
		}

		minStr := strings.TrimSpace(parts[0])
		maxStr := strings.TrimSpace(parts[1])

		min, err := strconv.ParseFloat(minStr, 64)
		if err != nil {
			return false
		}

		max, err := strconv.ParseFloat(maxStr, 64)
		if err != nil {
			return false
		}

		return ahr999Value >= min && ahr999Value < max
	}

	// 处理单个值 "1.0" 格式
	value, err := strconv.ParseFloat(rangeStr, 64)
	if err != nil {
		return false
	}
	return ahr999Value == value
}

// ValidateTimerTable 验证倍数表配置的有效性
func ValidateTimerTable(timerTable Ahr999TimerTable) error {
	if len(timerTable) == 0 {
		return nil // 空表是有效的
	}

	for rangeStr, multiplier := range timerTable {
		// 验证倍数是否为正数
		if multiplier < 0 {
			return fmt.Errorf("倍数不能为负数，区间 %s 的倍数为 %f", rangeStr, multiplier)
		}

		// 验证区间格式
		if !isValidRangeFormat(rangeStr) {
			return fmt.Errorf("无效的区间格式: %s", rangeStr)
		}
	}

	return nil
}

// isValidRangeFormat 验证区间格式是否有效
func isValidRangeFormat(rangeStr string) bool {
	rangeStr = strings.TrimSpace(rangeStr)

	// 检查 "<0.4" 格式
	if strings.HasPrefix(rangeStr, "<") {
		valueStr := strings.TrimPrefix(rangeStr, "<")
		_, err := strconv.ParseFloat(valueStr, 64)
		return err == nil
	}

	// 检查 ">1.8" 格式
	if strings.HasPrefix(rangeStr, ">") {
		valueStr := strings.TrimPrefix(rangeStr, ">")
		_, err := strconv.ParseFloat(valueStr, 64)
		return err == nil
	}

	// 检查 "0.4-0.6" 格式
	if strings.Contains(rangeStr, "-") {
		parts := strings.Split(rangeStr, "-")
		if len(parts) != 2 {
			return false
		}

		minStr := strings.TrimSpace(parts[0])
		maxStr := strings.TrimSpace(parts[1])

		_, err1 := strconv.ParseFloat(minStr, 64)
		_, err2 := strconv.ParseFloat(maxStr, 64)

		return err1 == nil && err2 == nil
	}

	// 检查单个值 "1.0" 格式
	_, err := strconv.ParseFloat(rangeStr, 64)
	return err == nil
}

// GetRecommendedAmount 获取推荐的定投金额（包含详细信息）
func GetRecommendedAmount(baseAmount float64, ahr999Value float64, timerTable Ahr999TimerTable) (amount float64, multiplier float64, rangeStr string, err error) {
	if baseAmount <= 0 {
		return 0, 0, "", fmt.Errorf("基准金额必须大于0，当前值: %f", baseAmount)
	}

	if len(timerTable) == 0 {
		return 0, 0, "", fmt.Errorf("没有配置倍数表，拒接执行")
	}

	// 查找对应的区间和倍数
	for rangeStr, mult := range timerTable {
		if isInRange(ahr999Value, rangeStr) {
			amount = baseAmount * mult
			return amount, mult, rangeStr, nil
		}
	}

	// 如果没有找到匹配的区间，使用默认倍数
	return 0, 0, "", fmt.Errorf("没有找到匹配的区间，拒接执行")
}
