package autobuy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const apihost = "https://api.coingecko.com/api/v3"

// CoingeckoPriceResponse 价格响应结构
type CoingeckoPriceResponse struct {
	Bitcoin map[string]float64 `json:"bitcoin"`
}

// CoingeckoMarketChartResponse 市场图表响应结构
type CoingeckoMarketChartResponse struct {
	Prices [][]float64 `json:"prices"`
}

// CoingeckoPrice 获取当前价格
// 其中 id="bitcoin", vs_currencies="usd"
func CoingeckoPrice(id, vs_currencies string) (float64, error) {
	api := apihost + "/simple/price" + "?ids=" + id + "&vs_currencies=" + vs_currencies

	resp, err := http.Get(api)
	if err != nil {
		return 0, fmt.Errorf("请求价格API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %w", err)
	}

	var response CoingeckoPriceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("解析价格响应失败: %w", err)
	}

	if response.Bitcoin == nil {
		return 0, fmt.Errorf("未找到比特币价格数据")
	}

	price, exists := response.Bitcoin[vs_currencies]
	if !exists {
		return 0, fmt.Errorf("未找到%s价格数据", vs_currencies)
	}

	return price, nil
}

// CoingeckoMarketChartRange 获取过去一段历史数据
// 其中 from/to 格式均为 strconv.Itoa(int(time.Now().Unix()))
func CoingeckoMarketChartRange(id, vs_currency, from, to string) ([][]float64, error) {
	api := apihost + "/coins/" + id + "/market_chart/range" + "?vs_currency=" + vs_currency + "&from=" + from + "&to=" + to

	resp, err := http.Get(api)
	if err != nil {
		return nil, fmt.Errorf("请求市场图表API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var response CoingeckoMarketChartResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析市场图表响应失败: %w", err)
	}

	return response.Prices, nil
}

// GetCurrentBitcoinPrice 获取当前比特币价格
func GetCurrentBitcoinPrice() (float64, error) {
	return CoingeckoPrice("bitcoin", "usd")
}

// GetBitcoinPriceHistory 获取比特币价格历史数据
func GetBitcoinPriceHistory(days int) ([][]float64, error) {
	now := time.Now()
	from := now.AddDate(0, 0, -days)

	fromUnix := strconv.FormatInt(from.Unix(), 10)
	toUnix := strconv.FormatInt(now.Unix(), 10)

	return CoingeckoMarketChartRange("bitcoin", "usd", fromUnix, toUnix)
}
