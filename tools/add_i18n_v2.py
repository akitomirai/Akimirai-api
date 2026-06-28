"""Properly add imageGen i18n keys to zh.ts and en.ts"""
import os

BASE = r"F:\Akimirai-api\frontend\src\i18n\locales"

image_gen_zh = """  imageGen: {
    title: '生图账号',
    searchPlaceholder: '搜索邮箱...',
    importAccounts: '导入账号',
    importTitle: '导入免费账号',
    importFormat: '账号格式',
    importFormatHint: '每行一个：email@domain.com:password',
    importPlaceholder: 'mail001@domain.com:password001\\nmail002@domain.com:password002',
    importBtn: '导入',
    importFirst: '导入第一个账号',
    noAccounts: '暂无生图账号',
    noAccountsHint: '点击"导入账号"添加免费 ChatGPT 账号用于逆向生图',
    groupLabel: '目标分组',
    enableConversation: '开启会话模式',
    disableConversation: '关闭会话模式',
    conversationOn: '会话模式',
    conversationOff: '普通模式',
    conversationEnabled: '会话模式已开启',
    conversationDisabled: '会话模式已关闭',
    tokenOk: 'Token 有效',
    tokenNone: '无 Token',
    confirmDelete: '确定删除账号 {email}？',
    deleted: '账号已删除',
    imported: '成功导入 {count} 个账号',
  },

"""

image_gen_en = """  imageGen: {
    title: 'Image Gen Accounts',
    searchPlaceholder: 'Search email...',
    importAccounts: 'Import Accounts',
    importTitle: 'Import Free Accounts',
    importFormat: 'Account Format',
    importFormatHint: 'One per line: email@domain.com:password',
    importPlaceholder: 'mail001@domain.com:password001\\nmail002@domain.com:password002',
    importBtn: 'Import',
    importFirst: 'Import First Account',
    noAccounts: 'No Image Gen Accounts',
    noAccountsHint: 'Click "Import Accounts" to add free ChatGPT accounts for image generation',
    groupLabel: 'Target Group',
    enableConversation: 'Enable Conversation Mode',
    disableConversation: 'Disable Conversation Mode',
    conversationOn: 'Conversation',
    conversationOff: 'Normal',
    conversationEnabled: 'Conversation mode enabled',
    conversationDisabled: 'Conversation mode disabled',
    tokenOk: 'Token OK',
    tokenNone: 'No Token',
    confirmDelete: 'Delete account {email}?',
    deleted: 'Account deleted',
    imported: 'Imported {count} accounts',
  },

"""

for fname, nav_old, nav_new, image_gen in [
    ("zh.ts", "    accounts: '账号管理',",
     "    accounts: '账号管理',\n    accountsList: '账号',\n    imageGenAccounts: '生图账号',",
     image_gen_zh),
    ("en.ts", "    accounts: 'Accounts',",
     "    accounts: 'Accounts',\n    accountsList: 'Accounts',\n    imageGenAccounts: 'Image Gen',",
     image_gen_en),
]:
    fpath = os.path.join(BASE, fname)
    with open(fpath, "r", encoding="utf-8") as f:
        lines = f.readlines()

    # Find the closing brace line (last line with just "}")
    close_idx = None
    for i in range(len(lines) - 1, -1, -1):
        if lines[i].strip() == "}":
            close_idx = i
            break

    if close_idx is None:
        print(f"ERROR: Could not find closing }} in {fname}")
        continue

    # Add nav keys
    for i, line in enumerate(lines):
        if line.strip() == nav_old.strip():
            lines[i] = nav_new + "\n"
            break

    # Insert imageGen section before closing }
    lines.insert(close_idx, image_gen)

    with open(fpath, "w", encoding="utf-8") as f:
        f.writelines(lines)
    print(f"{fname} updated: nav keys + imageGen section ({close_idx})")
