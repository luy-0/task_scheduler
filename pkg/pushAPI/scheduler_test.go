package pushAPI

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScheduledPush(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建推送API实例
	api := NewPushAPI()

	// 配置
	cfg := DefaultConfig()
	cfg.WorkingDir = tempDir
	cfg.FlushInterval = 1 * time.Second

	// 初始化
	if err := api.Initialize(cfg, Logger); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	defer func() {
		if impl, ok := api.(*PushAPIImpl); ok {
			impl.Stop()
		}
	}()

	// 创建测试消息
	message := NewNormalMessage("test_app", "定时测试消息", "这是一条定时测试消息")
	options := PushOptions{
		Receivers: []string{"user1"},
		Priority:  5,
		Retry:     2,
	}

	// 安排10秒后发送
	scheduledTime := time.Now().Add(10 * time.Second)
	if err := api.PushAt(*message, options, scheduledTime); err != nil {
		t.Fatalf("安排定时推送失败: %v", err)
	}

	// 检查文件是否创建
	currentFile := getCurrentTimeSlotFile(scheduledTime, tempDir)
	if _, err := os.Stat(currentFile); os.IsNotExist(err) {
		t.Errorf("定时消息文件应该存在: %s", currentFile)
	}

	// 等待一段时间让定时推送处理器处理
	time.Sleep(20 * time.Second)

	// 检查文件是否被清空（消息已发送）
	data, err := os.ReadFile(currentFile)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	// 打印文件内容用于调试
	t.Logf("文件内容: %s", string(data))

	// 文件应该只包含空数组或不存在
	if len(data) > 0 && string(data) != "[]" {
		t.Error("定时消息应该已被发送并从文件中移除")
	}
}

func TestTimeSlotFileNaming(t *testing.T) {
	// 测试不同时间段的文件名生成
	testCases := []struct {
		time     time.Time
		expected string
	}{
		{
			time:     time.Date(2024, 1, 1, 2, 30, 0, 0, time.UTC),
			expected: "scheduled_20240101_00.json", // 0-4点时间段
		},
		{
			time:     time.Date(2024, 1, 1, 6, 15, 0, 0, time.UTC),
			expected: "scheduled_20240101_04.json", // 4-8点时间段
		},
		{
			time:     time.Date(2024, 1, 1, 10, 45, 0, 0, time.UTC),
			expected: "scheduled_20240101_08.json", // 8-12点时间段
		},
	}

	for _, tc := range testCases {
		filename := getCurrentTimeSlotFile(tc.time, "./tmp/working")
		expectedPath := filepath.Join("./tmp/working", tc.expected)

		if filename != expectedPath {
			t.Errorf("时间 %v 期望文件名 %s，实际 %s", tc.time, expectedPath, filename)
		}
	}
}

// 辅助函数：获取当前时间段的文件名
func getCurrentTimeSlotFile(t time.Time, workingDir string) string {
	// 计算4小时时间段的开始时间
	hour := t.Hour()
	slotStart := hour - (hour % 4)

	// 生成文件名：scheduled_YYYYMMDD_HH.json
	dateStr := t.Format("20060102")
	timeStr := fmt.Sprintf("%02d", slotStart)

	return filepath.Join(workingDir, fmt.Sprintf("scheduled_%s_%s.json", dateStr, timeStr))
}

func TestScheduledPushShort(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建推送API实例
	api := NewPushAPI()

	// 配置
	cfg := DefaultConfig()
	cfg.WorkingDir = tempDir
	cfg.FlushInterval = 1 * time.Second

	// 初始化
	if err := api.Initialize(cfg, Logger); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	defer func() {
		if impl, ok := api.(*PushAPIImpl); ok {
			impl.Stop()
		}
	}()

	// 创建测试消息
	message := NewNormalMessage("test_app", "定时测试消息", "这是一条定时测试消息")
	options := PushOptions{
		Receivers: []string{"user1"},
		Priority:  5,
		Retry:     2,
	}

	// 安排下一个整分钟发送
	now := time.Now()
	scheduledTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
	if err := api.PushAt(*message, options, scheduledTime); err != nil {
		t.Fatalf("安排定时推送失败: %v", err)
	}

	// 检查文件是否创建
	currentFile := getCurrentTimeSlotFile(scheduledTime, tempDir)
	if _, err := os.Stat(currentFile); os.IsNotExist(err) {
		t.Errorf("定时消息文件应该存在: %s", currentFile)
	}

	// 等待70秒让定时推送处理器处理
	time.Sleep(70 * time.Second)

	// 检查文件是否被清空（消息已发送）
	data, err := os.ReadFile(currentFile)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	// 打印文件内容用于调试
	t.Logf("文件内容: %s", string(data))

	// 文件应该只包含空数组或不存在
	if len(data) > 0 && string(data) != "[]" {
		t.Error("定时消息应该已被发送并从文件中移除")
	}
}

func TestDelayMessageMergeAndCleanup(t *testing.T) {
	tempDir := t.TempDir()
	api := NewPushAPI()
	cfg := DefaultConfig()
	cfg.WorkingDir = tempDir
	cfg.FlushInterval = 1 * time.Second
	if err := api.Initialize(cfg, Logger); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	defer func() {
		if impl, ok := api.(*PushAPIImpl); ok {
			impl.Stop()
		}
	}()

	options := PushOptions{Receivers: []string{"user1"}, Priority: 1, Retry: 1}
	// 入队多条延迟消息
	for i := 0; i < 3; i++ {
		msg := NewNormalMessage("app1", fmt.Sprintf("延迟%d", i+1), fmt.Sprintf("内容%d", i+1))
		if err := api.Enqueue(*msg, options); err != nil {
			t.Fatalf("入队失败: %v", err)
		}
	}
	// 手动触发合并发送
	if err := api.FlushQueue(); err != nil {
		t.Fatalf("刷新队列失败: %v", err)
	}
	// 检查延迟消息文件是否被清空
	pattern := filepath.Join(tempDir, "delay_*.json")
	files, _ := filepath.Glob(pattern)
	for _, file := range files {
		data, _ := os.ReadFile(file)
		if len(data) > 0 && string(data) != "[]" {
			t.Errorf("延迟消息文件未被清空: %s", file)
		}
	}
}

func TestScheduledTriggerDelaySend(t *testing.T) {
	tempDir := t.TempDir()
	api := NewPushAPI()
	cfg := DefaultConfig()
	cfg.WorkingDir = tempDir
	cfg.FlushInterval = 1 * time.Second
	if err := api.Initialize(cfg, Logger); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	defer func() {
		if impl, ok := api.(*PushAPIImpl); ok {
			impl.Stop()
		}
	}()

	options := PushOptions{Receivers: []string{"user1"}, Priority: 1, Retry: 1}
	// 入队延迟消息
	msg := NewNormalMessage("app1", "延迟X", "内容X")
	if err := api.Enqueue(*msg, options); err != nil {
		t.Fatalf("入队失败: %v", err)
	}
	// 安排1分钟后定时推送
	now := time.Now()
	scheduledTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
	scheduledMsg := NewNormalMessage("app1", "定时触发", "定时内容")
	if err := api.PushAt(*scheduledMsg, options, scheduledTime); err != nil {
		t.Fatalf("定时推送失败: %v", err)
	}
	// 等待70秒让定时推送和延迟消息合并发送
	time.Sleep(70 * time.Second)
	// 检查延迟消息文件是否被清空
	pattern := filepath.Join(tempDir, "delay_*.json")
	files, _ := filepath.Glob(pattern)
	for _, file := range files {
		data, _ := os.ReadFile(file)
		if len(data) > 0 && string(data) != "[]" {
			t.Errorf("定时触发后延迟消息文件未被清空: %s", file)
		}
	}
}

func TestScheduledMessageExpirationCheck(t *testing.T) {
	tempDir := t.TempDir()
	api := NewPushAPI()
	cfg := DefaultConfig()
	cfg.WorkingDir = tempDir
	cfg.FlushInterval = 1 * time.Second
	if err := api.Initialize(cfg, Logger); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	defer func() {
		if impl, ok := api.(*PushAPIImpl); ok {
			impl.Stop()
		}
	}()

	options := PushOptions{Receivers: []string{"user1"}, Priority: 1, Retry: 1}

	// 安排多条定时消息，其中一些已经过期
	now := time.Now()

	// 1分钟后发送
	scheduledTime1 := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
	msg1 := NewNormalMessage("app1", "定时消息1", "内容1")
	if err := api.PushAt(*msg1, options, scheduledTime1); err != nil {
		t.Fatalf("安排定时推送失败: %v", err)
	}

	// 2分钟后发送
	scheduledTime2 := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+2, 0, 0, now.Location())
	msg2 := NewNormalMessage("app1", "定时消息2", "内容2")
	if err := api.PushAt(*msg2, options, scheduledTime2); err != nil {
		t.Fatalf("安排定时推送失败: %v", err)
	}

	// 等待70秒让定时推送处理器处理
	time.Sleep(70 * time.Second)

	// 检查定时消息文件是否被正确更新（过期的消息应该被移除）
	pattern := filepath.Join(tempDir, "scheduled_*.json")
	files, _ := filepath.Glob(pattern)
	for _, file := range files {
		data, _ := os.ReadFile(file)
		if len(data) > 0 {
			// 检查文件中是否还有未过期的消息
			var messages []interface{}
			if err := json.Unmarshal(data, &messages); err == nil {
				if len(messages) > 0 {
					t.Logf("文件中还有 %d 条未过期的定时消息", len(messages))
				}
			}
		}
	}
}

func TestHistoryRecordWithSendTime(t *testing.T) {
	tempDir := t.TempDir()
	historyDir := filepath.Join(tempDir, "history")
	api := NewPushAPI()
	cfg := DefaultConfig()
	cfg.WorkingDir = filepath.Join(tempDir, "working")
	cfg.HistoryDir = historyDir
	cfg.FlushInterval = 1 * time.Second
	if err := api.Initialize(cfg, Logger); err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	defer func() {
		if impl, ok := api.(*PushAPIImpl); ok {
			impl.Stop()
		}
	}()

	options := PushOptions{Receivers: []string{"user1"}, Priority: 1, Retry: 1}

	// 立即推送一条消息
	msg := NewNormalMessage("app1", "测试消息", "测试内容")
	if err := api.PushNow(*msg, options); err != nil {
		t.Fatalf("立即推送失败: %v", err)
	}

	// 等待一段时间让历史记录写入
	time.Sleep(2 * time.Second)

	// 检查历史记录文件是否存在
	monthStr := time.Now().Format("200601")
	successFile := filepath.Join(historyDir, fmt.Sprintf("success_send_%s.json", monthStr))

	if _, err := os.Stat(successFile); os.IsNotExist(err) {
		t.Errorf("成功发送历史记录文件应该存在: %s", successFile)
	} else {
		// 读取历史记录文件
		data, err := os.ReadFile(successFile)
		if err != nil {
			t.Fatalf("读取历史记录文件失败: %v", err)
		}

		var records []map[string]interface{}
		if err := json.Unmarshal(data, &records); err != nil {
			t.Fatalf("解析历史记录失败: %v", err)
		}

		if len(records) == 0 {
			t.Error("历史记录应该包含至少一条记录")
		} else {
			// 检查记录是否包含时间戳
			record := records[0]
			if _, exists := record["timestamp"]; !exists {
				t.Error("历史记录应该包含timestamp字段")
			}
			if _, exists := record["message_id"]; !exists {
				t.Error("历史记录应该包含message_id字段")
			}
			if _, exists := record["title"]; !exists {
				t.Error("历史记录应该包含title字段")
			}
		}
	}
}
