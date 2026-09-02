package handler

import (
	"day_4_1/pkg/errcode"
	"day_4_1/pkg/response"
	"day_4_1/service"

	"github.com/gin-gonic/gin"
)

type AgentReq struct {
	SessionID      string `json:"session_id" binding:"required,min=1,max=128"`
	Message        string `json:"message" binding:"required,min=1,max=4000"`
	ConfirmDraftID string `json:"confirm_draft_id"`
}

var agentService = service.NewAgentService()

func AgentChat(c *gin.Context) {
	var req AgentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BindError(c, err)
		return
	}
	userid, exists := c.Get("id")
	if !exists {
		c.Error(errcode.ErrUnauthorized)
		return
	}
	resp, err := agentService.Chat(userid.(uint), req.SessionID, req.Message, req.ConfirmDraftID)
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, resp)
}
