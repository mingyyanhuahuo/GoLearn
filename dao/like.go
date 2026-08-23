package dao

import (
	"day_4_1/model"

	"gorm.io/gorm"
)

func ToggleLike(postId uint, userId uint) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var count int64
	err := tx.Model(&model.Like{}).Where("post_id = ? AND user_id = ?", postId, userId).Count(&count).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	if count > 0 {
		if err := tx.Model(&model.Post{}).
			Where("id = ?", postId).
			UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Where("post_id = ? AND user_id = ?", postId, userId).Delete(&model.Like{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Model(&model.Post{}).
			Where("id = ?", postId).
			UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Create(&model.Like{PostId: postId, UserId: userId}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
func GetLikePostIds(postids []uint, userId uint) ([]uint, error) {
	var likesposts []uint
	err := db.Model(&model.Like{}).Where("post_id IN ? AND user_id = ?", postids, userId).Pluck("post_id", &likesposts).Error
	if err != nil {
		return nil, err
	}
	return likesposts, nil
}
