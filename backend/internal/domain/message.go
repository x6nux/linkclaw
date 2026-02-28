package domain

import (
	"encoding/json"
	"time"
)

type MsgType string

const (
	MsgTypeText       MsgType = "text"
	MsgTypeSystem     MsgType = "system"
	MsgTypeTaskUpdate MsgType = "task_update"
)

// TaskMeta 嵌入 task_update 消息，供前端渲染进度卡片
type TaskMeta struct {
	TaskID     string       `json:"task_id"`
	Title      string       `json:"title"`
	Status     TaskStatus   `json:"status"`
	Priority   TaskPriority `json:"priority"`
	AssigneeID *string      `json:"assignee_id,omitempty"`
	DueAt      *time.Time   `json:"due_at,omitempty"`
	Result     *string      `json:"result,omitempty"`
}

type Message struct {
	ID         string          `gorm:"column:id"          json:"id"`
	CompanyID  string          `gorm:"column:company_id"  json:"company_id"`
	SenderID   *string         `gorm:"column:sender_id"   json:"sender_id"`
	ChannelID  *string         `gorm:"column:channel_id"  json:"channel_id"`
	ReceiverID *string         `gorm:"column:receiver_id" json:"receiver_id"`
	Content    string          `gorm:"column:content"     json:"content"`
	MsgType    MsgType         `gorm:"column:msg_type"    json:"msg_type"`
	TaskMeta   json.RawMessage `gorm:"column:task_meta"   json:"task_meta"`
	CreatedAt  time.Time       `gorm:"column:created_at"  json:"created_at"`
}

// IsDM 判断是否为私信
func (m *Message) IsDM() bool {
	return m.ReceiverID != nil && m.ChannelID == nil
}
