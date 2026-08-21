package main

import (
	"day_3_1/dao"
	"day_3_1/hashcom"
	"day_3_1/model"
	"fmt"
	"log"

	// 用于密码哈希

	"gorm.io/driver/mysql"
	"gorm.io/gorm" // Gorm 框架
)

func initDB() *gorm.DB {
	// 数据库连接信息
	dsn := "root:1433223@tcp(localhost:3306)/post_model?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	return db
}

var UserList []model.User = []model.User{
	{Name: "Alice", PhoneNumber: "12345678901", PassHash: "12155", Role: "user"},
	{Name: "Bob", PhoneNumber: "12345678902", PassHash: "12155", Role: "user"},
	{Name: "Charlie", PhoneNumber: "12345678903", PassHash: "12155", Role: "user"},
	{Name: "David", PhoneNumber: "12345678904", PassHash: "12155", Role: "admin"},
}
var CommentList []model.Comment = []model.Comment{
	{AuthorId: 1, PostId: 1, Content: "Great post!"},
	{AuthorId: 2, PostId: 1, Content: "Thanks for sharing."},
	{AuthorId: 3, PostId: 2, Content: "Interesting read."},
	{AuthorId: 4, PostId: 3, Content: "I learned"},
	{AuthorId: 1, PostId: 4, Content: "I disagree with your point."},
}
var PostList []model.Post = []model.Post{
	{Title: "First Post", Content: "This is the content of the first post.", AuthorId: 1},
	{Title: "Second Post", Content: "This is the content of the second post.", AuthorId: 2},
	{Title: "Third Post", Content: "This is the content of the third post.", AuthorId: 3},
	{Title: "Fourth Post", Content: "This is the content of the fourth post.", AuthorId: 4},
}

func main() {
	db := initDB()
	dao.InitDB(db)
	if err := db.AutoMigrate(&model.User{}, &model.Post{}, &model.Comment{}, &model.Like{}); err != nil {
		log.Fatalf("自动迁移失败: %w", err)
	} else {
		fmt.Println("数据库连接成功")
	}

	for _, user := range UserList {
		hashedPassword, err := hashcom.RegisterHashPassword(user.PassHash)
		if err != nil {
			log.Printf("密码哈希失败: %w", err)
			continue
		}
		user.PassHash = hashedPassword
		dao.CreateUser(&user)
	}
	// for _, comment := range CommentList {
	// 	err := dao.GenerateComment(&comment)
	// 	if err != nil {
	// 		log.Printf("生成评论失败: %w", err)
	// 	}
	// }
	// for _, post := range PostList {
	// 	err := dao.GeneratePost(&post)
	// 	if err != nil {
	// 		log.Printf("生成帖子失败: %w", err)
	// 	}
	// }
	// posts, error := dao.GetPostPage(2)
	// if error != nil {
	// 	log.Printf("获取帖子分页失败: %w", error)
	// } else {
	// 	for _, post := range posts {
	// 		fmt.Println("Post ID: ", post.Id, ", Title: ", post.Title, ", Author: ", post.Author.Name)
	// 	}
	// }
	// 在这里可以使用 db 进行数据库操作
	// 正常版:评论插进去
	normal := &model.Comment{PostId: 1, AuthorId: 1, Content: "正常评论"}
	err := dao.GenerateComment(normal)
	fmt.Println("正常版评论 id:", normal.Id) // 有 id = 成功

	// 故障版:postId 改成 9999(不存在的帖子),计数那步必失败 → 评论被回滚
	fault := &model.Comment{PostId: 9999, AuthorId: 1, Content: "应该被回滚"}
	err = dao.GenerateComment(fault)
	fmt.Println("故障版错误:", err)
	fmt.Println("故障版评论 id:", fault.Id) // 0 = 回滚生效 ✅

	hashword, _ := hashcom.RegisterHashPassword("12155")
	erro := hashcom.
		LogincompareHashAndPassword(hashword, "12155")
	if erro != nil {
		log.Printf("密码验证失败: %w", erro)
	} else {
		fmt.Println("密码验证成功")
	}
}
