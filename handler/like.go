package handler

import (
	"day_4_1/model"
	"day_4_1/pkg/errcode"
	"day_4_1/pkg/response"
	"day_4_1/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

//	type LikePostIdsResp struct {
//		PostIds []uint `json:"post_ids"`
//	}
type LikePostReq struct {
	PostIds []uint `json:"post_ids" binding:"required"`
}

const (
	MaxStatusPostIds = 100
)

type LikeStatusResp struct {
	PostIds []service.LikeStatusItem `json:"status"`
}
type PostItem struct {
	Id           uint       `json:"id"`
	Content      string     `json:"content"`
	Author       model.User `json:"author"`
	LikeCount    uint       `json:"like_count"`
	CommentCount uint       `json:"comment_count"`
	CreatedAt    time.Time  `json:"created_at"`
}

func ToggleLike(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		BindError(c, err)
		return
	}
	userId, exists := c.Get("id")
	if !exists {
		c.Error(errcode.ErrUnauthorized)
		return
	}
	liked, err := service.ToggleLike(uint(postId), userId.(uint))
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, gin.H{"post_id": postId, "is_liked": liked})
}
func GetLikePostIds(c *gin.Context) {
	var req LikePostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BindError(c, err)
		return
	}
	if len(req.PostIds) > MaxStatusPostIds {
		c.Error(errcode.New(400, 10000, "请求参数错误: post_ids 数量超过限制"))
		return
	}
	userId, exists := c.Get("id")
	if !exists {
		c.Error(errcode.ErrUnauthorized)
		return
	}
	status, err := service.GetLikeStatus(req.PostIds, userId.(uint))
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, LikeStatusResp{PostIds: status})
}
