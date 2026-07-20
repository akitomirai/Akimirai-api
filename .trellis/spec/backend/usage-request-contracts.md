# Usage Request Contracts

## Scenario: User Usage Request Detail

### 1. Scope / Trigger
- Trigger: user-facing usage/request APIs that return or filter request details, billing facts, API key display data, or error correlation fields.
- Applies to `GET /api/v1/usage`, `GET /api/v1/usage/requests`, `GET /api/v1/usage/:id`, `GET /api/v1/usage/errors`, and `GET /api/v1/usage/errors/:id`.
- These APIs are user-scoped. They must not expose admin-only request investigation data.

### 2. Signatures
- Usage list:
  - `GET /api/v1/usage?page=&page_size=&api_key_id=&request_id=&model=&request_type=&stream=&billing_type=&charged=&start_date=&end_date=&start_time=&end_time=&timezone=&sort_by=&sort_order=`
- Usage detail:
  - `GET /api/v1/usage/:id`
- Unified request log:
  - `GET /api/v1/usage/requests?page=&page_size=&kind=all|consumption|error&api_key_id=&group_id=&model=&request_id=&start_date=&end_date=&start_time=&end_time=&period=&timezone=&sort_by=created_at|duration_ms&sort_order=asc|desc`
- Error list:
  - `GET /api/v1/usage/errors?page=&page_size=&start_date=&end_date=&start_time=&end_time=&period=&timezone=&model=&api_key_id=&status_code=&category=&request_id=`
- Error detail:
  - `GET /api/v1/usage/errors/:id`

### 3. Contracts
- Authenticated user scope is mandatory:
  - `UsageHandler.List` must set `UsageLogFilters.UserID` from the auth subject.
  - `api_key_id` usage filters must verify API key ownership before querying.
  - user error listing must force `OpsErrorLogFilter.UserID` and keep deleted-key owner matching enabled for user-owned deleted keys.
- `request_id` filters are exact matches, not fuzzy search.
- Custom user ranges use local `YYYY-MM-DDTHH:mm` values plus `timezone` and become exact half-open `[start_time, end_time)` instants. Rolling presets continue to use `period`; explicit minute values take precedence when present.
- The unified request log has one server-side pagination owner after union and deduplication:
  - `usage_logs` owns completed consumption and billing facts.
  - `ops_error_logs` owns failed-request status and redacted error classification.
  - exact non-empty `(request_id, api_key_id)` matches correlate the two ledgers; timestamp/model similarity is never used as proof.
  - a correlated error row takes precedence and may be enriched with the matching usage row's key/group/model/reasoning/latency/token/cost fields.
  - error rows are deduplicated by effective non-empty request ID plus API key, keeping the newest row; blank request IDs remain distinct by error ID.
  - the final combined set is counted, sorted, and paginated globally. The client must not merge independently paginated lists.
- User error visibility is fail-closed on the unified endpoint:
  - `kind=error` is forbidden while the setting is disabled.
  - `kind=all` falls back to consumption-only while the setting is disabled.
- Unified rows support the modern seven-column user projection: time, key/group, model/reasoning effort, Token breakdown, cache-hit rate inputs, safe costs, and latency. Row details remain user-safe and are opened from the row rather than a dedicated detail column.
- Token fields are nullable `input_tokens`, `output_tokens`, `cache_creation_tokens`, `cache_read_tokens`, and `total_tokens`. Cost fields are nullable `input_cost`, `output_cost`, `cache_creation_cost`, `cache_read_cost`, `total_cost`, and `actual_cost`.
- A completed consumption row may populate those fields from `usage_logs`. An exactly correlated error may inherit them from its matched usage row. A standalone error keeps them null; null means unavailable and must not be converted to fabricated zero measurements.
- Unified rows must not expose raw prompts, request/response bodies, account credentials, full API keys, client IPs, User-Agent, admin hints, proxy/route diagnostics, or raw upstream error messages.
- `charged` usage filter is backed by real `usage_logs.actual_cost`:
  - `charged=true` means `actual_cost > 0`.
  - `charged=false` means `actual_cost <= 0`.
- Usage row additive fields:
  - `api_key_name`, `api_key_prefix`, `api_key_masked`
  - `total_tokens`
  - `cost_amount`, `cost_unit`
  - `charged`, `billing_type_label`, `status`
  - `http_status`, `error_code`, `explanation`, `suggestion`, `retryable`
  - `detail_url`
- Usage rows must use `null` or `unknown` where the `usage_logs` source cannot prove an error/explanation value. Do not substitute `0` or `false` to imply certainty.
- `detail_url` may link to `/usage/errors?request_id=<escaped request_id>`; it is a correlation hint, not proof that an error record exists.

### 4. Validation & Error Matrix
- Missing auth subject -> unauthorized response.
- `api_key_id` not parseable -> bad request.
- `api_key_id` not owned by the current user -> forbidden.
- `request_type` outside `unknown|sync|stream|ws_v2|cyber` -> bad request.
- `stream` not parseable as bool -> bad request.
- `charged` not parseable as bool -> bad request.
- `start_date` / `end_date` not `YYYY-MM-DD` -> bad request.
- `start_time` and `end_time` not supplied together -> bad request.
- `start_time` / `end_time` not `YYYY-MM-DDTHH:mm`, or end is not after start -> bad request.
- user error view disabled -> forbidden.
- unified log `kind` outside `all|consumption|error` -> bad request.
- unified log `sort_by` outside `created_at|duration_ms` -> bad request.
- unified log `sort_order` outside `asc|desc` -> bad request.
- user error detail id invalid -> bad request.
- user error detail not owned by the current user -> not found.

### 5. Good/Base/Bad Cases
- Good: user `42` requests `/usage?request_id=req-1&charged=true`; the repository receives `UserID=42`, `RequestID=req-1`, and `Charged=true`, and returns only exact matching real usage rows.
- Good: `start_time=2026-07-12T22:00&end_time=2026-07-13T22:15&timezone=Asia/Shanghai` reaches every user usage projection as the same exact half-open range.
- Good: an error row with an exact `(request_id, api_key_id)` usage match inherits the matched row's Token/cost metrics; the same error without that proof returns null metrics.
- Base: a zero-cost usage row returns `status="unknown"`, `charged=false`, and null error descriptor fields unless a separate `/usage/errors` row is queried.
- Bad: a DTO returns a full API key, raw prompt/messages/content, upstream credentials, admin hint, client IP, or account credential fields.

### 6. Tests Required
- Handler tests must assert user-scope filters are always set from the auth subject.
- Handler tests must cover `request_type`, legacy `stream`, `request_id`, and invalid `charged` parsing.
- Handler tests must cover valid minute ranges, missing pairs, invalid syntax, reversed ranges, timezone conversion, and rolling-period compatibility.
- Repository tests must prove `request_id` filters are exact matches.
- Unified-log repository tests must prove ledger scoping occurs before correlation, exact-key correlation, newest-error deduplication, blank-ID independence, error precedence/enrichment, errors-disabled behavior, and global count/pagination after the union.
- Unified-log repository/service tests must assert exact correlated Token/cost inheritance, standalone-error null metrics, and stable select/scan order.
- Unified-log handler/service tests must prove auth scope cannot be overridden, API key ownership is checked, error visibility is fail-closed, and invalid kind/sort values are rejected.
- DTO tests must prove:
  - token totals and cost fields are derived from real usage data;
  - masked API key output uses the existing masking path;
  - unsupported usage-row error fields remain null;
  - sensitive raw fields are absent from marshaled user DTO JSON.
- Service tests must prove user error listing preserves user-safe `request_id` filters while still forcing user ownership.

### 7. Wrong vs Correct
#### Wrong
```go
row.ErrorCode = "OK"
row.HTTPStatus = ptr(200)
row.Retryable = ptr(false)
```

This fabricates an error/status contract from a `usage_logs` row that does not store those facts.

#### Correct
```go
row.Status = usageLogStatus(row.Charged)
row.HTTPStatus = nil
row.ErrorCode = nil
row.Retryable = nil
```

The usage route returns only proven usage/billing facts and leaves error explanation fields to `/usage/errors`.

For unified request logs, the same rule applies to metrics: populate pointer fields only from a proven `usage_logs` row and leave standalone error metrics nil.

## Scenario: Administrator Request Diagnostics

### 1. Scope / Trigger
- Trigger: changing request timing collection, usage-log persistence, administrator usage filters, route snapshots, retry/failover reporting, or the admin usage diagnostics UI.
- Applies to completed usage rows in `usage_logs`, `GET /api/v1/admin/usage`, `GET /api/v1/admin/usage/stats`, and `GET /api/v1/admin/usage/:id/diagnostics`.
- Failed-request bodies and classifications remain owned by `ops_error_logs`; diagnostics correlate to that owner by exact `request_id`.

### 2. Signatures
- Admin list/stat filters:
  - `route_kind=direct|proxy`
  - `proxy_id=<positive integer>`
  - `retry_only=true|false`
  - `min_request_total_ms=<non-negative integer>`
  - `min_request_first_token_ms=<non-negative integer>`
  - `min_upstream_first_byte_ms=<non-negative integer>`
- Admin detail: `GET /api/v1/admin/usage/:id/diagnostics`.
- Diagnostic storage fields:
  - `request_started_at`, `request_total_ms`, `request_body_read_ms`, `request_body_bytes`
  - `upstream_request_written_ms`, `upstream_first_byte_ms`, `request_first_token_ms`
  - `route_kind`, `proxy_id_snapshot`, `proxy_name_snapshot`, `proxy_protocol_snapshot`, `route_fingerprint`
  - `final_upstream_status`, `retry_count`, `account_switch_count`, `attempt_timeline`
- Route indexes: `(proxy_id_snapshot, created_at DESC)` and `(route_kind, created_at DESC)`, both partial on non-null values.
- Attempt event cap: `service.RequestDiagnosticsAttemptLimit == 32`.

### 3. Contracts
- `usage_logs` remains the only completed-request usage and billing ledger. Diagnostic columns are additive evidence on that row, not a second request store.
- `created_at` remains the completion/log-write timestamp. `request_started_at` is the only field that proves when the request began.
- Nullable phase timings mean unavailable, not zero. The UI must render unavailable phases explicitly and must not infer DNS, TCP, TLS, upload, or queue sub-phases that were not measured.
- For supported OpenAI HTTP paths (`/v1/responses`, `/v1/messages`, and `/v1/chat/completions`), the request-scoped collector owns the diagnostic snapshot, but not every numeric field has the same origin:
  - `request_body_read_ms` is the duration of reading the client request body itself, not an offset from request start.
  - `routing_latency_ms` is the existing handler routing-stage duration and may include concurrency waits, billing rechecks, failed attempts, retry backoff, and account switches.
  - `upstream_request_written_ms`, `upstream_first_byte_ms`, `request_first_token_ms`, and `request_total_ms` are request-start elapsed measurements for the final diagnostic snapshot.
- `upstream_request_written_ms` comes from `httptrace.WroteRequest`. It proves that the local HTTP transport completed its request write flow without reporting an error; it does not prove that the upstream received or processed the request.
- `upstream_first_byte_ms` comes from `httptrace.GotFirstResponseByte`; it proves that the first byte of the HTTP response headers became available, not that full headers, response body data, or a model token arrived.
- The HTTP transport may write a request and wait for a response concurrently, so the displayed conceptual order is not guaranteed to be strict wall-clock event order. The admin UI must describe each field's own basis and must not derive unsupported stage differences.
- `request_first_token_ms` re-anchors each protocol adapter's first-token result to the request collector. Detection points differ by adapter, so UI copy must not promise an identical displayable-token boundary across every upstream protocol.
- Historical `first_token_ms` and `duration_ms` are upstream-forwarding-relative legacy measurements. The diagnostics drawer must not silently relabel them as request-scoped `request_first_token_ms` or `request_total_ms`; missing request-scoped fields render unavailable.
- Route snapshots are immutable for the completed row. A later account proxy edit must not rewrite historical attribution.
- A proxy snapshot may contain the proxy ID, display name, protocol, and a sanitized fingerprint. It must not contain userinfo, credentials, query parameters, paths, authorization headers, full proxy URLs, prompts, request bodies, or response bodies.
- `attempt_timeline` is serialized only when `retry_count > 0` or `account_switch_count > 0`, is capped at 32 sanitized events, and reuses the upstream error sanitizer.
- Normal single-attempt successes keep `attempt_timeline` null/empty.
- Diagnostic collection, sanitization, or serialization failures are fail-open for forwarding and billing.
- Upstream request gzip is opt-in and limited to upstreams verified to accept `Content-Encoding: gzip`. Production canaries use exact hostnames; a verified host does not imply support for arbitrary subdomains.
- Payload-shape investigation remains scalar-only. Do not persist request bodies, image URLs, MIME types, content hashes, tool output, or `encrypted_content`; the gateway must not delete or rewrite opaque request content merely to reduce transfer size.
- The admin detail DTO may expose diagnostic fields. User usage DTOs and exports must not expose route fingerprints, proxy snapshots, attempt timelines, or admin-only phase timings.
- The existing admin usage view owns four tabs: usage rows, errors, user ranking, and request records. Request records reuse the same usage list rows and render the route/timing/status summary as configurable inline columns; they do not require a per-row drawer action or create another route, sidebar entry, API, or persistence ledger. The ordinary usage tab may retain its diagnostics drawer for the full attempt timeline.
- Request-record column visibility is client-local UI state. User and request-start time remain visible; optional columns are persisted under `usage-diagnostics-hidden-columns`, and the column-settings control is placed immediately before refresh.
- Diagnostic route/proxy/latency query parameters remain supported by the admin API but are not rendered in the main usage filter bar. The visible filter bar stays focused on user, API key, model, group, and account.
- Diagnostic rows follow the existing usage retention policy; no separate high-volume event table is introduced.
- Handwritten usage insert/select contracts must remain atomic: `usageLogInsertArgTypes`, every SQL column/value list, `prepareUsageLogInsert().args`, and `usageLogSelectColumns`/scan order are updated together. Batch capacities and test argument counts derive from `len(usageLogInsertArgTypes)` rather than a historical literal.

### 4. Validation & Error Matrix
- Detail `id` missing, non-numeric, or non-positive -> bad request.
- Detail row not found -> existing not-found response from `UsageService.GetByID`.
- `route_kind` outside `direct|proxy` -> bad request.
- `proxy_id` non-numeric or non-positive -> bad request.
- `retry_only` not parseable as a boolean -> bad request.
- Any latency minimum non-numeric or negative -> bad request.
- Historical row without diagnostics -> successful response with nullable timings and zero retry/switch counters; legacy `first_token_ms` and `duration_ms` do not fill request-scoped diagnostic fields.
- Unsupported timing phase -> null/unavailable, never a fabricated duration.
- Failed request with an operations record -> correlate by exact `request_id`; do not copy the operations error body into `usage_logs`.

### 5. Good/Base/Bad Cases
- Good: a request starts before a proxy change, completes afterward, and its usage row still shows the original proxy snapshot plus independent first-byte and first-token timings.
- Good: a retry succeeds on another account; the usage row records final route counters and at most 32 sanitized attempt events, while the errors tab remains the owner of failure details.
- Base: a historical or unsupported-path row opens in the drawer with unavailable request-scoped values and no fabricated fallback from legacy forwarding-relative timings.
- Bad: infer that a row started at `created_at`, display null as `0 ms`, or attribute historical rows from the account's current proxy assignment.
- Bad: persist a full proxy URL, proxy password, API key, authorization header, prompt, messages, request body, or response body.

### 6. Tests Required
- Collector tests must cover direct/proxy snapshots, credential removal, monotonic timing order, same-account retry, account switch, 32-event truncation, concurrency, and nil-collector no-op behavior.
- Repository tests must cover all insert paths, arg/type count equality, JSON round trip, nullable historical rows, filter SQL, and scan order.
- Handler tests must cover admin authorization through route registration, invalid detail IDs, not found, filter parsing, and exact request-ID correlation.
- DTO tests must marshal both admin and user responses and assert diagnostic fields exist only in the admin contract.
- Frontend tests must cover loading/error/empty states, null timing rendering, the six business checkpoint labels and descriptions, refusal to fall back to legacy `first_token_ms`/`duration_ms`, retry/switch indicators, inline request-record columns, persisted column toggles, drawer detail loading from the ordinary usage tab, and the errors-tab request-ID action.
- Full verification includes `go test ./...`, `go vet ./...`, frontend full tests, type-check, lint, production build, migration coverage, secret scan, and `git diff --check`.

### 7. Wrong vs Correct
#### Wrong
```go
args := make([]any, 0, len(rows)*58)
mock.ExpectQuery("INSERT INTO usage_logs").WithArgs(/* 58 handwritten values */)
```

This duplicates an old column count and silently drifts when diagnostic columns are added.

#### Correct
```go
args := make([]any, 0, len(rows)*len(usageLogInsertArgTypes))
expectedLog := *log
expected := prepareUsageLogInsert(&expectedLog)
mock.ExpectQuery("INSERT INTO usage_logs").WithArgs(anySliceToDriverValues(expected.args)...)
```

The production insert owner defines the ordered argument contract, while separate assertions verify the specific behavior under test.
