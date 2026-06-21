# POST /responses 延迟分析报告

**日期**: 2026-06-21
**背景**: Codex Desktop 的 POST /responses 请求较慢，记录显示首 token 6.32s / 总耗时 7.47s。
**范围**: 只做分析和低风险观测增强，不做协议重构。

---

## 1. 完整链路梳理

### 1.1 路由注册 (routes)

文件: `backend/internal/server/routes/gateway.go`

```
POST /v1/responses          → h.OpenAIGateway.Responses(c)
POST /v1/responses/*subpath → h.OpenAIGateway.Responses(c)
POST /responses             → h.OpenAIGateway.Responses(c)
POST /backend-api/codex/responses → h.OpenAIGateway.Responses(c)
GET  /v1/responses          → h.OpenAIGateway.ResponsesWebSocket(c)
```

中间件链：`bodyLimit → clientRequestID → opsErrorLogger → endpointNorm → apiKeyAuth → requireGroupAnthropic`

关键点：当 API Key 所属 Group 的 Platform 是 `openai` 时，调用 `h.OpenAIGateway.Responses(c)`，否则 fallback 到 `h.Gateway.Responses(c)`。

### 1.2 Handler 层

文件: `backend/internal/handler/openai_gateway_handler.go:137`

`Responses()` 方法流程：
1. **setOpenAIClientTransportHTTP(c)** — HTTP 入站固定标记为 `OpenAIClientTransportHTTP`
2. 提取 API Key / User / 请求体
3. 校验 model、stream、previous_response_id
4. 内容审核检查
5. **并发槽位获取**: user slot → account slot（含 wait queue）
6. **账号选择**: `SelectAccountWithSchedulerForCapability()` 带 failover 循环
7. **转发**: `gatewayService.Forward()` — 此时设置 `OpsUpstreamLatencyMsKey`
8. **用量记录**: `gatewayService.RecordUsage()` — worker pool 异步提交
9. **设置各阶段延迟**: `OpsAuthLatencyMsKey`, `OpsRoutingLatencyMsKey`, `OpsUpstreamLatencyMsKey`, `OpsResponseLatencyMsKey`, `OpsTimeToFirstTokenMsKey`

### 1.3 Service 层

文件: `backend/internal/service/openai_gateway_service.go:2371`

`Forward()` 方法：
- 构建上游 HTTP 请求（URL、Headers、Body）
- 发送请求到上游 OpenAI API
- 流式响应：SSE 逐 token 转发，记录 `FirstTokenMs`
- 非流式响应：JSON body 一次性读取
- 返回 `OpenAIForwardResult{Duration, FirstTokenMs, OpenAIWSMode, ...}`

文件: `backend/internal/service/openai_gateway_service.go:5893`

`RecordUsage()` 方法：
- 计算实际 input tokens（减去 cache_read）
- 调用 pricing service 计算费用
- 构建 `UsageLog` 结构体写入数据库

### 1.4 用量记录

文件: `backend/ent/schema/usage_log.go`

usage_logs 表字段（性能相关）:
- `duration_ms` (int, nullable)
- `first_token_ms` (int, nullable)
- `request_type` (通过 stream + openai_ws_mode 推断)
- `openai_ws_mode` (bool)

文件: `backend/internal/handler/dto/mappers.go:575`

`usageLogFromServiceUser()` 将 service.UsageLog 转为前端 DTO UsageLog JSON。

### 1.5 前端展示

文件: `frontend/src/views/user/UsageView.vue`

使用记录表格列: model, reasoning_effort, endpoint, type, billing_mode, tokens, cost, first_token, duration, time, user_agent

---

## 2. 后端已有但前端未展示的性能字段

| 字段 | 数据源 | 存储位置 | DTO | 前端类型 | 前端展示 |
|------|--------|----------|-----|----------|----------|
| `first_token_ms` | ForwardResult | usage_logs | ✅ | ✅ | ✅ 已展示 |
| `duration_ms` | ForwardResult | usage_logs | ✅ | ✅ | ✅ 已展示 |
| `openai_ws_mode` | ForwardResult | usage_logs | ✅ | ✅ | ❌ 未展示 |
| `client_transport` | gin context | ❌ 未存储 | ❌ | ❌ | ❌ |
| `auth_latency_ms` | gin context | ❌ 未存储 (仅 error_log) | ❌ | ❌ | ❌ |
| `routing_latency_ms` | gin context | ❌ 未存储 (仅 error_log) | ❌ | ❌ | ❌ |
| `upstream_latency_ms` | gin context | ❌ 未存储 (仅 error_log) | ❌ | ❌ | ❌ |
| `response_latency_ms` | gin context | ❌ 未存储 (仅 error_log) | ❌ | ❌ | ❌ |
| 单行缓存命中率 | computable | N/A (计算) | ❌ | ❌ | ❌ 仅聚合 |
| `account_id` | usage_logs | ✅ | ✅ (ID only) | ✅ | ❌ 用户不可见 |
| `inbound_endpoint` | usage_logs | ✅ | ✅ | ✅ | ✅ 已展示 |
| `upstream_endpoint` | usage_logs | ✅ | ✅ | ✅ | ❌ 用户表不展示 |
| `reasoning_effort` | usage_logs | ✅ | ✅ | ✅ | ✅ 已展示 |

---

## 3. 第一阶段实现 (低风险观测增强) — 已完成 2026-06-21

### 3.1 零 DB 变更项（已实现）

1. **前端展示 `openai_ws_mode`**: ✅ 在 type 列 badge 中增加 WS 标识
2. **前端单行缓存命中率**: ✅ 在 token 详情 tooltip 中计算并展示 `cache_read/(input+cache_read+cache_write)×100%`
3. **前端展示 `upstream_endpoint`**: ✅ 在 endpoint 列显示 `入站 → 上游` 链路

### 3.2 低风险 DB 变更项（已实现）

4. **usage_logs 增加 `client_transport`**: ✅ varchar(10), 记录 HTTP/WS 入站协议
5. **usage_logs 增加延迟分解字段**: ✅ `auth_latency_ms`, `routing_latency_ms`, `upstream_latency_ms`, `response_latency_ms`

### 3.3 变更文件清单

**后端**:
- `ent/schema/usage_log.go` — 新增 5 个字段
- `ent/` — 自动生成的 Ent 代码
- `internal/service/usage_log.go` — UsageLog struct 新增字段
- `internal/service/openai_gateway_service.go` — OpenAIRecordUsageInput + RecordUsage 填充
- `internal/handler/openai_gateway_handler.go` — Responses/Messages/WS 三条路径提取 gin context 延迟值
- `internal/handler/dto/types.go` — DTO UsageLog 新增 JSON 字段
- `internal/handler/dto/mappers.go` — DTO 映射新增字段

**前端**:
- `types/index.ts` — UsageLog 类型新增字段
- `views/user/UsageView.vue` — token tooltip 加缓存命中率, type badge 加 WS, endpoint 加 upstream, cost tooltip 加延迟分解
- `components/admin/usage/UsageTable.vue` — 同上（admin 视角）
- `i18n/locales/en.ts` — 新增 latencyBreakdown/clientTransport/authLatency/routingLatency/upstreamLatency/responseLatency
- `i18n/locales/zh.ts` — 同上（中文）

---

## 4. 架构要点

### 4.1 HTTP → WS 协议隔离

文件: `backend/internal/service/openai_client_transport.go`

- `OpenAIClientTransportHTTP` = "http"
- `OpenAIClientTransportWS` = "ws"
- handler 在入站时通过 `SetOpenAIClientTransport(c, transport)` 标记
- service 根据 transport 决定上游连接方式
- **HTTP 入站不会自动升级为 WS 上游**（明确的设计决策）

### 4.2 延迟计算

```
auth_start → auth_end     = auth_latency_ms
auth_end   → routing_end  = routing_latency_ms
routing_end → upstream_end = upstream_latency_ms
forwardDuration - upstream_latency = response_latency_ms
```

上游首 token (first_token_ms) 由 OpenAI API 报告，独立于上述分解。

---

## 5. 第二阶段实现 (低风险延迟优化) — 已完成 2026-06-21

### A. 请求体 gzip 压缩 ✅

**改动**: `backend/internal/service/openai_gateway_service.go`
- 新增 `maybeGzipCompressBody()` 方法：body > 1KB 且配置启用时做 gzip 压缩
- 修改 `buildUpstreamRequest` (L4200) 和 `buildUpstreamRequestOpenAIPassthrough` (L3411)
- 添加 `Content-Encoding: gzip` 头标识压缩
- 配置开关: `gateway.upstream_request_gzip` (默认关闭，显式启用)

**效果**: 大请求体 (如 81.7K tokens → ~100KB JSON) 可压缩至 ~25KB，网络传输减少 75%。压缩库 `compress/gzip` 为标准库，零额外依赖。

**测试**: `TestMaybeGzipCompressBody_Disabled`, `TestMaybeGzipCompressBody_SmallBody`, `TestMaybeGzipCompressBody_LargeBody`

### B. Compact 会话 ID 生成修复 ✅

**改动**: `backend/internal/service/openai_gateway_service.go`
- 修改 `resolveOpenAICompactSessionID`: 新增 `body []byte` 参数
- 无显式 session_id / conversation_id / prompt_cache_key 时，回退到 `deriveOpenAIContentSessionSeed(body)` 的内容摘要，而非 `uuid.NewString()`
- 同一 conversation 的连续请求产生相同 session_id，上游可复用缓存

**根因**: compact 路径对无会话标识的请求每次生成随机 UUID 作为 session_id，上游将每次请求视为全新会话，缓存命中率 ~2.3%。

**效果**: 对于未显式传递 session_id 的客户端，相同 content 的请求共享上游缓存键。缓存命中率预期从 2.3% 提升至合理水平（取决于 conversation 重用频率）。

**测试**: `TestResolveOpenAICompactSessionID_ContentFallback`, `TestResolveOpenAICompactSessionID_UsesHeaderFirst`, `TestResolveOpenAICompactSessionID_StableSameBody`

### 变更文件清单

| 文件 | 改动 |
|------|------|
| `internal/config/config.go` | 新增 `UpstreamRequestGzip` 配置字段 |
| `internal/service/openai_gateway_service.go` | gzip 压缩 + compact session ID 修复 |
| `internal/service/openai_content_session_seed_test.go` | 6 个新测试函数 |

---

## 6. 下一轮可调优项 (Phase 3)

1. **HTTP 连接池预热**: HTTP 传输池无预热机制，新 (账号, 代理) 组合首次请求需完整 TCP+TLS 握手。WS 连接池已有预热（`openai_ws_pool.go`），可参考实现 HTTP 版。
2. **非 Compact OAuth 路径缓存头修复**: `buildUpstreamRequest` 在 OAuth 账号下删除客户端 `session_id`/`conversation_id` 头后仅从 body `prompt_cache_key` 重新设置。纯 header 传递缓存的客户端会丢失缓存标识。
3. **首 token 优化**: 收集 `first_token_ms` 与模型/推理强度/输入长度的关联数据，量化各因素贡献。
4. **HTTP → WS 桥接预研**: 评估 WS upstream 的延迟收益。需处理 `openai_client_transport.go` 的 HTTP→WS 协议隔离约束。
5. **gzip 压缩效果验证**: 在生产环境启用 `gateway.upstream_request_gzip` 后，监控 `upstream_latency_ms` 的变化，验证压缩收益。
