# Product Commercialization Audit

审计日期：2026-06-29

目标：在不做大规模重构、不破坏现有支付/转发/后台功能的前提下，识别 Akimirai-api / sub2api 距离商业级 API 聚合平台的关键差距，并给出可落地的 P0/P1/P2 实施顺序。

## 当前能力

### 用户与认证

- 已有用户注册/登录、JWT、管理员权限、2FA/TOTP、邮箱/通知邮箱、OAuth 绑定与 pending session 流程。
- 后端路由在 `backend/internal/server/routes/user.go`、`backend/internal/server/routes/auth.go`、`backend/internal/server/routes/admin.go` 分离用户、认证、后台入口。
- `security_secrets` 表与启动引导逻辑已用于持久化 JWT secret，避免默认 JWT secret 漂移。

涉及文件：

- `backend/ent/schema/user.go`
- `backend/ent/schema/auth_identity.go`
- `backend/ent/schema/pending_auth_session.go`
- `backend/ent/schema/security_secret.go`
- `backend/internal/repository/security_secret_bootstrap.go`
- `backend/internal/handler/auth_*.go`
- `backend/internal/server/routes/user.go`
- `backend/internal/server/routes/admin.go`

### API Key

- 用户可创建、查询、更新、删除 API Key，支持分组绑定、IP 白/黑名单、额度、有效期、5h/1d/7d 费用级限流。
- Gateway 鉴权从 `Authorization: Bearer`、`x-api-key`、`x-goog-api-key` 获取 key，并显式拒绝 query 参数传 key。
- 删除 API Key 时有 `deleted_api_key_audits` 反查能力，用于无效 key 归因。
- 用户端有 Keys 页面、EndpointPopover、UseKeyModal，具备端点展示、复制和多平台客户端接入示例。

主要差距：

- `api_keys.key` 明文唯一存储，服务层和 Redis auth cache 都围绕原始 key 查询/失效。
- 前端列表、复制、UseKeyModal 依赖后端持续返回完整 key，不符合“只显示一次”。
- 删除审计会保存被删除 key 明文。

涉及文件：

- `backend/ent/schema/api_key.go`
- `backend/internal/repository/api_key_repo.go`
- `backend/internal/repository/api_key_cache.go`
- `backend/internal/service/api_key.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/service/api_key_auth_cache_invalidate.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_google.go`
- `backend/internal/handler/api_key_handler.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/handler/dto/mappers.go`
- `frontend/src/views/user/KeysView.vue`
- `frontend/src/components/keys/UseKeyModal.vue`
- `frontend/src/api/keys.ts`
- `frontend/src/types`

### 模型、渠道与账号池

- 已有 group/model/channel 概念，支持不同平台 OpenAI/Gemini/Anthropic/Antigravity 等网关映射。
- 账号池以 `accounts`、`account_groups`、调度/限流字段承载，支持 api_key/oauth/cookie 等凭证形态、并发、优先级、速率/过载/临时不可调度、会话窗口、自动暂停过期账号。
- 后台有账号 CRUD、批量导入、账号测试、隐私设置、模型同步、代理、TLS 指纹模板等运维能力。

主要差距：

- `accounts.credentials` JSONB 当前直接存储 `api_key`、`access_token`、`refresh_token`、cookie/session/service account 等上游敏感凭证。
- 前端 DTO 已做凭证脱敏展示，但数据库与备份仍是明文敏感面。
- refresh token 候选查询依赖 JSONB key 存在，后续加密需要保持 key 结构不被破坏。

涉及文件：

- `backend/ent/schema/account.go`
- `backend/internal/repository/account_repo.go`
- `backend/internal/service/account.go`
- `backend/internal/service/account_credentials_redact.go`
- `backend/internal/service/account_credentials_persistence.go`
- `backend/internal/service/token_refresh_service.go`
- `backend/internal/service/*gateway_service.go`
- `backend/internal/handler/admin/account_handler.go`
- `backend/internal/handler/dto/mappers.go`
- `frontend/src/views/admin/accounts*`

### 计费、扣费与使用日志

- `usage_logs` 记录 user/api_key/account/group/subscription、request_id、模型、token、cost、billing mode/tier、stream/ws 标记、端点、延迟与客户端信息。
- `usage_logs` 通过 `(request_id, api_key_id)` 冲突忽略降低重复写入风险。
- API Key 有 quota/quota_used 与窗口 usage 字段，支持限额与限流。
- 用户与后台均有 usage 查询与 dashboard 统计接口。

主要差距：

- 扣费流水可追溯性基本具备，但商业化对账还需要明确“用户余额变更流水”和 usage/payment/subscription/redeem 的统一关联视图。
- usage log 不存 prompt/body 是优点；需持续防止错误日志、系统日志旁路保存请求正文。

涉及文件：

- `backend/ent/schema/usage_log.go`
- `backend/internal/service/usage_log.go`
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/handler/usage_handler.go`
- `backend/internal/server/routes/user.go`
- `backend/internal/server/routes/admin.go`

### 支付、订单、回调与履约

- 支持 EasyPay、Alipay、Wxpay、Stripe、Airwallex 等 provider。
- Webhook 入口先按 out_trade_no 解析订单/provider instance，再调用 provider `VerifyNotification` 验签。
- `payment_orders.out_trade_no` 有唯一约束，订单包含支付金额、优惠、provider snapshot、状态、退款字段、用户快照等。
- `PaymentService.HandlePaymentNotification` 根据 out_trade_no 查单，校验 provider、metadata、金额，状态迁移使用条件更新，减少重复履约。
- 余额充值履约使用 redeem code 幂等判断；订阅履约用 audit log 防止重复延长。
- `PaymentOrderExpiryService` 定期执行过期处理并主动对账近期微信待支付订单；用户也可按 out_trade_no 主动 verify。
- 后台有支付 dashboard、订单详情、取消、重试履约、个人码确认、退款、计划与 provider instance 管理。

主要差距：

- 支付 provider 配置目前新记录按明文 JSON 存储，旧 AES 只做兼容解密；商用场景应恢复敏感字段加密。
- 回调 raw body 在验签失败时 debug 日志保留截断内容，需确保生产日志级别和脱敏策略一致。
- 对账能力当前偏微信 pending reconcile；其他 provider 的周期性主动对账/差异报表可作为 P1。

涉及文件：

- `backend/ent/schema/payment_order.go`
- `backend/ent/schema/payment_audit_log.go`
- `backend/ent/schema/payment_provider_instance.go`
- `backend/internal/handler/payment_webhook_handler.go`
- `backend/internal/service/payment_fulfillment.go`
- `backend/internal/service/payment_order_lifecycle.go`
- `backend/internal/service/payment_order_expiry_service.go`
- `backend/internal/service/payment_config_providers.go`
- `backend/internal/payment/provider/*.go`
- `backend/internal/server/routes/payment.go`

### 日志、错误解释与后台监控

- `usage_logs` 不保存原始 prompt/body。
- Gateway 路由已接入 `OpsErrorLoggerMiddleware`，记录失败请求、recovered upstream error、upstream events、延迟分段、request/upstream endpoint、request type。
- OpsService 已在持久化前对 `error_body`、`upstream_error_message`、`upstream_error_detail`、`upstream_errors` 做清洗和截断。
- 后台有 Ops dashboard、request errors、upstream errors、system logs、alerts、realtime traffic、account availability、channel monitor。
- 用户侧有 `/usage/errors`，可查看与自己有关的错误请求。

主要差距：

- 脱敏逻辑存在，但目前分散在 gateway、ops、upstream message helper 中，缺少统一的“商业级敏感词/凭证模式”入口和测试覆盖。
- 错误解释层以分类字段为主，用户端缺少稳定、可本地化的 `explanation/actionable_hint` 字段来解释失败请求、流式中断、上游错误。
- Ops 错误中的 API Key prefix 依赖 `apiKey.Key`，API Key 哈希/只显示一次后需要改为稳定 prefix 字段。
- 后台操作审计存在支付审计与删除 key 审计，但一般后台配置/账号/渠道操作缺少统一操作审计表。

涉及文件：

- `backend/internal/handler/ops_error_logger.go`
- `backend/internal/service/ops_service.go`
- `backend/internal/service/ops_port.go`
- `backend/internal/service/ops_upstream_context.go`
- `backend/internal/repository/ops*.go`
- `backend/internal/server/routes/gateway.go`
- `backend/internal/server/routes/admin.go`
- `frontend/src/views/admin/ops*`
- `frontend/src/views/user/UsageView.vue`

### 配置与安全项

- 配置体系覆盖支付、网关、隐私过滤、Ops、渠道监控、邮件通知、内容风控、代理、TLS 指纹等。
- `PrivacyFilter` 可按配置改写 JSON 请求体，减少转发给上游的隐私字段。
- TOTP secret 与 Channel Monitor API Key 已使用 `SecretEncryptor` / AES-GCM 模式，可复用到账号凭证。
- OpenAI/Gemini/Antigravity OAuth 与 token refresh 体系已较完整。

主要差距：

- SecretEncryptor 当前依赖 TOTP encryption key；如果未配置，敏感加密链路会不可用。商业部署应把“主加密密钥必须配置”作为启动/健康检查项。
- 支付配置和账号凭证加密策略不统一。
- 缺少统一密钥轮换/重加密任务与“哪些字段已加密”的健康检查。

涉及文件：

- `backend/internal/repository/aes_encryptor.go`
- `backend/internal/service/totp_service.go`
- `backend/internal/service/channel_monitor_service.go`
- `backend/internal/payment/crypto.go`
- `backend/internal/service/payment_config_providers.go`
- `backend/internal/server/middleware/privacy_filter.go`
- `backend/internal/service/setting_service.go`

## 风险点与优先级

| 优先级 | 风险/差距 | 当前判断 | 影响 | 涉及文件 |
| --- | --- | --- | --- | --- |
| P0 | API Key 明文存储且持续返回 | `api_keys.key` 为唯一明文字段，列表/详情/复制/UseKeyModal 依赖完整 key | 数据库、备份、Redis key、接口响应泄露后可直接调用 | `backend/ent/schema/api_key.go`, `backend/internal/repository/api_key_repo.go`, `backend/internal/service/api_key_service.go`, `frontend/src/views/user/KeysView.vue` |
| P0 | 上游敏感凭证明文存储 | `accounts.credentials` JSONB 保存 api_key/token/refresh_token/cookie/service account | 数据库/备份泄露导致上游账号直接失控 | `backend/ent/schema/account.go`, `backend/internal/repository/account_repo.go`, `backend/internal/service/account_credentials_persistence.go` |
| P0 | 错误日志脱敏与解释不够统一 | usage log 不存 prompt；ops 已清洗，但规则分散且用户可解释字段不足 | 敏感错误体旁路入库；用户无法判断失败原因或下一步 | `backend/internal/handler/ops_error_logger.go`, `backend/internal/service/ops_service.go`, `backend/internal/service/ops_upstream_context.go` |
| P0 | API Key 删除审计保存明文 key | `deleted_api_key_audits` 插入被删 key 原文 | 删除后仍保留可用历史密钥 | `backend/internal/repository/api_key_repo.go`, `backend/migrations/145_deleted_api_key_audit.sql` |
| P1 | 支付 provider 配置明文 | 新 provider config 明文 JSON，旧 AES fallback only | 支付商户 secret/API key 暴露 | `backend/internal/service/payment_config_providers.go`, `backend/internal/payment/crypto.go` |
| P1 | 支付对账范围不完整 | 微信 pending 周期 reconcile；其他 provider 主要依赖回调/用户 verify | 回调丢失时部分渠道恢复慢 | `backend/internal/service/payment_order_lifecycle.go`, `backend/internal/service/payment_order_expiry_service.go` |
| P1 | 统一后台操作审计不足 | 支付 audit/key 删除 audit 存在，但账号/渠道/配置修改没有统一审计面 | 商业后台缺少“谁在何时改了什么”的追溯 | `backend/internal/server/routes/admin.go`, `backend/internal/handler/admin/*` |
| P1 | 扣费/余额统一流水视图不足 | usage/payment/redeem/subscription 各自有记录，但缺少统一 reconciliation view | 财务客服排障成本高 | `backend/ent/schema/usage_log.go`, `backend/ent/schema/payment_order.go`, `backend/internal/service/redeem_service.go` |
| P1 | 加密密钥健康检查不足 | AES-GCM 已有实现，但没有针对商业敏感存储统一强制检查/轮换 | 部署误配置时安全特性失效 | `backend/internal/repository/aes_encryptor.go`, `backend/internal/service/setting_service.go` |
| P2 | 新手接入路径可进一步产品化 | 已有 guide、endpoint、use key 示例，但一次显示后需新增保存提醒和再生成路径 | 降低首次 API 调用转化 | `frontend/src/views/user/KeysView.vue`, `frontend/src/components/keys/UseKeyModal.vue`, `frontend/src/components/Guide/steps.ts` |
| P2 | 渠道健康与账号池容量解释可增强 | Channel monitor/Ops 已有，需更聚合地展示可用容量、错误归因、建议动作 | 运维效率提升 | `backend/internal/service/channel_monitor_service.go`, `frontend/src/views/admin/*monitor*` |

## 非破坏性迁移方案

### API Key 哈希存储

建议新增字段：

- `api_keys.key_hash`：可空字符串，保存 HMAC-SHA256 后的版本化 hash，例如 `ak_hmac_sha256_v1:<hex>`。
- `api_keys.key_prefix`：可空字符串，保存展示/排障前缀，例如前 8 位。

迁移：

1. 新增 nullable 字段和唯一索引：`key_hash IS NOT NULL AND key_hash <> ''`。
2. 新建 key 只保存 hash/prefix；`key` 旧字段保留非敏感占位值，避免破坏旧 schema 与唯一约束。
3. 鉴权先按 hash 查找，未命中再按 legacy 明文 `key` 查找。
4. legacy 命中时可机会式回填 hash/prefix，但不立即清空 legacy 明文，避免破坏仍依赖旧字段的逻辑。
5. API 响应只在创建成功的当次响应返回原始 key；列表/详情/更新返回 masked key 与 prefix。

回滚：

- 保留 legacy `key` 查询逻辑即可回滚代码。
- 新字段可不删除；若必须回滚 DB，先确认没有仅 hash 存储的新 key，否则这些 key 无法按旧代码认证。

### 上游敏感凭证加密

建议不新增字段，沿用 `accounts.credentials` JSONB，按 key 保持原结构，仅加密敏感字段值。

迁移：

1. 使用现有 `SecretEncryptor` AES-GCM，增加版本前缀，如 `enc:v1:<ciphertext>`。
2. 仅加密敏感 key：`api_key`、`access_token`、`refresh_token`、`id_token`、`session_key`、`cookie`、`aws_secret_access_key`、`aws_session_token`、`service_account_json`、`service_account`、`private_key` 等。
3. 仓储写入前加密，读出后解密给 service；已是 `enc:v1:` 的值不重复加密。
4. 未加密 legacy 明文读出兼容，下一次更新/刷新 token 时自动加密。
5. 保持 JSONB 顶层 key 不变，避免破坏 `refresh_token` 候选查询。

回滚：

- 代码保留 decrypt-read 兼容；如需回滚到旧代码，必须先运行离线解密脚本把 `enc:v1:` 值恢复为明文。
- 不建议删除加密前缀或批量覆盖 credentials；优先代码级回滚。

### 请求日志脱敏与错误解释层

建议不新增字段先完成 P0：

1. 把 ops 错误写入前的敏感信息清洗收敛为单一 helper。
2. 扩展敏感模式：Authorization/Bearer/API key/OAuth token/refresh token/cookie/session/private key/service account/email-like secret URL 参数。
3. 增加错误解释 helper，把 status/error_type/upstream status/stream/request_type 映射为稳定 explanation/actionable hint。
4. 在用户错误详情 DTO 中输出解释字段；后台仍保留脱敏后的 detail。

回滚：

- 字段为衍生展示，不改变数据库；回滚只需移除 DTO 字段或保留为空。

## 建议实现顺序

1. `P0-1 API Key 哈希存储与只显示一次`
   - 先加 additive migration 与 ent 字段。
   - 后端支持 hash auth + legacy fallback。
   - API 响应改为 create-only raw key，其他场景 masked。
   - 前端新增 create secret 一次显示提示；列表禁用复制 masked key，UseKeyModal 只在有原始 key 时可复制。

2. `P0-2 上游敏感凭证加密存储`
   - 复用现有 AES-GCM `SecretEncryptor`。
   - 在 account repository 边界做读解密/写加密，避免扩散到 gateway/token refresh。
   - 添加 unit test 覆盖明文兼容、加密写入、避免双重加密、refresh_token key 保持。

3. `P0-3 请求日志脱敏与错误解释层`
   - 强化 ops 清洗 helper 与测试。
   - 给用户错误详情增加 explanation/actionable hint。
   - 确认 usage_log 仍不保存 prompt/body，ops error body 不含 token/prompt-like sensitive payload。

4. P1：支付 provider 配置敏感字段加密。
5. P1：多 provider 主动对账、差异报表、后台统一操作审计。
6. P2：优化新手接入 UX、渠道健康容量说明、后台排障工作台。

## 验证要求

每完成一个 P0 项必须运行现有可验证命令，并记录结果：

- 后端优先：`cd backend && go test ./...`
- 后端构建：`cd backend && make build` 或等价 `go build ./cmd/server`
- 前端涉及时：`cd frontend && pnpm run lint:check && pnpm run typecheck && pnpm run test:run && pnpm run build`
- 根目录可选全量：`make test`、`make build`、`make secret-scan`

如某条命令因环境缺少工具或测试依赖不可运行，需要记录实际执行命令、失败原因和替代验证结果。
