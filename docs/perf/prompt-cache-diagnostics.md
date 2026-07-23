# OpenAI 提示缓存诊断

## 目的

使用记录中的缓存命中率只能说明本次输入 Token 有多少来自上游缓存，不能单独解释为什么某次请求突然变冷。本诊断链路为管理员记录以下脱敏指纹，用于区分会话切换、稳定前缀变化和工具或系统提示变化：

| 字段 | 含义 |
| --- | --- |
| `prompt_cache_key_hash` | 上游实际生效的缓存身份 SHA-256；从不保存原始 key |
| `prompt_cache_key_source` | `client_header`、`client_body`、`compat_derived` 或 `none` |
| `prompt_cache_prefix_hash` | 可复用稳定前缀的指纹 |
| `prompt_cache_tools_hash` | `tools` / `functions` 定义及顺序的指纹 |
| `prompt_cache_system_hash` | `instructions`、system 和 developer 内容的指纹 |

这些字段只通过管理员使用记录 DTO 返回。普通用户的使用记录响应不包含这些字段，数据库也不保存原始提示词或原始缓存 key。

## 缓存身份规则

1. 请求 body 已有非空 `prompt_cache_key` 时，保留并以它作为实际缓存身份。
2. body 没有缓存 key，但请求显式携带 `session_id` 或 `conversation_id` 时，OAuth Responses 路径把该会话身份写入 `prompt_cache_key`。
3. Chat Completions 兼容路径缺少显式身份时，沿用现有的稳定派生 key，并标记为 `compat_derived`。
4. 不从用户 ID、模型名或全局常量生成缓存 key。缺少可靠会话身份时保持 `none`，避免并行会话互相污染或挤占缓存。
5. `/responses/compact` 不注入该字段，保持 compact 上游协议不变。

客户端应为每个会话生成稳定且唯一的 `prompt_cache_key`，并让 model、推理强度、tools 顺序、instructions/system 和历史前缀保持稳定。新消息只追加到末尾；动态时间、临时状态等放在稳定前缀之后。上下文压缩后的第一条请求通常是冷请求，后续同一摘要前缀应恢复命中。

## 排障方法

按同一用户和模型查看相邻请求，并优先比较 `prompt_cache_key_hash`：

- key 指纹变化：客户端切换了会话身份，或没有复用同一会话 key。
- key 不变但 prefix 指纹变化：稳定前缀被重排、插入或改写。
- tools 或 system 指纹变化：工具定义、顺序、系统提示或 instructions 发生变化。
- 所有指纹稳定但单次命中率低，下一次恢复：通常是上游冷缓存、缓存过期或压缩后的首个请求。
- 同一组指纹连续多次低命中：再检查请求路由、上游账号和缓存策略；告警应以这一条件为准，而不是对单次冷请求报警。

缓存命中率公式和账号调度器不参与本功能改动。

## 验证

- 后端测试覆盖 append-only 历史稳定性、tools/system 变化、body/header 优先级、兼容层派生和空身份。
- 仓储测试覆盖单条、批量和 best-effort 插入参数契约。
- DTO 测试保证诊断字段仅管理员可见。
- 前端测试覆盖诊断抽屉的来源和截断指纹展示。
