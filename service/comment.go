package service

import (
	"day_4_1/dao"
	"day_4_1/model"
	"fmt"
	"time"
)

func GenerateComment(userId, postId uint, content string) (model.Comment, error) {
	if _, err := dao.GetPostById(postId); err != nil {
		return model.Comment{}, fmt.Errorf("帖子不存在")
	}
	comment := &model.Comment{
		Content:   content,
		PostId:    postId,
		AuthorId:  userId,
		CreatedAt: time.Now(),
	}
	err := dao.GenerateComment(comment)
	if err != nil {
		return model.Comment{}, fmt.Errorf("创建评论失败")
	}
	return *comment, nil
}
func DeleteComment(commentId uint, role string, userId uint) error {
	comment, err := dao.GetCommentById(commentId)
	if err != nil {
		return fmt.Errorf("评论不存在")
	}
	authorId := comment.AuthorId
	if role != "admin" && authorId != userId {
		return fmt.Errorf("无权限删除评论")
	}
	if err := dao.DeleteComment(commentId); err != nil {
		return fmt.Errorf("删除评论失败")
	}
	return nil
}

//废弃掉了不想删...
// func DeleteCommentByPostId(postId uint) error {
// 	comments, err := dao.GetCommentsByPostId(postId)
// 	if err != nil {
// 		return fmt.Errorf("获取帖子评论失败")
// 	}
// 	for _, comment := range comments {
// 		if err := dao.DeleteComment(comment.Id); err != nil {
// 			return fmt.Errorf("删除评论失败")
// 		}
// 	}
// 	return nil
// }
