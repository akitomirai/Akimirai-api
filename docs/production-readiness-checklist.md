# Production Readiness Checklist

> 适用范围：Sub2API/Akimirai-api 商业化上线前检查。每次生产发布前逐项打勾；涉及数据库或密钥的变更必须先在预发环境演练。

## 1. 发布前停止条件

- [ ] 没有未评估的破坏性数据库迁移。
- [ ] 已完成数据库全量备份，并验证至少一次恢复流程。
- [ ] 已确认生产 `totp.encryption_key` 为固定 64 hex 字符串，不使用自动生成密钥。
- [ ] 已确认 `jwt.secret` 为强随机值；API Key hash 使用稳定密钥源。
- [ ] 已确认日志、错误详情、前端用户错误页不会展示原始 prompt/messages/token/cookie/private_key/api_key。
- [ ] 支付回调、订单验签、幂等、主动查询/对账路径在预发环境通过。
- [ ] 管理员账号、支付商户账号、上游账号、数据库、Redis 均已更换默认密码/密钥。

## 2. 数据库迁移顺序

1. 停止自动发布，记录当前运行镜像/二进制版本。
2. 对 PostgreSQL 做全量备份，并记录备份文件校验值。
3. 在预发环境按 `backend/migrations` 的数字顺序执行迁移。
4. 重点确认 `158_add_api_key_hash_columns.sql` 为增量迁移：
   - `api_keys.key_hash`
   - `api_keys.key_prefix`
   - `deleted_api_key_audits.key_hash`
   - `deleted_api_key_audits.key_prefix`
   - `api_keys.key_hash` 唯一索引与 key prefix 辅助索引
5. 部署新后端，启动后执行 smoke test：
   - 登录后台。
   - 创建 API Key，确认完整 key 只显示一次。
   - 使用新 key 调用 `/v1/chat/completions` 或目标兼容接口。
   - 删除 key 后确认用户错误归属仍可按 key prefix/hash 查到。
6. 部署前端，确认用户端“API Key -> 接入配置示例 -> 用量/错误 -> 充值”路径可达。

## 3. 回滚方案

- [ ] 优先采用“回滚二进制/镜像，保留新增字段”的方式；`158` 为加字段迁移，旧版本通常可忽略新增列。
- [ ] 如果已经创建了 hash 存储的新 API Key，回滚到旧二进制后旧逻辑可能无法认证这些新 key；回滚时应通知用户重新创建 key，或恢复发布前备份。
- [ ] 如果已经写入 `sub2api_enc_v1:` 账号凭证密文，回滚到不支持解密的旧二进制会导致上游账号不可用；必须先继续使用新版本导出/重写凭证，或恢复发布前备份。
- [ ] 不建议在生产回滚时删除新增列；只有在确认没有新版本数据依赖后，才可在维护窗口执行 drop index/drop column。
- [ ] 支付相关回滚不得删除订单、流水、回调日志；如需回滚支付配置，先导出当前 provider 配置和订单状态。

## 4. 环境变量与配置

- [ ] `server.mode` 为 `release`。
- [ ] `server.frontend_url`、公开 API Base URL 与反向代理域名一致。
- [ ] `server.trusted_proxies` 仅包含真实反向代理 IP/CIDR。
- [ ] `cors.allowed_origins` 只配置正式前端域名；`allow_credentials=true` 时不得使用通配来源。
- [ ] `jwt.secret` 为强随机值，并在多实例间一致。
- [ ] `totp.encryption_key` 为固定 64 hex 字符串，并进入密钥管理/备份流程。
- [ ] PostgreSQL 使用独立生产账号，最小权限，开启备份和慢查询/连接数监控。
- [ ] Redis 配置密码、持久化策略和内存上限；多实例共享同一 Redis 时隔离 key prefix。
- [ ] 邮件、OAuth、支付、上游账号凭证均通过生产密钥管理配置，不写入公开仓库。

## 5. API Key 与上游凭证

- [ ] 新 API Key 只在创建响应中展示一次，列表/详情只展示 prefix/masked 形式。
- [ ] `api_keys.key_hash` 对新 key 非空，`api_keys.key` 不保存明文 secret。
- [ ] legacy 明文 key 首次认证或删除前会回填 hash/prefix。
- [ ] 删除审计对 hash key 记录 `key_hash/key_prefix`，不新增明文 key 审计。
- [ ] 上游账号敏感字段写入时加密，覆盖 `api_key`、`access_token`、`refresh_token`、`cookie`、`private_key`、`service_account`、`client_secret` 等常见别名。
- [ ] 配置变更前确认旧明文凭证可读取；配置变更后确认新写入凭证在数据库中为 `sub2api_enc_v1:` 前缀。

## 6. 日志与错误详情

- [ ] `log_upstream_error_body` 默认关闭；如临时开启，必须配置最大字节数和保留时间。
- [ ] ops 错误日志入库前执行脱敏，敏感 JSON 字段和文本 token/cookie/key 被替换。
- [ ] 用户错误列表/详情只返回白名单字段，不包含 IP、UA、上游 endpoint、账号 ID、上游凭证、原始 prompt/messages。
- [ ] 统一错误解释包含 `error_code`、`explanation`、`suggestion`、`retryable`、`charged`、`http_status`。
- [ ] 管理员排障使用 ops 后台和内部日志，不把 `admin_hint` 暴露给终端用户。

## 7. 支付与账务

- [ ] 支付 provider 配置已加密或置于生产密钥管理，后台导出不显示 secret 明文。
- [ ] 回调 URL 仅通过 HTTPS 暴露，并在支付平台侧配置为正式域名。
- [ ] 每个 provider 的回调验签在预发环境通过。
- [ ] 订单 `out_trade_no` 唯一，重复回调幂等，不重复入账。
- [ ] 用户主动 verify、订单过期任务、pending 对账任务可运行。
- [ ] 退款、取消、失败、超时订单不会改变已入账余额，或具备对应冲正记录。
- [ ] 上线前导出一份订单/余额/用量对账样例，确认客服可追踪“订单 -> 余额变更 -> 用量消耗”。

## 8. HTTPS、代理与网络

- [ ] 反向代理强制 HTTPS，配置 HSTS。
- [ ] `/api`、`/v1`、静态前端资源的 body size、timeout、streaming timeout 与后端一致。
- [ ] WebSocket/streaming 代理禁用缓冲或配置适当 flush。
- [ ] 只信任明确的 `X-Forwarded-For` 来源；未在 `trusted_proxies` 内的请求不使用伪造客户端 IP。
- [ ] 上游 URL allowlist/代理配置符合业务需求；不允许用户任意指定内网地址。

## 9. 管理员、风控与运营

- [ ] 初始化管理员账号已改密，关闭默认密码。
- [ ] 后台操作审计、支付审计、错误日志查询可用。
- [ ] 渠道健康、账号可调度状态、冷却/禁用原因可在后台定位。
- [ ] 用户端能完成：注册/登录 -> 创建 API Key -> 复制 Base URL 与 Key -> 选择模型 -> 查看调用示例 -> 查看用量/错误 -> 充值。
- [ ] 注册、登录、API Key 创建、支付创建、支付取消等关键路径有速率限制或风控开关。
- [ ] 内容安全/隐私过滤策略已按上线地区和客户类型确认。

## 10. 备份恢复与上线后观察

- [ ] 数据库备份保留策略已配置，恢复演练通过。
- [ ] Redis 故障预案已确认，限流/缓存降级行为可接受。
- [ ] 日志保留时间、敏感日志清理策略、异常告警渠道已确认。
- [ ] 上线后 24 小时观察：
  - API 2xx/4xx/5xx 比例。
  - 上游 401/403/429/5xx。
  - 流式中断率。
  - 余额扣费与 usage 总额差异。
  - 支付成功率、回调延迟、pending 订单数。
  - API Key 创建量、首调成功率、新手错误分布。
