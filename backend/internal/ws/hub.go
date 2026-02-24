package ws

import (
	"encoding/json"

	"github.com/linkclaw/backend/internal/event"
)

// Hub 管理所有 WebSocket 连接，并将事件广播给同公司的客户端
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMsg
}

type broadcastMsg struct {
	CompanyID string
	Msg       interface{}
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client, 16),
		unregister: make(chan *Client, 16),
		broadcast:  make(chan broadcastMsg, 256),
	}
}

// Run 启动 Hub 主循环（需在 goroutine 中运行）
func (h *Hub) Run() {
	h.subscribeEvents()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case msg := <-h.broadcast:
			for client := range h.clients {
				if client.CompanyID != msg.CompanyID {
					continue
				}
				client.SendJSON(msg.Msg)
			}
		}
	}
}

// Broadcast 向指定公司的所有 WS 客户端广播消息
func (h *Hub) Broadcast(companyID string, msg interface{}) {
	select {
	case h.broadcast <- broadcastMsg{CompanyID: companyID, Msg: msg}:
	default:
	}
}

// subscribeEvents 将 event.Bus 上的事件转发给 WS 客户端
func (h *Hub) subscribeEvents() {
	forward := func(e event.Event) {
		companyID := extractCompanyID(e)
		if companyID == "" {
			return
		}
		h.Broadcast(companyID, WSMessage{
			Type: string(e.Type),
			Data: e.Payload,
		})
	}

	for _, t := range []event.Type{
		event.AgentOnline, event.AgentOffline, event.AgentStatus,
		event.TaskCreated, event.TaskUpdated,
		event.MessageNew,
	} {
		event.Global.Subscribe(t, forward)
	}
}

func extractCompanyID(e event.Event) string {
	switch e.Type {
	case event.AgentOnline, event.AgentOffline, event.AgentStatus:
		var p event.AgentStatusPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return p.CompanyID
		}
	case event.TaskCreated, event.TaskUpdated:
		var p event.TaskUpdatedPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return p.CompanyID
		}
	case event.MessageNew:
		var p event.MessageNewPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return p.CompanyID
		}
	}
	return ""
}
