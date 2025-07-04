package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"task_scheduler/pkg/pushAPI/base"
	"time"
)

// HistoryHandler 历史消息记录处理器
type HistoryHandler struct {
	historyDir string
	mu         sync.RWMutex
}

// NewHistoryHandler 创建历史消息记录处理器
func NewHistoryHandler(historyDir string) *HistoryHandler {
	return &HistoryHandler{
		historyDir: historyDir,
	}
}

// RecordSuccess 记录成功发送的消息
func (h *HistoryHandler) RecordSuccess(msg base.Message, pusherName string, options base.PushOptions) error {
	record := base.NewSuccessHistoryRecord(msg, pusherName, options)
	return h.writeRecord(record, "success_send")
}

// RecordFailure 记录发送失败的消息
func (h *HistoryHandler) RecordFailure(msg base.Message, pusherName string, options base.PushOptions, errorReason string) error {
	record := base.NewFailedHistoryRecord(msg, pusherName, options, errorReason)
	return h.writeRecord(record, "failed_send")
}

// writeRecord 写入历史记录到文件
func (h *HistoryHandler) writeRecord(record *base.HistoryRecord, recordType string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 创建历史记录目录
	if err := os.MkdirAll(h.historyDir, 0755); err != nil {
		return fmt.Errorf("创建历史记录目录失败: %w", err)
	}

	// 生成文件名：按月份组织
	monthStr := record.Timestamp.Format("200601") // YYYYMM格式
	filename := fmt.Sprintf("%s_%s.json", recordType, monthStr)
	filePath := filepath.Join(h.historyDir, filename)

	// 读取现有记录
	var records []*base.HistoryRecord
	if data, err := os.ReadFile(filePath); err == nil {
		if err := json.Unmarshal(data, &records); err != nil {
			// 如果解析失败，创建新的记录数组
			records = []*base.HistoryRecord{}
		}
	}

	// 添加新记录
	records = append(records, record)

	// 写入文件
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化历史记录失败: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入历史记录文件失败: %w", err)
	}

	return nil
}

// GetSuccessRecords 获取成功发送记录
func (h *HistoryHandler) GetSuccessRecords(yearMonth string) ([]*base.HistoryRecord, error) {
	return h.getRecords("success_send", yearMonth)
}

// GetFailedRecords 获取失败发送记录
func (h *HistoryHandler) GetFailedRecords(yearMonth string) ([]*base.HistoryRecord, error) {
	return h.getRecords("failed_send", yearMonth)
}

// getRecords 获取指定类型和月份的历史记录
func (h *HistoryHandler) getRecords(recordType, yearMonth string) ([]*base.HistoryRecord, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	filename := fmt.Sprintf("%s_%s.json", recordType, yearMonth)
	filePath := filepath.Join(h.historyDir, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*base.HistoryRecord{}, nil
		}
		return nil, fmt.Errorf("读取历史记录文件失败: %w", err)
	}

	var records []*base.HistoryRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("解析历史记录失败: %w", err)
	}

	return records, nil
}

// GetAvailableMonths 获取可用的历史记录月份
func (h *HistoryHandler) GetAvailableMonths() ([]string, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	entries, err := os.ReadDir(h.historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("读取历史记录目录失败: %w", err)
	}

	months := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 解析文件名获取月份
		// 格式：success_send_202401.json 或 failed_send_202401.json
		name := entry.Name()
		if len(name) >= 15 {
			var month string
			if name[:12] == "success_send_" {
				month = name[12:18] // success_send_ 后面取6位
			} else if name[:11] == "failed_send_" {
				month = name[11:17] // failed_send_ 后面取6位
			}
			if month != "" {
				months[month] = true
			}
		}
	}

	// 转换为切片
	result := make([]string, 0, len(months))
	for month := range months {
		result = append(result, month)
	}

	return result, nil
}

// CleanupOldRecords 清理旧的历史记录（保留指定月数）
func (h *HistoryHandler) CleanupOldRecords(keepMonths int) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	entries, err := os.ReadDir(h.historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取历史记录目录失败: %w", err)
	}

	cutoffTime := time.Now().AddDate(0, -keepMonths, 0)
	cutoffMonth := cutoffTime.Format("200601")

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if len(name) >= 15 {
			var month string
			if name[:12] == "success_send_" {
				month = name[12:18] // success_send_ 后面取6位
			} else if name[:11] == "failed_send_" {
				month = name[11:17] // failed_send_ 后面取6位
			}

			// 如果月份早于截止时间，删除文件
			if month != "" && month < cutoffMonth {
				filePath := filepath.Join(h.historyDir, name)
				if err := os.Remove(filePath); err != nil {
					return fmt.Errorf("删除旧历史记录文件失败: %w", err)
				}
			}
		}
	}

	return nil
}
