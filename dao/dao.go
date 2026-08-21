package dao

import (
	"day_3_1/model"
	"fmt"

	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB(database *gorm.DB) {
	db = database
}
func CreateUser(user *model.User) error {
	return db.Create(user).Error
}
func GetUserById(id uint) (model.User, error) {
	var user model.User
	result := db.First(&user, id)
	return user, result.Error
}
func GenerateComment(comment *model.Comment) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	error := tx.Create(comment).Error
	if error != nil {
		tx.Rollback()
		return fmt.Errorf("创建评论失败: %w", error)
	}
	err := tx.Model(&model.Post{}).
		Where("id = ?", comment.PostId).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("更新评论计数失败: %w", err)
	}
	return tx.Commit().Error
}
func GeneratePost(post *model.Post) error {
	return db.Create(post).Error
}
func MakeLike(postId uint) error {
	return db.Model(&model.Post{}).
		Where("id = ?", postId).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
}
func GetPostPage(page uint) ([15]model.Post, error) {
	var posts [15]model.Post
	result := db.Preload("Author").Order("created_at asc").Limit(15).Offset(int((page - 1) * 15)).Find(&posts)
	return posts, result.Error
}
