package autobuy

import (
	"log"
	"time"
)

func ExampleUsage() {
	exampleGetAhr999()
	// exampleGetAhr999At()
	// exampleCalculateAmount()
}

func exampleGetAhr999() {
	curr_btc_price, ahr999_value, err := GetAhr999()
	if err != nil {
		log.Fatalf("获取AHR999失败: %v", err)
	}
	log.Printf("当前比特币价格: %f, AHR999指标: %f", curr_btc_price, ahr999_value)
}

func exampleGetAhr999At() {
	test_day_1 := "2025-07-05"
	test_day_2 := "2025-05-01"
	test_day_3 := "2025-04-01"
	test_day_4 := "2025-01-01"

	for _, test_day := range []string{test_day_1, test_day_2, test_day_3, test_day_4} {
		when, err := time.Parse("2006-01-02", test_day)
		if err != nil {
			log.Fatalf("解析日期失败: %v", err)
		}
		curr_btc_price, ahr999_value, err := GetAhr999At(when)
		if err != nil {
			log.Fatalf("获取AHR999失败: %v", err)
		}
		log.Printf("日期: %s, 当前比特币价格: %f, AHR999指标: %f", test_day, curr_btc_price, ahr999_value)
	}
}

func exampleCalculateAmount() {
	log.Println("=== 定投金额计算示例 ===")

	// 示例1: 使用配置的倍数表
	log.Println("1. 使用配置的倍数表计算定投金额:")

	// 模拟配置文件中的倍数表
	timerTable := Ahr999TimerTable{
		"<0.4":    4,
		"0.4-0.6": 3,
		"0.6-0.8": 2,
		"0.8-1.2": 1,
		"1.2-1.4": 0.6,
		"1.4-1.6": 0.3,
		"1.6-1.8": 0.15,
	}

	baseAmount := 1000.0

	// 测试不同AHR999值的定投金额
	testAhr999Values := []float64{0.3, 0.5, 0.7, 1.0, 1.3, 1.5, 1.7, 2.0}

	for _, ahr999 := range testAhr999Values {
		_, err := CalculateAmount(baseAmount, ahr999, timerTable)
		if err != nil {
			log.Printf("计算失败 AHR999=%.1f: %v", ahr999, err)
			continue
		}

		// 获取详细信息
		recommendedAmount, multiplier, rangeStr, _ := GetRecommendedAmount(baseAmount, ahr999, timerTable)

		log.Printf("AHR999=%.1f | 区间: %s | 倍数: %.2f | 定投金额: $%.2f",
			ahr999, rangeStr, multiplier, recommendedAmount)
	}

	// 示例2: 验证倍数表配置
	log.Println("\n2. 验证倍数表配置:")

	err := ValidateTimerTable(timerTable)
	if err != nil {
		log.Printf("倍数表验证失败: %v", err)
	} else {
		log.Println("倍数表配置有效 ✓")
	}

	// 示例3: 测试无效配置
	log.Println("\n3. 测试无效配置:")

	invalidTable := Ahr999TimerTable{
		"invalid-range": 1,
		"0.4-0.6":       -1, // 负数倍数
	}

	err = ValidateTimerTable(invalidTable)
	if err != nil {
		log.Printf("正确捕获无效配置: %v ✓", err)
	} else {
		log.Println("应该捕获无效配置但未捕获 ✗")
	}

	// 示例4: 空倍数表的情况
	log.Println("\n4. 空倍数表的情况:")

	emptyTable := Ahr999TimerTable{}
	emptyAmount, err := CalculateAmount(baseAmount, 0.5, emptyTable)
	if err != nil {
		log.Printf("空倍数表计算失败: %v", err)
	} else {
		log.Printf("空倍数表时，定投金额等于基准金额: $%.2f ✓", emptyAmount)
	}

	// 示例5: 不同基准金额的对比
	log.Println("\n5. 不同基准金额的对比:")

	differentBaseAmounts := []float64{500, 1000, 2000}
	ahr999Value := 0.5 // 严重低估

	for _, baseAmount := range differentBaseAmounts {
		amount, err := CalculateAmount(baseAmount, ahr999Value, timerTable)
		if err != nil {
			log.Printf("计算失败 基准金额=%.0f: %v", baseAmount, err)
			continue
		}

		log.Printf("基准金额: $%.0f | AHR999=%.1f | 定投金额: $%.2f",
			baseAmount, ahr999Value, amount)
	}

	log.Println("\n=== 定投金额计算示例完成 ===")
}
