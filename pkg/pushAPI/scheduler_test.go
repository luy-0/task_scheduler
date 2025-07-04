package pushAPI

import (
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
