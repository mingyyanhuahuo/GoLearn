package service

import (
	"context"
	"day_4_1/dao"
	"day_4_1/model"
	"day_4_1/pkg/deepseek"
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
	DraftID   string `json:"draft_id"`
	Action    string `json:"action"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	ExpiresAt string `json:"expires_at"`
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
	sort := "time"
	if s, ok := args["sort"].(string); ok {
		sort = s
	}
	var posts []model.Post
	var err error
	if sort == "hot" {
		posts, _, err = dao.GetPostPageHot(page, 20)
	} else {
		posts, _, err = dao.GetPostPage(page, 20)
	}
	if err != nil {
		return nil, err
	}
	sortName := "时间"
	if sort == "hot" {
		sortName = "热门"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "当前是%s排序,查询到第 %d 页的帖子列表:\n", sortName, page)
	for index, post := range posts {
		fmt.Fprintf(&sb, "%d. %s (ID: %d) - 点赞数量: %d, 评论数: %d\n", index+1, post.Title, post.Id, post.LikeCount, post.CommentCount)
	}
	sb.WriteString("想换排序,可以直接说「热门」或「时间」")
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
	Action  string `json:"action"`
}

func (t *CreatePostTool) Call(ctx context.Context, args map[string]any) (any, error) {
	userId := uint(args["user_id"].(float64))
	title := args["title"].(string)
	if title == "" {
		title = "新帖子"
	}
	content := args["content"].(string)
	darftID := fmt.Sprintf("draft-%d-%d", userId, time.Now().UnixNano())
	key := "agent:draft:" + darftID
	expiresAt := time.Now().Add(5 * time.Minute).Format(time.RFC3339)
	payload, _ := json.Marshal(draftPayload{
		UserId:  userId,
		Title:   title,
		Content: content,
		Action:  "create_post",
	})
	if err := redisdb.Rdb.Set(ctx, key, payload, 5*time.Minute).Err(); err != nil {
		return nil, errcode.ErrInternalServerError
	}
	return &PendingAction{
		Action:    "create_post",
		Title:     title,
		Content:   content,
		ExpiresAt: expiresAt,
		DraftID:   darftID,
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
		return "", errcode.ErrDraftNotFound
	}
	var draft draftPayload
	if err := json.Unmarshal([]byte(payload), &draft); err != nil {
		return "", errcode.ErrDraftInvalid
	}
	if draft.UserId != userid {
		return "", errcode.ErrDraftInvalid
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
func parseToolCall(s string) (string, map[string]any, bool) {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end <= start {
		return "", nil, false
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(s[start:end+1]), &obj); err != nil {
		return "", nil, false
	}
	tool, ok := obj["tool"].(string)
	if !ok {
		return "", nil, false
	}
	args, _ := obj["args"].(map[string]any)
	return tool, args, true
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
	if deepseek.Enabled() {
		msgs := []deepseek.Message{
			{Role: "system", Content: `你是"招新助手",服务于精弘社团招新论坛。根据用户意图,选择调用工具或直接回复。

【工具调用】当且仅当用户意图匹配以下三种情况之一时,输出一个 JSON 对象。输出工具调用时,不要输出任何其他文字,不要用 markdown 代码块:
1. 用户想看帖子列表、翻页、分页、有哪些帖子 → {"tool":"get_posts","args":{"page":页码数字,"sort":"hot"或"time"}}
2. 用户想看某个帖子的详情、内容或评论 → {"tool":"get_post_detail","args":{"post_id":帖子ID数字}}
3. 用户想发帖、发一条、写帖子、写一条、发布内容 → {"tool":"create_post_draft","args":{"title":"从内容提炼的简短标题(20字内)","content":"用户想发布的完整内容"}}

【工具调用细节】
- args 里 page、post_id 必须是数字(不要加引号),content 必须是字符串。
- 用户没提页码时 page 默认填 1;提到"第X页"时填 X。
- 帖子ID从用户消息里提取;提取不到但前文提到过帖子,用前文的帖子ID。
- 用户表达发帖意图但内容不完整时,也要调用 create_post_draft,把已有的内容放进去,绝不能用文字回复代替工具调用。
- 调用 create_post_draft 时,必须从用户内容里提炼一个简短标题填进 args.title,不能省略。
- 用户确认草稿的操作在客户端完成;聊天中用户提到确认时,用文字回复确认说明即可,不要调用工具。
- get_posts 的 sort 参数:"hot"=按热门(点赞评论热度),"time"=按时间,不填默认 time。
- 用户要看列表但没说排序方式时,先用文字问「你想要热门排序还是按时间排序?」,等用户回答后再调用 get_posts。
【直接回复】其余情况(打招呼、咨询社团/招新活动、闲聊、感谢等),用热情的中文直接回复,可以适当使用 emoji 和换行,风格像社团招新宣传,200字以内,不要调用任何工具。例如用户问"介绍一下你们社团",直接介绍招新活动,不要调用工具。

【多轮对话】你有会话上下文,保持话题连贯,不要重复已回答过的内容。
`},
		}
		for _, h := range sees.History {
			msgs = append(msgs, deepseek.Message{
				Role:    h.Role,
				Content: h.Content,
			})
		}
		msgs = append(msgs, deepseek.Message{
			Role:    "user",
			Content: message,
		})
		resp, err := deepseek.Chat(msgs)
		if err == nil && resp != "" {
			if tool, args, ok := parseToolCall(resp); ok {
				switch tool {
				case "get_posts":
					page := uint(1)
					if p, ok := args["page"].(float64); ok {
						page = uint(p)
					}
					sort, _ := args["sort"].(string)
					if sort == "" {
						sort = "time"
					}
					result, err := s.tools["get_posts"].Call(context.Background(), map[string]any{
						"page": float64(page),
						"sort": sort,
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

				case "get_post_detail":
					if id, ok := args["post_id"].(float64); ok && id > 0 {
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
				case "create_post_draft":
					content, _ := args["content"].(string)
					title, _ := args["title"].(string)
					if content != "" && title != "" {
						result, err := s.tools["create_post_draft"].Call(context.Background(), map[string]any{
							"title":   title,
							"content": content,
							"user_id": float64(userid),
						})
						if err != nil {
							return &ChatResponse{}, err
						}
						pa := result.(*PendingAction)
						reply := fmt.Sprintf("我已为你生成了一个帖子草稿，标题: %s, 内容: %s\n确认发布请回复 confirm 并带上确认编号 %s（5 分钟内有效）。",
							pa.Title, pa.Content, pa.DraftID)
						s.appendHistory(sees, "user", message)
						s.appendHistory(sees, "assistant", reply)
						sees.LastTool = ""
						return &ChatResponse{
							SessionID:     sessionID,
							Reply:         reply,
							PendingAction: pa,
						}, nil
					}
				}
			} else {
				if strings.Contains(message, "发") || strings.Contains(message, "创建") || strings.Contains(message, "写一条") {
					runes := []rune(message)
					if len(runes) > 20 {
						runes = runes[:20]
					}
					title := string(runes)
					result, err := s.tools["create_post_draft"].Call(context.Background(), map[string]any{
						"title":   title,
						"content": message,
						"user_id": float64(userid),
					})
					if err != nil {
						return &ChatResponse{}, err
					}
					pa := result.(*PendingAction)
					reply := fmt.Sprintf("我已为你生成了一个帖子草稿，标题: %s, 内容: %s\n确认发布请回复 confirm 并带上确认编号 %s（5 分钟内有效）。",
						pa.Title, pa.Content, pa.DraftID)
					s.appendHistory(sees, "user", message)
					s.appendHistory(sees, "assistant", reply)
					sees.LastTool = ""
					return &ChatResponse{
						SessionID:     sessionID,
						Reply:         reply,
						PendingAction: pa,
					}, nil
				}
				s.appendHistory(sees, "user", message)
				s.appendHistory(sees, "assistant", resp)
				sees.LastTool = ""
				return &ChatResponse{
					SessionID:     sessionID,
					Reply:         resp,
					PendingAction: nil,
				}, nil
			}
		}
	}
	if strings.Contains(message, "发") || strings.Contains(message, "创建") || strings.Contains(message, "写一条") {
		runes := []rune(message)
		if len(runes) > 20 {
			runes = runes[:20]
		}
		title := string(runes)
		result, err := s.tools["create_post_draft"].Call(context.Background(), map[string]any{
			"title":   title,
			"content": message,
			"user_id": float64(userid),
		})
		if err != nil {
			return &ChatResponse{}, err
		}
		pa := result.(*PendingAction)
		reply := fmt.
			Sprintf("我已为你生成了一个帖子草稿，标题: %s, 内容: %s\n确认发布请回复 confirm 并带上确认编号 %s（5 分钟内有效）。",
				pa.Title, pa.Content, pa.DraftID)
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
		if strings.Contains(message, "热门") || strings.Contains(message, "时间") || strings.Contains(message, "最新") {
			sort := "time"
			if strings.Contains(message, "热门") {
				sort = "hot"
			}
			result, err := s.tools["get_posts"].Call(context.Background(), map[string]any{
				"page": float64(1),
				"sort": sort,
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
		sort := "time"
		if strings.Contains(message, "热门") {
			sort = "hot"
		}
		result, err := s.tools["get_posts"].Call(context.Background(), map[string]any{
			"page": float64(1),
			"sort": sort,
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
	reply := "📌【技术部暑期招新火热进行中】\n\n✨ 加入我们,你可以:\n- 参与真实项目,积累实战经验\n- 学习新技能,结识志同道合的伙伴\n- 一起办活动、搞创作、玩转校园\n\n🎯 我们希望你:有热情、有责任心、愿意尝试新事物\n\n📅 报名时间:即日起,招满为止\n📍 报名方式:评论区留言「我要报名」或私信获取报名表\n\n💬 你也可以直接告诉我:\n- 「看看帖子列表」→ 浏览招新帖子(可选热门或时间排序)\n- 「看看96号帖子详情」→ 查看帖子内容和评论\n- 「帮我写一条帖子」→ 让我帮你起草招新帖"
	s.appendHistory(sees, "user", message)
	s.appendHistory(sees, "assistant", reply)
	return &ChatResponse{
		SessionID:     sessionID,
		Reply:         reply,
		PendingAction: nil,
	}, nil
}
