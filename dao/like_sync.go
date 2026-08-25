package dao

import (
	"context"
	"day_4_1/model"
	"day_4_1/pkg/logger"
	"day_4_1/pkg/redisdb"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func ImportLikesFromDB() error {
	var likes []model.Like
	if err := db.Find(&likes).Error; err != nil {
		return err
	}
	ctx := context.Background()
	pipe := redisdb.Rdb.Pipeline()
	for _, like := range likes {
		postId := fmt.Sprintf("%d", like.PostId)
		pipe.SAdd(ctx, "post:likes:"+postId, like.UserId)
	}
	_, err := pipe.Exec(ctx)
	return err
}
func SyncLikesToDB() error {
	ctx := context.Background()
	var cursor uint64
	for {
		keys, next, err := redisdb.Rdb.Scan(ctx, cursor, "post:likes:*", 100).Result()
		if err != nil {
			return err
		}
		for _, key := range keys {
			err := syncOneKey(ctx, key)
			if err != nil {
				logger.Logger.Error("同步点赞数据到数据库失败", zap.String("key", key), zap.Error(err))
				continue
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return syncZeroCounts(ctx)
}
func syncOneKey(ctx context.Context, key string) error {
	postId, err := strconv.ParseUint(strings.TrimPrefix(key, "post:likes:"), 10, 64)
	if err != nil {
		return err
	}
	members, err := redisdb.Rdb.SMembers(ctx, key).Result()
	if err != nil {
		return err
	}
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var postCount int64
	if err := tx.Model(&model.Post{}).Where("id = ?", postId).Count(&postCount).Error; err != nil {
		tx.Rollback()
		return err
	}
	if postCount == 0 {
		redisdb.Rdb.Del(ctx, key)
		tx.Rollback()
		return nil
	}
	if err := tx.Model(&model.Post{}).
		Where("id = ?", postId).Update("like_count", len(members)).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("post_id = ?", postId).
		Delete(&model.Like{}).
		Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, member := range members {
		userId, err := strconv.ParseUint(member, 10, 64)
		if err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Create(&model.Like{
			PostId: uint(postId),
			UserId: uint(userId),
		}).Error; err != nil {
			tx.Rollback()
			return err
		}

	}
	return tx.Commit().Error
}
func syncZeroCounts(ctx context.Context) error {
	var postIds []uint
	if err := db.Model(&model.Like{}).Distinct("post_id").Pluck("post_id", &postIds).Error; err != nil {
		return err
	}
	pipe := redisdb.Rdb.Pipeline()
	existsCmds := make([]*redis.IntCmd, len(postIds))
	for i, postId := range postIds {
		key := fmt.Sprintf("post:likes:%d", postId)
		existsCmds[i] = pipe.Exists(ctx, key)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	for i, cmd := range existsCmds {
		if cmd.Val() == 0 {
			if err := zeroOnePost(postIds[i]); err != nil {
				logger.Logger.Error("同步点赞数据到数据库失败", zap.Uint("postId", postIds[i]), zap.Error(err))
				continue
			}
		}
	}
	return nil
}
func zeroOnePost(postId uint) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Model(&model.Post{}).Where("id = ?", postId).Update("like_count", 0).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("post_id = ?", postId).Delete(&model.Like{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
