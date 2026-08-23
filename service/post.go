package service

import (
	"day_4_1/dao"
	"day_4_1/model"
	"fmt"

	"errors"

	"gorm.io/gorm"
)

func CreatePost(userId uint, title, content string) (model.Post, error) {
	if title == "" || content == "" {
		return model.Post{}, fmt.Errorf("标题或内容不能为空")
	}
	post := &model.Post{
		Title:    title,
		Content:  content,
		AuthorId: userId,
	}
	if err := dao.GeneratePost(post); err != nil {
		return model.Post{}, fmt.Errorf("创建帖子失败")
	}
	return *post, nil
}
func ListPosts(page uint) ([]model.Post, error) {
	posts, err := dao.GetPostPage(page)
	if err != nil {
		return []model.Post{}, fmt.Errorf("获取帖子列表失败")
	}
	return posts, nil
}
func DeletePost(postId uint, role string, userId uint) error {
	post, err := dao.GetPostById(postId)
	if err != nil {
		return fmt.Errorf("帖子不存在")
	}
	authorId := post.AuthorId
	if role != "admin" && authorId != userId {
		return fmt.Errorf("无权限删除帖子")
	}
	if err := dao.DeletePost(postId); err != nil {
		return fmt.Errorf("删除帖子失败")
	}
	if err = dao.DeleteCommentsByPostId(postId); err != nil {
		return fmt.Errorf("删除帖子评论失败")
	}
	return nil

}
func GetDetailedPostById(postId uint) (model.Post, error) {
	post, err := dao.GetDetailedPostById(postId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Post{}, fmt.Errorf("帖子不存在")
	}
	if err != nil {
		return model.Post{}, fmt.Errorf("获取帖子详情失败")
	}
	return post, nil
}

// func GetPostAuthorId(postId uint) (uint, error) {
// 	post, err := GetPostbyId(postId)
// 	if err != nil {
// 		return 0, fmt.Errorf("获取帖子作者ID失败")
// 	}
// 	return post.AuthorId, nil
// }
// func GetPostbyId(postId uint) (model.Post, error) {
// 	post, err := dao.GetPostById(postId)
// 	if err != nil {
// 		return model.Post{}, fmt.Errorf("获取帖子失败")
// 	}
// 	return post, nil
// }
