package service

import (
	"day_4_1/dao"
	"day_4_1/model"

	"day_4_1/pkg/errcode"
	"errors"

	"gorm.io/gorm"
)

func CreatePost(userId uint, title, content string) (model.Post, error) {
	if title == "" || content == "" {
		return model.Post{}, errcode.ErrBadRequest
	}
	post := &model.Post{
		Title:    title,
		Content:  content,
		AuthorId: userId,
	}
	if err := dao.GeneratePost(post); err != nil {
		return model.Post{}, errcode.ErrDatabase
	}
	if user, err := dao.GetUserById(userId); err == nil {
		post.Author = user
	}
	return *post, nil
}
func ListPosts(page uint, pageSize uint) ([]model.Post, int64, error) {
	posts, total, err := dao.GetPostPage(page, pageSize)
	if err != nil {
		return []model.Post{}, 0, errcode.ErrDatabase
	}
	return posts, total, nil
}
func DeletePost(postId uint, role string, userId uint) error {
	post, err := dao.GetPostById(postId)
	if err != nil {
		return errcode.ErrNotFoundPost
	}
	authorId := post.AuthorId
	if !IsAdmin(role) && authorId != userId {
		return errcode.ErrForbidden
	}
	if err := dao.DeletePost(postId); err != nil {
		return errcode.ErrDatabase
	}
	if err = dao.DeleteCommentsByPostId(postId); err != nil {
		return errcode.ErrDatabase
	}
	return nil

}
func GetDetailedPostById(postId uint) (model.Post, error) {
	post, err := dao.GetDetailedPostById(postId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Post{}, errcode.ErrNotFoundPost
	}
	if err != nil {
		return model.Post{}, errcode.ErrDatabase
	}
	return post, nil
}
func GetPostPageHot(page uint, pageSize uint) ([]model.Post, int64, error) {
	posts, total, err := dao.GetPostPageHot(page, pageSize)
	if err != nil {
		return []model.Post{}, 0, errcode.ErrDatabase
	}
	return posts, total, nil
}
