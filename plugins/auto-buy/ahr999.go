package autobuy

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	// AHR999公开数据API
	ahr999APIURL = "https://dncapi.flink1.com/api/v2/index/arh999?code=bitcoin&webp=1"
	// 缓存目录
	historyDir = "plugins/auto-buy/ahr999_history"
	// 月份文件模板
	monthFileTemplate = "%04d-%02d.json"
)

// Ahr999DataPoint 单条AHR999数据点
type Ahr999DataPoint struct {
	Date      string  `json:"date"`
	Timestamp int64   `json:"timestamp"`
	Ahr999    float64 `json:"ahr999"`
	BtcPrice  float64 `json:"btc_price"`
}

// GetAhr999 获取当前AHR999（优先缓存，无则API并写入）
func GetAhr999() (curr_btc_price, ahr999_value float64, err error) {
	today := time.Now().Format("2006-01-02")
	point, err := getAhr999FromCache(today)
	if err == nil && point != nil {
		return point.BtcPrice, point.Ahr999, nil
	}
	// 缓存无则查API
	points, err := fetchAhr999FromAPI()
	if err != nil {
		return 0, 0, err
	}
	// 写入新数据到缓存
	if err := updateAhr999Cache(points); err != nil {
		fmt.Printf("警告：写入AHR999缓存失败: %v\n", err)
	}
	// 再查一次缓存
	point, err = getAhr999FromCache(today)
	if err == nil && point != nil {
		return point.BtcPrice, point.Ahr999, nil
	}
	return 0, 0, fmt.Errorf("未获取到有效的AHR999数据")
}

// GetAhr999At 获取指定日期AHR999（优先缓存，无则API并写入）
func GetAhr999At(when time.Time) (btc_price, ahr999_value float64, err error) {
	dateStr := when.Format("2006-01-02")
	point, err := getAhr999FromCache(dateStr)
	if err == nil && point != nil {
		return point.BtcPrice, point.Ahr999, nil
	}
	// 缓存无则查API
	points, err := fetchAhr999FromAPI()
	if err != nil {
		return 0, 0, err
	}
	if err := updateAhr999Cache(points); err != nil {
		fmt.Printf("警告：写入AHR999缓存失败: %v\n", err)
	}
	point, err = getAhr999FromCache(dateStr)
	if err == nil && point != nil {
		return point.BtcPrice, point.Ahr999, nil
	}
	return 0, 0, fmt.Errorf("未找到%s的AHR999数据", dateStr)
}

// fetchAhr999FromAPI 获取API数据，返回每天最早一条，按日期分组
func fetchAhr999FromAPI() ([]Ahr999DataPoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ahr999APIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求API失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	var response struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data [][]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析API响应失败: %w", err)
	}
	if response.Code != 200 {
		return nil, fmt.Errorf("API返回错误: %s", response.Msg)
	}
	// 按天分组，选取每天最早一条
	dayMap := make(map[string]Ahr999DataPoint)
	for _, row := range response.Data {
		if len(row) >= 5 {
			timestamp, ok1 := row[0].(float64)
			ahr999, ok2 := row[1].(float64)
			btcPrice, ok3 := row[2].(float64)
			if ok1 && ok2 && ok3 {
				t := time.Unix(int64(timestamp), 0)
				date := t.Format("2006-01-02")
				point := Ahr999DataPoint{
					Date:      date,
					Timestamp: int64(timestamp),
					Ahr999:    ahr999,
					BtcPrice:  btcPrice,
				}
				// 只保留当天最早的
				if old, ok := dayMap[date]; !ok || point.Timestamp < old.Timestamp {
					dayMap[date] = point
				}
			}
		}
	}
	// 转为slice并按时间逆序
	var points []Ahr999DataPoint
	for _, v := range dayMap {
		points = append(points, v)
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].Timestamp > points[j].Timestamp
	})
	return points, nil
}

// updateAhr999Cache 按月写入新数据，每天只保留一条（最早的）
// 只记录 2024.01 以来的数据，之前的忽略
func updateAhr999Cache(points []Ahr999DataPoint) error {
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return err
	}
	// 按月分组
	monthMap := make(map[string][]Ahr999DataPoint)
	for _, p := range points {
		month := p.Date[:7] // yyyy-MM
		if month < "2024-01" {
			continue
		}
		monthMap[month] = append(monthMap[month], p)
	}
	for month, pts := range monthMap {
		filePath := filepath.Join(historyDir, month+".json")
		// 读取已有数据
		existing := make(map[string]Ahr999DataPoint)
		f, err := os.Open(filePath)
		if err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				var dp Ahr999DataPoint
				if err := json.Unmarshal(scanner.Bytes(), &dp); err == nil {
					existing[dp.Date] = dp
				}
			}
			f.Close()
		}
		// 合并新数据（只保留最早的）
		for _, p := range pts {
			if p.Date < "2024-01-01" {
				continue
			}
			if old, ok := existing[p.Date]; !ok || p.Timestamp < old.Timestamp {
				existing[p.Date] = p
			}
		}
		// 写回文件（按日期逆序）
		var all []Ahr999DataPoint
		for _, v := range existing {
			all = append(all, v)
		}
		sort.Slice(all, func(i, j int) bool {
			return all[i].Timestamp > all[j].Timestamp
		})
		f, err = os.Create(filePath)
		if err != nil {
			return err
		}
		w := bufio.NewWriter(f)
		for _, v := range all {
			b, _ := json.Marshal(v)
			w.WriteString(string(b) + "\n")
		}
		w.Flush()
		f.Close()
	}
	return nil
}

// getAhr999FromCache 按日期查找缓存
func getAhr999FromCache(date string) (*Ahr999DataPoint, error) {
	month := date[:7]
	filePath := filepath.Join(historyDir, month+".json")
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var dp Ahr999DataPoint
		if err := json.Unmarshal(scanner.Bytes(), &dp); err == nil {
			if dp.Date == date {
				return &dp, nil
			}
		}
	}
	return nil, fmt.Errorf("缓存中无%s的数据", date)
}
