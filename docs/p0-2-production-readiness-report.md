# P0-2 Production Readiness Report

日期：2026-06-29

## 1. 本轮完成内容

### API Key 安全验收

- 新建 API Key 走 HMAC-SHA256 hash 存储，`api_keys.key` 写入 hash placeholder，`key_hash/key_prefix` 保存可认证与可展示字段。
- 用户列表/详情通过 `DisplayKey()` 输出 masked key；只有创建响应中的 `key_visible_once=true` 会返回完整 key。
- 创建 API Key 的幂等缓存响应会脱敏，避免重放完整 key。
- legacy 明文 key 仍可认证；首次认证或删除路径会回填 `key_hash/key_prefix`。
- 删除审计对 hash key 写入 `key_hash/key_prefix`，不再为新 key 新增明文审计。

### 上游敏感凭证加密验收

- 账号 credentials 写入时对敏感字段做 `sub2api_enc_v1:` AES-GCM envelope 加密，读取时自动解密。
- 覆盖字段扩展到 `api_key/access_token/refresh_token/id_token/client_secret/cookie/session_token/private_key/service_account/service_account_json` 及常见 camelCase/lowercase 别名。
- legacy 明文凭证保持可读；新写入或更新的敏感字段会加密。

### 日志脱敏与错误解释层

- ops/upstream error 入库前继续走 `logredact` 脱敏，覆盖 token/cookie/key/private_key 与 prompt/messages/content/input/text 等字段。
- 移除了 TOTP 调试日志中的 secret/decrypted prefix 输出，只保留长度与匹配状态。
- 增加用户侧错误码枚举与解释元数据：`user_message/admin_hint/retryable/charged/http_status/suggestion`。
- OpenAI-compatible 网关响应未改动；用户错误列表/详情只展示 `error_code/explanation/suggestion/retryable/charged/http_status` 等用户可见字段，不暴露 `admin_hint`。

### 用户端接入闭环

- 用户 API Key 页面新增“接入配置示例”，覆盖 OpenAI SDK、Codex、Cursor、Cherry Studio。
- Base URL 和示例配置支持一键复制，使用占位 key，不展示或缓存真实 secret。

### 生产上线文档

- 新增 `docs/production-readiness-checklist.md`，覆盖迁移顺序、回滚方案、环境变量、密钥、PostgreSQL/Redis、支付回调、HTTPS、CORS、trusted proxy、日志脱敏、备份恢复、管理员账号、风控开关和上线观察指标。

## 2. 主要涉及文件

- API Key hash/一次展示：`backend/internal/service/api_key_hash.go`、`backend/internal/service/api_key.go`、`backend/internal/service/api_key_service.go`、`backend/internal/repository/api_key_repo.go`、`backend/internal/handler/api_key_handler.go`、`backend/internal/handler/dto/mappers.go`、`backend/migrations/158_add_api_key_hash_columns.sql`
- 上游凭证加密：`backend/internal/repository/account_credentials_crypto.go`、`backend/internal/repository/account_repo.go`、`backend/internal/service/account_credentials_redact.go`
- 日志脱敏：`backend/internal/service/ops_service.go`、`backend/internal/util/logredact/redact.go`、`backend/internal/service/totp_service.go`
- 错误解释层：`backend/internal/service/ops_user_error.go`、`backend/internal/service/ops_user_error_test.go`、`backend/internal/service/ops_user_error_cyber_test.go`
- 用户端展示：`frontend/src/components/user/UserErrorRequestsTable.vue`、`frontend/src/components/user/UserErrorDetailModal.vue`、`frontend/src/types/index.ts`
- 接入示例：`frontend/src/views/user/KeysView.vue`、`frontend/src/i18n/locales/zh.ts`、`frontend/src/i18n/locales/en.ts`
- 文档与版本控制：`.gitignore`、`docs/production-readiness-checklist.md`、`docs/p0-2-production-readiness-report.md`

## 3. 验证结果

| 命令 | 结果 | 备注 |
| --- | --- | --- |
| `go test ./...` | 通过 | 后端全量测试通过 |
| `go build -o "$env:TEMP\\sub2api-server-p0-2.exe" ./cmd/server` | 通过 | 后端可构建 |
| `go test ./internal/service ...` | 通过 | 错误解释、API Key、ops redaction 相关窄测 |
| `go test ./internal/repository ...` | 通过 | API Key/账号凭证加密相关窄测 |
| `go test ./internal/util/logredact -count=1` | 通过 | 日志脱敏单包测试 |
| `go test -tags unit ./internal/service/account_credentials_redact.go ./internal/service/account_credentials_redact_test.go` | 通过 | 文件级验证敏感凭证 key 识别；完整 `-tags unit ./internal/service` 被既有测试桩阻断 |
| `pnpm typecheck` | 阻断 | pnpm ignored-builds 策略拒绝 esbuild/vue-demi build scripts |
| `pnpm lint:check` | 阻断 | 同上 |
| `./node_modules/.bin/vue-tsc.CMD --noEmit` | 通过 | pnpm 阻断后的等价 typecheck |
| `./node_modules/.bin/eslint.CMD ...` | 通过 | 单独重跑，避开 Vitest 临时文件竞态 |
| `./node_modules/.bin/vue-tsc.CMD -b` | 通过 | frontend build type phase |
| `./node_modules/.bin/vite.CMD build` | 通过 | 仅有既有 chunk/import 警告 |
| `./node_modules/.bin/vitest.CMD run` | 失败 | 117 个文件中 7 个失败、17 个断言失败，失败点与本轮改动无关 |
| `./node_modules/.bin/vitest.CMD run src/components/__tests__/ApiKeyCreate.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts src/components/keys/__tests__/EndpointPopover.spec.ts src/components/user/profile/__tests__/ProfilePasswordForm.spec.ts` | 通过 | 用户接入/API Key 相关关键用例 13 个通过 |
| `git diff --check` | 通过 | 仅提示 `.gitignore` 下次触碰会 LF->CRLF |

## 4. 未解决风险

- 生产必须配置稳定的 `totp.encryption_key`。账号凭证加密复用该密钥；如果使用自动生成密钥，重启后已写入密文可能不可解。
- 回滚到不支持 hash/encryption 的旧二进制存在数据格式风险：新 API Key 可能无法认证，`sub2api_enc_v1:` 凭证会被旧逻辑当作普通字符串。
- legacy 明文 API Key 仍依赖首次认证/删除路径完成回填；尚未提供后台批量回填任务。
- 支付核心本轮未改动；仍建议下一轮补全 provider 配置加密验收、跨 provider 周期性主动对账、订单/余额/usage 统一 reconciliation view。
- 前端全量 Vitest 存在既有失败：平台配额测试仍按 4 平台断言、AccountUsageCell spy 参数断言未适配当前实现、部分图表测试缺少 cost 字段兜底、page size 测试期望与当前默认值不一致。

## 5. 下一轮建议

1. 增加生产启动前 readiness guard：生产模式下若 `totp.encryption_key` 非手动配置则拒绝启动或进入只读维护模式。
2. 增加 legacy API Key 批量 hash 回填后台任务，回填后再禁止明文 key 认证路径。
3. 补支付 provider 配置加密和订单/余额/usage 对账视图。
4. 修复前端既有 Vitest 失败，恢复全量前端测试绿线。
5. 为用户错误解释 UI 增加组件级测试，覆盖 `suggestion/retryable/charged/error_code` 展示。
