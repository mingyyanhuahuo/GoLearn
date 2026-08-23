package handler

import (
	"day_4_1/pkg/response"
	"day_4_1/service"
	"net/http"
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
		response.Err(c, http.StatusBadRequest, "帖子ID格式错误")
		return
	}
	userId, exists := c.Get("id")
	if !exists {
		response.Err(c, http.StatusUnauthorized, "用户未登录")
		return
	}
	err = service.ToggleLike(uint(postId), userId.(uint))
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, "操作成功")
}
func GetLikePostIds(c *gin.Context) {
	var req LikePostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}
	userId, exists := c.Get("id")
	if !exists {
		response.Err(c, http.StatusUnauthorized, "用户未登录")
		return
	}
	likePostIds, err := service.GetLikePostIds(req.PostIds, userId.(uint))
	if err != nil {
		response.Err(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, LikePostIdsResp{PostIds: likePostIds})
}
