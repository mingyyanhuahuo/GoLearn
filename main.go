package main

import (
	"day_4_1/config"
	"day_4_1/dao"
	"day_4_1/middleware"
	"day_4_1/model"
	"day_4_1/pkg/jwtutil"
	"day_4_1/pkg/logger"
	"day_4_1/router"
	"log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initDB() *gorm.DB {
	dsn := config.GetConfig().Mysql.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	return db
}

func main() {
	logger.InitLogger()
	logger.Logger.Info("日志服务启动", zap.String("port", config.GetConfig().Server.Port))
	db := initDB()
	// 自动迁移数据库表
	err := db.AutoMigrate(&model.User{}, &model.Post{}, &model.Comment{}, &model.Like{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	dao.InitDB(db)
	if err := jwtutil.Init(config.GetConfig().JWT.Secret); err != nil {
		log.Fatalf("JWT初始化失败: %v", err)
	}
	r := gin.Default()
	r.Use(middleware.AccessLog())
	r.Use(middleware.ErrorMiddleware())
	router.InitRouter(r)
	port := config.GetConfig().Server.Port
	log.Printf("服务器启动，监听端口: %s", port)
	r.Run(":" + port)
}
