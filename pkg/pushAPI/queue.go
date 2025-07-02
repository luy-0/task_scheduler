package pushAPI

import (
	"sync"
)

// MemoryQueue 内存消息队列实现
type MemoryQueue struct {
	messages []Message
	mu       sync.RWMutex
	maxSize  int
}

// NewMemoryQueue 创建新的内存队列
func NewMemoryQueue(maxSize int) *MemoryQueue {
	return &MemoryQueue{
		messages: make([]Message, 0),
		maxSize:  maxSize,
	}
}

// Enqueue 入队
func (q *MemoryQueue) Enqueue(msg Message) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.messages) >= q.maxSize {
		// 队列已满，移除最旧的消息
		q.messages = q.messages[1:]
	}

	q.messages = append(q.messages, msg)
	return nil
}

// DequeueAll 出队所有消息
func (q *MemoryQueue) DequeueAll() ([]Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.messages) == 0 {
		return []Message{}, nil
	}

	messages := make([]Message, len(q.messages))
	copy(messages, q.messages)
	q.messages = q.messages[:0] // 清空队列

	return messages, nil
}

// Clear 清空队列
func (q *MemoryQueue) Clear() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.messages = q.messages[:0]
	return nil
}

// Size 获取队列大小
func (q *MemoryQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return len(q.messages)
}
