package service

import (
	"day_4_1/dao"
	"day_4_1/pkg/errcode"
)

type LikeStatusItem struct {
	PostId uint `json:"post_id"`
	Liked  bool `json:"liked"`
}

func ToggleLike(postId, userId uint) (bool, error) {
	if _, err := dao.GetPostById(postId); err != nil {
		return false, errcode.ErrNotFoundPost
	}
	liked, err := dao.ToggleLike(postId, userId)
	if err != nil {
		return false, errcode.ErrDatabase
	}
	return liked, nil
}

func GetLikeStatus(postids []uint, userId uint) ([]LikeStatusItem, error) {
	valid, err := dao.GetExistingLikePostIds(postids)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if len(valid) == 0 {
		return []LikeStatusItem{}, nil
	}
	likeMap, err := dao.GetLikesStatus(valid, userId)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	items := make([]LikeStatusItem, 0, len(valid))
	for _, id := range valid {
		items = append(items, LikeStatusItem{
			PostId: id,
			Liked:  likeMap[id],
		})
	}
	return items, nil
}
