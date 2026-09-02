package dao

import (
	"day_4_1/model"

	"gorm.io/gorm"
)

func GeneratePost(post *model.Post) error {
	return db.Create(post).Error
}
func MakeLike(postId uint) error {
	return db.Model(&model.Post{}).
		Where("id = ?", postId).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
}
func GetPostPage(page uint, pageSize uint) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	if err := db.Model(&model.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	result := db.Preload("Author").Order("created_at desc").
		Limit(int(pageSize)).Offset(int((page - 1) * pageSize)).Find(&posts)
	return posts, total, result.Error
}
func GetPostById(postId uint) (model.Post, error) {
	var post model.Post
	erro := db.First(&post, postId).Error
	return post, erro
}
func DeletePost(postId uint) error {
	return db.Delete(&model.Post{}, postId).Error
}
func GetDetailedPostById(postId uint) (model.Post, error) {
	var post model.Post
	err := db.Preload("Author").Preload("Comments", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc")
	}).Preload("Comments.Author").First(&post, postId).Error
	return post, err
}
func GetPostPageHot(page, pageSize uint) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	if err := db.Model(&model.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	result := db.Preload("Author").
		Order("(like_count*2 + comment_count*5)/POWER(TIMESTAMPDIFF(HOUR, created_at, NOW()) + 2, 1.5) DESC").
		Limit(int(pageSize)).Offset(int((page - 1) * pageSize)).Find(&posts)
	return posts, total, result.Error
}
