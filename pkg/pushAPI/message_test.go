package pushAPI

import (
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	// 测试创建消息
	message := NewMessage("app1", "测试标题", "测试内容", Emergency)

	// 验证基本字段
	if message.AppID != "app1" {
		t.Errorf("期望AppID为'app1'，实际为'%s'", message.AppID)
	}

	if message.Title != "测试标题" {
		t.Errorf("期望Title为'测试标题'，实际为'%s'", message.Title)
	}

	if message.Content != "测试内容" {
		t.Errorf("期望Content为'测试内容'，实际为'%s'", message.Content)
	}

	if message.Level != Emergency {
		t.Errorf("期望Level为Emergency，实际为%v", message.Level)
	}

	// 验证ID格式
	if message.ID == "" {
		t.Error("消息ID不能为空")
	}

	// 验证时间字段
	if message.CreatedAt.IsZero() {
		t.Error("创建时间不能为零值")
	}

	// 验证状态
	if message.SendStatus != StatusInitialized {
		t.Errorf("期望初始状态为StatusInitialized，实际为%v", message.SendStatus)
	}
}

func TestNewNormalMessage(t *testing.T) {
	// 测试使用默认级别创建消息
	message := NewNormalMessage("app2", "默认标题", "默认内容")

	if message.Level != Normal {
		t.Errorf("期望默认级别为Normal，实际为%v", message.Level)
	}
}

func TestMessageIDGeneration(t *testing.T) {
	// 测试ID生成格式
	message1 := NewMessage("app1", "标题1", "内容1", Normal)
	message2 := NewMessage("app1", "标题2", "内容2", Normal)

	// 验证ID格式：{app_id}_YYMMDD_{gen_id}
	if len(message1.ID) == 0 {
		t.Error("消息ID不能为空")
	}

	// 验证ID不重复
	if message1.ID == message2.ID {
		t.Error("不同消息的ID应该不同")
	}

	// 验证ID格式
	expectedPrefix := "app1_"
	if len(message1.ID) < len(expectedPrefix) || message1.ID[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("ID应该以'%s'开头，实际为'%s'", expectedPrefix, message1.ID)
	}
}

func TestSetMetadata(t *testing.T) {
	message := NewNormalMessage("app1", "标题", "内容")

	// 测试设置元数据
	message.SetMetadata("key1", "value1")
	message.SetMetadata("key2", 123)

	// 测试获取元数据
	if value, exists := message.GetMetadata("key1"); !exists || value != "value1" {
		t.Errorf("期望获取到'value1'，实际为%v", value)
	}

	if value, exists := message.GetMetadata("key2"); !exists || value != 123 {
		t.Errorf("期望获取到123，实际为%v", value)
	}

	// 测试不存在的键
	if value, exists := message.GetMetadata("nonexistent"); exists {
		t.Errorf("不存在的键应该返回false，实际为%v", value)
	}
}

func TestSetSendStatus(t *testing.T) {
	message := NewNormalMessage("app1", "标题", "内容")

	// 测试状态变化
	message.SetSendStatus(StatusPending)
	if message.SendStatus != StatusPending {
		t.Errorf("期望状态为StatusPending，实际为%v", message.SendStatus)
	}

	message.SetSendStatus(StatusSuccess)
	if message.SendStatus != StatusSuccess {
		t.Errorf("期望状态为StatusSuccess，实际为%v", message.SendStatus)
	}
}

func TestSetSentAt(t *testing.T) {
	message := NewNormalMessage("app1", "标题", "内容")

	// 测试设置发送时间
	sentTime := time.Now()
	message.SetSentAt(sentTime)

	if !message.SentAt.Equal(sentTime) {
		t.Errorf("期望发送时间为%v，实际为%v", sentTime, message.SentAt)
	}
}

func TestMessageLevelString(t *testing.T) {
	// 测试消息级别字符串表示
	if Normal.String() != "normal" {
		t.Errorf("期望Normal.String()返回'normal'，实际为'%s'", Normal.String())
	}

	if Emergency.String() != "emergency" {
		t.Errorf("期望Emergency.String()返回'emergency'，实际为'%s'", Emergency.String())
	}
}

func TestSendStatusString(t *testing.T) {
	// 测试发送状态字符串表示
	if StatusInitialized.String() != "initialized" {
		t.Errorf("期望StatusInitialized.String()返回'initialized'，实际为'%s'", StatusInitialized.String())
	}

	if StatusPending.String() != "pending" {
		t.Errorf("期望StatusPending.String()返回'pending'，实际为'%s'", StatusPending.String())
	}

	if StatusSuccess.String() != "success" {
		t.Errorf("期望StatusSuccess.String()返回'success'，实际为'%s'", StatusSuccess.String())
	}

	if StatusFailed.String() != "failed" {
		t.Errorf("期望StatusFailed.String()返回'failed'，实际为'%s'", StatusFailed.String())
	}
}
