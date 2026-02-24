package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

const msgColumns = `id, company_id, sender_id, channel_id, receiver_id, content, msg_type,
	COALESCE(task_meta::text, 'null')::json as task_meta, created_at`

type messageRepo struct {
	db *gorm.DB
}

func NewMessageRepo(db *gorm.DB) MessageRepo {
	return &messageRepo{db: db}
}

func (r *messageRepo) Create(ctx context.Context, m *domain.Message) error {
	var createdAt time.Time
	result := r.db.WithContext(ctx).Raw(
		`INSERT INTO messages
		(id, company_id, sender_id, channel_id, receiver_id, content, msg_type, task_meta)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`,
		m.ID, m.CompanyID, m.SenderID, m.ChannelID, m.ReceiverID,
		m.Content, string(m.MsgType), m.TaskMeta).Scan(&createdAt)
	if result.Error != nil {
		return fmt.Errorf("message create: %w", result.Error)
	}
	m.CreatedAt = createdAt
	return nil
}

func (r *messageRepo) ListByChannel(ctx context.Context, channelID string, limit int, beforeID string) ([]*domain.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	var msgs []*domain.Message
	var result *gorm.DB
	if beforeID != "" {
		result = r.db.WithContext(ctx).Raw(
			`SELECT `+msgColumns+` FROM messages
			WHERE channel_id = $1
			  AND created_at < (SELECT created_at FROM messages WHERE id = $2)
			ORDER BY created_at DESC LIMIT $3`,
			channelID, beforeID, limit,
		).Scan(&msgs)
	} else {
		result = r.db.WithContext(ctx).Raw(
			`SELECT `+msgColumns+` FROM messages WHERE channel_id = $1 ORDER BY created_at DESC LIMIT $2`,
			channelID, limit,
		).Scan(&msgs)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("message list channel: %w", result.Error)
	}
	return msgs, nil
}

func (r *messageRepo) ListDM(ctx context.Context, agentA, agentB string, limit int, beforeID string) ([]*domain.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	var msgs []*domain.Message
	var result *gorm.DB
	if beforeID != "" {
		result = r.db.WithContext(ctx).Raw(
			`SELECT `+msgColumns+` FROM messages
			WHERE channel_id IS NULL
			  AND ((sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1))
			  AND created_at < (SELECT created_at FROM messages WHERE id = $3)
			ORDER BY created_at DESC LIMIT $4`,
			agentA, agentB, beforeID, limit,
		).Scan(&msgs)
	} else {
		result = r.db.WithContext(ctx).Raw(
			`SELECT `+msgColumns+` FROM messages
			WHERE channel_id IS NULL
			  AND ((sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1))
			ORDER BY created_at DESC LIMIT $3`,
			agentA, agentB, limit,
		).Scan(&msgs)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("message list dm: %w", result.Error)
	}
	return msgs, nil
}

func (r *messageRepo) MarkRead(ctx context.Context, agentID string, messageIDs []string) error {
	if len(messageIDs) == 0 {
		return nil
	}
	// 构建批量 INSERT ... ON CONFLICT DO NOTHING
	var sb strings.Builder
	args := make([]any, 0, len(messageIDs)+1)
	args = append(args, agentID) // $1 = agentID
	sb.WriteString("INSERT INTO message_reads (message_id, agent_id) VALUES ")
	for i, id := range messageIDs {
		if i > 0 {
			sb.WriteString(",")
		}
		fmt.Fprintf(&sb, "($%d, $1)", i+2)
		args = append(args, strings.TrimSpace(id))
	}
	sb.WriteString(" ON CONFLICT DO NOTHING")
	result := r.db.WithContext(ctx).Exec(sb.String(), args...)
	if result.Error != nil {
		return fmt.Errorf("message mark read: %w", result.Error)
	}
	return nil
}

func (r *messageRepo) ListUnreadForAgent(ctx context.Context, agentID, companyID string) ([]*domain.Message, error) {
	var msgs []*domain.Message
	result := r.db.WithContext(ctx).Raw(
		`WITH target AS (
			SELECT ?::uuid AS aid, ?::uuid AS cid,
			       (SELECT name FROM agents WHERE id = ?::uuid) AS aname
		)
		SELECT m.id, m.company_id, m.sender_id, m.channel_id, m.receiver_id,
		       m.content, m.msg_type,
		       COALESCE(m.task_meta::text, 'null')::json as task_meta,
		       m.created_at
		FROM messages m
		CROSS JOIN target t
		LEFT JOIN message_reads mr ON m.id = mr.message_id AND mr.agent_id = t.aid
		LEFT JOIN agents sender ON m.sender_id = sender.id
		WHERE m.company_id = t.cid
		  AND mr.message_id IS NULL
		  AND (
		      m.receiver_id = t.aid
		      OR (m.channel_id IS NOT NULL AND (
		          sender.id IS NULL
		          OR sender.is_human = true
		          OR m.content LIKE '%@' || t.aname || '%'
		      ))
		  )
		  AND (m.sender_id IS NULL OR m.sender_id != t.aid)
		ORDER BY m.created_at ASC`,
		agentID, companyID, agentID,
	).Scan(&msgs)
	if result.Error != nil {
		return nil, fmt.Errorf("message list unread: %w", result.Error)
	}
	return msgs, nil
}
