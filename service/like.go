package service

import (
	"day_4_1/dao"
	"fmt"
)

func ToggleLike(postId, userId uint) error {
	if _, err := dao.GetPostById(postId); err != nil {
		return fmt.Errorf("帖子不存在")
	}
	err := dao.ToggleLike(postId, userId)
	if err != nil {
		return fmt.Errorf("点赞/取消点赞失败")
	}
	return nil
}
func GetLikePostIds(postids []uint, userId uint) ([]uint, error) {
	likePostIds, err := dao.GetLikePostIds(postids, userId)
	if err != nil {
		return []uint{}, fmt.Errorf("获取点赞帖子ID失败")
	}
	return likePostIds, nil
}
