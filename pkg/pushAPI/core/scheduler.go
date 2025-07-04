package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"task_scheduler/pkg/pushAPI/base"
	"task_scheduler/pkg/pushAPI/push_method"
	"time"
)

// Scheduler 定时推送处理器
type Scheduler struct {
	workingDir     string
	pusher         push_method.IPusher
	historyHandler *HistoryHandler
	stopChan       chan struct{}
	wg             sync.WaitGroup
	mu             sync.Mutex
}

// NewScheduler 创建定时推送处理器
func NewScheduler(workingDir string, pusher push_method.IPusher, historyHandler *HistoryHandler) *Scheduler {
	return &Scheduler{
		workingDir:     workingDir,
		pusher:         pusher,
		historyHandler: historyHandler,
		stopChan:       make(chan struct{}),
	}
}

// Start 启动定时推送处理器
func (s *Scheduler) Start() error {
	// 创建必要的目录
	if err := os.MkdirAll(s.workingDir, 0755); err != nil {
		return fmt.Errorf("创建工作目录失败: %w", err)
	}

	s.wg.Add(1)
	go s.scheduleLoop()
	return nil
}

// Stop 停止定时推送处理器
func (s *Scheduler) Stop() error {
	close(s.stopChan)
	s.wg.Wait()
	return nil
}

// scheduleLoop 定时推送循环
func (s *Scheduler) scheduleLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.processScheduledMessages(); err != nil {
				log.Printf("处理定时消息失败: %v", err)
			}
		}
	}
}

// processScheduledMessages 处理定时消息
func (s *Scheduler) processScheduledMessages() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取当前时间（精确到分钟）
	now := time.Now().Truncate(time.Minute)

	// 获取当前4小时时间段的文件名
	currentFile := s.getCurrentTimeSlotFile(now)

	// 读取当前时间段的定时消息
	scheduledMessages, err := s.readScheduledMessages(currentFile)
	if err != nil {
		return fmt.Errorf("读取定时消息失败: %w", err)
	}

	// 处理到期的消息
	var remainingMessages []*base.ScheduledMessage
	for _, scheduledMsg := range scheduledMessages {
		// 检查是否到了发送时间
		if scheduledMsg.ScheduledAt.Truncate(time.Minute).Before(now) {
			// 发送消息
			if err := s.pusher.Push(scheduledMsg.Message); err != nil {
				// 记录发送失败
				if s.historyHandler != nil {
					s.historyHandler.RecordFailure(scheduledMsg.Message, s.pusher.GetName(), scheduledMsg.Options, fmt.Sprintf("定时推送失败: %v", err))
				}
				log.Printf("定时消息发送失败: %v", err)
			} else {
				// 记录发送成功
				if s.historyHandler != nil {
					s.historyHandler.RecordSuccess(scheduledMsg.Message, s.pusher.GetName(), scheduledMsg.Options)
				}
				log.Printf("定时消息发送成功: %s", scheduledMsg.Message.ID)
			}
		} else {
			// 未到发送时间，保留消息
			remainingMessages = append(remainingMessages, scheduledMsg)
		}
	}

	// 更新文件内容
	if err := s.writeScheduledMessages(currentFile, remainingMessages); err != nil {
		return fmt.Errorf("更新定时消息文件失败: %w", err)
	}

	return nil
}

// ScheduleMessage 安排定时消息
func (s *Scheduler) ScheduleMessage(msg base.Message, options base.PushOptions, scheduledAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建必要的目录
	if err := os.MkdirAll(s.workingDir, 0755); err != nil {
		return fmt.Errorf("创建工作目录失败: %w", err)
	}

	// 创建定时消息
	scheduledMsg := &base.ScheduledMessage{
		Message:     msg,
		Options:     options,
		ScheduledAt: scheduledAt, // 精确到分钟
	}

	// 获取目标时间段的文件名
	targetFile := s.getCurrentTimeSlotFile(scheduledAt)

	// 读取现有消息
	existingMessages, err := s.readScheduledMessages(targetFile)
	if err != nil {
		return fmt.Errorf("读取现有定时消息失败: %w", err)
	}

	// 添加新消息
	existingMessages = append(existingMessages, scheduledMsg)

	// 写入文件
	if err := s.writeScheduledMessages(targetFile, existingMessages); err != nil {
		return fmt.Errorf("写入定时消息失败: %w", err)
	}

	log.Printf("定时消息已安排: %s -> %s", msg.ID, scheduledAt.Format("2006-01-02 15:04"))
	return nil
}

// getCurrentTimeSlotFile 获取当前4小时时间段的文件名
func (s *Scheduler) getCurrentTimeSlotFile(t time.Time) string {
	// 计算4小时时间段的开始时间
	hour := t.Hour()
	slotStart := hour - (hour % 4)

	// 生成文件名：scheduled_YYYYMMDD_HH.json
	dateStr := t.Format("20060102")
	timeStr := fmt.Sprintf("%02d", slotStart)

	return filepath.Join(s.workingDir, fmt.Sprintf("scheduled_%s_%s.json", dateStr, timeStr))
}

// readScheduledMessages 读取定时消息文件
func (s *Scheduler) readScheduledMessages(filePath string) ([]*base.ScheduledMessage, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*base.ScheduledMessage{}, nil
		}
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	var messages []*base.ScheduledMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("解析定时消息失败: %w", err)
	}

	return messages, nil
}

// writeScheduledMessages 写入定时消息文件
func (s *Scheduler) writeScheduledMessages(filePath string, messages []*base.ScheduledMessage) error {
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化定时消息失败: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入定时消息文件失败: %w", err)
	}

	return nil
}
