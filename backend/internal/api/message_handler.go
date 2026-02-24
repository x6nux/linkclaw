package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/linkclaw/backend/internal/service"
)

type messageHandler struct {
	messageSvc *service.MessageService
}

func (h *messageHandler) list(c *gin.Context) {
	agent := currentAgent(c)
	channel    := c.Query("channel")
	receiverID := c.Query("receiver_id")
	beforeID   := c.Query("before_id")
	limit, _   := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 1000 {
		limit = 500
	}

	var msgs interface{}
	var err error

	switch {
	case channel != "":
		msgs, err = h.messageSvc.GetChannelMessages(c.Request.Context(), agent.CompanyID, channel, limit, beforeID)
	case receiverID != "":
		msgs, err = h.messageSvc.GetDMMessages(c.Request.Context(), agent.ID, receiverID, limit, beforeID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "specify channel or receiver_id"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": msgs})
}

type sendMessageRequest struct {
	Channel    string `json:"channel"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content" binding:"required"`
}

func (h *messageHandler) send(c *gin.Context) {
	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	agent := currentAgent(c)
	msg, err := h.messageSvc.Send(c.Request.Context(), service.SendMessageInput{
		CompanyID:  agent.CompanyID,
		SenderID:   agent.ID,
		Channel:    req.Channel,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, msg)
}

func (h *messageHandler) listChannels(c *gin.Context) {
	agent    := currentAgent(c)
	channels, err := h.messageSvc.ListChannels(c.Request.Context(), agent.CompanyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": channels, "total": len(channels)})
}
