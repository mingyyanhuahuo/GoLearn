package handler

import (
	"day_4_1/pkg/response"
	"day_4_1/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CommentReq struct {
	Content string `json:"content" binding:"required,max=150"`
	PostId  uint   `json:"post_id" binding:"required"`
}
type CommentResp struct {
	AuthorId  uint      `json:"author_id"`
	Content   string    `json:"content"`
	PostId    uint      `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateComment(c *gin.Context) {
	var req CommentReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Err(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}
	UserId, exists := c.Get("id")
	if !exists {
		response.Err(c, http.StatusUnauthorized, "用户未登录")
		return
	}
	comment, err := service.GenerateComment(UserId.(uint), req.PostId, req.Content)
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, CommentResp{
		AuthorId:  comment.AuthorId,
		Content:   comment.Content,
		PostId:    comment.PostId,
		CreatedAt: comment.CreatedAt,
	})
}
func DeleteComment(c *gin.Context) {
	commentIdStr := c.Param("comment_id")
	commentId, err := strconv.ParseUint(commentIdStr, 10, 64)
	if err != nil {
		response.Err(c, http.StatusBadRequest, "无效的评论ID")
		return
	}
	UserId, exists := c.Get("id")
	if !exists {
		response.Err(c, http.StatusUnauthorized, "用户未登录")
		return
	}
	err = service.DeleteComment(uint(commentId), c.GetString("role"), UserId.(uint))
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, "评论删除成功")
}
