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

type PostReq struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required,max=800"`
}
type PostResp struct {
	Id        uint      `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	AuthorId  uint      `json:"author_id"`
}
type ListPostsResp struct {
	Posts []model.Post `json:"posts"`
	Page  uint         `json:"page"`
}
type DetailedPostResp struct {
	Id         uint            `json:"id"`
	Title      string          `json:"title"`
	Content    string          `json:"content"`
	CreatedAt  time.Time       `json:"created_at"`
	LikeCount  uint            `json:"like_count"`
	AuthorId   uint            `json:"author_id"`
	AuthorName string          `json:"author_name"`
	Comments   []model.Comment `json:"comments"`
}

func CreatePost(c *gin.Context) {
	var req PostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errcode.New(400, 10000, "请求参数错误: "+err.Error()))
		return
	}
	UserId, exists := c.Get("id")
	if !exists {
		c.Error(errcode.ErrUnauthorized)
		return
	}
	post, err := service.CreatePost(UserId.(uint), req.Title, req.Content)
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, PostResp{
		Id:        post.Id,
		Title:     post.Title,
		Content:   post.Content,
		AuthorId:  post.AuthorId,
		CreatedAt: post.CreatedAt,
	})
}
func ListPosts(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)

	if err != nil {
		c.Error(errcode.New(400, 10000, "无效的页码: "+err.Error()))
		return
	}
	if page < 1 {
		c.Error(errcode.New(400, 10000, "无效的页码"))
		return
	}
	sort := c.DefaultQuery("sort", "")
	if sort == "hot" {
		posts, err := service.GetPostPageHot(uint(page))
		if err != nil {
			c.Error(err)
			return
		}
		response.OK(c, ListPostsResp{
			Posts: posts,
			Page:  uint(page),
		})
		return
	}
	posts, err := service.ListPosts(uint(page))
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, ListPostsResp{
		Posts: posts,
		Page:  uint(page),
	})

}
func DeletePost(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		c.Error(errcode.New(400, 10000, "帖子ID格式错误: "+err.Error()))
		return
	}
	role, _ := c.Get("role")
	userId, exists := c.Get("id")
	if !exists {
		c.Error(errcode.ErrUnauthorized)
		return
	}
	err = service.DeletePost(uint(postId), role.(string), userId.(uint))
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, "删除帖子成功")
}
func GetDetailedPost(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		c.Error(errcode.New(400, 10000, "id格式错误: "+err.Error()))
		return
	}
	post, err := service.GetDetailedPostById(uint(postId))

	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, DetailedPostResp{
		Id:         post.Id,
		Title:      post.Title,
		Content:    post.Content,
		CreatedAt:  post.CreatedAt,
		AuthorId:   post.AuthorId,
		LikeCount:  post.LikeCount,
		AuthorName: post.Author.Username,
		Comments:   post.Comments,
	})
}
