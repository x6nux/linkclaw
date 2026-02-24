package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 50 * time.Second
	maxMsgSize = 8192
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client 代表一个 WebSocket 连接
type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	CompanyID  string
	AgentID    string
	agent      *domain.Agent
	agentRepo  repository.AgentRepo
	messageSvc *service.MessageService
}

// WSMessage 推送给前端的事件格式
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// incomingFrame 前端发来的消息帧
type incomingFrame struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// sendMessageData 前端发送消息的数据
type sendMessageData struct {
	Channel    string `json:"channel"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
}

// Upgrade 升级 HTTP 连接为 WebSocket，并绑定 agent 身份与服务依赖
func Upgrade(c *gin.Context, hub *Hub, agent *domain.Agent, agentRepo repository.AgentRepo, messageSvc *service.MessageService) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	client := &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		CompanyID:  agent.CompanyID,
		AgentID:    agent.ID,
		agent:      agent,
		agentRepo:  agentRepo,
		messageSvc: messageSvc,
	}
	hub.register <- client

	// 标记在线
	ctx := context.Background()
	_ = agentRepo.UpdateStatus(ctx, agent.ID, domain.StatusOnline)
	_ = agentRepo.UpdateLastSeen(ctx, agent.ID)

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		// 断开即离线
		_ = c.agentRepo.UpdateStatus(context.Background(), c.AgentID, domain.StatusOffline)
	}()
	c.conn.SetReadLimit(maxMsgSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		_ = c.agentRepo.UpdateLastSeen(context.Background(), c.AgentID)
		return nil
	})
	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		c.handleFrame(raw)
	}
}

func (c *Client) handleFrame(raw []byte) {
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
		if c.messageSvc == nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = c.messageSvc.Send(ctx, service.SendInput{
			CompanyID:  c.CompanyID,
			SenderID:   c.AgentID,
			Channel:    d.Channel,
			ReceiverID: d.ReceiverID,
			Content:    d.Content,
		})
	case "ping":
		_ = c.agentRepo.UpdateLastSeen(context.Background(), c.AgentID)
		c.SendJSON(WSMessage{Type: "pong", Data: struct{}{}})
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendJSON 将任意值序列化为 JSON 并推入发送队列
func (c *Client) SendJSON(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	select {
	case c.send <- b:
	default:
	}
}
