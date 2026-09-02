package dao

import (
	"context"

	"day_4_1/model"
	"day_4_1/pkg/redisdb"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func GetLikesStatus(postids []uint, userId uint) (map[uint]bool, error) {
	ctx := context.Background()
	pipe := redisdb.Rdb.Pipeline()
	cmds := make([]*redis.BoolCmd, len(postids))
	for i, postId := range postids {
		key := fmt.Sprintf("post:likes:%d", postId)
		cmds[i] = pipe.SIsMember(ctx, key, userId)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[uint]bool, len(postids))
	for i, cmd := range cmds {
		isMember, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		result[postids[i]] = isMember
	}
	return result, nil
}
func GetExistingLikePostIds(postids []uint) ([]uint, error) {
	var ids []uint
	err := db.Model(&model.Post{}).Where("id IN ?", postids).Pluck("id", &ids).Error
	return ids, err
}
func ToggleLike(postId uint, userId uint) (bool, error) {
	keys := fmt.Sprintf("post:likes:%d", postId)
	added, err := redisdb.Rdb.SAdd(context.Background(), keys, userId).Result()
	if err != nil {
		return false, err
	}
	if added == 0 {
		redisdb.Rdb.SRem(context.Background(), keys, userId)
		return false, nil
	}
	return true, nil
}

//这是先前学mysql数据库操作时写的db操作,已经注释掉,功能与上方一致但为mysql操作,不适合like操作,下方函数也是同道理
// func ToggleLike(postId uint, userId uint) error {
// 	tx := db.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 		}
// 	}()
// 	var count int64
// 	err := tx.Model(&model.Like{}).Where("post_id = ? AND user_id = ?", postId, userId).Count(&count).Error
// 	if err != nil {
// 		tx.Rollback()
// 		return err
// 	}
// 	if count > 0 {
// 		if err := tx.Model(&model.Post{}).
// 			Where("id = ?", postId).
// 			UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
// 			tx.Rollback()
// 			return err
// 		}
// 		if err := tx.Where("post_id = ? AND user_id = ?", postId, userId).Delete(&model.Like{}).Error; err != nil {
// 			tx.Rollback()
// 			return err
// 		}
// 	} else {
// 		if err := tx.Model(&model.Post{}).
// 			Where("id = ?", postId).
// 			UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
// 			tx.Rollback()
// 			return err
// 		}
// 		if err := tx.Create(&model.Like{PostId: postId, UserId: userId}).Error; err != nil {
// 			tx.Rollback()
// 			return err
// 		}
// 	}
// 	return tx.Commit().Error
// }

//
// func GetLikePostIds(postids []uint, userId uint) ([]uint, error) {
// 	var likesposts []uint
// 	err := db.Model(&model.Like{}).Where("post_id IN ? AND user_id = ?", postids, userId).Pluck("post_id", &likesposts).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return likesposts, nil
// }
