# -*- coding: utf-8 -*-
"""Edge 渲染 Apifox 分享文档:密码 -> 抓树 -> 抓全部接口详情 JSON + 页面全文。"""
import json, os, time
from playwright.sync_api import sync_playwright

ID = "c645e2be-6bcd-42bf-8108-c052607fb260"
PWD = "zjutjh2026"
BASE = f"https://s.apifox.cn/{ID}"
URL = f"{BASE}/485816007e0"
OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "apifox_capture")
os.makedirs(OUT, exist_ok=True)

logs = []
def on_response(resp):
    u = resp.url
    if "/api/" in u or "/shared-doc" in u or u.endswith(".md"):
        try:
            body = resp.text()[:600]
        except Exception:
            body = ""
        logs.append({"url": u, "method": resp.request.method, "status": resp.status})
        print(f"[{resp.status}] {resp.request.method} {u}")

with sync_playwright() as p:
    browser = p.chromium.launch(channel="msedge", headless=True)
    ctx = browser.new_context(viewport={"width": 1440, "height": 2600})
    page = ctx.new_page()
    page.on("response", on_response)

    # 1. 访问页面;若出现密码框,输入密码
    page.goto(URL, wait_until="domcontentloaded", timeout=60000)
    time.sleep(2)
    txt = page.inner_text("body")
    print("step1 body:", txt[:60].replace("\n", " / "))
    if "密码" in txt:
        # 找输入框与按钮
        try:
            inputs = page.locator("input").all()
            print("inputs:", len(inputs))
            for inp in inputs:
                try:
                    inp.fill(PWD)
                except Exception:
                    pass
            page.keyboard.press("Enter")
        except Exception as e:
            print("fill err", e)
        time.sleep(3)
        # 兜底:按"访问文档/确定"按钮
        for label in ["访问文档", "确定", "确认", "登录"]:
            try:
                btn = page.get_by_text(label, exact=True)
                if btn.count() > 0:
                    btn.first.click()
                    break
            except Exception:
                pass
    # 2. 等待文档内容渲染
    for i in range(40):
        txt = page.inner_text("body")
        if "注册用户" in txt and "登录" in txt:
            print("rendered at", i, "s")
            break
        time.sleep(1)
    time.sleep(2)

    # 3. 保存页面当前文本(目录/摘要)
    page_text = page.inner_text("body")
    with open(os.path.join(OUT, "page_text.txt"), "w", encoding="utf-8") as f:
        f.write(page_text)

    # 4. 在浏览器上下文里直接把 12 个 .md/详情全文抓下来(带鉴权头,浏览器会补)
    results = {}
    nodes = ["7348302m0", "485816007e0", "485816008e0", "485816010e0", "485816011e0",
             "485816012e0", "485816013e0", "485816014e0", "485846243e0", "485816016e0",
             "485816017e0", "485816018e0"]
    for node in nodes:
        try:
            data = page.evaluate("""
                async (u) => {
                    const r = await fetch(u, {headers: {"Accept": "text/markdown"}});
                    return {status: r.status, text: await r.text()};
                }
            """, f"{BASE}/{node}.md")
            results[node] = data
            print(f"[{data['status']}] md/{node} len={len(data['text'])}")
        except Exception as e:
            print("fetch err", node, e)
    # 5. 若 fetch 不到 markdown,直接抓页面 DOM 全文
    try:
        body_html = page.content()
        with open(os.path.join(OUT, "page_full.html"), "w", encoding="utf-8") as f:
            f.write(body_html)
    except Exception:
        pass
    with open(os.path.join(OUT, "md_results.json"), "w", encoding="utf-8") as f:
        json.dump(results, f, ensure_ascii=False, indent=1)
    browser.close()

print("log entries:", len(logs))
with open(os.path.join(OUT, "network.json"), "w", encoding="utf-8") as f:
    json.dump(logs, f, ensure_ascii=False, indent=1)
