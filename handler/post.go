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
	Title   string `json:"title" binding:"min=1,max=150"`
	Content string `json:"content" binding:"omitempty,max=2000"`
}
type CreatePostResp struct {
	Id        uint       `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Author    model.User `json:"author"`
	CreatedAt string     `json:"created_at"`
}
type PageMeta struct {
	Page     uint  `json:"page"`
	PageSize uint  `json:"page_size"`
	Total    int64 `json:"total"`
}

type ListPostsResp struct {
	Items []PostItem `json:"items"`
	Meta  PageMeta   `json:"meta"`
}

type DetailedPostResp struct {
	Id           uint            `json:"id"`
	Title        string          `json:"title"`
	Content      string          `json:"content"`
	Author       model.User      `json:"author"`
	LikeCount    uint            `json:"like_count"`
	CommentCount uint            `json:"comment_count"`
	CreatedAt    time.Time       `json:"created_at"`
	Comments     []model.Comment `json:"comments"`
}

func CreatePost(c *gin.Context) {
	var req PostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BindError(c, err)
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
	response.Created(c, CreatePostResp{
		Id:        post.Id,
		Title:     post.Title,
		Content:   post.Content,
		Author:    post.Author,
		CreatedAt: post.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}
func ListPosts(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.Error(errcode.New(400, 10000, "无效的页码"))
		return
	}
	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		c.Error(errcode.New(400, 10000, "无效的页码大小"))
		return
	}
	sort := c.DefaultQuery("sort", "")
	var posts []model.Post
	var total int64
	if sort == "hot" {
		posts, total, err = service.GetPostPageHot(uint(page), uint(pageSize))
	} else {
		posts, total, err = service.ListPosts(uint(page), uint(pageSize))
	}
	if err != nil {
		c.Error(err)
		return
	}
	items := make([]PostItem, 0, len(posts))
	for _, post := range posts {
		items = append(items, PostItem{
			Id:           post.Id,
			Content:      post.Content,
			Author:       post.Author,
			LikeCount:    post.LikeCount,
			CommentCount: post.CommentCount,
			CreatedAt:    post.CreatedAt,
		})
	}
	response.OK(c, ListPostsResp{
		Items: items,
		Meta: PageMeta{
			Page:     uint(page),
			PageSize: uint(pageSize),
			Total:    total,
		},
	})

}
func DeletePost(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		c.Error(errcode.New(400, 10000, "帖子ID格式错误"))
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
	response.OK(c, nil)
}
func GetDetailedPost(c *gin.Context) {
	postIdStr := c.Param("post_id")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		c.Error(errcode.New(400, 10000, "id格式错误"))
		return
	}
	post, err := service.GetDetailedPostById(uint(postId))

	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, DetailedPostResp{
		Id:           post.Id,
		Title:        post.Title,
		Content:      post.Content,
		Author:       post.Author,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		CreatedAt:    post.CreatedAt,
		Comments:     post.Comments,
	})
}
