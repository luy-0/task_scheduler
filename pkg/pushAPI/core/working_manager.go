package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"task_scheduler/pkg/pushAPI/base"
	"task_scheduler/pkg/pushAPI/push_method"
	"time"
)

// WorkingManager 工作目录管理器
type WorkingManager struct {
	workingDir     string
	pusher         push_method.IPusher
	historyHandler *HistoryHandler
	stopChan       chan struct{}
	wg             sync.WaitGroup
	mu             sync.Mutex
}

// NewWorkingManager 创建工作目录管理器
func NewWorkingManager(workingDir string, pusher push_method.IPusher, historyHandler *HistoryHandler) *WorkingManager {
	return &WorkingManager{
		workingDir:     workingDir,
		pusher:         pusher,
		historyHandler: historyHandler,
		stopChan:       make(chan struct{}),
	}
}

// Start 启动工作目录管理器
func (wm *WorkingManager) Start() error {
	if err := os.MkdirAll(wm.workingDir, 0755); err != nil {
		return fmt.Errorf("创建工作目录失败: %w", err)
	}

	wm.wg.Add(1)
	go wm.periodicSendLoop()
	return nil
}

// Stop 停止工作目录管理器
func (wm *WorkingManager) Stop() error {
	close(wm.stopChan)
	wm.wg.Wait()
	return nil
}

// periodicSendLoop 定期发送循环（每4小时检查一次）
func (wm *WorkingManager) periodicSendLoop() {
	defer wm.wg.Done()

	// 每4小时检查一次延迟消息
	delayTicker := time.NewTicker(4 * time.Hour)
	defer delayTicker.Stop()

	// 每分钟检查一次定时消息
	scheduledTicker := time.NewTicker(1 * time.Minute)
	defer scheduledTicker.Stop()

	for {
		select {
		case <-wm.stopChan:
			return
		case <-delayTicker.C:
			if err := wm.sendAllDelayMessages(); err != nil {
				log.Printf("定期发送延迟消息失败: %v", err)
			}
		case <-scheduledTicker.C:
			if err := wm.ProcessScheduledMessages(); err != nil {
				log.Printf("处理定时消息失败: %v", err)
			}
		}
	}
}

// AddDelayMessage 添加延迟消息
func (wm *WorkingManager) AddDelayMessage(msg base.Message, options base.PushOptions) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	delayMsg := &base.DelayMessage{
		Message:   msg,
		Options:   options,
		CreatedAt: time.Now(),
	}

	currentFile := wm.getDelayFileName(time.Now())
	existingMessages, err := wm.readDelayMessages(currentFile)
	if err != nil {
		return fmt.Errorf("读取现有延迟消息失败: %w", err)
	}

	existingMessages = append(existingMessages, delayMsg)

	if err := wm.writeDelayMessages(currentFile, existingMessages); err != nil {
		return fmt.Errorf("写入延迟消息失败: %w", err)
	}

	log.Printf("延迟消息已添加: %s", msg.ID)
	return nil
}

// AddScheduledMessage 添加定时消息
func (wm *WorkingManager) AddScheduledMessage(msg base.Message, options base.PushOptions, scheduledAt time.Time) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	scheduledMsg := &base.ScheduledMessage{
		Message:     msg,
		Options:     options,
		ScheduledAt: scheduledAt.Truncate(time.Minute),
	}

	targetFile := wm.getScheduledFileName(scheduledAt)
	existingMessages, err := wm.readScheduledMessages(targetFile)
	if err != nil {
		return fmt.Errorf("读取现有定时消息失败: %w", err)
	}

	existingMessages = append(existingMessages, scheduledMsg)

	if err := wm.writeScheduledMessages(targetFile, existingMessages); err != nil {
		return fmt.Errorf("写入定时消息失败: %w", err)
	}

	log.Printf("定时消息已安排: %s -> %s", msg.ID, scheduledAt.Format("2006-01-02 15:04"))
	return nil
}

// ProcessScheduledMessages 处理定时消息
func (wm *WorkingManager) ProcessScheduledMessages() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	now := time.Now().Truncate(time.Minute)
	currentFile := wm.getScheduledFileName(now)

	scheduledMessages, err := wm.readScheduledMessages(currentFile)
	if err != nil {
		return fmt.Errorf("读取定时消息失败: %w", err)
	}

	var remainingMessages []*base.ScheduledMessage
	var messagesToSend []*base.ScheduledMessage

	// 检查所有定时消息，找出需要发送的消息
	for _, scheduledMsg := range scheduledMessages {
		if scheduledMsg.ScheduledAt.Truncate(time.Minute).Before(now) || scheduledMsg.ScheduledAt.Truncate(time.Minute).Equal(now) {
			// 需要发送的消息
			messagesToSend = append(messagesToSend, scheduledMsg)
		} else {
			// 未到发送时间，保留消息
			remainingMessages = append(remainingMessages, scheduledMsg)
		}
	}

	// 如果有需要发送的定时消息
	if len(messagesToSend) > 0 {
		log.Printf("发现 %d 条过期的定时消息需要发送", len(messagesToSend))

		// 发送所有过期的定时消息
		for _, scheduledMsg := range messagesToSend {
			// 设置发送时间
			sentTime := time.Now()
			scheduledMsg.Message.SetSentAt(sentTime)
			scheduledMsg.Message.SetSendStatus(base.StatusSuccess)

			if err := wm.pusher.Push(scheduledMsg.Message); err != nil {
				// 设置失败状态
				scheduledMsg.Message.SetSendStatus(base.StatusFailed)
				// 记录发送失败
				if wm.historyHandler != nil {
					wm.historyHandler.RecordFailure(scheduledMsg.Message, wm.pusher.GetName(), scheduledMsg.Options, fmt.Sprintf("定时推送失败: %v", err))
				}
				log.Printf("定时消息发送失败: %v", err)
			} else {
				// 记录发送成功
				if wm.historyHandler != nil {
					wm.historyHandler.RecordSuccess(scheduledMsg.Message, wm.pusher.GetName(), scheduledMsg.Options)
				}
				log.Printf("定时消息发送成功: %s", scheduledMsg.Message.ID)
			}
		}

		// 发送定时消息后，检查并发送所有延迟消息
		if err := wm.sendAllDelayMessages(); err != nil {
			log.Printf("发送延迟消息失败: %v", err)
		}
	}

	// 更新文件内容
	if err := wm.writeScheduledMessages(currentFile, remainingMessages); err != nil {
		return fmt.Errorf("更新定时消息文件失败: %w", err)
	}

	return nil
}

// SendAllDelayMessages 发送所有延迟消息
func (wm *WorkingManager) SendAllDelayMessages() error {
	return wm.sendAllDelayMessages()
}

// sendAllDelayMessages 发送所有延迟消息（内部方法）
func (wm *WorkingManager) sendAllDelayMessages() error {
	delayFiles, err := wm.getAllDelayFiles()
	if err != nil {
		return fmt.Errorf("获取延迟消息文件失败: %w", err)
	}

	if len(delayFiles) == 0 {
		return nil
	}

	var allDelayMessages []*base.DelayMessage
	for _, file := range delayFiles {
		messages, err := wm.readDelayMessages(file)
		if err != nil {
			log.Printf("读取延迟消息文件失败 %s: %v", file, err)
			continue
		}
		allDelayMessages = append(allDelayMessages, messages...)
	}

	if len(allDelayMessages) == 0 {
		return nil
	}

	mergedMessage := wm.mergeDelayMessages(allDelayMessages)
	mergedOptions := wm.mergeDelayOptions(allDelayMessages)

	// 设置发送时间
	sentTime := time.Now()
	mergedMessage.SetSentAt(sentTime)
	mergedMessage.SetSendStatus(base.StatusSuccess)

	if err := wm.pusher.Push(mergedMessage); err != nil {
		// 设置失败状态
		mergedMessage.SetSendStatus(base.StatusFailed)
		if wm.historyHandler != nil {
			wm.historyHandler.RecordFailure(mergedMessage, wm.pusher.GetName(), mergedOptions, fmt.Sprintf("延迟消息推送失败: %v", err))
		}
		return fmt.Errorf("发送延迟消息失败: %w", err)
	}

	if wm.historyHandler != nil {
		wm.historyHandler.RecordSuccess(mergedMessage, wm.pusher.GetName(), mergedOptions)
	}

	log.Printf("延迟消息发送成功: %d条消息已合并发送", len(allDelayMessages))

	// 清空所有延迟消息文件
	for _, file := range delayFiles {
		if err := wm.writeDelayMessages(file, []*base.DelayMessage{}); err != nil {
			log.Printf("清空延迟消息文件失败 %s: %v", file, err)
		}
	}

	wm.cleanupOldDelayFiles()
	return nil
}

// mergeDelayMessages 合并延迟消息
func (wm *WorkingManager) mergeDelayMessages(messages []*base.DelayMessage) base.Message {
	if len(messages) == 0 {
		return base.Message{}
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})

	var titles []string
	var contents []string
	var appIDs []string

	for _, msg := range messages {
		titles = append(titles, msg.Message.Title)
		contents = append(contents, fmt.Sprintf("[%s] %s", msg.Message.Title, msg.Message.Content))
		appIDs = append(appIDs, msg.Message.AppID)
	}

	uniqueAppIDs := make(map[string]bool)
	for _, appID := range appIDs {
		uniqueAppIDs[appID] = true
	}

	var uniqueAppIDList []string
	for appID := range uniqueAppIDs {
		uniqueAppIDList = append(uniqueAppIDList, appID)
	}

	mergedMessage := base.NewMessage(
		strings.Join(uniqueAppIDList, ","),
		fmt.Sprintf("%d条延迟消息", len(messages)),
		strings.Join(contents, "\n"),
		base.Normal,
	)

	mergedMessage.SetMetadata("merged_count", len(messages))
	mergedMessage.SetMetadata("original_messages", len(messages))
	mergedMessage.SetMetadata("merge_time", time.Now())

	return *mergedMessage
}

// mergeDelayOptions 合并延迟消息的推送选项
func (wm *WorkingManager) mergeDelayOptions(messages []*base.DelayMessage) base.PushOptions {
	if len(messages) == 0 {
		return base.PushOptions{}
	}

	receiversMap := make(map[string]bool)
	var maxPriority int
	var maxRetry int

	for _, msg := range messages {
		for _, receiver := range msg.Options.Receivers {
			receiversMap[receiver] = true
		}
		if msg.Options.Priority > maxPriority {
			maxPriority = msg.Options.Priority
		}
		if msg.Options.Retry > maxRetry {
			maxRetry = msg.Options.Retry
		}
	}

	var receivers []string
	for receiver := range receiversMap {
		receivers = append(receivers, receiver)
	}

	return base.PushOptions{
		Receivers: receivers,
		Priority:  maxPriority,
		Retry:     maxRetry,
	}
}

// getDelayFileName 获取延迟消息文件名
func (wm *WorkingManager) getDelayFileName(t time.Time) string {
	hour := t.Hour()
	slotStart := hour - (hour % 4)
	dateStr := t.Format("20060102")
	timeStr := fmt.Sprintf("%02d", slotStart)
	return filepath.Join(wm.workingDir, fmt.Sprintf("delay_%s_%s.json", dateStr, timeStr))
}

// getScheduledFileName 获取定时消息文件名
func (wm *WorkingManager) getScheduledFileName(t time.Time) string {
	hour := t.Hour()
	slotStart := hour - (hour % 4)
	dateStr := t.Format("20060102")
	timeStr := fmt.Sprintf("%02d", slotStart)
	return filepath.Join(wm.workingDir, fmt.Sprintf("scheduled_%s_%s.json", dateStr, timeStr))
}

// getAllDelayFiles 获取所有延迟消息文件
func (wm *WorkingManager) getAllDelayFiles() ([]string, error) {
	pattern := filepath.Join(wm.workingDir, "delay_*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("查找延迟消息文件失败: %w", err)
	}
	return files, nil
}

// readDelayMessages 读取延迟消息文件
func (wm *WorkingManager) readDelayMessages(filePath string) ([]*base.DelayMessage, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*base.DelayMessage{}, nil
		}
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	var messages []*base.DelayMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("解析延迟消息失败: %w", err)
	}

	return messages, nil
}

// writeDelayMessages 写入延迟消息文件
func (wm *WorkingManager) writeDelayMessages(filePath string, messages []*base.DelayMessage) error {
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化延迟消息失败: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入延迟消息文件失败: %w", err)
	}

	return nil
}

// readScheduledMessages 读取定时消息文件
func (wm *WorkingManager) readScheduledMessages(filePath string) ([]*base.ScheduledMessage, error) {
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
func (wm *WorkingManager) writeScheduledMessages(filePath string, messages []*base.ScheduledMessage) error {
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化定时消息失败: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入定时消息文件失败: %w", err)
	}

	return nil
}

// cleanupOldDelayFiles 清理旧的延迟消息文件
func (wm *WorkingManager) cleanupOldDelayFiles() {
	files, err := wm.getAllDelayFiles()
	if err != nil {
		log.Printf("获取延迟消息文件失败: %v", err)
		return
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("读取文件失败 %s: %v", file, err)
			continue
		}

		if len(data) == 0 || string(data) == "[]" {
			if err := os.Remove(file); err != nil {
				log.Printf("删除空文件失败 %s: %v", file, err)
			} else {
				log.Printf("已删除空文件: %s", file)
			}
		}
	}
}
