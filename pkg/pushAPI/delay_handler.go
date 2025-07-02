package pushAPI

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DelayMessage 延迟消息结构
type DelayMessage struct {
	Message Message     `json:"message"`
	Options PushOptions `json:"options"`
}

// FileDelayHandler 文件延迟处理器
type FileDelayHandler struct {
	delayDir     string
	processedDir string
	pusher       Pusher
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
}

// NewFileDelayHandler 创建文件延迟处理器
func NewFileDelayHandler(delayDir, processedDir string, pusher Pusher) *FileDelayHandler {
	return &FileDelayHandler{
		delayDir:     delayDir,
		processedDir: processedDir,
		pusher:       pusher,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动延迟处理器
func (h *FileDelayHandler) Start() error {
	// 创建必要的目录
	if err := os.MkdirAll(h.delayDir, 0755); err != nil {
		return fmt.Errorf("创建延迟目录失败: %w", err)
	}
	if err := os.MkdirAll(h.processedDir, 0755); err != nil {
		return fmt.Errorf("创建已处理目录失败: %w", err)
	}

	h.wg.Add(1)
	go h.processLoop()
	return nil
}

// Stop 停止延迟处理器
func (h *FileDelayHandler) Stop() error {
	close(h.stopChan)
	h.wg.Wait()
	return nil
}

// processLoop 处理循环
func (h *FileDelayHandler) processLoop() {
	defer h.wg.Done()

	ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-h.stopChan:
			return
		case <-ticker.C:
			if err := h.ProcessDelayFiles(); err != nil {
				log.Printf("处理延迟文件失败: %v", err)
			}
		}
	}
}

// ProcessDelayFiles 处理延迟文件
func (h *FileDelayHandler) ProcessDelayFiles() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	entries, err := os.ReadDir(h.delayDir)
	if err != nil {
		return fmt.Errorf("读取延迟目录失败: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasPrefix(entry.Name(), "delay_") || !strings.HasSuffix(entry.Name(), ".msg") {
			continue
		}

		if err := h.processDelayFile(entry); err != nil {
			log.Printf("处理延迟文件 %s 失败: %v", entry.Name(), err)
		}
	}

	return nil
}

// processDelayFile 处理单个延迟文件
func (h *FileDelayHandler) processDelayFile(entry fs.DirEntry) error {
	filePath := filepath.Join(h.delayDir, entry.Name())

	// 尝试获取文件锁
	lockFile := filePath + ".lock"
	lock, err := os.Create(lockFile)
	if err != nil {
		return fmt.Errorf("创建锁文件失败: %w", err)
	}
	defer os.Remove(lockFile)
	defer lock.Close()

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 解析延迟消息
	var delayMsg DelayMessage
	if err := json.Unmarshal(data, &delayMsg); err != nil {
		return fmt.Errorf("解析延迟消息失败: %w", err)
	}

	// 推送消息
	if err := h.pusher.Push(delayMsg.Message); err != nil {
		return fmt.Errorf("推送消息失败: %w", err)
	}

	// 移动到已处理目录
	processedPath := filepath.Join(h.processedDir, entry.Name())
	if err := os.Rename(filePath, processedPath); err != nil {
		return fmt.Errorf("移动文件失败: %w", err)
	}

	log.Printf("延迟文件处理成功: %s", entry.Name())
	return nil
}

// WriteDelayFile 写入延迟文件
func (h *FileDelayHandler) WriteDelayFile(msg Message, options PushOptions) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 创建必要的目录
	if err := os.MkdirAll(h.delayDir, 0755); err != nil {
		return fmt.Errorf("创建延迟目录失败: %w", err)
	}

	// 生成文件名
	timestamp := time.Now().Format("20060102_1504")
	filename := fmt.Sprintf("delay_%s.msg", timestamp)
	filePath := filepath.Join(h.delayDir, filename)

	// 创建延迟消息
	delayMsg := DelayMessage{
		Message: msg,
		Options: options,
	}

	// 序列化为JSON
	data, err := json.Marshal(delayMsg)
	if err != nil {
		return fmt.Errorf("序列化延迟消息失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入延迟文件失败: %w", err)
	}

	log.Printf("延迟文件写入成功: %s", filename)
	return nil
}
