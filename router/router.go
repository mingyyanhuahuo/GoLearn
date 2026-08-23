package router

import (
	"day_4_1/handler"
	"day_4_1/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	api := r.Group("/api/v1")
	api.POST("/register", handler.Register)
	api.POST("/login", handler.Login)
	api.GET("/posts", handler.ListPosts)
	api.GET("/posts/:post_id", handler.GetDetailedPost)
	auth := api.Group("", middleware.AuthMiddleware())
	auth.POST("/posts", handler.CreatePost)
	auth.DELETE("/posts/:post_id", handler.DeletePost)
	auth.POST("/comments", handler.CreateComment)
	auth.DELETE("/comments/:comment_id", handler.DeleteComment)
	auth.POST("/posts/:post_id/like", handler.ToggleLike)
	auth.POST("/posts/likes", handler.GetLikePostIds)
	admin := auth.Group("admin", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	admin.DELETE("/posts/:post_id", handler.DeletePost)
}
