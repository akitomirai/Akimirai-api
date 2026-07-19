#!/usr/bin/env node
import { createHash } from 'node:crypto'
import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { createRequire } from 'node:module'

const require = createRequire(import.meta.url)

const __dirname = dirname(fileURLToPath(import.meta.url))
const defaultProfileDir = resolve(__dirname, '.xianyu-browser-profile')
const defaultStateFile = resolve(__dirname, '.xianyu-auto-fulfill-state.json')
const defaultAPIBase = 'https://miraiapi.cloud'
const defaultItemID = '1061537515030'
const defaultManualURL = 'https://lcn1w5cr3w90.feishu.cn/wiki/KvSDwhig9iTSa5k3g6qcm8U4nrd'
const supportedAmounts = [5, 10, 20, 50, 100]

const targetStatuses = [
  '待发货',
  '买家已付款',
  '等待卖家发货',
  '付款成功',
  '已付款'
]
const completedStatus = '交易成功'

function usage() {
  console.log(`Usage:
  node tools/xianyu_auto_fulfill.mjs [options]

Modes:
  --scan-only          Only scan Goofish IM candidates. This is the default.
  --login              Open Goofish IM and wait for login in the persistent profile.
  --fulfill           Create the Bakaai redeem code, but do not send Goofish IM.
  --send              Create the redeem code and send the delivery message.
  --watch             Keep polling. Without this, the script scans once.

Important options:
  --api-base URL      Backend API base. Default: ${defaultAPIBase}
  --api-key KEY       Admin API key. Prefer AKIMIRAI_ADMIN_API_KEY.
  --api-key-file PATH Read the admin API key from a local file.
  --profile-dir PATH  Persistent browser profile. Default: ${defaultProfileDir}
  --state-file PATH   Sent/fulfilled state file. Default: ${defaultStateFile}
  --item-id ID        Goofish item id. Default: ${defaultItemID}
  --item-title TEXT   Extra chat-detail text required before fulfilling.
  --poll-ms N         Watch polling interval. Default: 15000
  --login-wait-ms N   Login wait time for --login. Default: 600000
  --limit N           Max visible conversations to inspect per pass. Default: 8
  --include-completed Include "交易成功" chats for dry-run selector testing.
  --headless          Run browser headless. Not recommended until login works.

Examples:
  node tools/xianyu_auto_fulfill.mjs --scan-only
  node tools/xianyu_auto_fulfill.mjs --login
  set AKIMIRAI_ADMIN_API_KEY=...
  node tools/xianyu_auto_fulfill.mjs --fulfill --once
  node tools/xianyu_auto_fulfill.mjs --send --watch
`)
}

function parseArgs(argv) {
  const args = {
    apiBase: process.env.AKIMIRAI_API_BASE || defaultAPIBase,
    apiKey: process.env.AKIMIRAI_ADMIN_API_KEY || '',
    apiKeyFile: process.env.AKIMIRAI_ADMIN_API_KEY_FILE || '',
    profileDir: defaultProfileDir,
    stateFile: defaultStateFile,
    itemID: defaultItemID,
    itemTitle: '',
    pollMs: 15000,
    loginWaitMs: 600000,
    limit: 8,
    mode: 'scan',
    watch: false,
    headless: false,
    includeCompleted: false
  }

  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i]
    const next = () => {
      if (i + 1 >= argv.length) throw new Error(`${arg} requires a value`)
      return argv[++i]
    }
    switch (arg) {
      case '-h':
      case '--help':
        args.help = true
        break
      case '--scan-only':
        args.mode = 'scan'
        break
      case '--login':
        args.mode = 'login'
        args.watch = false
        break
      case '--fulfill':
        args.mode = 'fulfill'
        break
      case '--send':
        args.mode = 'send'
        break
      case '--watch':
        args.watch = true
        break
      case '--once':
        args.watch = false
        break
      case '--api-base':
        args.apiBase = next()
        break
      case '--api-key':
        args.apiKey = next()
        break
      case '--api-key-file':
        args.apiKeyFile = next()
        break
      case '--profile-dir':
        args.profileDir = resolve(next())
        break
      case '--state-file':
        args.stateFile = resolve(next())
        break
      case '--item-id':
        args.itemID = next()
        break
      case '--item-title':
        args.itemTitle = next()
        break
      case '--poll-ms':
        args.pollMs = Number.parseInt(next(), 10)
        break
      case '--login-wait-ms':
        args.loginWaitMs = Number.parseInt(next(), 10)
        break
      case '--limit':
        args.limit = Number.parseInt(next(), 10)
        break
      case '--include-completed':
        args.includeCompleted = true
        break
      case '--headless':
        args.headless = true
        break
      default:
        throw new Error(`unknown option: ${arg}`)
    }
  }

  if (!Number.isFinite(args.pollMs) || args.pollMs < 3000) {
    throw new Error('--poll-ms must be at least 3000')
  }
  if (!Number.isFinite(args.loginWaitMs) || args.loginWaitMs < 5000) {
    throw new Error('--login-wait-ms must be at least 5000')
  }
  if (!Number.isFinite(args.limit) || args.limit < 1) {
    throw new Error('--limit must be positive')
  }
  if (!args.apiKey && args.apiKeyFile) {
    args.apiKey = readFileSync(resolve(args.apiKeyFile), 'utf8').trim()
  }
  if ((args.mode === 'fulfill' || args.mode === 'send') && !args.apiKey) {
    throw new Error('AKIMIRAI_ADMIN_API_KEY, --api-key, or --api-key-file is required for --fulfill/--send')
  }
  return args
}

function loadPlaywright() {
  try {
    return require('playwright')
  } catch {
    const candidateRoots = [
      ...String(process.env.NODE_PATH || '').split(process.platform === 'win32' ? ';' : ':'),
      process.env.APPDATA ? resolve(process.env.APPDATA, 'npm', 'node_modules') : '',
      process.env.ProgramFiles ? resolve(process.env.ProgramFiles, 'nodejs', 'node_modules') : '',
      'E:/nodejs/node_global/node_modules',
      'C:/Users/Lenovo/AppData/Roaming/npm/node_modules'
    ].filter(Boolean)

    try {
      const globalRoot = execFileSync('npm root -g', {
        encoding: 'utf8',
        shell: true,
        stdio: ['ignore', 'pipe', 'ignore']
      }).trim()
      candidateRoots.unshift(globalRoot)
    } catch {
      // The fixed candidate list above covers this workstation's Node layout.
    }

    for (const moduleRoot of candidateRoots) {
      try {
        const packagePath = resolve(moduleRoot, 'playwright', 'package.json')
        if (!existsSync(packagePath)) continue
        const globalRequire = createRequire(packagePath)
        return globalRequire('playwright')
      } catch {
        // Try the next possible global module root.
      }
    }

    console.error('Missing dependency: playwright')
    console.error('Install it once with: npm install -g playwright && npx playwright install chromium')
    process.exit(1)
  }
}

function loadState(path) {
  if (!existsSync(path)) return { handled: {} }
  try {
    const parsed = JSON.parse(readFileSync(path, 'utf8'))
    if (parsed && typeof parsed === 'object' && parsed.handled) return parsed
  } catch {
    // Ignore corrupt state and start fresh; the old file remains for inspection.
  }
  return { handled: {} }
}

function saveState(path, state) {
  mkdirSync(dirname(path), { recursive: true })
  writeFileSync(path, JSON.stringify(state, null, 2) + '\n')
}

function stableHash(input) {
  return createHash('sha256').update(input).digest('hex').slice(0, 20)
}

function parseAmount(text) {
  const matches = [...String(text).matchAll(/¥\s*([0-9]+(?:\.[0-9]+)?)/g)]
  for (const match of matches) {
    const value = Number.parseFloat(match[1])
    const rounded = Math.round(value)
    if (Math.abs(value - rounded) < 0.001 && supportedAmounts.includes(rounded)) {
      return rounded
    }
  }
  return null
}

function statusWanted(text, includeCompleted) {
  if (targetStatuses.some(status => text.includes(status))) return true
  return includeCompleted && text.includes(completedStatus)
}

function alreadyDelivered(text) {
  return text.includes('您的兑换码：') ||
    text.includes('兑换入口：') ||
    text.includes(defaultManualURL)
}

async function getConversationRows(page, includeCompleted) {
  return page.evaluate(({ includeCompletedArg, statuses, completed }) => {
    const rows = []
    const elements = Array.from(document.querySelectorAll('div[class*="conversation-item"]'))
    for (let index = 0; index < elements.length; index++) {
      const el = elements[index]
      const rect = el.getBoundingClientRect()
      const text = (el.innerText || '').replace(/\s+/g, ' ').trim()
      if (!text || rect.width < 100 || rect.height < 30) continue
      const wanted = statuses.some(status => text.includes(status)) ||
        (includeCompletedArg && text.includes(completed))
      rows.push({
        index,
        text,
        wanted,
        rect: {
          x: Math.round(rect.left),
          y: Math.round(rect.top),
          w: Math.round(rect.width),
          h: Math.round(rect.height)
        }
      })
    }
    return rows
  }, { includeCompletedArg: includeCompleted, statuses: targetStatuses, completed: completedStatus })
}

async function getIMDiagnostics(page, rows) {
  return page.evaluate(({ statuses, completed, knownRows }) => {
    const bodyText = (document.body?.innerText || '').replace(/\r/g, '').trim()
    const loginHints = ['登录', '扫码登录', '账号登录', '安全验证', '请完成验证']
    const statusHits = statuses.filter(status => bodyText.includes(status))
    const bodySample = bodyText.split('\n')
      .map(line => line.trim())
      .filter(Boolean)
      .slice(0, 12)
    return {
      title: document.title,
      url: location.href,
      looksLikeLogin: loginHints.some(hint => bodyText.includes(hint)),
      rowCount: knownRows.length,
      statusHits,
      completedHits: bodyText.includes(completed),
      bodySample,
      rowSample: knownRows.slice(0, 5).map(row => row.text)
    }
  }, { statuses: targetStatuses, completed: completedStatus, knownRows: rows })
}

async function clickConversation(page, row) {
  await page.mouse.click(row.rect.x + Math.min(120, row.rect.w / 2), row.rect.y + row.rect.h / 2)
  await page.waitForTimeout(1200)
}

async function readChatDetail(page) {
  return page.evaluate(() => {
    const main = document.querySelector('main')
    const scope = main || document.body
    const text = (scope.innerText || '').replace(/\r/g, '').trim()
    const orderLinks = Array.from(scope.querySelectorAll('a'))
      .map(a => (a.innerText || '').replace(/\s+/g, ' ').trim())
      .filter(t => t.includes('¥'))
    const headerLines = text.split('\n').map(s => s.trim()).filter(Boolean).slice(0, 8)
    return { text, orderLinks, headerLines }
  })
}

function buildOrderIdentity(row, detail, amount) {
  const seed = [
    row.text,
    amount,
    detail.orderLinks.join('|'),
    detail.headerLines.join('|')
  ].join('\n')
  return `xianyu-chat-${stableHash(seed)}`
}

async function createFulfillment(args, orderID, buyerRef, amount) {
  const payload = {
    platform: 'xianyu',
    platform_order_id: orderID,
    buyer_ref: buyerRef,
    sku_code: `bakaai-balance-${amount}`,
    amount,
    currency: 'CNY',
    notify_feishu: true
  }
  const response = await fetch(`${args.apiBase.replace(/\/$/, '')}/api/v1/admin/payment/external-fulfillments`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'x-api-key': args.apiKey
    },
    body: JSON.stringify(payload)
  })
  const body = await response.json().catch(() => ({}))
  if (!response.ok || body.code !== 0) {
    throw new Error(`fulfillment failed: http=${response.status} body=${JSON.stringify(body).slice(0, 800)}`)
  }
  const fulfillment = body.data?.fulfillment
  if (!fulfillment?.delivery_message) {
    throw new Error('fulfillment response did not include delivery_message')
  }
  return { replay: Boolean(body.data?.replay), fulfillment }
}

async function sendGoofishMessage(page, message) {
  const textarea = page.locator('textarea[placeholder*="请输入消息"]')
  await textarea.waitFor({ state: 'visible', timeout: 10000 })
  await textarea.fill(message)
  const sendButton = page.getByRole('button', { name: '发 送' })
  await sendButton.waitFor({ state: 'visible', timeout: 10000 })
  await sendButton.click()
  await page.waitForTimeout(1500)
}

function buyerRefFromDetail(detail, row) {
  const first = detail.headerLines.find(line => line && !line.includes('¥') && !line.includes('闲鱼号'))
  return first || row.text.split(' ')[0] || ''
}

async function scanOnce(page, args, state) {
  await page.goto('https://www.goofish.com/im', { waitUntil: 'domcontentloaded' })
  await page.waitForTimeout(2000)

  const allRows = await getConversationRows(page, args.includeCompleted)
  const rows = allRows
    .filter(row => row.wanted)
    .slice(0, args.limit)

  if (rows.length === 0) {
    const diagnostics = await getIMDiagnostics(page, allRows)
    console.log(`[${new Date().toISOString()}] no paid-order conversations found`)
    console.log(JSON.stringify(diagnostics, null, 2))
    if (diagnostics.looksLikeLogin && args.watch) {
      console.log('Goofish login is required in the opened browser window.')
      await waitForLogin(page, args)
    }
    return
  }

  for (const row of rows) {
    if (!statusWanted(row.text, args.includeCompleted)) continue
    await clickConversation(page, row)
    const detail = await readChatDetail(page)
    if (args.itemTitle && !detail.text.includes(args.itemTitle)) {
      console.log(`skip: item title marker missing (${args.itemTitle}) row="${row.text}"`)
      continue
    }
    if (!statusWanted(detail.text, args.includeCompleted)) {
      console.log(`skip: status marker missing after open row="${row.text}"`)
      continue
    }
    if (alreadyDelivered(detail.text)) {
      console.log(`skip: delivery text already present row="${row.text}"`)
      continue
    }

    const amount = parseAmount(`${detail.orderLinks.join(' ')} ${detail.text}`)
    if (!amount) {
      console.log(`skip: supported amount not detected row="${row.text}"`)
      continue
    }

    const orderID = buildOrderIdentity(row, detail, amount)
    if (state.handled[orderID]) {
      console.log(`skip: already handled ${orderID}`)
      continue
    }

    const buyerRef = buyerRefFromDetail(detail, row)
    console.log(`candidate: order=${orderID} amount=${amount} buyer="${buyerRef}" row="${row.text}"`)

    if (args.mode === 'scan') {
      continue
    }

    const result = await createFulfillment(args, orderID, buyerRef, amount)
    console.log(`fulfilled: id=${result.fulfillment.id} replay=${result.replay} notify=${result.fulfillment.notify_status}`)

    if (args.mode === 'send') {
      await sendGoofishMessage(page, result.fulfillment.delivery_message)
      console.log(`sent: order=${orderID}`)
      state.handled[orderID] = {
        amount,
        buyerRef,
        fulfillmentID: result.fulfillment.id,
        sentAt: new Date().toISOString()
      }
      saveState(args.stateFile, state)
    } else {
      console.log('delivery_message:')
      console.log(result.fulfillment.delivery_message)
      state.handled[orderID] = {
        amount,
        buyerRef,
        fulfillmentID: result.fulfillment.id,
        fulfilledAt: new Date().toISOString()
      }
      saveState(args.stateFile, state)
    }
  }
}

async function waitForLogin(page, args) {
  await page.goto('https://www.goofish.com/im', { waitUntil: 'domcontentloaded' })
  const startedAt = Date.now()
  console.log(`login mode: waiting up to ${args.loginWaitMs}ms for Goofish IM login`)
  while (Date.now() - startedAt < args.loginWaitMs) {
    await page.waitForTimeout(3000)
    const rows = await getConversationRows(page, true)
    const diagnostics = await getIMDiagnostics(page, rows)
    if (!diagnostics.looksLikeLogin && rows.length > 0) {
      console.log('login ready:')
      console.log(JSON.stringify(diagnostics, null, 2))
      return
    }
    console.log(`waiting for login... title="${diagnostics.title}" rows=${rows.length} login=${diagnostics.looksLikeLogin}`)
  }
  throw new Error('login wait timed out; rerun --login after finishing Goofish login')
}

async function main() {
  const args = parseArgs(process.argv.slice(2))
  if (args.help) {
    usage()
    return
  }

  const { chromium } = loadPlaywright()
  mkdirSync(args.profileDir, { recursive: true })
  const state = loadState(args.stateFile)
  const context = await chromium.launchPersistentContext(args.profileDir, {
    headless: args.headless,
    channel: process.env.XIANYU_BROWSER_CHANNEL || 'msedge',
    viewport: { width: 1280, height: 900 }
  })

  const page = context.pages()[0] || await context.newPage()
  console.log(`mode=${args.mode} watch=${args.watch} profile=${args.profileDir}`)
  console.log('If Goofish asks for login, finish login in the opened browser window first.')

  try {
    if (args.mode === 'login') {
      await waitForLogin(page, args)
    } else {
      do {
        await scanOnce(page, args, state)
        if (args.watch) await page.waitForTimeout(args.pollMs)
      } while (args.watch)
    }
  } finally {
    await context.close()
  }
}

main().catch(err => {
  console.error(err?.stack || err?.message || String(err))
  process.exit(1)
})
