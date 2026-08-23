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
func GetPostPage(page uint) ([]model.Post, error) {
	var posts []model.Post
	result := db.Preload("Author").Order("created_at desc").Limit(15).Offset(int((page - 1) * 15)).Find(&posts)
	return posts, result.Error
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
		return db.Order("created_at desc")
	}).Preload("Comments.Author").First(&post, postId).Error
	return post, err
}
