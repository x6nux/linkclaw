package ws

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/event"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

// agentConns 全局 Agent 连接注册表，同一 Agent 只保留最新连接
var agentConns sync.Map // map[agentID]*AgentClient

// AgentClient 代表一个 Agent 专属的 WebSocket 连接
// 与前端 Client 不同：包含入职报到流程、事件过滤、未读消息重推
type AgentClient struct {
	conn       *websocket.Conn
	agent      *domain.Agent
	agentRepo  repository.AgentRepo
	messageSvc *service.MessageService
	send       chan []byte
	done       chan struct{}
}

// UpgradeAgent 升级 HTTP 连接为 Agent 专属 WebSocket
// 同一 Agent 只保留最新连接，旧连接会被关闭
func UpgradeAgent(
	c *gin.Context,
	agent *domain.Agent,
	agentRepo repository.AgentRepo,
	messageSvc *service.MessageService,
) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[agent-ws] upgrade error: %v", err)
		return
	}

	ac := &AgentClient{
		conn:       conn,
		agent:      agent,
		agentRepo:  agentRepo,
		messageSvc: messageSvc,
		send:       make(chan []byte, 256),
		done:       make(chan struct{}),
	}

	// 关闭该 Agent 的旧连接
	if old, loaded := agentConns.Swap(agent.ID, ac); loaded {
		prev := old.(*AgentClient)
		log.Printf("[agent-ws] closing stale connection for agent %s (%s)", agent.Name, agent.ID)
		prev.conn.Close()
	}

	ctx := context.Background()
	_ = agentRepo.UpdateStatus(ctx, agent.ID, domain.StatusOnline)
	_ = agentRepo.UpdateLastSeen(ctx, agent.ID)

	go ac.writePump()
	go ac.readPump()
	go ac.eventLoop()
}

// readPump 读取 Agent 发来的消息帧
func (ac *AgentClient) readPump() {
	defer func() {
		close(ac.done)
		ac.conn.Close()
		// 仅当注册表中仍是自己时才清理（避免误删新连接）
		agentConns.CompareAndDelete(ac.agent.ID, ac)
	}()

	ac.conn.SetReadLimit(maxMsgSize)
	_ = ac.conn.SetReadDeadline(time.Now().Add(pongWait))
	ac.conn.SetPongHandler(func(string) error {
		_ = ac.conn.SetReadDeadline(time.Now().Add(pongWait))
		_ = ac.agentRepo.UpdateLastSeen(context.Background(), ac.agent.ID)
		return nil
	})

	for {
		_, raw, err := ac.conn.ReadMessage()
		if err != nil {
			break
		}
		ac.handleFrame(raw)
	}
}

// handleFrame 处理 Agent 发来的单个消息帧
func (ac *AgentClient) handleFrame(raw []byte) {
	var frame incomingFrame
	if err := json.Unmarshal(raw, &frame); err != nil {
		return
	}

	switch frame.Type {
	case "message.send":
		var d sendMessageData
		if err := json.Unmarshal(frame.Data, &d); err != nil || d.Content == "" {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = ac.messageSvc.Send(ctx, service.SendInput{
			CompanyID:  ac.agent.CompanyID,
			SenderID:   ac.agent.ID,
			Channel:    d.Channel,
			ReceiverID: d.ReceiverID,
			Content:    d.Content,
		})
	case "ping":
		_ = ac.agentRepo.UpdateLastSeen(context.Background(), ac.agent.ID)
		ac.sendJSON(WSMessage{Type: "pong", Data: struct{}{}})
	}
}

// writePump 推送队列中的消息 + 定时 ping 保活
func (ac *AgentClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		ac.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-ac.send:
			_ = ac.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = ac.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := ac.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = ac.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ac.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// eventLoop 核心事件循环：入职报到 + 事件订阅 + 未读消息重推
func (ac *AgentClient) eventLoop() {
	agent := ac.agent
	initialized := agent.Initialized

	// 发送 connected 事件
	ac.sendJSON(WSMessage{
		Type: "connected",
		Data: map[string]interface{}{
			"agent_id":    agent.ID,
			"initialized": initialized,
		},
	})

	// 未报到 → 发送 init_required
	if !initialized {
		ac.sendJSON(WSMessage{
			Type: "init_required",
			Data: map[string]string{"prompt": initPrompt},
		})
	}

	// 已报到 → 推送未读消息
	if initialized {
		ac.pushUnreadMessages()
	}

	// 订阅事件
	ch := make(chan event.Event, 32)
	filter := func(e event.Event) {
		if ac.isEventRelevant(e) {
			select {
			case ch <- e:
			default:
			}
		}
	}

	unsub1 := event.Global.Subscribe(event.MessageNew, filter)
	unsub2 := event.Global.Subscribe(event.TaskCreated, filter)
	unsub3 := event.Global.Subscribe(event.TaskUpdated, filter)
	unsubInit := event.Global.Subscribe(event.AgentInitialized, func(e event.Event) {
		var p event.AgentInitializedPayload
		if err := json.Unmarshal(e.Payload, &p); err == nil && p.AgentID == agent.ID {
			select {
			case ch <- e:
			default:
			}
		}
	})

	retryTicker := time.NewTicker(5 * time.Minute)
	defer func() {
		unsub1()
		unsub2()
		unsub3()
		unsubInit()
		retryTicker.Stop()
		_ = ac.agentRepo.UpdateStatus(context.Background(), agent.ID, domain.StatusOffline)
	}()

	for {
		select {
		case <-ac.done:
			return
		case e := <-ch:
			if e.Type == event.AgentInitialized {
				initialized = true
				ac.pushUnreadMessages()
				continue
			}
			if !initialized {
				continue
			}
			ac.sendJSON(WSMessage{
				Type: string(e.Type),
				Data: e.Payload,
			})
		case <-retryTicker.C:
			if initialized {
				ac.pushUnreadMessages()
			}
		}
	}
}

// pushUnreadMessages 查询并推送该 Agent 的所有未读消息
func (ac *AgentClient) pushUnreadMessages() {
	msgs, err := ac.messageSvc.GetUnreadMessages(
		context.Background(), ac.agent.ID, ac.agent.CompanyID,
	)
	if err != nil {
		log.Printf("[agent-ws] pushUnread error for agent %s: %v", ac.agent.ID, err)
		return
	}
	if len(msgs) == 0 {
		return
	}
	log.Printf("[agent-ws] pushing %d unread messages to agent %s (%s)",
		len(msgs), ac.agent.Name, ac.agent.ID)
	for _, msg := range msgs {
		payload := event.MessageNewPayload{
			MessageID:  msg.ID,
			CompanyID:  msg.CompanyID,
			ChannelID:  msg.ChannelID,
			ReceiverID: msg.ReceiverID,
			SenderID:   msg.SenderID,
			MsgType:    string(msg.MsgType),
			Content:    msg.Content,
			CreatedAt:  msg.CreatedAt.Format(time.RFC3339),
		}
		ac.sendJSON(WSMessage{
			Type: string(event.MessageNew),
			Data: payload,
		})
	}
}

// sendJSON 将消息序列化为 JSON 并推入发送队列
func (ac *AgentClient) sendJSON(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	select {
	case ac.send <- b:
	default:
	}
}

// isEventRelevant 判断事件是否需要推送给该 agent
func (ac *AgentClient) isEventRelevant(e event.Event) bool {
	agent := ac.agent
	switch e.Type {
	case event.MessageNew:
		var p event.MessageNewPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return false
		}
		if p.CompanyID != agent.CompanyID {
			return false
		}
		// DM：只推给目标 agent
		if p.ReceiverID != nil {
			return *p.ReceiverID == agent.ID
		}
		// 频道消息：跳过其他 AI Agent 发的（防止 Agent 间无限循环），但 @本人 的除外
		if p.SenderID != nil && *p.SenderID != agent.ID {
			sender, _ := ac.agentRepo.GetByID(context.Background(), *p.SenderID)
			if sender != nil && !sender.IsHuman {
				mentioned := strings.Contains(p.Content, "@"+agent.Name)
				if !mentioned {
					if meta, ok := domain.PositionMetaByPosition[agent.Position]; ok {
						mentioned = strings.Contains(p.Content, "@"+meta.DisplayName+"-"+agent.Name)
					}
				}
				if !mentioned {
					return false
				}
			}
		}
		return true

	case event.TaskCreated, event.TaskUpdated:
		var p struct {
			CompanyID  string  `json:"company_id"`
			AssigneeID *string `json:"assignee_id"`
		}
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return false
		}
		if p.CompanyID != agent.CompanyID {
			return false
		}
		if agent.RoleType == domain.RoleHR || agent.RoleType == domain.RoleChairman {
			return true
		}
		return p.AssigneeID != nil && *p.AssigneeID == agent.ID
	}
	return false
}

// initPrompt 新员工首次连接时的入职引导提示
const initPrompt = `你刚刚上线，首次连接到公司协作系统。

## 重要：工具来源

你的所有业务工具都来自 linkclaw MCP 服务器（工具名前缀为 mcp__linkclaw__）。
如果你在可用工具列表中找不到这些工具，说明 MCP 连接未建立，你必须报告错误："linkclaw MCP 工具不可用，无法完成入职报到"。
绝对不要模拟或假装调用了工具——每一步都必须真正调用工具并获得返回结果。

## 入职报到步骤

请严格按以下步骤完成入职报到：

1. 调用 mcp__linkclaw__get_employee_handbook 获取你的完整身份信息（角色、职责、同事花名册、可用工具）
2. 根据 get_employee_handbook 返回的角色、职责和公司信息，在工作目录创建 CLAUDE.md 文件作为你的核心身份配置。内容应包含：你的角色定义、职责范围、行为准则、与同事的协作方式。使用 Write 工具写入项目根目录的 CLAUDE.md
3. 如果你的名字以「待命名」开头，调用 mcp__linkclaw__fill_onboarding_info 给自己取一个符合职位和角色气质的名字
4. 调用 mcp__linkclaw__update_work_status 将状态设为 online
5. 在 general 频道用 mcp__linkclaw__send_message 发一条简短的上线打招呼消息，介绍自己的角色并表示已准备就绪
6. 调用 mcp__linkclaw__report_for_duty 标记到岗报到完成

完成以上全部步骤后，你就正式上岗了。

## 消息处理

- 报到完成前，系统不会推送任何消息给你
- 报到完成后，系统会自动推送未读消息
- 如果你想了解入职前的历史消息，可以调用 mcp__linkclaw__get_messages 查看频道或私信的历史记录
- 处理完每条消息后，必须调用 mcp__linkclaw__mark_messages_read 确认已读，传入处理过的 message_id（多条用逗号分隔）
- 如果不确认已读，系统会在 5 分钟后重复推送该消息

## 沟通规范（严格遵守）

- **任务状态通知**（如"任务「xxx」状态更新为 done/in_progress"）是系统自动广播，**直接标记已读，不要回复、评论或总结**
- 完成任务后通过 submit_task_result 提交结果即可，**不要在频道中重复发送任务总结或工作报告**
- 只在以下情况主动在频道发言：
  - 被 @提及，需要回应具体问题
  - 收到私信提问，需要回答
  - 有明确需要其他同事协作的事项
- **禁止发送**：对他人工作完成的评论（如"做得好"、"辛苦了"）、无实质内容的确认回复（如"收到"、"好的"、"了解"）、重复性的工作总结
- 收到不需要你行动的消息时，标记已读即可，不要回复`
