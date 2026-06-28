#!/usr/bin/env python3
"""解析 GPT 和 Grok 文件夹中的账号，构建 Sub2API 导入 JSON"""
import json
import os
import glob
from pathlib import Path

GPT_DIR = r"D:\Users\Lenovo\Desktop\api\gpt"
GROK_DIR = r"D:\Users\Lenovo\Desktop\api\grok"
OUTPUT = r"D:\Users\Lenovo\Desktop\api\sub2api_import_all.json"

GPT_GROUP_ID = 8   # 逆向生图（gpt）
GROK_GROUP_ID = 7  # 逆向生图（grok）

def parse_gpt_accounts():
    """解析 GPT 文件夹中所有有效账号"""
    accounts = []
    seen = set()

    # 找到所有非 failed、非 cookies 的 gpt_mail*.json 文件
    json_files = glob.glob(os.path.join(GPT_DIR, "gpt_mail*.json"))

    for f in sorted(json_files):
        basename = os.path.basename(f)
        # 跳过 failed 和 cookies 文件
        if "_failed" in basename or "_cookies" in basename:
            continue
        if basename.startswith("gpt_1_"):
            # gpt_1_ 类型，可能没有密码
            pass

        try:
            with open(f, 'r', encoding='utf-8') as fh:
                data = json.load(fh)
        except Exception as e:
            print(f"  SKIP {basename}: {e}")
            continue

        email = data.get("email", "").strip()
        password = data.get("password", "").strip()
        session_token = data.get("session_token", "").strip()
        access_token = data.get("access_token", "").strip()
        name = data.get("name", data.get("full_name", ""))
        plan = data.get("plan", "free")

        if not email:
            print(f"  SKIP {basename}: no email")
            continue

        if email in seen:
            continue
        seen.add(email)

        account = {
            "name": email,
            "platform": "openai",
            "type": "password",
            "credentials": {
                "email": email,
                "password": password if password else ""
            },
            "extra": {
                "note": f"GPT {'free' if plan == 'free' else plan} plan"
            },
            "concurrency": 5,
            "priority": 80,
            "rate_multiplier": 1,
            "auto_pause_on_expired": False
        }

        # 如果有 session_token，加入 extra
        token = session_token or access_token
        if token:
            account["extra"]["session_token"] = token
            account["extra"]["note"] += " | has session_token"

        accounts.append(account)
        status = "OK" if password else "NO_PASSWORD"
        print(f"  [{status}] {email}")

    return accounts


def parse_grok_accounts():
    """解析 Grok sub2api_import.json"""
    import_file = os.path.join(GROK_DIR, "sub2api_import.json")
    if not os.path.exists(import_file):
        print("Grok import file not found")
        return []

    with open(import_file, 'r', encoding='utf-8') as f:
        data = json.load(f)

    accounts = data.get("data", {}).get("accounts", [])
    print(f"\nGrok: {len(accounts)} accounts in sub2api_import.json")
    return accounts


def main():
    print("=== Parsing GPT accounts ===")
    gpt_accounts = parse_gpt_accounts()
    print(f"\nGPT total valid: {len(gpt_accounts)}")

    print("\n=== Parsing Grok accounts ===")
    grok_accounts = parse_grok_accounts()

    # 构建导入数据
    # GPT → group 8 (逆向生图gpt)
    # Grok → group 7 (逆向生图grok)

    gpt_with_passwords = [a for a in gpt_accounts if a["credentials"]["password"]]
    gpt_without_passwords = [a for a in gpt_accounts if not a["credentials"]["password"]]

    print(f"\n=== Summary ===")
    print(f"GPT with passwords:  {len(gpt_with_passwords)}")
    print(f"GPT without passwords: {len(gpt_without_passwords)}")
    print(f"Grok with passwords: {len(grok_accounts)}")
    print(f"TOTAL importable:    {len(gpt_with_passwords) + len(grok_accounts)}")

    # 分开构建两个 import payload（因为要导入不同分组）

    # GPT import
    gpt_payload = {
        "data": {
            "type": "sub2api-data",
            "version": 1,
            "proxies": [],
            "accounts": [gpt_with_passwords, {}]  # placeholder
        },
        "skip_default_group_bind": True,
        "group_ids": [GPT_GROUP_ID]
    }

    # Actually the API format is accounts array directly
    # Let me build the correct format

    output = {
        "description": "GPT + Grok accounts import for 逆向生图",
        "gpt": {
            "group_id": GPT_GROUP_ID,
            "group_name": "逆向生图（gpt）",
            "accounts": gpt_with_passwords,
            "count": len(gpt_with_passwords)
        },
        "grok": {
            "group_id": GROK_GROUP_ID,
            "group_name": "逆向生图（grok）",
            "accounts": grok_accounts,
            "count": len(grok_accounts)
        }
    }

    with open(OUTPUT, 'w', encoding='utf-8') as f:
        json.dump(output, f, ensure_ascii=False, indent=2)

    print(f"\nOutput written to: {OUTPUT}")

    # Also write individual import files for the API
    gpt_api_file = os.path.join(os.path.dirname(OUTPUT), "gpt_import_api.json")
    gpt_api_payload = {
        "data": {
            "type": "sub2api-data",
            "version": 1,
            "proxies": [],
            "accounts": gpt_with_passwords
        },
        "skip_default_group_bind": True
    }
    with open(gpt_api_file, 'w', encoding='utf-8') as f:
        json.dump(gpt_api_payload, f, ensure_ascii=False, indent=2)
    print(f"GPT API import file: {gpt_api_file}")

    grok_api_file = os.path.join(os.path.dirname(OUTPUT), "grok_import_api.json")
    grok_api_payload = {
        "data": {
            "type": "sub2api-data",
            "version": 1,
            "proxies": [],
            "accounts": grok_accounts
        },
        "skip_default_group_bind": True
    }
    with open(grok_api_file, 'w', encoding='utf-8') as f:
        json.dump(grok_api_payload, f, ensure_ascii=False, indent=2)
    print(f"Grok API import file: {grok_api_file}")

    # Print accounts without passwords for review
    if gpt_without_passwords:
        print(f"\n=== Accounts WITHOUT passwords (not imported) ===")
        for a in gpt_without_passwords:
            print(f"  {a['name']}")


if __name__ == "__main__":
    main()
