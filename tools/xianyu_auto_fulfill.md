# Xianyu Auto Delivery

This document describes the first browser-based delivery helper for Xianyu /
Goofish orders.

## Current Boundary

Goofish web can open IM conversations and send text messages, but the web
profile page currently shows the sold-order list as app-only. Because the real
sold order ID is not exposed on the web page, this helper uses:

- the visible IM conversation status,
- the order card amount,
- and a local state file

to avoid duplicate handling.

If Goofish Open Platform order/message permissions become available later, use
the official order ID and webhook/event flow instead of this browser fallback.

## Preconditions

1. Production backend has external fulfillment SKUs:
   - `bakaai-balance-5`
   - `bakaai-balance-10`
   - `bakaai-balance-20`
   - `bakaai-balance-50`
   - `bakaai-balance-100`
2. Each SKU has the delivery template containing:
   - redeem code,
   - `https://miraiapi.cn/redeem`,
   - Feishu manual link,
   - QQ support group `1048099894`.
3. `EXTERNAL_FULFILLMENT_FEISHU_WEBHOOK_URL` is set on the backend if Feishu
   operator notifications are desired.
4. Node.js and Playwright are installed on the machine that keeps the Goofish
   browser logged in.

Install Playwright once:

```bash
npm install -g playwright
npx playwright install chromium
```

## Safe Dry Run

The helper uses its own persistent browser profile, not the Codex in-app browser
session. Log in once before scanning:

```bash
node tools/xianyu_auto_fulfill.mjs --login
```

If Goofish still asks for login after the helper is restarted, use the target
watch command directly and finish login in that same browser window. Watch mode
waits for login before continuing.

Dry-run scanning does not create redeem codes and does not send Goofish IM:

```bash
node tools/xianyu_auto_fulfill.mjs --scan-only
```

For selector testing against old completed chats:

```bash
node tools/xianyu_auto_fulfill.mjs --scan-only --include-completed
```

## Create Code Without Sending

Set an admin API key outside the repository:

```bash
set AKIMIRAI_ADMIN_API_KEY=your-admin-api-key
```

Then create the fulfillment record and print the generated delivery message,
without sending it to Goofish:

```bash
node tools/xianyu_auto_fulfill.mjs --fulfill
```

## Enable Automatic Sending

After dry-run output looks correct:

```bash
node tools/xianyu_auto_fulfill.mjs --send --watch
```

The script opens a persistent browser profile under
`tools/.xianyu-browser-profile`. If Goofish asks for login, finish login in the
opened browser window; watch mode will continue after login is detected.

## Important Options

- `--api-base`: backend API base, default `https://miraiapi.cn`.
- `--api-key-file`: read admin API key from a local file.
- `--profile-dir`: browser profile directory.
- `--state-file`: duplicate-protection state file.
- `--poll-ms`: watch interval, default 15000.
- `--limit`: max visible conversations to inspect per pass.
- `--item-title`: require a text marker in the opened chat before fulfillment.
  Leave empty if the Goofish chat card does not expose the item title.

## Operational Notes

- The script only handles amounts `5`, `10`, `20`, `50`, and `100`.
- If amount cannot be detected from the order card, the conversation is skipped.
- If the chat already contains the delivery manual/redeem text, it is skipped.
- `--send` is the only mode that clicks the Goofish send button.
