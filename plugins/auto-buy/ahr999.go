package autobuy

import (
	"fmt"
	"math"
	"time"
)

// GetAhr999 获取当前AHR999指标
func GetAhr999() (curr_btc_price, ahr999_value float64, err error) {
	// 1. 获取当前价格
	curr_btc_price, err = GetCurrentBitcoinPrice()
	if err != nil {
		return 0, 0, fmt.Errorf("获取当前比特币价格失败: %w", err)
	}

	// 2. 计算AHR999指标
	ahr999_value, err = calculateAhr999(curr_btc_price)
	if err != nil {
		return 0, 0, fmt.Errorf("计算AHR999指标失败: %w", err)
	}

	return curr_btc_price, ahr999_value, nil
}

// GetAhr999At 获取指定时间的AHR999指标
func GetAhr999At(when time.Time) (old_btc_price, old_ahr999_value float64, err error) {
	// 注意：这个函数需要历史数据，目前简化实现
	// 实际应用中可能需要查询历史数据库或缓存

	// 获取当前价格作为近似值（实际应该查询历史价格）
	old_btc_price, err = GetCurrentBitcoinPrice()
	if err != nil {
		return 0, 0, fmt.Errorf("获取历史比特币价格失败: %w", err)
	}

	// 计算历史AHR999指标
	old_ahr999_value, err = calculateAhr999AtTime(old_btc_price, when)
	if err != nil {
		return 0, 0, fmt.Errorf("计算历史AHR999指标失败: %w", err)
	}

	return old_btc_price, old_ahr999_value, nil
}

// calculateAhr999 计算AHR999指标
func calculateAhr999(currentPrice float64) (float64, error) {
	// 1. 获取时间范围
	now := time.Now()
	// before := now.AddDate(0, 0, -200) // 200天前 - 暂时未使用

	// 2. 获取比特币价格数据
	prices, err := GetBitcoinPriceHistory(200)
	if err != nil {
		return 0, fmt.Errorf("获取比特币价格历史数据失败: %w", err)
	}

	// 3. 计算200日均价（使用调和平均数）
	avg, err := calculateHarmonicMean(prices)
	if err != nil {
		return 0, fmt.Errorf("计算200日均价失败: %w", err)
	}

	// 4. 计算彩虹带模型的理论价格
	logprice, err := calculateRainbowPrice(now)
	if err != nil {
		return 0, fmt.Errorf("计算彩虹带价格失败: %w", err)
	}

	// 5. 计算AHR999指标
	ahr999 := (currentPrice / avg) * (currentPrice / logprice)

	// 保留3位小数
	ahr999 = math.Round(ahr999*1000) / 1000

	return ahr999, nil
}

// calculateAhr999AtTime 计算指定时间的AHR999指标
func calculateAhr999AtTime(priceAtTime float64, when time.Time) (float64, error) {
	// 计算指定时间200天前的日期
	// before := when.AddDate(0, 0, -200) // 暂时未使用

	// 获取历史价格数据（简化实现，实际应该查询历史数据）
	prices, err := GetBitcoinPriceHistory(200)
	if err != nil {
		return 0, fmt.Errorf("获取历史比特币价格数据失败: %w", err)
	}

	// 计算200日均价
	avg, err := calculateHarmonicMean(prices)
	if err != nil {
		return 0, fmt.Errorf("计算历史200日均价失败: %w", err)
	}

	// 计算彩虹带模型的理论价格
	logprice, err := calculateRainbowPrice(when)
	if err != nil {
		return 0, fmt.Errorf("计算历史彩虹带价格失败: %w", err)
	}

	// 计算AHR999指标
	ahr999 := (priceAtTime / avg) * (priceAtTime / logprice)

	// 保留3位小数
	ahr999 = math.Round(ahr999*1000) / 1000

	return ahr999, nil
}

// calculateHarmonicMean 计算调和平均数
func calculateHarmonicMean(prices [][]float64) (float64, error) {
	if len(prices) == 0 {
		return 0, fmt.Errorf("价格数据为空")
	}

	var sum float64
	count := 0

	for _, priceData := range prices {
		if len(priceData) >= 2 {
			price := priceData[1] // 价格在数组的第二个位置
			if price > 0 {
				sum += 1.0 / price
				count++
			}
		}
	}

	if count == 0 {
		return 0, fmt.Errorf("没有有效的价格数据")
	}

	// 调和平均数 = n / (1/x1 + 1/x2 + ... + 1/xn)
	harmonicMean := float64(count) / sum

	return harmonicMean, nil
}

// calculateRainbowPrice 计算彩虹带模型的理论价格
func calculateRainbowPrice(when time.Time) (float64, error) {
	// 比特币诞生时间：2009-01-03
	bitcoinBirth := time.Date(2009, 1, 3, 0, 0, 0, 0, time.UTC)

	// 计算比特币年龄（天数）
	duration := when.Sub(bitcoinBirth)
	age := duration.Hours() / 24.0

	if age <= 0 {
		return 0, fmt.Errorf("无效的比特币年龄")
	}

	// 彩虹带模型公式：logprice = 10^(5.84 * log10(age) - 17.01)
	logAge := math.Log10(age)
	logprice := math.Pow(10, 5.84*logAge-17.01)

	return logprice, nil
}
