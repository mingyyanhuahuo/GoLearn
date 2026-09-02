package router

import (
	"day_4_1/handler"
	"day_4_1/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	api := r.Group("/api/v1/auth")
	api.POST("/register", handler.Register)
	api.POST("/login", handler.Login)
	auth := r.Group("/api/v1")
	auth.Use(middleware.AuthMiddleware())
	auth.GET("/posts", handler.ListPosts)
	auth.GET("/posts/:post_id", handler.GetDetailedPost)

	auth.POST("/posts", handler.CreatePost)
	auth.DELETE("/posts/:post_id", handler.DeletePost)
	auth.POST("/posts/:post_id/comments", handler.CreateComment)
	auth.DELETE("posts/:post_id/comments/:comment_id", handler.DeleteComment)
	auth.POST("/posts/:post_id/like", middleware.RateLimitMiddleware(), handler.ToggleLike)
	auth.POST("/posts/likes", handler.GetLikePostIds)
	auth.POST("/agent/chat", handler.AgentChat)
	admin := auth.Group("admin", middleware.AdminMiddleware())
	admin.DELETE("/posts/:post_id", handler.DeletePost)
}
