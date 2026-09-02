package handler

import (
	"day_4_1/pkg/errcode"
	"day_4_1/pkg/response"
	"day_4_1/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentReq struct {
	Content string `json:"content" binding:"min=1,max=1000"`
}

// type CommentResp struct {
// 	AuthorId  uint      `json:"author_id"`
// 	Content   string    `json:"content"`
// 	PostId    uint      `json:"post_id"`
// 	CreatedAt time.Time `json:"created_at"`
// }

func CreateComment(c *gin.Context) {
	postIdStr := c.Param("post_id")
	PostId, err := strconv.Atoi(postIdStr)
	if err != nil || PostId < 1 {
		c.Error(errcode.New(400, 10000, "帖子ID格式错误"))
		return
	}
	var req CommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BindError(c, err)
		return
	}
	UserId, exists := c.Get("id")
	if !exists {
		c.Error(errcode.ErrUnauthorized)
		return
	}
	comment, err := service.GenerateComment(UserId.(uint), uint(PostId), req.Content)
	if err != nil {
		c.Error(err)
		return
	}
	response.Created(c, comment)
}
func DeleteComment(c *gin.Context) {
	commentIdStr := c.Param("comment_id")
	commentId, err := strconv.ParseUint(commentIdStr, 10, 64)
	if err != nil {
		c.Error(errcode.New(400, 10000, "无效评论ID"))
		return
	}
	UserId, exists := c.Get("id")
	if !exists {
		c.Error(errcode.ErrUnauthorized)
		return
	}
	err = service.DeleteComment(uint(commentId), c.GetString("role"), UserId.(uint))
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, nil)
}
