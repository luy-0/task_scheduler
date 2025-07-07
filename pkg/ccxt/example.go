package ccxt

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

// ExampleUsage CCXT使用示例
func ExampleUsage() {
	fmt.Println("=== CCXT 交易所接口使用示例 ===")

	// 创建无需认证的客户端（用于公开接口）
	client := NewClientWithoutAuth("")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. 测试连通性
	fmt.Println("\n1. 测试交易所连通性")
	if err := client.Ping(ctx); err != nil {
		log.Printf("Ping失败: %v", err)
	} else {
		fmt.Println("✓ 交易所连通性正常")
	}

	// 2. 获取服务器时间
	fmt.Println("\n2. 获取服务器时间")
	serverTime, err := client.GetServerTime(ctx)
	if err != nil {
		log.Printf("获取服务器时间失败: %v", err)
	} else {
		fmt.Printf("✓ 服务器时间: %s\n", serverTime.Format("2006-01-02 15:04:05"))
	}

	// 3. 获取BTC最新价格
	fmt.Println("\n3. 获取BTC最新价格")
	btcPrice, err := client.GetBTCPrice(ctx)
	if err != nil {
		log.Printf("获取BTC价格失败: %v", err)
	} else {
		fmt.Printf("✓ BTC价格: $%.2f\n", btcPrice)
	}

	// 4. 获取ETH最新价格
	fmt.Println("\n4. 获取ETH最新价格")
	ethPrice, err := client.GetETHPrice(ctx)
	if err != nil {
		log.Printf("获取ETH价格失败: %v", err)
	} else {
		fmt.Printf("✓ ETH价格: $%.2f\n", ethPrice)
	}

	// 5. 获取BTC历史价格数据（用于AHR999计算）
	fmt.Println("\n5. 获取BTC历史价格数据")
	historyPrices, err := client.GetBTCHistoryPrices(ctx, 200) // 获取200天数据
	if err != nil {
		log.Printf("获取BTC历史价格失败: %v", err)
	} else {
		fmt.Printf("✓ 获取到 %d 条历史价格数据\n", len(historyPrices))

		// 显示最近几天的价格
		for i, priceData := range historyPrices {
			if i >= 30 { // 只显示最近30天
				break
			}
			timestamp := time.Unix(int64(priceData[0])/1000, 0)
			price := priceData[1]
			fmt.Printf("  %s: $%.2f\n", timestamp.Format("01-02"), price)
		}
	}

	// 6. 健康检查
	fmt.Println("\n6. 健康检查")
	if err := client.HealthCheck(ctx); err != nil {
		log.Printf("健康检查失败: %v", err)
	} else {
		fmt.Println("✓ 所有功能正常")
	}

	fmt.Println("\n=== 示例完成 ===")
}

func ExampleUsageWithAuth() {
	fmt.Println("\n=== ExampleUsageBuyCoin 示例开始 ===")
	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET_KEY")

	client := NewClient(apiKey, secretKey, "")
	err := client.Ping(context.Background())
	if err != nil {
		fmt.Println("Ping失败: ", err)
	} else {
		fmt.Println("Ping成功")
	}

	fmt.Println("\n=== 获取账户金额信息 ===")
	fmt.Println(client.GetAccountBalance(context.Background()))

	fmt.Println("\n=== 购买BTC ===")
	// fmt.Println(client.BuyCoinByMarketPrice(context.Background(), "BTCUSDT", 10))

	fmt.Println("\n=== 获取BTC订单盘口 ===")
	bestSellPrice, bestBuyPrice, _ := client.GetBestPrice(context.Background(), "BTCUSDT")
	fmt.Printf("✓ 最优卖价: $%.2f, 最优买价: $%.2f\n", bestSellPrice, bestBuyPrice)

	fmt.Println("\n=== 最优价购买BTC ===")
	// fmt.Println(client.BuyCoinByBestPrice(context.Background(), "BTCUSDT", 11))

	fmt.Println("\n=== 获取BTC余额 ===")
	fmt.Println(client.GetBTCBalance(context.Background()))

	fmt.Println("\n=== ExampleUsageBuyCoin 示例完成 ===")
}
