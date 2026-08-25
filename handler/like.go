package handler

import (
	"day_4_1/pkg/errcode"
	"day_4_1/pkg/response"
	"day_4_1/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LikePostIdsResp struct {
	PostIds []uint `json:"post_ids"`
}
type LikePostReq struct {
	PostIds []uint `json:"post_ids" binding:"required"`
}

func ToggleLike(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		c.Error(errcode.New(400, 10000, "帖子ID格式错误: "+err.Error()))
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
		c.Error(errcode.New(400, 10000, "请求参数错误: "+err.Error()))
		return
	}
	userId, exists := c.Get("id")
	if !exists {
		c.Error(errcode.ErrUnauthorized)
		return
	}
	likePostIds, err := service.GetLikePostIds(req.PostIds, userId.(uint))
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, LikePostIdsResp{PostIds: likePostIds})
}
