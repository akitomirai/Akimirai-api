#!/usr/bin/env python3
"""清理并重新导入账号：GPT session_token -> oauth 类型"""
import json
import os
import glob
import psycopg2
from datetime import datetime, timezone

GPT_DIR = r"D:\Users\Lenovo\Desktop\api\gpt"
GROK_DIR = r"D:\Users\Lenovo\Desktop\api\grok"

GPT_GROUP_ID = 8
GROK_GROUP_ID = 7

DB_CONFIG = {
    "host": "localhost", "port": 5432, "dbname": "sub2api",
    "user": "sub2api", "password": "sub2api"
}


def parse_gpt():
    """解析 GPT，按 session_token 有无分两类"""
    oauth = []   # 有 token → oauth
    pwd = []     # 无 token → password
    seen = set()

    for f in sorted(glob.glob(os.path.join(GPT_DIR, "gpt_mail*.json"))):
        basename = os.path.basename(f)
        if "_failed" in basename or "_cookies" in basename:
            continue
        if basename.startswith("gpt_1_"):
            continue

        try:
            with open(f, 'r', encoding='utf-8') as fh:
                d = json.load(fh)
        except Exception:
            continue

        email = (d.get("email") or "").strip()
        pwd_val = (d.get("password") or "").strip()
        token = (d.get("session_token") or d.get("access_token") or "").strip()

        if not email or not pwd_val or email in seen:
            continue
        seen.add(email)

        note = f"GPT {d.get('plan','free')} | reg: {d.get('registered_at','')}"
        if token and len(token) > 20:
            note += f" | session_token OK"
            oauth.append({"email": email, "password": pwd_val, "token": token, "note": note})
        else:
            note += " | NO session_token"
            pwd.append({"email": email, "password": pwd_val, "token": None, "note": note})

    return oauth, pwd


def parse_grok():
    fpath = os.path.join(GROK_DIR, "sub2api_import.json")
    if not os.path.exists(fpath):
        return []
    with open(fpath, 'r', encoding='utf-8') as f:
        data = json.load(f)
    accounts = []
    for a in data.get("data", {}).get("accounts", []):
        c = a.get("credentials", {})
        email = (c.get("email") or "").strip()
        pwd_val = (c.get("password") or "").strip()
        if email and pwd_val:
            accounts.append({
                "email": email,
                "password": pwd_val,
                "token": None,
                "note": "Grok | " + a.get("extra", {}).get("note", "Login at x.ai/api"),
            })
    return accounts


def cleanup():
    conn = psycopg2.connect(**DB_CONFIG)
    cur = conn.cursor()
    cur.execute("DELETE FROM account_groups")
    cur.execute("DELETE FROM accounts")
    conn.commit()
    cur.close()
    conn.close()
    print("  Deleted all accounts and group bindings")


def import_oauth(accounts, group_id, platform, label):
    """导入 oauth 类型账号（session_token → access_token）"""
    conn = psycopg2.connect(**DB_CONFIG)
    cur = conn.cursor()
    cnt = 0
    try:
        now = datetime.now(timezone.utc)
        for i, a in enumerate(accounts):
            creds = json.dumps({
                "email": a["email"],
                "password": a["password"],
                "access_token": a["token"],  # session_token 用作 access_token
            })
            extra = json.dumps({"note": a["note"], "codex_image_generation_bridge_enabled": True})
            cur.execute(
                """INSERT INTO accounts
                (name, platform, type, credentials, extra, notes,
                 concurrency, priority, rate_multiplier, status,
                 auto_pause_on_expired, created_at, updated_at)
                VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
                RETURNING id""",
                (a["email"], platform, "oauth", creds, extra, a["note"],
                 5, 80, 1.0, "active", False, now, now))
            aid = cur.fetchone()[0]
            cur.execute(
                "INSERT INTO account_groups (account_id, group_id, priority, created_at) VALUES (%s,%s,%s,%s) ON CONFLICT DO NOTHING",
                (aid, group_id, 50, now))
            cnt += 1
            print(f"  [{cnt}/{len(accounts)}] oauth id={aid} {a['email']}")
        conn.commit()
        print(f"  {label}: {cnt} oauth accounts imported")
    except Exception as e:
        conn.rollback()
        print(f"  ERROR: {e}")
        raise
    finally:
        cur.close()
        conn.close()


def import_password(accounts, group_id, platform, label):
    """导入 password 类型（无 token 的备份存储）"""
    if not accounts:
        return
    conn = psycopg2.connect(**DB_CONFIG)
    cur = conn.cursor()
    cnt = 0
    try:
        now = datetime.now(timezone.utc)
        for i, a in enumerate(accounts):
            creds = json.dumps({"email": a["email"], "password": a["password"]})
            extra = json.dumps({"note": a["note"]})
            cur.execute(
                """INSERT INTO accounts
                (name, platform, type, credentials, extra, notes,
                 concurrency, priority, rate_multiplier, status,
                 auto_pause_on_expired, created_at, updated_at)
                VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
                RETURNING id""",
                (a["email"], platform, "password", creds, extra, a["note"],
                 5, 80, 1.0, "active", False, now, now))
            aid = cur.fetchone()[0]
            cur.execute(
                "INSERT INTO account_groups (account_id, group_id, priority, created_at) VALUES (%s,%s,%s,%s) ON CONFLICT DO NOTHING",
                (aid, group_id, 50, now))
            cnt += 1
            print(f"  [{cnt}/{len(accounts)}] password id={aid} {a['email']}")
        conn.commit()
        print(f"  {label}: {cnt} password accounts imported")
    except Exception as e:
        conn.rollback()
        print(f"  ERROR: {e}")
        raise
    finally:
        cur.close()
        conn.close()


def main():
    oauth_gpt, pwd_gpt = parse_gpt()
    grok = parse_grok()

    print(f"GPT oauth (has token):  {len(oauth_gpt)}")
    print(f"GPT password (no token): {len(pwd_gpt)}")
    print(f"Grok password:          {len(grok)}")

    print("\n=== Cleaning up ===")
    cleanup()

    if oauth_gpt:
        print(f"\n=== Importing {len(oauth_gpt)} GPT -> oauth -> group {GPT_GROUP_ID} ===")
        import_oauth(oauth_gpt, GPT_GROUP_ID, "openai", "GPT-oauth")

    if pwd_gpt:
        print(f"\n=== Importing {len(pwd_gpt)} GPT -> password -> group {GPT_GROUP_ID} ===")
        import_password(pwd_gpt, GPT_GROUP_ID, "openai", "GPT-password")

    if grok:
        print(f"\n=== Importing {len(grok)} Grok -> password -> group {GROK_GROUP_ID} ===")
        import_password(grok, GROK_GROUP_ID, "grok", "Grok-password")

    total = len(oauth_gpt) + len(pwd_gpt) + len(grok)
    print(f"\n=== Done! Total: {total} accounts ===")


if __name__ == "__main__":
    main()
