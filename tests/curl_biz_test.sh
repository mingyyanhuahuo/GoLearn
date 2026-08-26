#!/bin/bash
# ============================================================
# 业务 curl 全流程测试(对应 01-功能清单 11 接口)
# 用法:bash curl_biz_test.sh
# 重置:运行前自动清空数据库(仅保留正式账号 stu001 及其帖子)和 Redis 残留
# 幂等性:每次运行使用随机用户名,可重复执行
# 中文安全:请求体写 UTF-8 文件 + curl -d @file(不经过代码页转换)
# ============================================================
BASE=http://localhost:8080/api/v1
# 脚本目录定位(从任意位置运行都找得到 config.yaml)
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
DIR=/tmp/curl_biz
mkdir -p $DIR
PJ() { python -c "import sys,json;sys.stdout.reconfigure(encoding='utf-8');d=json.loads(sys.stdin.buffer.read().decode('utf-8'));print($1)"; }

# ===== 运行前重置数据库 / Redis(密码从本地 config.yaml 读取,不进仓库) =====
MYSQL_EXE=$(command -v mysql || echo "/c/Program Files/MySQL/MySQL Server 8.0/bin/mysql.exe")
REDIS_CLI=$(command -v redis-cli || echo "/c/Users/22254/Redis-x64-5.0.14.1/redis-cli.exe")
MYSQL_PWD=$(grep -E '^\s+password:' "$SCRIPT_DIR/../config/config.yaml" | head -1 | awk '{print $2}')
MYSQL_DB=$(grep -E '^\s+database:' "$SCRIPT_DIR/../config/config.yaml" | head -1 | awk '{print $2}')
echo "########## 0. 重置数据库(仅保留 stu001) ##########"
MYSQL_PWD="$MYSQL_PWD" "$MYSQL_EXE" -uroot -e "
USE $MYSQL_DB;
SET @sid := (SELECT id FROM users WHERE username='stu001' LIMIT 1);
DELETE FROM likes WHERE user_id <> COALESCE(@sid, -1) OR post_id NOT IN (SELECT id FROM posts WHERE author_id=COALESCE(@sid, -1));
DELETE FROM comments WHERE author_id <> COALESCE(@sid, -1) OR post_id NOT IN (SELECT id FROM posts WHERE author_id=COALESCE(@sid, -1));
DELETE FROM posts WHERE author_id <> COALESCE(@sid, -1);
DELETE FROM users WHERE username <> 'stu001';" && echo "数据库已重置"
"$REDIS_CLI" --scan --pattern "post:likes:*" 2>/dev/null | xargs -r "$REDIS_CLI" del >/dev/null 2>&1
"$REDIS_CLI" --scan --pattern "agent:*" 2>/dev/null | xargs -r "$REDIS_CLI" del >/dev/null 2>&1
"$REDIS_CLI" --scan --pattern "rate_limit:*" 2>/dev/null | xargs -r "$REDIS_CLI" del >/dev/null 2>&1

# 随机后缀:注册全新用户,可重复跑
TS=$(date +%s)
U1="curl_user${TS}1"
U2="curl_user${TS}2"
echo "== 本次测试用户:$U1 / $U2 =="

echo "########## ① 注册 POST /api/v1/register ##########"
cat > $DIR/reg1.json <<EOF
{"username":"$U1","password":"123456","phone":"137${TS: -8}1"}
EOF
cat > $DIR/reg2.json <<EOF
{"username":"$U2","password":"123456","phone":"137${TS: -8}2"}
EOF
echo "-- 注册用户1 --"
curl -s -X POST $BASE/register -H "Content-Type: application/json" -d @$DIR/reg1.json; echo
echo "-- 注册用户2 --"
curl -s -X POST $BASE/register -H "Content-Type: application/json" -d @$DIR/reg2.json; echo
echo "-- 重复注册(应400) --"
curl -s -X POST $BASE/register -H "Content-Type: application/json" -d @$DIR/reg1.json; echo

echo; echo "########## ② 登录 POST /api/v1/login ##########"
cat > $DIR/login1.json <<EOF
{"username":"$U1","password":"123456"}
EOF
cat > $DIR/login2.json <<EOF
{"username":"$U2","password":"123456"}
EOF
echo "-- 登录用户1 --"
RESP=$(curl -s -X POST $BASE/login -H "Content-Type: application/json" -d @$DIR/login1.json)
echo "$RESP"; echo
TOKEN1=$(echo "$RESP" | PJ "d['data']['token']")
echo "-- 登录用户2 --"
RESP=$(curl -s -X POST $BASE/login -H "Content-Type: application/json" -d @$DIR/login2.json)
echo "$RESP"; echo
TOKEN2=$(echo "$RESP" | PJ "d['data']['token']")
echo "-- 错误密码(应400) --"
cat > $DIR/loginbad.json <<EOF
{"username":"$U1","password":"wrong"}
EOF
curl -s -X POST $BASE/login -H "Content-Type: application/json" -d @$DIR/loginbad.json; echo

echo; echo "########## ③ 发帖 POST /api/v1/posts ##########"
cat > $DIR/post1.json <<'EOF'
{"title":"业务回归帖1","content":"招新报名火热进行中,欢迎加入精弘网络技术部"}
EOF
cat > $DIR/post2.json <<'EOF'
{"title":"业务回归帖2","content":"技术部等你来,周五面试不见不散"}
EOF
echo "-- 用户1发帖 --"
RESP=$(curl -s -X POST $BASE/posts -H "Authorization: Bearer $TOKEN1" -H "Content-Type: application/json" -d @$DIR/post1.json)
echo "$RESP"; echo
PID1=$(echo "$RESP" | PJ "d['data']['id']")
echo "-- 用户2发帖 --"
RESP=$(curl -s -X POST $BASE/posts -H "Authorization: Bearer $TOKEN2" -H "Content-Type: application/json" -d @$DIR/post2.json)
echo "$RESP"; echo
PID2=$(echo "$RESP" | PJ "d['data']['id']")
echo "-- 空标题(应400) --"
cat > $DIR/postbad.json <<'EOF'
{"title":"","content":"x"}
EOF
curl -s -X POST $BASE/posts -H "Authorization: Bearer $TOKEN1" -H "Content-Type: application/json" -d @$DIR/postbad.json; echo

echo; echo "########## ④ 帖子列表 GET /api/v1/posts ##########"
curl -s "$BASE/posts?page=1" | PJ "json.dumps(d['data'],ensure_ascii=False)[:300]"; echo
echo "-- 热门排序 sort=hot --"
curl -s "$BASE/posts?sort=hot" | PJ "d['code']"; echo
echo "-- 页码0(应400) --"
curl -s "$BASE/posts?page=0" | PJ "d['code']"; echo

echo; echo "########## ⑤ 帖子详情 GET /api/v1/posts/:id ##########"
curl -s "$BASE/posts/$PID1" | PJ "json.dumps(d['data'],ensure_ascii=False)[:400]"; echo
echo "-- 不存在的帖(应404) --"
curl -s "$BASE/posts/999999" | PJ "d['code']"; echo

echo; echo "########## ⑨ 评论 POST /api/v1/comments(挂在帖子下) ##########"
cat > $DIR/comment.json <<EOF
{"content":"已报名,期待面试","post_id":$PID2}
EOF
echo "-- 用户1评论用户2的帖 --"
curl -s -X POST $BASE/comments -H "Authorization: Bearer $TOKEN1" -H "Content-Type: application/json" -d @$DIR/comment.json; echo
echo "-- 评论不存在的帖(应404) --"
cat > $DIR/commentbad.json <<'EOF'
{"content":"x","post_id":999999}
EOF
curl -s -X POST $BASE/comments -H "Authorization: Bearer $TOKEN1" -H "Content-Type: application/json" -d @$DIR/commentbad.json; echo

echo; echo "########## ⑦ 点赞/取消 POST /api/v1/posts/:id/like ##########"
echo "-- 点赞(应 is_liked:true) --"
curl -s -X POST $BASE/posts/$PID1/like -H "Authorization: Bearer $TOKEN1" | PJ "json.dumps(d['data'],ensure_ascii=False)"; echo
echo "-- 再点(应 is_liked:false) --"
curl -s -X POST $BASE/posts/$PID1/like -H "Authorization: Bearer $TOKEN1" | PJ "json.dumps(d['data'],ensure_ascii=False)"; echo
echo "-- 点赞不存在的帖(应404) --"
curl -s -X POST $BASE/posts/999999/like -H "Authorization: Bearer $TOKEN1" | PJ "d['code']"; echo

echo; echo "########## ⑧ 批量点赞 POST /api/v1/posts/likes ##########"
cat > $DIR/likes.json <<EOF
{"post_ids":[$PID1,$PID2]}
EOF
curl -s -X POST $BASE/posts/likes -H "Authorization: Bearer $TOKEN1" -H "Content-Type: application/json" -d @$DIR/likes.json; echo

echo; echo "########## ⑥ 删帖 DELETE /api/v1/posts/:id ##########"
echo "-- 用户1删用户2的帖(应403) --"
curl -s -X DELETE $BASE/posts/$PID2 -H "Authorization: Bearer $TOKEN1" | PJ "d['code']"; echo
echo "-- 用户2删自己的帖(应200) --"
curl -s -X DELETE $BASE/posts/$PID2 -H "Authorization: Bearer $TOKEN2" | PJ "d['code']"; echo

echo; echo "########## ⑩ admin DELETE /api/v1/admin/posts/:id ##########"
echo "-- 普通用户访问(应403) --"
curl -s -X DELETE $BASE/admin/posts/$PID1 -H "Authorization: Bearer $TOKEN1" | PJ "d['code']"; echo

echo; echo "########## ⑪ Agent POST /api/v1/agent/chat ##########"
cat > $DIR/msg1.json <<'EOF'
{"session_id":"biz1","message":"看看帖子列表"}
EOF
cat > $DIR/msg2.json <<'EOF'
{"session_id":"biz1","message":"第 2 页"}
EOF
cat > $DIR/msg4.json <<'EOF'
{"session_id":"biz1","message":"帮我发一条招新帖:业务测试草稿"}
EOF
echo "-- 列表 --"
curl -s -X POST $BASE/agent/chat -H "Authorization: Bearer $TOKEN1" -H "Content-Type: application/json" -d @$DIR/msg1.json | PJ "json.dumps(d['data'],ensure_ascii=False)[:150]"; echo
echo "-- 多轮续接:第 2 页 --"
curl -s -X POST $BASE/agent/chat -H "Authorization: Bearer $TOKEN1" -H "Content-Type: application/json" -d @$DIR/msg2.json | PJ "json.dumps(d['data'],ensure_ascii=False)[:150]"; echo
echo "-- 发草稿(记 confirm_draft_id) --"
RESP=$(curl -s -X POST $BASE/agent/chat -H "Authorization: Bearer $TOKEN1" -H "Content-Type: application/json" -d @$DIR/msg4.json)
echo "$RESP" | PJ "json.dumps(d['data'],ensure_ascii=False)[:250]"; echo
DRAFT=$(echo "$RESP" | PJ "d['data']['pending_action']['confirm_draft_id']")
echo "-- 确认发布 --"
cat > $DIR/confirm.json <<EOF
{"session_id":"biz1","message":"confirm","confirm_draft_id":"$DRAFT"}
EOF
curl -s -X POST $BASE/agent/chat -H "Authorization: Bearer $TOKEN1" -H "Content-Type: application/json" -d @$DIR/confirm.json | PJ "json.dumps(d['data'],ensure_ascii=False)[:150]"; echo
echo "-- 重复确认(应404) --"
curl -s -X POST $BASE/agent/chat -H "Authorization: Bearer $TOKEN1" -H "Content-Type: application/json" -d @$DIR/confirm.json | PJ "d['code']"; echo
echo "-- 未登录(应401) --"
curl -s -X POST $BASE/agent/chat -H "Content-Type: application/json" -d @$DIR/msg1.json | PJ "d['code']"; echo

echo; echo "########## 收尾:用户1删自己的帖1(应200) ##########"
curl -s -X DELETE $BASE/posts/$PID1 -H "Authorization: Bearer $TOKEN1" | PJ "d['code']"; echo
echo "========== 全部结束 =========="
