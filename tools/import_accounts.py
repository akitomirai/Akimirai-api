#!/usr/bin/env python3
"""直接将 GPT 和 Grok 账号导入 Sub2API PostgreSQL 数据库"""
import json
import os
import glob
import psycopg2
from datetime import datetime, timezone

GPT_DIR = r"D:\Users\Lenovo\Desktop\api\gpt"
GROK_DIR = r"D:\Users\Lenovo\Desktop\api\grok"

GPT_GROUP_ID = 8   # 逆向生图（gpt）
GROK_GROUP_ID = 7  # 逆向生图（grok）

DB_CONFIG = {
    "host": "localhost",
    "port": 5432,
    "dbname": "sub2api",
    "user": "sub2api",
    "password": "sub2api"
}


def parse_gpt_accounts():
    """解析 GPT 文件夹中所有有密码的账号"""
    accounts = []
    seen = set()

    for f in sorted(glob.glob(os.path.join(GPT_DIR, "gpt_mail*.json"))):
        basename = os.path.basename(f)
        if "_failed" in basename or "_cookies" in basename:
            continue
        if basename.startswith("gpt_1_"):
            continue

        try:
            with open(f, 'r', encoding='utf-8') as fh:
                data = json.load(fh)
        except Exception:
            continue

        email = (data.get("email") or "").strip()
        password = (data.get("password") or "").strip()
        token = (data.get("session_token") or data.get("access_token") or "").strip()

        if not email or not password or email in seen:
            continue
        seen.add(email)

        note = f"GPT {data.get('plan', 'free')} | registered: {data.get('registered_at', '')}"
        if token:
            note += " | has session_token"

        accounts.append({
            "email": email,
            "password": password,
            "session_token": token or None,
            "note": note,
        })
    return accounts


def parse_grok_accounts():
    """解析 Grok sub2api_import.json"""
    fpath = os.path.join(GROK_DIR, "sub2api_import.json")
    if not os.path.exists(fpath):
        print(f"Not found: {fpath}")
        return []

    with open(fpath, 'r', encoding='utf-8') as f:
        data = json.load(f)

    accounts = []
    for a in data.get("data", {}).get("accounts", []):
        creds = a.get("credentials", {})
        email = (creds.get("email") or a.get("name") or "").strip()
        password = (creds.get("password") or "").strip()
        if not email or not password:
            continue
        accounts.append({
            "email": email,
            "password": password,
            "session_token": None,
            "note": "Grok | " + a.get("extra", {}).get("note", "Login at x.ai/api"),
        })
    return accounts


def import_accounts(accounts, group_id, platform, label):
    """导入账号到 DB 并绑定分组"""
    conn = psycopg2.connect(**DB_CONFIG)
    cur = conn.cursor()
    inserted = updated = 0

    try:
        for i, acct in enumerate(accounts):
            email = acct["email"]
            pwd = acct["password"]
            note = acct["note"]
            creds = json.dumps({"email": email, "password": pwd})
            extra = {"note": note}
            if acct.get("session_token"):
                extra["session_token"] = acct["session_token"]
            extra_json = json.dumps(extra)
            now = datetime.now(timezone.utc)

            # 检查是否已存在
            cur.execute(
                "SELECT id FROM accounts WHERE name = %s AND deleted_at IS NULL",
                (email,))
            row = cur.fetchone()

            if row:
                account_id = row[0]
                cur.execute(
                    "UPDATE accounts SET notes=%s, credentials=%s, extra=%s, updated_at=%s WHERE id=%s",
                    (note, creds, extra_json, now, account_id))
                updated += 1
                action = "UPDATE"
            else:
                cur.execute(
                    """INSERT INTO accounts
                    (name, platform, type, credentials, extra, notes,
                     concurrency, priority, rate_multiplier, status,
                     auto_pause_on_expired, created_at, updated_at)
                    VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
                    RETURNING id""",
                    (email, platform, "password", creds, extra_json, note,
                     5, 80, 1.0, "active", False, now, now))
                account_id = cur.fetchone()[0]
                inserted += 1
                action = "INSERT"

            # 绑定分组
            cur.execute(
                """INSERT INTO account_groups (account_id, group_id, priority, created_at)
                VALUES (%s,%s,%s,%s) ON CONFLICT (account_id, group_id) DO NOTHING""",
                (account_id, group_id, 50, now))

            print(f"  [{i+1}/{len(accounts)}] {action} id={account_id} {email}")

        conn.commit()
        print(f"  {label}: {inserted} new, {updated} updated, {len(accounts)} total")
    except Exception as e:
        conn.rollback()
        print(f"  ERROR: {e}")
        raise
    finally:
        cur.close()
        conn.close()


def main():
    print("=== Parsing GPT accounts ===")
    gpt = parse_gpt_accounts()
    print(f"  Found {len(gpt)} with passwords")

    print("\n=== Parsing Grok accounts ===")
    grok = parse_grok_accounts()
    print(f"  Found {len(grok)} with passwords")

    if gpt:
        print(f"\n=== Importing GPT -> group {GPT_GROUP_ID} ===")
        import_accounts(gpt, GPT_GROUP_ID, "openai", "GPT")

    if grok:
        print(f"\n=== Importing Grok -> group {GROK_GROUP_ID} ===")
        import_accounts(grok, GROK_GROUP_ID, "grok", "Grok")

    print(f"\n=== Done! GPT:{len(gpt)} Grok:{len(grok)} ===")


if __name__ == "__main__":
    main()
