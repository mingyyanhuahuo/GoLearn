# GoLearn — 精弘网络大作业(Go 后端)


#打开docs可以查看我的开发历程以及开发的学习安排
#当然啦这份read.me和developmnet是我拿ai总结的

招新社团大作业:基于 Gin + GORM + Redis 的论坛后端,**11 个接口全部完成 + 4 项进阶全部落地**。

## 技术栈

| 组件 | 用途 |
|---|---|
| Go 1.26 | 语言 |
| Gin | HTTP 框架(路由 / 中间件 / 参数绑定) |
| GORM + MySQL | ORM 与数据库(users / posts / comments / likes) |
| go-redis (v9) | Redis:点赞计数、Agent 会话与草稿 |
| Viper | yaml 配置读取 |
| golang-jwt/v5 | 登录鉴权(HS256) |
| zap + lumberjack | 日志文件按大小轮转 |

## 快速启动

1. 启动 MySQL 与 Redis(默认 localhost:3306 / 6379)
2. 编辑 `config/config.yaml`(数据库密码、jwt secret 等)
3. 运行 `go run .`

> `config/config.yaml` 含密码,已加入 `.gitignore` 不进仓库。

## 接口清单(11 个)

| # | 方法 | 路径 | 说明 | 鉴权 |
|---|---|---|---|---|
| 1 | POST | `/api/v1/register` | 注册 | 无 |
| 2 | POST | `/api/v1/login` | 登录,返回 JWT | 无 |
| 3 | GET | `/api/v1/posts` | 帖子列表,`page` 分页、`sort=hot` 热门排序 | 无 |
| 4 | GET | `/api/v1/posts/:post_id` | 帖子详情(作者 + 评论列表) | 无 |
| 5 | POST | `/api/v1/posts` | 发帖 | ✅ |
| 6 | DELETE | `/api/v1/posts/:post_id` | 删帖(作者本人) | ✅ |
| 7 | POST | `/api/v1/comments` | 评论 | ✅ |
| 8 | DELETE | `/api/v1/comments/:comment_id` | 删评论(作者本人) | ✅ |
| 9 | POST | `/api/v1/posts/:post_id/like` | 点赞 / 取消点赞切换 | ✅ |
| 10 | POST | `/api/v1/posts/likes` | 批量查询点赞状态(post_ids 数组,一次 IN 查询) | ✅ |
| 11 | POST | `/api/v1/agent/chat` | Agent 多轮对话 | ✅ |

管理员:`DELETE /api/v1/admin/posts/:post_id`(Auth + Admin 双中间件,可删任意帖)。

## 四项进阶设计说明

### 1. 全局错误处理 + 日志文件

- `errcode` 统一错误码:`New(httpStatus, code, msg)`,错误自带 HTTP 状态码与业务码,handler 层不再手猜状态码
- 全局错误中间件统一响应格式,日志通过 zap 写入 `logs/app.log`,lumberjack 按大小轮转

### 2. Agent 服务(接口 11,规则 Agent 实现)

需求原文为 Eino 方案(选做);按计划文档附录 B 的降级授权,用**手写规则 Agent** 实现同等能力,不引入额外框架:

- **多轮会话**:`session_id` 隔离会话,历史 JSON 持久化到 Redis(`agent:session:{id}`,30 分钟滑动 TTL)
- **上下文续接**:`LastTool` 记录最近工具,无关键词消息自动接续(「第 2 页」接着翻页、只说数字查详情)
- **查询类工具**:`get_posts` / `get_post_detail` 后端执行、结果回填对话
- **写操作安全**:发帖请求只生成草稿(`agent:draft:{id}`,TTL 300s),返回 `pending_action` + `confirm_draft_id`,用户二次确认后才落库,确认即销毁草稿防重复发帖;TTL 到期自动过期
- 工具层通过 `AgentTool` 接口(`Name/Description/Call`)抽象,后续接入真实 LLM(Eino)时只需替换意图判断层,工具与持久化不动

### 3. Redis 点赞优化,定时落库

- 点赞状态实时走 Redis(`SADD`/`SREM` + 原子计数),读接口零数据库压力
- `LikeSyncer` 后台协程定时对账落库 MySQL,进程退出时优雅关闭兜底同步
- 对账归零保证 Redis 与 MySQL 计数最终一致

### 4. 热门排序(时间衰减)

- `GET /posts?sort=hot`:热度 = 点赞 / 评论加权 × 时间衰减,新帖短暂靠前、随时间自然回落
- 与默认时间倒序(`sort=latest`)并存,设计见 `04-进阶要求设计方案.md` §4

## 目录结构

```text
4/1/
├── main.go          # 启动:配置 / 日志 / Redis / 路由
├── config/          # Viper 配置读取(config.yaml 不入库)
├── router/          # 路由注册
├── middleware/      # Auth / Admin 中间件
├── handler/         # HTTP 层:绑定参数、响应
├── service/         # 业务层:规则、Agent 会话/工具、点赞同步
├── dao/             # GORM 数据访问
├── model/           # 数据模型
└── pkg/             # errcode / response / redisdb / jwt
```

## 测试

测试脚本在 `tests/` 目录,均为**幂等脚本**(每次运行随机生成全新测试用户,可反复执行):

```bash
# 1. 先启动服务(go run .),然后:
python tests/full_regression.py   # 全功能回归 45/45:注册/登录/帖子/评论/点赞/删除/权限/admin/Agent
bash tests/curl_biz_test.sh       # 业务 curl 全流程:11 接口走查 + Agent(列表→翻页→草稿→确认→重复确认404)
```

实测结果:全功能回归 **45/45 通过**,Agent 业务 curl 全流程(列表 → 多轮续接翻页 → 详情 → 草稿生成 → 二次确认落库 → 重复确认 404 → 未登录 401)全部通过。
