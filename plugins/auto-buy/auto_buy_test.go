package autobuy

import (
	"testing"
	"time"
)

func TestGetCurrentBitcoinPrice(t *testing.T) {
	price, err := GetCurrentBitcoinPrice()
	if err != nil {
		t.Fatalf("获取比特币价格失败: %v", err)
	}

	if price <= 0 {
		t.Errorf("比特币价格应该大于0，实际: %f", price)
	}

	t.Logf("当前比特币价格: $%.2f", price)
}

func TestGetBitcoinPriceHistory(t *testing.T) {
	prices, err := GetBitcoinPriceHistory(7) // 获取7天历史数据
	if err != nil {
		t.Fatalf("获取比特币价格历史失败: %v", err)
	}

	if len(prices) == 0 {
		t.Errorf("历史价格数据不应该为空")
	}

	t.Logf("获取到 %d 条历史价格数据", len(prices))

	// 检查数据格式
	for i, priceData := range prices {
		if len(priceData) < 2 {
			t.Errorf("价格数据格式错误，索引 %d: %v", i, priceData)
		}
		if priceData[1] <= 0 {
			t.Errorf("价格应该大于0，索引 %d: %f", i, priceData[1])
		}
	}
}

func TestGetAhr999(t *testing.T) {
	price, ahr999, err := GetAhr999()
	if err != nil {
		t.Fatalf("获取AHR999指标失败: %v", err)
	}

	if price <= 0 {
		t.Errorf("比特币价格应该大于0，实际: %f", price)
	}

	if ahr999 < 0 {
		t.Errorf("AHR999指标不应该为负数，实际: %f", ahr999)
	}

	t.Logf("当前比特币价格: $%.2f", price)
	t.Logf("当前AHR999指标: %.3f", ahr999)
}

func TestCalculateRainbowPrice(t *testing.T) {
	now := time.Now()
	logprice, err := calculateRainbowPrice(now)
	if err != nil {
		t.Fatalf("计算彩虹带价格失败: %v", err)
	}

	if logprice <= 0 {
		t.Errorf("彩虹带价格应该大于0，实际: %f", logprice)
	}

	t.Logf("彩虹带理论价格: $%.2f", logprice)
}

func TestCalculateHarmonicMean(t *testing.T) {
	// 测试数据
	testPrices := [][]float64{
		{1640995200000, 100.0}, // 时间戳, 价格
		{1641081600000, 200.0},
		{1641168000000, 150.0},
	}

	harmonicMean, err := calculateHarmonicMean(testPrices)
	if err != nil {
		t.Fatalf("计算调和平均数失败: %v", err)
	}

	// 调和平均数应该小于算术平均数
	arithmeticMean := (100.0 + 200.0 + 150.0) / 3.0
	if harmonicMean >= arithmeticMean {
		t.Errorf("调和平均数应该小于算术平均数，调和平均: %f, 算术平均: %f", harmonicMean, arithmeticMean)
	}

	t.Logf("调和平均数: %.2f", harmonicMean)
	t.Logf("算术平均数: %.2f", arithmeticMean)
}
