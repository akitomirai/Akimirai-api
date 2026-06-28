"""Add imageGen i18n keys to zh.ts and en.ts"""
import re

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

# Update zh.ts
with open('F:/Akimirai-api/frontend/src/i18n/locales/zh.ts', 'r', encoding='utf-8') as f:
    zh = f.read()

zh = zh.replace(
    "    accounts: '账号管理',",
    "    accounts: '账号管理',\n    accountsList: '账号',\n    imageGenAccounts: '生图账号',"
)
zh = zh.rstrip()
if not zh.endswith('}'):
    zh += '\n}'
zh = zh[:-2] + image_gen_zh + '\n}\n'

with open('F:/Akimirai-api/frontend/src/i18n/locales/zh.ts', 'w', encoding='utf-8') as f:
    f.write(zh)
print('zh.ts updated')

# Update en.ts
with open('F:/Akimirai-api/frontend/src/i18n/locales/en.ts', 'r', encoding='utf-8') as f:
    en = f.read()

en = en.replace(
    "    accounts: 'Accounts',",
    "    accounts: 'Accounts',\n    accountsList: 'Accounts',\n    imageGenAccounts: 'Image Gen',"
)
en = en.rstrip()
if not en.endswith('}'):
    en += '\n}'
en = en[:-2] + image_gen_en + '\n}\n'

with open('F:/Akimirai-api/frontend/src/i18n/locales/en.ts', 'w', encoding='utf-8') as f:
    f.write(en)
print('en.ts updated')
