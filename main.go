package main

import (
	"day_4_1/config"
	"day_4_1/dao"
	"day_4_1/middleware"
	"day_4_1/model"
	"day_4_1/pkg/deepseek"
	"day_4_1/pkg/jwtutil"
	"day_4_1/pkg/logger"
	"day_4_1/pkg/redisdb"
	"day_4_1/router"
	"log"
	"os"
	"os/signal"
	"time"

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
	redisdb.InitRedis()
	if err := jwtutil.Init(config.GetConfig().JWT.Secret); err != nil {
		log.Fatalf("JWT初始化失败: %v", err)
	}
	deepseek.Init(config.GetConfig().DeepSeek.ApiKey, config.GetConfig().DeepSeek.BaseUrl, config.GetConfig().DeepSeek.ModelFlash)
	r := gin.Default()
	r.Use(middleware.AccessLog())
	r.Use(middleware.ErrorMiddleware())
	router.InitRouter(r)
	if err := dao.ImportLikesFromDB(); err != nil {
		logger.Logger.Error("从数据库导入点赞数据到Redis失败", zap.Error(err))
	}
	exitChan := make(chan struct{})
	doneChan := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := dao.SyncLikesToDB(); err != nil {
					logger.Logger.Error("同步点赞数据到数据库失败", zap.Error(err))
				}
			case <-exitChan:
				if err := dao.SyncLikesToDB(); err != nil {
					logger.Logger.Error("同步点赞数据到数据库失败", zap.Error(err))
				}
				close(doneChan)
				return
			}
		}
	}()
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, os.Kill)
		<-quit
		logger.Logger.Info("接收到退出信号，正在同步点赞数据到数据库...")
		close(exitChan)
		<-doneChan
		logger.Logger.Info("点赞数据同步完成，程序退出")
		os.Exit(0)
	}()

	port := config.GetConfig().Server.Port
	log.Printf("服务器启动，监听端口: %s", port)
	r.Run(":" + port)
}
