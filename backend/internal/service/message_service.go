package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/event"
	"github.com/linkclaw/backend/internal/repository"
)

type MessageService struct {
	messageRepo repository.MessageRepo
	companyRepo repository.CompanyRepo
}

func NewMessageService(messageRepo repository.MessageRepo, companyRepo repository.CompanyRepo) *MessageService {
	return &MessageService{messageRepo: messageRepo, companyRepo: companyRepo}
}

type SendMessageInput = SendInput

type SendInput struct {
	CompanyID  string
	SenderID   string
	Channel    string // 群聊频道名，与 ReceiverID 二选一
	ReceiverID string // DM 目标，与 Channel 二选一
	Content    string
}

type MessageOut = domain.Message

func (s *MessageService) Send(ctx context.Context, in SendInput) (*domain.Message, error) {
	msg := &domain.Message{
		ID:        uuid.New().String(),
		CompanyID: in.CompanyID,
		Content:   in.Content,
		MsgType:   domain.MsgTypeText,
	}
	if in.SenderID != "" {
		msg.SenderID = &in.SenderID
	}

	switch {
	case in.Channel != "":
		ch, err := s.companyRepo.GetChannelByName(ctx, in.CompanyID, in.Channel)
		if err != nil {
			return nil, err
		}
		if ch == nil {
			return nil, fmt.Errorf("channel %q not found", in.Channel)
		}
		msg.ChannelID = &ch.ID

	case in.ReceiverID != "":
		if in.SenderID != "" && in.SenderID == in.ReceiverID {
			return nil, fmt.Errorf("cannot send message to yourself")
		}
		msg.ReceiverID = &in.ReceiverID

	default:
		return nil, fmt.Errorf("must specify channel or receiver_id")
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	// 发布新消息事件（供 WS Hub 实时推送给前端）
	var channelName *string
	if in.Channel != "" {
		channelName = &in.Channel
	}
	event.Global.Publish(event.NewEvent(event.MessageNew, event.MessageNewPayload{
		MessageID:   msg.ID,
		CompanyID:   msg.CompanyID,
		ChannelID:   msg.ChannelID,
		ChannelName: channelName,
		ReceiverID:  msg.ReceiverID,
		SenderID:    msg.SenderID,
		MsgType:     string(msg.MsgType),
		Content:     msg.Content,
		CreatedAt:   msg.CreatedAt.Format(time.RFC3339),
	}))
	return msg, nil
}

func (s *MessageService) GetChannelMessages(ctx context.Context, companyID, channelName string, limit int, beforeID string) ([]*domain.Message, error) {
	ch, err := s.companyRepo.GetChannelByName(ctx, companyID, channelName)
	if err != nil || ch == nil {
		return nil, fmt.Errorf("channel %q not found", channelName)
	}
	return s.messageRepo.ListByChannel(ctx, ch.ID, limit, beforeID)
}

func (s *MessageService) GetDMMessages(ctx context.Context, agentA, agentB string, limit int, beforeID string) ([]*domain.Message, error) {
	return s.messageRepo.ListDM(ctx, agentA, agentB, limit, beforeID)
}

func (s *MessageService) ListChannels(ctx context.Context, companyID string) ([]*domain.Channel, error) {
	return s.companyRepo.GetChannels(ctx, companyID)
}

func (s *MessageService) MarkRead(ctx context.Context, agentID string, messageIDs []string) error {
	return s.messageRepo.MarkRead(ctx, agentID, messageIDs)
}

func (s *MessageService) GetUnreadMessages(ctx context.Context, agentID, companyID string) ([]*domain.Message, error) {
	return s.messageRepo.ListUnreadForAgent(ctx, agentID, companyID)
}

// SendRaw 直接持久化并发布一条已构造好的消息（供 partner 注入使用）
func (s *MessageService) SendRaw(ctx context.Context, msg *domain.Message) error {
	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return err
	}
	event.Global.Publish(event.NewEvent(event.MessageNew, event.MessageNewPayload{
		MessageID:  msg.ID,
		CompanyID:  msg.CompanyID,
		ChannelID:  msg.ChannelID,
		ReceiverID: msg.ReceiverID,
		SenderID:   msg.SenderID,
		MsgType:    string(msg.MsgType),
		Content:    msg.Content,
		CreatedAt:  msg.CreatedAt.Format(time.RFC3339),
	}))
	return nil
}
