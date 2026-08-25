package service

import (
	"context"
	"day_4_1/dao"
	"day_4_1/model"
	"day_4_1/pkg/errcode"
	"day_4_1/pkg/redisdb"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ChatMessage struct {
	Role    string
	Content string
}
type Session struct {
	History  []ChatMessage
	LastTool string
}
type AgentTool interface {
	Name() string
	Description() string
	Call(ctx context.Context, args map[string]any) (any, error)
}
type PendingAction struct {
	Action         string `json:"action"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	ExpiresAt      int64  `json:"expires_at"`
	ConfirmDraftID string `json:"confirm_draft_id"`
}
type ChatResponse struct {
	SessionID     string         `json:"session_id"`
	Reply         string         `json:"reply"`
	PendingAction *PendingAction `json:"pending_action"`
}
type AgentService struct {
	tools map[string]AgentTool
}

func NewAgentService() *AgentService {
	service := &AgentService{
		tools: make(map[string]AgentTool),
	}
	service.tools["get_posts"] = &GetPostsTool{}
	service.tools["get_post_detail"] = &GetPostDetailTool{}
	service.tools["create_post_draft"] = &CreatePostTool{}
	return service
}

type GetPostsTool struct {
}

func (t *GetPostsTool) Name() string {
	return "get_posts"
}
func (t *GetPostsTool) Description() string {
	return "查询分页帖子列表"
}
func (t *GetPostsTool) Call(ctx context.Context, args map[string]any) (any, error) {
	page := uint(1)
	if p, ok := args["page"].(float64); ok {
		page = uint(p)
	}
	posts, err := dao.GetPostPage(page)
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "查询到第 %d 页的帖子列表:\n", page)
	for index, post := range posts {
		fmt.Fprintf(&sb, "%d. %s (ID: %d) - 点赞数量: %d, 评论数: %d\n", index+1, post.Title, post.Id, post.LikeCount, post.CommentCount)
	}
	return sb.String(), nil
}

type GetPostDetailTool struct {
}

func (t *GetPostDetailTool) Name() string { return "get_post_detail" }
func (t *GetPostDetailTool) Description() string {
	return "按照id查询帖子详情以及评论"
}
func (t *GetPostDetailTool) Call(ctx context.Context, args map[string]any) (any, error) {
	id := uint(args["post_id"].(float64))
	post, err := dao.GetDetailedPostById(id)
	if err != nil {
		return nil, errcode.ErrNotFoundPost
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "<<%s>>\n作者: %s\n点赞数量: %d\n", post.Title, post.Author.Username, post.LikeCount)
	fmt.Fprintf(&sb, "内容: %s\n", post.Content)
	for index, comment := range post.Comments {
		fmt.Fprintf(&sb, "%d. %s (ID: %d) --- 评论内容: %s\n", index+1, comment.Author.Username, comment.Id, comment.Content)
		fmt.Fprintf(&sb, "   评论时间: %s\n", comment.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return sb.String(), nil
}

type CreatePostTool struct {
}

func (t *CreatePostTool) Name() string { return "create_post_draft" }
func (t *CreatePostTool) Description() string {
	return "创建帖子(需确认才可发布)"
}

type draftPayload struct {
	UserId  uint   `json:"user_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (t *CreatePostTool) Call(ctx context.Context, args map[string]any) (any, error) {
	userId := uint(args["user_id"].(float64))
	title := args["title"].(string)
	content := args["content"].(string)
	darftID := fmt.Sprintf("draft-%d-%d", userId, time.Now().UnixNano())
	key := "agent:draft:" + darftID
	expiresAt := time.Now().Add(5 * time.Minute).Unix()
	payload, _ := json.Marshal(draftPayload{
		UserId:  userId,
		Title:   title,
		Content: content,
	})
	if err := redisdb.Rdb.Set(ctx, key, payload, 5*time.Minute).Err(); err != nil {
		return nil, errcode.ErrInternalServerError
	}
	return &PendingAction{
		Action:         "create_post",
		Title:          title,
		Content:        content,
		ExpiresAt:      expiresAt,
		ConfirmDraftID: darftID,
	}, nil
}
func (s *AgentService) getSession(sessionID string) (*Session, error) {
	ctx := context.Background()
	key := "agent:session:" + sessionID
	payload, err := redisdb.Rdb.Get(ctx, key).Result()
	if err == nil {
		var sess Session
		if err := json.Unmarshal([]byte(payload), &sess); err == nil {
			return &sess, nil
		}
	}
	return &Session{}, nil
}
func (s *AgentService) saveSession(sessionID string, sess *Session) error {
	payload, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	return redisdb.Rdb.Set(context.Background(),
		"agent:session:"+sessionID, payload, 30*time.Minute).Err()
}
func (s *AgentService) appendHistory(sess *Session, role, content string) {
	sess.History = append(sess.History, ChatMessage{
		Role:    role,
		Content: content,
	})
	if len(sess.History) > 15 {
		sess.History = sess.History[len(sess.History)-15:]
	}
}
func (s *AgentService) confirmDraft(userid uint, draftID string) (string, error) {
	ctx := context.Background()
	key := "agent:draft:" + draftID
	payload, err := redisdb.Rdb.Get(ctx, key).Result()
	if err != nil {
		return "", errcode.New(404, 10009, "草稿不存在或已过期")
	}
	var draft draftPayload
	if err := json.Unmarshal([]byte(payload), &draft); err != nil {
		return "", errcode.New(404, 10009, "草稿数据解析失败")
	}
	if draft.UserId != userid {
		return "", errcode.New(404, 10009, "无权限确认此草稿")
	}
	post := &model.Post{
		Title:    draft.Title,
		Content:  draft.Content,
		AuthorId: userid,
	}

	if err := dao.GeneratePost(post); err != nil {
		return "", errcode.ErrDatabase
	}
	redisdb.Rdb.Del(ctx, key)
	return fmt.Sprintf("帖子《%s》已成功发布，ID: %d", post.Title, post.Id), nil

}
func extractPostId(msg string) uint {
	for _, word := range strings.Fields(msg) {
		if n, err := strconv.Atoi(word); err == nil {
			return uint(n)
		}
	}
	return 0
}

// mainchat
// ----------------------------
func (s *AgentService) Chat(userid uint, sessionID string, message string, confirmID string) (*ChatResponse, error) {
	sees, err := s.getSession(sessionID)
	if err != nil {
		return &ChatResponse{}, err
	}
	defer s.saveSession(sessionID, sees)
	if confirmID != "" {
		reply, err := s.confirmDraft(userid, confirmID)
		if err != nil {
			return &ChatResponse{}, err
		}
		s.appendHistory(sees, "user", "[confirm]"+confirmID)
		s.appendHistory(sees, "assistant", reply)
		return &ChatResponse{
			SessionID:     sessionID,
			Reply:         reply,
			PendingAction: nil,
		}, nil
	}
	if strings.Contains(message, "发") || strings.Contains(message, "创建") || strings.Contains(message, "写一条") {
		result, err := s.tools["create_post_draft"].Call(context.Background(), map[string]any{
			"title":   "新帖子",
			"content": message,
			"user_id": float64(userid),
		})
		if err != nil {
			return &ChatResponse{}, err
		}
		pa := result.(*PendingAction)
		reply := fmt.Sprintf("我已为你生成了一个帖子草稿，标题: %s, 内容: %s\n确认发布请回复 confirm 并带上确认编号 %s（5 分钟内有效）。", pa.Title, pa.Content, pa.ConfirmDraftID)
		s.appendHistory(sees, "user", message)
		s.appendHistory(sees, "assistant", reply)
		sees.LastTool = ""
		return &ChatResponse{
			SessionID:     sessionID,
			Reply:         reply,
			PendingAction: pa,
		}, nil
	}
	if strings.Contains(message, "详情") || strings.Contains(message, "查看") || strings.Contains(message, "内容") {
		if id := extractPostId(message); id > 0 {
			result, err := s.tools["get_post_detail"].Call(context.Background(), map[string]any{
				"post_id": float64(id),
			})
			if err != nil {
				return &ChatResponse{}, err
			}
			reply := result.(string)
			s.appendHistory(sees, "user", message)
			s.appendHistory(sees, "assistant", reply)
			sees.LastTool = "get_post_detail"
			return &ChatResponse{
				SessionID:     sessionID,
				Reply:         reply,
				PendingAction: nil,
			}, nil
		}
	}
	if sees.LastTool == "get_posts" {
		if strings.Contains(message, "页") {
			if page := extractPostId(message); page > 0 {
				result, err := s.tools["get_posts"].Call(context.Background(), map[string]any{
					"page": float64(page),
				})
				if err != nil {
					return &ChatResponse{}, err
				}
				reply := result.(string)
				s.appendHistory(sees, "user", message)
				s.appendHistory(sees, "assistant", reply)
				sees.LastTool = "get_posts"
				return &ChatResponse{
					SessionID:     sessionID,
					Reply:         reply,
					PendingAction: nil,
				}, nil
			}
		}
		if id := extractPostId(message); id > 0 {
			result, err := s.tools["get_post_detail"].Call(context.Background(), map[string]any{
				"post_id": float64(id),
			})
			if err != nil {
				return &ChatResponse{}, err
			}
			reply := result.(string)
			s.appendHistory(sees, "user", message)
			s.appendHistory(sees, "assistant", reply)
			sees.LastTool = "get_post_detail"
			return &ChatResponse{
				SessionID:     sessionID,
				Reply:         reply,
				PendingAction: nil,
			}, nil
		}
	}
	if strings.Contains(message, "有哪些") || strings.Contains(message, "列表") || strings.Contains(message, "分页") {
		result, err := s.tools["get_posts"].Call(context.Background(), map[string]any{
			"page": float64(1),
		})
		if err != nil {
			return &ChatResponse{}, err
		}
		reply := result.(string)
		s.appendHistory(sees, "user", message)
		s.appendHistory(sees, "assistant", reply)
		sees.LastTool = "get_posts"
		return &ChatResponse{
			SessionID:     sessionID,
			Reply:         reply,
			PendingAction: nil,
		}, nil
	}
	reply := "我是招新助手,可以帮你:1.查看帖子列表 2.查看帖子详情 3.起草发帖(需你确认)。试试说「看看帖子列表」?"
	s.appendHistory(sees, "user", message)
	s.appendHistory(sees, "assistant", reply)
	return &ChatResponse{
		SessionID:     sessionID,
		Reply:         reply,
		PendingAction: nil,
	}, nil
}
