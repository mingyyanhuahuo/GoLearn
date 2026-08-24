package service

import (
	"day_4_1/dao"
	"day_4_1/pkg/errcode"
)

func ToggleLike(postId, userId uint) error {
	if _, err := dao.GetPostById(postId); err != nil {
		return errcode.ErrNotFoundPost
	}
	err := dao.ToggleLike(postId, userId)
	if err != nil {
		return errcode.ErrDatabase
	}
	return nil
}
func GetLikePostIds(postids []uint, userId uint) ([]uint, error) {
	likePostIds, err := dao.GetLikePostIds(postids, userId)
	if err != nil {
		return []uint{}, errcode.ErrDatabase
	}
	return likePostIds, nil
}
