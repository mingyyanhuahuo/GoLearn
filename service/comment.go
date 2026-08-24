package service

import (
	"day_4_1/dao"
	"day_4_1/model"
	"day_4_1/pkg/errcode"
	"time"
)

func GenerateComment(userId, postId uint, content string) (model.Comment, error) {
	if _, err := dao.GetPostById(postId); err != nil {
		return model.Comment{}, errcode.ErrNotFoundPost
	}
	comment := &model.Comment{
		Content:   content,
		PostId:    postId,
		AuthorId:  userId,
		CreatedAt: time.Now(),
	}
	err := dao.GenerateComment(comment)
	if err != nil {
		return model.Comment{}, errcode.ErrDatabase
	}
	return *comment, nil
}
func DeleteComment(commentId uint, role string, userId uint) error {
	comment, err := dao.GetCommentById(commentId)
	if err != nil {
		return errcode.ErrNotFoundComment
	}
	authorId := comment.AuthorId
	if role != "admin" && authorId != userId {
		return errcode.ErrForbidden
	}
	if err := dao.DeleteComment(commentId); err != nil {
		return errcode.ErrDatabase
	}
	return nil
}
