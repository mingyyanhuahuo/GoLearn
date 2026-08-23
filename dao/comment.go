package dao

import (
	"day_4_1/model"

	"gorm.io/gorm"
)

func GenerateComment(comment *model.Comment) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(comment).Error; err != nil {
		tx.Rollback()
		return err
	}
	err := tx.Model(&model.Post{}).
		Where("id = ?", comment.PostId).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
func GetCommentById(commentId uint) (model.Comment, error) {
	var comment model.Comment
	err := db.First(&comment, commentId).Error
	if err != nil {
		return model.Comment{}, err
	}
	return comment, err
}
func DeleteComment(commentId uint) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	comment, err := GetCommentById(commentId)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&comment).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&model.Post{}).
		Where("id = ?", comment.PostId).
		UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1)).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
func DeleteCommentsByPostId(postId uint) error {
	return db.Where("post_id = ?", postId).Delete(&model.Comment{}).Error
}
func GetCommentsByPostId(postId uint) ([]model.Comment, error) {
	var comments []model.Comment
	err := db.Where("post_id = ?", postId).
		Order("created_at desc").
		Find(&comments).Error
	if err != nil {
		return []model.Comment{}, err
	}
	return comments, nil
}
