# -*- coding: utf-8 -*-
# 全功能回归测试:注册/登录/帖子/评论/点赞/删除/权限/admin/Agent
# 用法:先启动服务(go run .),然后 python tests/full_regression.py
# 重置:运行前自动清空数据库(仅保留正式账号 stu001 及其帖子)和 Redis 残留
# 幂等性:每次运行使用随机用户名,可重复执行
#这份测试代码curl是我让ai根据后端功能生成的测试代码,所以ai味会很重
import json, os, re, shutil, subprocess, sys, time, urllib.request, urllib.error

sys.stdout.reconfigure(encoding="utf-8")

BASE = "http://localhost:8080/api/v1"


# ===== 运行前重置数据库 / Redis =====
def _read_cfg():
    """从本地 config/config.yaml 读取数据库密码和库名(密码不进脚本、不进仓库)"""
    cfg = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "config", "config.yaml")
    with open(cfg, encoding="utf-8") as f:
        txt = f.read()
    return (re.search(r"^\s+password:\s*(\S+)", txt, re.M).group(1),
            re.search(r"^\s+database:\s*(\S+)", txt, re.M).group(1))


def _find_cmd(names, fallback):
    for n in names:
        p = shutil.which(n)
        if p:
            return p
    return fallback if os.path.exists(fallback) else None


def reset_db():
    """清空测试数据,仅保留 stu001 及其帖子(回到交付时的干净状态)"""
    exe = _find_cmd(["mysql"], r"C:\Program Files\MySQL\MySQL Server 8.0\bin\mysql.exe")
    if not exe:
        print("[跳过] 未找到 mysql 客户端,跳过数据库重置")
        return
    pwd, db = _read_cfg()
    sql = ("SET @sid := (SELECT id FROM users WHERE username='stu001' LIMIT 1);"
           "DELETE FROM likes WHERE user_id <> COALESCE(@sid, -1) "
           "OR post_id NOT IN (SELECT id FROM posts WHERE author_id=COALESCE(@sid, -1));"
           "DELETE FROM comments WHERE author_id <> COALESCE(@sid, -1) "
           "OR post_id NOT IN (SELECT id FROM posts WHERE author_id=COALESCE(@sid, -1));"
           "DELETE FROM posts WHERE author_id <> COALESCE(@sid, -1);"
           "DELETE FROM users WHERE username <> 'stu001';")
    env = dict(os.environ, MYSQL_PWD=pwd)  # 密码走环境变量,不进命令行
    r = subprocess.run([exe, "-uroot", "-e", "USE " + db + "; " + sql],
                       env=env, capture_output=True, text=True, encoding="utf-8", errors="replace")
    if r.returncode != 0:
        print("[警告] 数据库重置失败(测试仍继续):", r.stderr.strip()[:200])
    else:
        print("[重置] 数据库已清空,仅保留正式账号 stu001")


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
# 随机后缀:保证每次运行注册的都是全新用户(可重复跑)
suf = str(int(time.time() * 1000))[-6:]
uA = "curl_regA" + suf
uB = "curl_regB" + suf
phoneA = "139" + suf + "0" + suf[-1]
phoneB = "139" + suf + "1" + suf[-1]
results = []


def req(path, method="POST", body=None, token=None):
    data = json.dumps(body, ensure_ascii=False).encode("utf-8") if body is not None else None
    r = urllib.request.Request(BASE + path, data=data, method=method,
        headers={"Content-Type": "application/json"})
    if token:
        r.add_header("Authorization", "Bearer " + token)
    try:
        with urllib.request.urlopen(r) as resp:
            return resp.status, json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read().decode("utf-8"))


def chat(sid, msg, token, confirm=None):
    body = {"session_id": sid, "message": msg}
    if confirm:
        body["confirm_draft_id"] = confirm
    return req("/agent/chat", body=body, token=token)


def T(name, cond, detail=""):
    results.append(cond)
    print(("[PASS] " if cond else "[FAIL] ") + name + ((" | " + str(detail)) if detail != "" else ""))


# ================= A. 注册/登录 =================
st, r = req("/register", body={"username": uA, "password": "123456", "phone": phoneA})
T("A1 注册新用户A", st == 200, r.get("msg"))
st, r = req("/register", body={"username": uA, "password": "123456", "phone": phoneA})
T("A2 重复注册", st == 400, (st, r.get("code")))
st, _ = req("/register", body={"username": uB, "password": "123456", "phone": phoneB})
T("A3 注册新用户B", st == 200, st)
st, r = req("/login", body={"username": uA, "password": "123456"})
tkA = r["data"]["token"] if st == 200 else ""
T("A4 登录A", st == 200 and bool(tkA), st)
st, r = req("/login", body={"username": uB, "password": "123456"})
tkB = r["data"]["token"] if st == 200 else ""
T("A5 登录B", st == 200 and bool(tkB), st)
st, r = req("/login", body={"username": uA, "password": "wrongpass"})
T("A6 错误密码登录", st in (400, 401), (st, r.get("code")))
st, r = req("/register", body={"username": "curl_bad_phone", "password": "123456", "phone": "123"})
T("A7 手机号非11位", st == 400, (st, r.get("code")))

# ================= B. 帖子 =================
st, r = req("/posts", token=tkA, body={"title": "回归测试帖A", "content": "A发的帖子内容"})
postA = r["data"]["id"] if st == 200 else 0
T("B1 A发帖", st == 200 and postA > 0, postA)
st, r = req("/posts", token=tkB, body={"title": "回归测试帖B", "content": "B发的帖子内容"})
postB = r["data"]["id"] if st == 200 else 0
T("B2 B发帖", st == 200 and postB > 0, postB)
st, _ = req("/posts", token=tkA, body={"title": "", "content": "x"})
T("B3 空标题发帖", st == 400, st)
st, _ = req("/posts", body={"title": "x", "content": "y"})
T("B4 未登录发帖", st == 401, st)
st, r = req("/posts", method="GET")
T("B5 帖子列表", st == 200 and len(r["data"]["posts"]) > 0, (st, len(r["data"]["posts"])))
st, _ = req("/posts?page=0", method="GET")
T("B6 页码0", st == 400, st)
st, _ = req("/posts?page=abc", method="GET")
T("B7 页码非数字", st == 400, st)
st, r = req("/posts/%d" % postA, method="GET")
T("B8 帖子详情(作者名)", st == 200 and r["data"]["author_name"] == uA, st)
st, _ = req("/posts/999999", method="GET")
T("B9 详情不存在", st == 404, st)
st, r = req("/posts?sort=hot", method="GET")
T("B10 热门排序列表", st == 200, st)

# ================= C. 评论 =================
st, r = req("/comments", token=tkA, body={"content": "A评B的帖", "post_id": postB})
T("C1 A评论B的帖", st == 200, st)
st, r = req("/comments", token=tkB, body={"content": "B评A的帖", "post_id": postA})
T("C2 B评论A的帖", st == 200, st)
st, _ = req("/comments", token=tkA, body={"content": "", "post_id": postB})
T("C3 空评论", st == 400, st)
st, r = req("/comments", token=tkA, body={"content": "评论不存在的帖", "post_id": 999999})
T("C4 评论不存在的帖", st == 404, (st, r.get("code")))

# ================= D. 点赞(Redis链路) =================
st, r = req("/posts/%d/like" % postB, token=tkA)
T("D1 点赞", st == 200 and r["data"]["is_liked"] is True, r.get("data"))
st, r = req("/posts/%d/like" % postB, token=tkA)
T("D2 再点取消点赞", st == 200 and r["data"]["is_liked"] is False, r.get("data"))
req("/posts/%d/like" % postB, token=tkA)  # 重新点上
req("/posts/%d/like" % postA, token=tkA)  # A也赞A的帖
st, r = req("/posts/likes", token=tkA, body={"post_ids": [postA, postB]})
liked = r["data"]["post_ids"] if st == 200 else []
T("D3 批量查点赞状态", st == 200 and postA in liked and postB in liked, (st, liked))
st, _ = req("/posts/%d/like" % postA)
T("D4 未登录点赞", st == 401, st)
st, r = req("/posts/999999/like", token=tkA)
T("D5 点赞不存在的帖(应404)", st == 404, (st, r.get("code")))
st, r = req("/posts/likes", token=tkA, body={})
T("D6 批量查缺参数(应400)", st == 400, (st, r.get("code")))

# ================= E. 删除 =================
st, _ = req("/posts/%d" % postA, method="DELETE", token=tkA)
T("E1 A删自己的帖", st == 200, st)
st, r = req("/posts/%d" % postB, method="DELETE", token=tkA)
T("E2 A删B的帖(应403)", st == 403, (st, r.get("code")))
st, _ = req("/posts/999999", method="DELETE", token=tkA)
T("E3 删不存在的帖", st == 404, st)
st, _ = req("/posts/%d" % postA, method="DELETE")
T("E4 未登录删帖", st == 401, st)
st, r = req("/posts/%d" % postA, method="GET")
T("E5 删后详情404", st == 404, st)

# 评论删除:先让 B 评 B 自己的帖(这样 postB 上才有 B 的评论,供 E8 使用)
req("/comments", token=tkB, body={"content": "B评B的帖", "post_id": postB})
st, r = req("/posts/%d" % postB, method="GET")
comments = r["data"]["comments"] if st == 200 else []
cidA = None
for c in comments:
    if c.get("author", {}).get("username") == uA:
        cidA = c.get("id")
        break
T("E6 详情含评论", cidA is not None, comments[:1])
st, _ = req("/comments/%d" % cidA, method="DELETE", token=tkA) if cidA else (0, None)
T("E7 A删自己的评论", st == 200, st)
cidB = None
for c in comments:
    if c.get("author", {}).get("username") == uB:
        cidB = c.get("id")
        break
st, r = req("/comments/%d" % cidB, method="DELETE", token=tkA) if cidB else (0, None)
T("E8 A删B的评论(应403)", st == 403, (st, r.get("code") if r else None))
st, _ = req("/comments/999999", method="DELETE", token=tkA)
T("E9 删不存在的评论", st == 404, st)

# ================= F. admin 权限 =================
st, r = req("/admin/posts/%d" % postB, method="DELETE", token=tkA)
T("F1 普通用户访问admin(应403)", st == 403, (st, r.get("code")))
# 重置后仅保留正式账号 stu001(role=admin),用它验证管理员删帖
st, r = req("/login", body={"username": "stu001", "password": "123456"})
tkAdmin = r["data"]["token"] if st == 200 else ""
if tkAdmin:
    st, r = req("/posts", token=tkAdmin, body={"title": "给admin删的帖", "content": "admin测试"})
    pid = r["data"]["id"] if st == 200 else 0
    st, _ = req("/admin/posts/%d" % pid, method="DELETE", token=tkAdmin)
    T("F2 admin删任意帖", st == 200, st)
else:
    print("[SKIP] F2 正式admin账号 stu001 登录失败(密码被改?),跳过")

# ================= G. Agent 回归 =================
st, r = chat("sa1", "看看帖子列表", tkA)
T("G1 列表", st == 200 and "列表" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa1", "第 2 页", tkA)
T("G2 多轮续接翻页", st == 200 and "第 2 页" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa1", "帖子 %d 的详情" % postB, tkA)
T("G3 详情", st == 200 and "作者" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa1", "帮我发一条招新帖:回归测试草稿", tkA)
pa = r["data"].get("pending_action") or {}
draft = pa.get("confirm_draft_id", "")
T("G4 发草稿(不落库)", st == 200 and bool(draft) and pa.get("action") == "create_post", draft)
st, r = chat("sa1", "confirm", tkA, confirm=draft)
T("G5 确认发布", st == 200 and "已成功发布" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa1", "confirm", tkA, confirm=draft)
T("G6 重复确认(应404)", st == 404, (st, r.get("code")))
st, r = chat("sb1", "第 2 页", tkA)
T("G7 session隔离", st == 200 and "招新助手" in r["data"]["reply"], r["data"]["reply"][:40])

# ---- 新增 G8-G20:关键词变体 / LastTool 续接边界 / 草稿权限 / 会话隔离 / 兜底 ----
st, r = chat("sa2", "有哪些帖子", tkA)
T("G8 列表关键词「有哪些」", st == 200 and "查询到第 1 页" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa3", "看看帖子列表", tkA)
st, r = chat("sa3", str(postB), tkA)
T("G9 列表后纯数字续接详情", st == 200 and "回归测试帖B" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa3", "帖子 %d 的详情" % postB, tkA)
st, r = chat("sa3", "2", tkA)
T("G10 详情后纯数字(应兜底)", st == 200 and "招新助手" in r["data"]["reply"], r["data"]["reply"][:40])
st, r = chat("sa4", "帖子 999999 的详情", tkA)
T("G11 查询不存在帖子详情(应404)", st == 404, (st, r.get("code")))
st, r = chat("sa5", "帮我发一条:草稿A", tkA)
st, r = chat("sa5", "第 2 页", tkA)
T("G12 发草稿后 LastTool 清空,翻页应兜底", st == 200 and "招新助手" in r["data"]["reply"], r["data"]["reply"][:40])
st, r = chat("sa6", "帮我发一条:草稿B", tkA)
draft2 = (r["data"].get("pending_action") or {}).get("confirm_draft_id", "")
st, r = chat("sa6", "confirm", tkB, confirm=draft2)
T("G13 B确认A的草稿(应404无权限)", st == 404 and "无权限" in (str(r.get("msg")) or ""), (st, r.get("code")))
st, r = chat("sa6", "confirm", tkA, confirm=draft2)
T("G14 A本人确认草稿", st == 200 and "已成功发布" in r["data"]["reply"], r["data"]["reply"][:60])
m = re.search(r"ID: (\d+)", r["data"]["reply"]) if st == 200 else None
pid2 = int(m.group(1)) if m else 0
st, r = req("/posts/%d" % pid2, method="GET")
T("G15 确认发布后帖子可查", st == 200 and r["data"]["title"] == "新帖子", (st, pid2))
st, r = chat("sa7", "看看帖子列表", tkA)
st, r = chat("sa8", "第 2 页", tkA)
T("G16 会话间 LastTool 隔离(新会话应兜底)", st == 200 and "招新助手" in r["data"]["reply"], r["data"]["reply"][:40])
st, r = chat("sa9", "你好呀", tkA)
T("G17 新会话兜底自我介绍", st == 200 and "招新助手" in r["data"]["reply"], r["data"]["reply"][:40])
st, r = chat("sa9", "confirm", tkA)
T("G18 confirm不带编号(应兜底)", st == 200 and "招新助手" in r["data"]["reply"], r["data"]["reply"][:40])
st, r = chat("sa10", "看看帖子列表", tkA)
st, r = chat("sa10", "第 3 页", tkA)
T("G19 续接翻页到第3页", st == 200 and "第 3 页" in r["data"]["reply"], r["data"]["reply"][:60])
st, r = chat("sa11", "查看帖子 %d" % postB, tkA)
T("G20 关键词「查看」详情", st == 200 and "作者" in r["data"]["reply"], r["data"]["reply"][:60])

print()
print("=" * 40)
print("通过: %d / %d" % (sum(results), len(results)))
print("=" * 40)
