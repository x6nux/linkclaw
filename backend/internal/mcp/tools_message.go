package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/service"
)

func (h *Handler) toolSendMessage(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Channel    string `json:"channel"`
		ReceiverID string `json:"receiver_id"`
		Content    string `json:"content"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Content == "" {
		return ErrorResult("参数错误：需要 content")
	}
	if p.Channel == "" && p.ReceiverID == "" {
		return ErrorResult("参数错误：需要指定 channel 或 receiver_id")
	}

	// receiver_id 支持传名字，自动解析为 ID
	dir := h.buildDirectory(ctx, sess.Agent.CompanyID)
	receiverID := p.ReceiverID
	if receiverID != "" {
		receiverID = dir.resolve(receiverID)
	}

	_, err := h.messageSvc.Send(ctx, service.SendMessageInput{
		CompanyID:  sess.Agent.CompanyID,
		SenderID:   sess.Agent.ID,
		Channel:    p.Channel,
		ReceiverID: receiverID,
		Content:    p.Content,
	})
	if err != nil {
		return ErrorResult("发送消息失败: " + err.Error())
	}

	target := "#" + p.Channel
	if receiverID != "" {
		if label, ok := dir.labels[receiverID]; ok {
			target = "私信 " + label
		} else {
			target = "私信 " + receiverID
		}
	}
	return TextResult(fmt.Sprintf("消息已发送到 %s", target))
}

func (h *Handler) toolGetMessages(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Channel    string `json:"channel"`
		ReceiverID string `json:"receiver_id"`
		Limit      string `json:"limit"`
		BeforeID   string `json:"before_id"`
	}
	json.Unmarshal(args, &p) //nolint:errcheck

	limit := 20
	if p.Limit != "" {
		if l, err := strconv.Atoi(p.Limit); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// 构建花名册，用于把 sender_id 显示为"职位-名字"
	dir := h.buildDirectory(ctx, sess.Agent.CompanyID)

	// receiver_id 支持按名字查找
	receiverID := p.ReceiverID
	if receiverID != "" {
		receiverID = dir.resolve(receiverID)
	}

	var msgs []string
	if p.Channel != "" {
		ms, err := h.messageSvc.GetChannelMessages(ctx, sess.Agent.CompanyID, p.Channel, limit, p.BeforeID)
		if err != nil {
			return ErrorResult(err.Error())
		}
		for _, m := range ms {
			sender := "系统"
			if m.SenderID != nil {
				if label, ok := dir.labels[*m.SenderID]; ok {
					sender = label
				} else {
					sender = *m.SenderID
				}
			}
			msgs = append(msgs, fmt.Sprintf("[%s] %s: %s",
				m.CreatedAt.Format("15:04"), sender, m.Content))
		}
	} else if receiverID != "" {
		ms, err := h.messageSvc.GetDMMessages(ctx, sess.Agent.ID, receiverID, limit, p.BeforeID)
		if err != nil {
			return ErrorResult(err.Error())
		}
		for _, m := range ms {
			sender := "对方"
			if m.SenderID != nil {
				if *m.SenderID == sess.Agent.ID {
					sender = "你"
				} else if label, ok := dir.labels[*m.SenderID]; ok {
					sender = label
				}
			}
			msgs = append(msgs, fmt.Sprintf("[%s] %s: %s",
				m.CreatedAt.Format("15:04"), sender, m.Content))
		}
	} else {
		return ErrorResult("需要指定 channel 或 receiver_id")
	}

	if len(msgs) == 0 {
		return TextResult("暂无消息")
	}
	return TextResult(strings.Join(msgs, "\n"))
}

func (h *Handler) toolMarkMessagesRead(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		MessageIDs string `json:"message_ids"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.MessageIDs == "" {
		return ErrorResult("参数错误：需要 message_ids")
	}
	raw := strings.Split(p.MessageIDs, ",")
	ids := make([]string, 0, len(raw))
	for _, id := range raw {
		id = strings.TrimSpace(id)
		if id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return ErrorResult("参数错误：message_ids 为空")
	}
	if err := h.messageSvc.MarkRead(ctx, sess.Agent.ID, ids); err != nil {
		return ErrorResult("标记已读失败: " + err.Error())
	}
	return TextResult(fmt.Sprintf("已标记 %d 条消息为已读", len(ids)))
}

// agentDirectory 同公司 Agent 花名册，支持多种方式查找
type agentDirectory struct {
	labels  map[string]string // id → "职位-名字"
	byName  map[string]string // 名字 → id
	byLabel map[string]string // "职位-名字" → id
}

func (h *Handler) buildDirectory(ctx context.Context, companyID string) *agentDirectory {
	agents, err := h.agentSvc.ListByCompany(ctx, companyID)
	if err != nil {
		return &agentDirectory{
			labels: map[string]string{}, byName: map[string]string{}, byLabel: map[string]string{},
		}
	}
	dir := &agentDirectory{
		labels:  make(map[string]string, len(agents)),
		byName:  make(map[string]string, len(agents)),
		byLabel: make(map[string]string, len(agents)),
	}
	for _, a := range agents {
		meta, ok := domain.PositionMetaByPosition[a.Position]
		label := a.Name
		if ok {
			label = meta.DisplayName + "-" + a.Name
		}
		dir.labels[a.ID] = label
		dir.byName[a.Name] = a.ID
		dir.byLabel[label] = a.ID
	}
	return dir
}

// resolveAgentID 将 receiver（名字、"职位-名字"、或 ID）解析为 agent ID
func (d *agentDirectory) resolve(receiver string) string {
	if _, ok := d.labels[receiver]; ok {
		return receiver // 已经是有效 ID
	}
	if id, ok := d.byName[receiver]; ok {
		return id
	}
	if id, ok := d.byLabel[receiver]; ok {
		return id
	}
	return receiver // 原样返回，让下游报错
}

func (h *Handler) toolListChannels(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	channels, err := h.messageSvc.ListChannels(ctx, sess.Agent.CompanyID)
	if err != nil {
		return ErrorResult("获取频道列表失败: " + err.Error())
	}
	if len(channels) == 0 {
		return TextResult("暂无频道")
	}
	var lines []string
	for _, ch := range channels {
		defaultMark := ""
		if ch.IsDefault {
			defaultMark = "（默认）"
		}
		lines = append(lines, fmt.Sprintf("  #%s%s — %s", ch.Name, defaultMark, ch.Description))
	}
	return TextResult("频道列表：\n" + strings.Join(lines, "\n"))
}
