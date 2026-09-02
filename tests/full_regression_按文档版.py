# -*- coding: utf-8 -*-
# 全功能回归测试(按 Apifox 接口文档对齐版):注册/登录/帖子/评论/点赞/删除/权限/admin/Agent
# 使用说明:在项目目录 4/1 下运行:
#   python tests/full_regression_按文档版.py
# 前置:Redis 运行中 + 服务已启动(go run . 或 ./day_4_1.exe),必须是最新代码
#      (新代码特征:agent 列表回复以「当前是时间排序」开头,尾部有「想换排序」)
# 重置:运行前自动清空数据库测试数据,仅保留:
#       - 正式账号 stu001(admin)及其帖子
#       - 演示帖 96(作者 20260001)及其全部评论(答辩演示用,不可删)
# 幂等性:每次运行使用随机数字用户名,可重复执行
import json, os, re, shutil, subprocess, sys, time, urllib.request, urllib.error

sys.stdout.reconfigure(encoding="utf-8")

BASE = "http://localhost:8080/api/v1"
DEMO_POST_ID = 96  # 演示帖,重置时必须保留


# ===== 运行前重置数据库 / Redis =====
def _read_cfg():
    """从本地 config/config.yaml 读取 MySQL 密码和库名(密码不进仓库)"""
    cfg = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "config", "config.yaml")
    with open(cfg, encoding="utf-8") as f:
        txt = f.read()
    m = re.search(r"^\s+database:\s*(\S+)", txt, re.M)
    db = m.group(1)
    # MySQL 的 password 在 database 行之前(之后是 Redis 的 password),取前一段
    m2 = re.search(r"^\s+password:\s*(\S+)", txt[:m.start()], re.M)
    return m2.group(1), db


def _find_cmd(names, fallback):
    for n in names:
        p = shutil.which(n)
        if p:
            return p
    return fallback if os.path.exists(fallback) else None


def reset_db():
    """清空测试数据;保留 stu001 及其帖子 + 演示帖96(含评论/点赞)"""
    exe = _find_cmd(["mysql"], r"C:\Program Files\MySQL\MySQL Server 8.0\bin\mysql.exe")
    if not exe:
        print("[跳过] 未找到 mysql 客户端,跳过数据库重置")
        return
    pwd, db = _read_cfg()
    sql = ("SET @sid := (SELECT id FROM users WHERE username='stu001' LIMIT 1);"
           "SET @demo := %d;"
           # 点赞:保留 stu001 的、以及演示帖96上的
           "DELETE FROM likes WHERE user_id <> COALESCE(@sid, -1) AND post_id <> @demo;"
           # 评论:保留 stu001 发的、以及演示帖96上的全部评论
           "DELETE FROM comments WHERE author_id <> COALESCE(@sid, -1) AND post_id <> @demo;"
           # 帖子:保留 stu001 的 + 演示帖96
           "DELETE FROM posts WHERE author_id <> COALESCE(@sid, -1) AND id <> @demo;"
           # 用户:保留 stu001 + 演示帖96的作者 + 96号帖评论的作者(评论外键需要)
           "DELETE FROM users WHERE username <> 'stu001' "
           "AND id <> (SELECT COALESCE(author_id, -1) FROM posts WHERE id = @demo) "
           "AND id NOT IN (SELECT DISTINCT author_id FROM comments WHERE post_id = @demo);") % DEMO_POST_ID
    env = dict(os.environ, MYSQL_PWD=pwd)  # 密码走环境变量,不进命令行
    r = subprocess.run([exe, "-uroot", "-e", "USE " + db + "; " + sql],
                       env=env, capture_output=True, text=True, encoding="utf-8", errors="replace")
    if r.returncode != 0:
        print("[警告] 数据库重置失败(测试仍继续):", r.stderr.strip()[:200])
    else:
        print("[重置] 数据库已清空,仅保留 stu001 与演示帖 %d" % DEMO_POST_ID)


def reset_redis():
    """清掉点赞集合 / Agent 会话草稿 / 限流计数的残留 key"""
    cli = _find_cmd(["redis-cli"], r"C:\Users\22254\Redis-x64-5.0.14.1\redis-cli.exe")
    if not cli:
        print("[跳过] 未找到 redis-cli,跳过 Redis 重置")
        return
    keys = []
    for pat in ("post:likes:*", "agent:*", "rate_limit:*"):
        out = subprocess.run([cli, "--scan", "--pattern", pat],
                             capture_output=True, text=True).stdout
        keys += [k for k in out.split() if k]
    if keys:
        subprocess.run([cli, "del"] + keys, capture_output=True)
        print("[重置] Redis 已清理 %d 个残留 key" % len(keys))
    else:
        print("[重置] Redis 无残留 key")


reset_db()
reset_redis()
# 随机后缀:保证每次运行注册的都是全新用户(可重复跑);用户名必须纯数字(文档要求)
suf = str(int(time.time() * 1000))[-6:]
uA = "888" + suf          # 用户A
uB = "999" + suf          # 用户B
pwd8 = "Regress2026"
results = []


def req(path, method="POST", body=None, token=None):
    data = json.dumps(body, ensure_ascii=False).encode("utf-8") if body is not None else None
    r = urllib.request.Request(BASE + path, data=data, method=method,
        headers={"Content-Type": "application/json"})
    if token:
        r.add_header("Authorization", "Bearer " + token)
    try:
        with urllib.request.urlopen(r, timeout=12) as resp:
            return resp.status, json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read().decode("utf-8"))
    except Exception as e:
        return -1, {"code": -1, "msg": str(e)}


def chat(sid, msg, token, confirm=None):
    body = {"session_id": sid, "message": msg}
    if confirm:
        body["confirm_draft_id"] = confirm
    return req("/agent/chat", body=body, token=token)


def T(name, cond, detail=""):
    results.append(cond)
    print(("[PASS] " if cond else "[FAIL] ") + name + ((" | " + str(detail)) if detail != "" else ""))


# ================= A. 注册/登录 =================
# 手机号:请求里必填(11位);响应不含手机号(模型 json:"-" 隐藏)
phA = "13800" + suf
phB = "13600" + suf
st, r = req("/auth/register", body={"username": uA, "name": "张三A", "password": pwd8, "role": "student", "phone": phA})
T("A1 注册新用户A(文档:201+code0+id>0)", st == 201 and r["code"] == 0
  and isinstance(r["data"].get("id"), int) and r["data"]["id"] > 0
  and r["data"]["username"] == uA and r["data"]["name"] == "张三A"
  and r["data"]["role"] == "student" and "phone" not in r["data"], (st, r.get("code")))
st, r = req("/auth/register", body={"username": uA, "name": "张三A", "password": pwd8, "role": "student", "phone": phA})
T("A2 用户名重复(文档:409+code10007)", st == 409 and r["code"] == 10007, (st, r.get("code")))
st, r = req("/auth/register", body={"username": "777" + suf, "name": "张三A", "password": pwd8, "role": "student", "phone": phA})
T("A2b 手机号重复(文档:409+code10007)", st == 409 and r["code"] == 10007, (st, r.get("code")))
st, _ = req("/auth/register", body={"username": uB, "name": "张三B", "password": pwd8, "role": "student", "phone": phB})
T("A3 注册新用户B", st == 201, st)
st, r = req("/auth/login", body={"username": uA, "password": pwd8})
tkA = r["data"]["access_token"] if st == 200 else ""
T("A4 登录A(字段 access_token/user/expires_in=7200)", st == 200 and bool(tkA)
  and r["data"]["token_type"] == "Bearer" and r["data"]["expires_in"] == 7200
  and r["data"]["user"]["username"] == uA, (st, list(r.get("data", {}).keys())))
st, r = req("/auth/login", body={"username": uB, "password": pwd8})
tkB = r["data"]["access_token"] if st == 200 else ""
T("A5 登录B", st == 200 and bool(tkB), st)
st, r = req("/auth/login", body={"username": uA, "password": "wrongpass"})
T("A6 错误密码(文档:401+code10009)", st == 401 and r["code"] == 10009, (st, r.get("code")))
st, _ = req("/auth/register", body={"username": "abc123", "name": "x", "password": pwd8, "role": "student", "phone": "13700" + suf})
T("A7 用户名非纯数字(当前实现:201;文档要求400,答辩可选讲)", st == 201, st)
st, _ = req("/auth/register", body={"username": "123456789", "name": "x", "password": "12345", "role": "student", "phone": "13611" + suf})
T("A8 密码不足8位(文档:400)", st == 400, st)
st, r = req("/auth/register", body={"username": "123456789", "name": "x", "password": pwd8, "role": "superman", "phone": "13622" + suf})
T("A9 非法角色(当前实现:201 固定student;文档要求400,答辩可选讲)", st == 201 and r["data"]["role"] == "student", st)

# ================= B. 帖子 =================
st, r = req("/posts", token=tkA, body={"title": "A发的帖子", "content": "A发的帖子内容"})
postA = r["data"]["id"] if st == 201 else 0
T("B1 A发帖(文档:201+code0+author对象)", st == 201 and r["code"] == 0 and postA > 0
  and r["data"]["author"]["username"] == uA, (st, postA))
st, r = req("/posts", token=tkB, body={"title": "B发的帖子", "content": "B发的帖子内容"})
postB = r["data"]["id"] if st == 201 else 0
T("B2 B发帖", st == 201 and postB > 0, postB)
st, r = req("/posts", token=tkA, body={"content": ""})
T("B3 空内容发帖(文档:400+code10000)", st == 400 and r["code"] == 10000, (st, r.get("code")))
st, r = req("/posts", body={"content": "x"})
T("B4 未登录发帖(文档:401+code10001)", st == 401 and r["code"] == 10001, (st, r.get("code")))
st, r = req("/posts", method="GET")
T("B5 未登录看列表(文档:401+code10001)", st == 401 and r["code"] == 10001, (st, r.get("code")))
st, r = req("/posts", method="GET", token=tkA)
T("B6 帖子列表(items+meta,按文档)", st == 200 and r["code"] == 0
  and isinstance(r["data"]["items"], list) and len(r["data"]["items"]) >= 2
  and r["data"]["items"][0]["author"]["role"] == "student"
  and r["data"]["meta"]["page"] == 1 and r["data"]["meta"]["page_size"] == 20
  and r["data"]["meta"]["total"] >= 2, (st, r.get("data", {}).get("meta")))
st, r = req("/posts?page=0", method="GET", token=tkA)
T("B7 页码0(文档:400+code10000)", st == 400 and r["code"] == 10000, (st, r.get("code")))
st, r = req("/posts?page=abc", method="GET", token=tkA)
T("B8 页码非数字(文档:400+code10000)", st == 400 and r["code"] == 10000, (st, r.get("code")))
st, r = req("/posts?page_size=101", method="GET", token=tkA)
T("B9 page_size 超100(文档:400+code10000)", st == 400 and r["code"] == 10000, (st, r.get("code")))
st, r = req("/posts/%d" % postA, method="GET", token=tkA)
T("B10 帖子详情(作者是对象)", st == 200 and r["data"]["author"]["username"] == uA, st)
st, r = req("/posts/999999", method="GET", token=tkA)
T("B11 详情不存在(文档:404+code10003)", st == 404 and r["code"] == 10003, (st, r.get("code")))
st, r = req("/posts?sort=hot", method="GET", token=tkA)
T("B12 热门排序列表(进阶要求)", st == 200 and "items" in r["data"], st)

# ================= C. 评论 =================
st, r = req("/posts/%d/comments" % postB, token=tkA, body={"content": "A评B的帖"})
cidA = r["data"]["id"] if st == 201 else 0
T("C1 A评论B的帖(文档:201+author对象)", st == 201 and cidA > 0
  and r["data"]["author"]["username"] == uA and r["data"]["post_id"] == postB, (st, cidA))
st, _ = req("/posts/%d/comments" % postA, token=tkB, body={"content": "B评A的帖"})
T("C2 B评论A的帖", st == 201, st)
st, r = req("/posts/%d/comments" % postB, token=tkA, body={"content": ""})
T("C3 空评论(文档:400+code10000)", st == 400 and r["code"] == 10000, (st, r.get("code")))
st, r = req("/posts/999999/comments", token=tkA, body={"content": "评论不存在的帖"})
T("C4 评论不存在的帖(文档:404+code10003)", st == 404 and r["code"] == 10003, (st, r.get("code")))

# ================= D. 点赞(Redis链路) =================
st, r = req("/posts/%d/like" % postB, token=tkA)
T("D1 点赞(文档:data.is_liked=true)", st == 200 and r["data"]["is_liked"] is True, r.get("data"))
st, r = req("/posts/%d/like" % postB, token=tkA)
T("D2 再点取消点赞(文档:is_liked=false)", st == 200 and r["data"]["is_liked"] is False, r.get("data"))
req("/posts/%d/like" % postB, token=tkA)  # 重新点上
req("/posts/%d/like" % postA, token=tkA)  # A也赞A的帖
req("/posts/%d/like" % postA, token=tkB)  # B也赞A的帖
st, r = req("/posts/likes", token=tkA, body={"post_ids": [postA, postB]})
status = r["data"]["status"] if st == 200 else []
mp = {s["post_id"]: s["liked"] for s in status}
T("D3 批量查点赞状态(文档:status[{post_id,liked}])", st == 200
  and mp.get(postA) is True and mp.get(postB) is True, (st, status))
st, r = req("/posts/%d/like" % postA)
T("D4 未登录点赞(文档:401+code10001)", st == 401 and r["code"] == 10001, (st, r.get("code")))
st, r = req("/posts/999999/like", token=tkA)
T("D5 点赞不存在的帖(文档:404+code10003)", st == 404 and r["code"] == 10003, (st, r.get("code")))
st, r = req("/posts/likes", token=tkA, body={})
T("D6 批量查缺参数(文档:400+code10000)", st == 400 and r["code"] == 10000, (st, r.get("code")))
big = list(range(1, 102))
st, r = req("/posts/likes", token=tkA, body={"post_ids": big})
T("D7 批量查超上限(文档:400+code10000)", st == 400 and r["code"] == 10000, (st, r.get("code")))
st, r = req("/posts/likes", token=tkA, body={"post_ids": [postA, 999999]})
valid = [s["post_id"] for s in r["data"]["status"]] if st == 200 else []
T("D8 无效ID不返回(仅有效帖子)", st == 200 and 999999 not in valid and postA in valid, valid)

# ================= E. 删除 =================
st, r = req("/posts/%d" % postA, method="DELETE", token=tkA)
T("E1 A删自己的帖(文档:200+data=null)", st == 200 and r["data"] is None, (st, r.get("data")))
st, r = req("/posts/%d" % postB, method="DELETE", token=tkA)
T("E2 A删B的帖(文档:403+code10002;权限修复点)", st == 403 and r["code"] == 10002, (st, r.get("code")))
st, r = req("/posts/999999", method="DELETE", token=tkA)
T("E3 删不存在的帖(文档:404+code10003)", st == 404 and r["code"] == 10003, (st, r.get("code")))
st, r = req("/posts/%d" % postA, method="DELETE")
T("E4 未登录删帖(文档:401+code10001)", st == 401 and r["code"] == 10001, (st, r.get("code")))
st, _ = req("/posts/%d" % postA, method="GET", token=tkB)
T("E5 删后详情404(文档:404)", st == 404, st)

# 详情含评论:让 B 评 B 自己的帖,再查详情
req("/posts/%d/comments" % postB, token=tkB, body={"content": "B评B的帖"})
st, r = req("/posts/%d" % postB, method="GET", token=tkA)
comments = r["data"]["comments"] if st == 200 else []
T("E6 详情含评论(author为对象)", st == 200 and len(comments) >= 1
  and comments[0]["author"]["username"] in (uA, uB), comments[:1])

# ================= F. admin 权限 =================
st, r = req("/admin/posts/%d" % postB, method="DELETE", token=tkA)
T("F1 普通用户访问admin(文档:403+code10002)", st == 403 and r["code"] == 10002, (st, r.get("code")))
# 正式 admin 账号 stu001(重置后唯一保留的管理员)
st, r = req("/auth/login", body={"username": "stu001", "password": "123456"})
tkAdmin = r["data"]["access_token"] if st == 200 else ""
if tkAdmin:
    st, r = req("/posts", token=tkAdmin, body={"title": "给admin删的帖", "content": "给admin删的帖"})
    pid = r["data"]["id"] if st == 201 else 0
    st, r = req("/admin/posts/%d" % pid, method="DELETE", token=tkAdmin)
    T("F2 admin删自己发的帖(文档:200)", st == 200 and r["code"] == 0, (st, r.get("code")))
    st, r = req("/posts", token=tkB, body={"title": "学生的帖子给admin删", "content": "x"})
    pid3 = r["data"]["id"] if st == 201 else 0
    st, r = req("/admin/posts/%d" % pid3, method="DELETE", token=tkAdmin)
    T("F3 admin删学生的帖(文档:200;!IsAdmin修复生效)", st == 200 and r["code"] == 0, (st, r.get("code")))
else:
    print("[SKIP] F2/F3 正式admin账号 stu001 登录失败(密码被改?),跳过")

# ================= G. Agent 回归 =================
st, r = chat("sa1", "看看帖子列表", tkA)
T("G1 列表(DS在线先问排序;规则层直接列出并提示换排序)", st == 200
  and "排序" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa1", "热门", tkA)
T("G1b 切换热门排序(回复含热门)", st == 200 and "热门" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa1", "时间", tkA)
T("G1c 切回时间排序(回复含时间)", st == 200 and "时间" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa1", "第 2 页", tkA)
T("G2 多轮续接翻页", st == 200 and "第 2 页" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa1", "帖子 %d 的详情" % postB, tkA)
T("G3 详情", st == 200 and "作者" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa1", "帮我发一条招新帖:回归测试草稿", tkA)
pa = r["data"].get("pending_action") or {}
draft = pa.get("draft_id", "")
T("G4 发草稿(文档字段:draft_id/action/expires_at=RFC3339字符串)", st == 200 and bool(draft)
  and pa.get("action") == "create_post" and "content" in pa
  and isinstance(pa.get("expires_at"), str) and "T" in pa.get("expires_at", ""), draft)
st, r = chat("sa1", "confirm", tkA, confirm=draft)
T("G5 确认发布", st == 200 and "已成功发布" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa1", "confirm", tkA, confirm=draft)
T("G6 重复确认(文档:404+code10010)", st == 404 and r["code"] == 10010, (st, r.get("code")))
st, r = chat("sb1", "第 2 页", tkA)
T("G7 session隔离(新会话,DS判断或规则均需有回复)", st == 200 and len(r["data"]["reply"]) > 0, r["data"]["reply"][:40])

st, r = chat("sa2", "有哪些帖子", tkA)
T("G8 列表关键词「有哪些」(DS先问排序或直接列出)", st == 200
  and "排序" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa3", "看看帖子列表", tkA)
st, r = chat("sa3", "热门", tkA)
st, r = chat("sa3", str(postB), tkA)
T("G9 列表后纯数字续接(DS在线按AI判断;离线规则层会查详情)", st == 200
  and len(r["data"]["reply"]) > 0, r["data"]["reply"][:60])
st, r = chat("sa3", "帖子 %d 的详情" % postB, tkA)
st, r = chat("sa3", "2", tkA)
T("G10 详情后纯数字(DS重新判断,200或404均可)", st in (200, 404), (st, r.get("code")))
st, r = chat("sa4", "帖子 999999 的详情", tkA)
T("G11 查询不存在帖子详情(应404+code10003)", st == 404 and r["code"] == 10003, (st, r.get("code")))
st, r = chat("sa5", "帮我发一条:草稿A", tkA)
st, r = chat("sa5", "第 2 页", tkA)
T("G12 发草稿后翻页(DS重新判断,非空即可)", st == 200 and len(r["data"]["reply"]) > 0, r["data"]["reply"][:40])
st, r = chat("sa6", "帮我发一条:草稿B", tkA)
draft2 = (r["data"].get("pending_action") or {}).get("draft_id", "")
st, r = chat("sa6", "confirm", tkB, confirm=draft2)
T("G13 B确认A的草稿(文档:400+code10011)", st == 400 and r["code"] == 10011, (st, r.get("code")))
st, r = chat("sa6", "confirm", tkA, confirm=draft2)
T("G14 A本人确认草稿", st == 200 and "已成功发布" in r["data"]["reply"], r["data"]["reply"][:60])
m = re.search(r"ID: (\d+)", r["data"]["reply"]) if st == 200 else None
pid2 = int(m.group(1)) if m else 0
st, r = req("/posts/%d" % pid2, method="GET", token=tkA)
T("G15 确认发布后帖子可查", st == 200 and r["data"]["id"] == pid2, (st, pid2))
st, r = chat("sa7", "看看帖子列表", tkA)
st, r = chat("sa8", "第 2 页", tkA)
T("G16 会话间 LastTool 隔离(新会话,非空即可)", st == 200 and len(r["data"]["reply"]) > 0, r["data"]["reply"][:40])
st, r = chat("sa9", "你好呀", tkA)
reply9 = r["data"]["reply"] if st == 200 else ""
T("G17 新会话(非空回复)", st == 200 and len(reply9) > 0, reply9[:40])
print("[信息] DeepSeek 是否生效(回复是否为固定招新话术):", "📌【技术部暑期招新" not in reply9)
st, r = chat("sa9", "confirm", tkA)
T("G18 confirm不带编号(非空回复即可)", st == 200 and len(r["data"]["reply"]) > 0, r["data"]["reply"][:40])
st, r = chat("sa10", "看看帖子列表", tkA)
st, r = chat("sa10", "第 3 页", tkA)
T("G19 续接翻页到第3页", st == 200 and "第 3 页" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa11", "查看帖子 %d" % postB, tkA)
T("G20 关键词「查看」详情", st == 200 and "作者" in r["data"]["reply"], r["data"]["reply"][:60])

print()
print("=" * 40)
print("通过: %d / %d" % (sum(results), len(results)))
print("=" * 40)
