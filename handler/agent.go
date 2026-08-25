package handler

import (
	"day_4_1/pkg/errcode"
	"day_4_1/pkg/response"
	"day_4_1/service"
	"fmt"

	"github.com/gin-gonic/gin"
)

type AgentReq struct {
	SessionID      string `json:"session_id"`
	Message        string `json:"message" binding:"required"`
	ConfirmDraftID string `json:"confirm_draft_id"`
}

var agentService = service.NewAgentService()

func AgentChat(c *gin.Context) {
	var req AgentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errcode.New(400, 10000, "请求参数错误: "+err.Error()))
		return
	}
	userid, exists := c.Get("id")
	if !exists {
		c.Error(errcode.ErrUnauthorized)
		return
	}
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("session_%d", userid)
	}
	resp, err := agentService.Chat(userid.(uint), req.SessionID, req.Message, req.ConfirmDraftID)
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, resp)
}
