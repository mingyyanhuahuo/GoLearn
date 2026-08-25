# -*- coding: utf-8 -*-
# 全功能回归测试:注册/登录/帖子/评论/点赞/删除/权限/admin/Agent
# 用法:先启动服务(go run .),然后 python tests/full_regression.py
# 幂等性:每次运行使用随机用户名,可重复执行(历史测试数据不影响结果)
import json, sys, time, urllib.request, urllib.error

sys.stdout.reconfigure(encoding="utf-8")

BASE = "http://localhost:8080/api/v1"
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
st, r = req("/login", body={"username": "admindemo", "password": "123456"})
tkAdmin = r["data"]["token"] if st == 200 else ""
if tkAdmin:
    st, r = req("/posts", token=tkAdmin, body={"title": "给admin删的帖", "content": "admin测试"})
    pid = r["data"]["id"] if st == 200 else 0
    st, _ = req("/admin/posts/%d" % pid, method="DELETE", token=tkAdmin)
    T("F2 admin删任意帖", st == 200, st)
else:
    print("[SKIP] F2 admin账号不可用(admindemo/123456 登录失败,跳过)")

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

print()
print("=" * 40)
print("通过: %d / %d" % (sum(results), len(results)))
print("=" * 40)
