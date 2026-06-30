# P1-3 / P1-3B Model Availability Report

## Goal

P1-3 improves the user path for choosing a model, understanding availability, understanding multiplier hints, and opening Quick Start with the selected model. P1-3B adds the backend aggregation endpoint that makes this path reusable across the dashboard, Quick Start, and the user model entry.

This round does not change payment core logic, channel scheduling, billing reconstruction, tickets, announcements, operations dashboards, or database schema.

## Current Model Data Audit

- User model entry remains `/available-channels`; it now renders a model catalog panel above the legacy available-channel detail table.
- Legacy user endpoint `GET /api/v1/channels/available` is kept for compatibility and still returns active user-visible channels, groups, supported models, and pricing details.
- New user endpoint `GET /api/v1/user/models/catalog` returns a user-safe aggregated model catalog DTO.
- The catalog aggregation uses real data from user-visible groups, user group rates, all channels, channel model pricing/mapping, and LiteLLM pricing metadata where available.
- Reliable fields: `model_id`, `provider`, user-visible groups, effective multiplier, pricing presence, available channel count, status, status reason, Quick Start URL, and updated time.
- Optional fields: `family`, `context_window`, `supports_streaming`, `supports_vision`, `supports_tools`, `supports_json`; these are `null` when no reliable source exists.
- Not reliable yet: real success rate, error rate, congestion, latency, and recommended use. These are not fabricated.

## Status Rules

- `available`: at least one user-visible active channel can serve the model.
- `maintenance`: the user can see a related model configuration, but only non-active channels currently expose it.
- `unavailable`: reserved for a real configured model trace with no currently available path.
- `unknown`: data is insufficient.

The endpoint supports all four status values, but real responses only show states supported by current data. It does not create fake rows just to display every state.

## User Model Entry

- `/available-channels` now loads `/user/models/catalog` for the model catalog and `/channels/available` for the legacy channel table.
- Each row shows the API `model` value, provider, status, status reason, multiplier description, channel/group source, capability hints when available, copy model, and Quick Start action.
- Loading, empty, and error states are present.
- The multiplier explanation remains: "倍率表示该模型按平台基础计费单位的倍数消耗，最终扣费以用量账单为准。"

## Quick Start

- Quick Start now loads `/user/models/catalog`.
- `/quick-start?model=<MODEL_NAME>` selects the model only when it is currently available.
- Missing or unavailable query models show a friendly fallback hint and use a recommended available model when possible.
- Examples continue to use `<YOUR_API_KEY>` in normal context.
- Base URL normalization still keeps exactly one `/v1`; examples avoid both missing `/v1` and duplicate `/v1/v1`.

## Dashboard Entry

- The user dashboard model card now loads the catalog endpoint.
- It shows real available model count when loaded, selects a recommended available model, can copy that model, and links to Quick Start with the model query.
- Recent `MODEL_DISABLED` and `NO_AVAILABLE_CHANNEL` errors link users back to `/available-channels`.

## Admin Hint

- The existing admin channel page already includes a lightweight hint explaining that user-facing model status depends on enabled channels, visible groups, and configured pricing.
- No admin model system or backend route rewrite was added.

## Files

- `backend/internal/handler/available_channel_handler.go`
- `backend/internal/handler/available_channel_handler_test.go`
- `backend/internal/server/routes/user.go`
- `backend/internal/service/channel_service.go`
- `backend/internal/service/pricing_service.go`
- `backend/cmd/server/wire_gen.go`
- `frontend/src/api/channels.ts`
- `frontend/src/utils/modelCatalog.ts`
- `frontend/src/components/channels/ModelCatalogPanel.vue`
- `frontend/src/views/user/AvailableChannelsView.vue`
- `frontend/src/views/user/QuickStartView.vue`
- `frontend/src/views/user/DashboardView.vue`
- `frontend/src/utils/__tests__/modelCatalog.spec.ts`
- `frontend/src/components/channels/__tests__/ModelCatalogPanel.spec.ts`
- `frontend/src/views/user/__tests__/QuickStartView.spec.ts`
- `docs/p1-model-availability-report.md`
- `docs/user-onboarding-flow.md`
- `docs/api-client-examples.md`

## Migrations

无。

## Test Results

- `git diff --check`: passed, with line-ending warnings for tracked docs only.
- `go test ./...`: passed.
- `go build -o $env:TEMP\sub2api-p1-3b-final.exe ./cmd/server`: passed.
- `node_modules\.bin\vue-tsc.cmd --noEmit --pretty false`: passed.
- `node_modules\.bin\eslint.cmd . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts`: passed.
- Key Vitest:
  `node_modules\.bin\vitest.cmd run src/utils/__tests__/modelCatalog.spec.ts src/components/channels/__tests__/ModelCatalogPanel.spec.ts src/views/user/__tests__/QuickStartView.spec.ts`: passed, 3 files / 10 tests.
- Full Vitest: `node_modules\.bin\vitest.cmd run`: passed, 122 files / 757 tests.
- `node_modules\.bin\vite.cmd build`: passed, with existing dynamic-import and chunk-size warnings.
- Docs ignore check: the three docs files match `.gitignore:131 docs/*`, but they are already tracked by Git.
- Migration check: no migration files were added or changed.

## Security Checks

- The catalog response uses a field whitelist and does not expose upstream API keys, upstream tokens, cookies, private keys, prompt originals, internal account names, channel ids, account ids, scheduler internals, or admin-only notes.
- Frontend catalog mapping copies only typed DTO fields and ignores unknown extra payload fields.
- Tests use obvious placeholders such as `sk-test-placeholder`.
- Documentation uses `<YOUR_API_KEY>`, `<BASE_URL>`, and `<MODEL_NAME>`.
- No plaintext historical API key is restored or persisted.
- Diff scan found only field names, placeholders, and documented negative-test strings; no real API key, token, cookie, private key, or prompt original was introduced.

## Remaining Risks

- `unavailable` and `unknown` rows require real configured-but-unavailable or insufficient-data sources; the system does not fabricate those rows.
- Real model health metrics, latency, congestion, and success rate are still not exposed to users.
- Capability metadata depends on LiteLLM pricing metadata being present for a model.
- Existing Vite dynamic-import and chunk-size warnings may remain.
- In this environment, `pnpm exec` can be blocked by ignored build scripts; use `node_modules/.bin` equivalents when needed.

## Recommendation

- Suggested commit message: `feat: improve model availability onboarding`
- If final validation passes and the working tree is clean after commit, Trellis finish-work / archive is allowed.

## Next Round

- Add authoritative model health metrics only if safe user-facing monitor data is available.
- Add richer capability metadata only after deciding the authoritative source.
- Consider a fuller model marketplace after the backend catalog contract has settled.
