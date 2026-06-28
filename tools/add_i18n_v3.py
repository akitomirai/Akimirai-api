"""Add missing i18n keys for imageGen page columns and test"""
import os

BASE = r"F:\Akimirai-api\frontend\src\i18n\locales"

zh_additions = """    colEmail: '邮箱',
    colPlatform: '平台',
    colType: '类型',
    colToken: 'Token',
    colMode: '模式',
    colStatus: '状态',
    test: '测试',
    testing: '测试中...',
    testTitle: '生图测试 - {email}',
    testPromptPlaceholder: '输入生图提示词...',
    testDefaultPrompt: '一只可爱的猫咪',
    testSend: '发送',
    testHint: '输入提示词后点击"发送"开始测试',
    testConnecting: '正在连接',
    testStarted: '开始测试',
    testImageReceived: '收到图片',
    testError: '测试出错',
    testComplete: '测试完成',
    testFailed: '测试失败',
    statusActive: '活跃',
"""

en_additions = """    colEmail: 'Email',
    colPlatform: 'Platform',
    colType: 'Type',
    colToken: 'Token',
    colMode: 'Mode',
    colStatus: 'Status',
    test: 'Test',
    testing: 'Testing...',
    testTitle: 'Image Test - {email}',
    testPromptPlaceholder: 'Enter image prompt...',
    testDefaultPrompt: 'A cute cat',
    testSend: 'Send',
    testHint: 'Enter a prompt and click "Send" to test',
    testConnecting: 'Connecting',
    testStarted: 'Test started',
    testImageReceived: 'Image received',
    testError: 'Test error',
    testComplete: 'Test complete',
    testFailed: 'Test failed',
    statusActive: 'Active',
"""

for fname, additions in [("zh.ts", zh_additions), ("en.ts", en_additions)]:
    fpath = os.path.join(BASE, fname)
    with open(fpath, "r", encoding="utf-8") as f:
        lines = f.readlines()

    # Find the imageGen section's "imported" line and add after it
    for i, line in enumerate(lines):
        if "imported:" in line and "imageGen" in "".join(lines[max(0,i-20):i+1]):
            # Find the next line that's just a closing or transition
            # Insert after the imported line
            lines.insert(i + 1, additions)
            break

    with open(fpath, "w", encoding="utf-8") as f:
        f.writelines(lines)
    print(f"{fname} updated")
