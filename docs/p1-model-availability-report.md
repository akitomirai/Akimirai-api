# P1-3 Model Availability Onboarding Report

## Goal

P1-3 improves the user path for choosing a model, understanding user-visible availability, understanding multiplier/pricing hints, and opening Quick Start with the selected model.

This round does not implement a full model marketplace, payment-core changes, a billing-system rewrite, announcements/tickets, an operations dashboard rewrite, or destructive database changes.

## Phase 0 Baseline

- `git status --short --branch`: clean at start, `main...origin/main`.
- `git log --oneline -5`: latest commit was `163438d3 feat: improve user onboarding dashboard flow`.
- Origin sync: `git rev-list --left-right --count origin/main...HEAD` returned `0 0`; no pre-development push was needed.
- `git diff --check`: passed.
- Backend `go test ./...`: passed.
- Backend build: `go build -o $env:TEMP\sub2api-p1-3-baseline.exe ./cmd/server`: passed.
- `pnpm exec vitest run` / `pnpm exec vue-tsc --noEmit`: blocked by `ERR_PNPM_IGNORED_BUILDS` for `esbuild@0.21.5` and `vue-demi@0.14.10`.
- Frontend fallback commands:
  - `.\node_modules\.bin\vitest.cmd run`: passed, 119 files / 746 tests.
  - `.\node_modules\.bin\vue-tsc.cmd --noEmit`: passed in baseline.
  - `.\node_modules\.bin\eslint.cmd . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts`: passed after running sequentially. A concurrent baseline attempt raced with a transient Vitest timestamp file.
  - `.\node_modules\.bin\vite.cmd build`: passed with existing dynamic-import and chunk-size warnings.

## Audit

### Existing Frontend

- User model entry existed at `/available-channels`.
- `AvailableChannelsView.vue` loaded `/channels/available` through `frontend/src/api/channels.ts`.
- The existing table showed channels, visible groups, supported models, and pricing chips.
- `QuickStartView.vue` existed, but selected its model from dashboard usage stats rather than current availability.
- `QuickStartExamples.vue` already normalized Base URL to exactly one `/v1` and used `<YOUR_API_KEY>` in normal context.
- `DashboardView.vue` already rendered the P1-2 commercial path and linked to `/available-channels` and `/quick-start`.

### Existing Backend

- User route: `GET /api/v1/channels/available`.
- Handler: `backend/internal/handler/available_channel_handler.go`.
- The handler authenticates the user, honors the available-channels feature flag, intersects channels with user-visible groups, filters by platform, and returns a field whitelist.
- User-safe returned fields include channel name/description, platform, user-visible groups, group rate multiplier, supported model name/platform, and pricing fields.
- The handler does not expose channel ids, account ids, upstream account names, upstream API keys, cookies, tokens, private keys, raw prompts, or scheduler internals.
- Channel service has active/inactive status and channel model pricing. The user API only returns currently visible active channels, so disabled model or no-channel states are not directly enumerable for users.

## Reliable Fields

- Model id: `supported_models[].name`.
- Platform/family: `supported_models[].platform` and section `platform`.
- User-visible source: channel `name`.
- User-visible groups: group `name`, `subscription_type`, `rate_multiplier`, `is_exclusive`.
- Pricing presence and token/per-request/image pricing fields when configured.

## Unreliable Or Unavailable Fields

- Real model health, congestion, recent success/error rate, and latency are not available in `/channels/available`.
- Full disabled model catalog is not available to ordinary users.
- Streaming support is not explicitly declared by the current user-visible model payload.
- Bound internal channel count is intentionally not exposed.

## Status Rules

- `available`: model is present in at least one user-visible available channel section.
- `maintenance`: only if real model/channel disabled data is available. The current user API does not expose this for catalog rows.
- `temporarily_unavailable`: only if real configured model data exists but no available channel is visible. The current user API does not expose such rows.
- `unknown`: insufficient evidence; do not claim availability.

This round therefore shows reliable `available` rows from real data and conservative empty/unknown messaging instead of fake status.

## Implemented

- Added `frontend/src/utils/modelCatalog.ts` as the single frontend derivation point for user model catalog rows, status rules, multiplier range formatting, Quick Start selection, and model-availability error detection.
- Added `frontend/src/components/channels/ModelCatalogPanel.vue` for the lightweight model catalog on `/available-channels`.
- Enhanced `AvailableChannelsView.vue` to show the model catalog, loading, empty, error, retry, copy-model, and Quick Start actions while keeping the existing channel detail table.
- Changed `QuickStartView.vue` to load `/channels/available`, read `?model=<MODEL_NAME>`, validate it against real available models, and fall back safely.
- Enhanced `QuickStartExamples.vue` with copy current model and model/channel common error codes.
- Enhanced `DashboardView.vue` to load real available model summary.
- Enhanced `UserDashboardCommercialPath.vue` to show real available model count when loaded, copy recommended model, open Quick Start with selected model, and route model/channel errors to `/available-channels`.
- Added a lightweight admin hint in `ChannelsView.vue` explaining that user-facing model status depends on enabled channels, visible groups, and pricing rules.

## Files

- `frontend/src/utils/modelCatalog.ts`
- `frontend/src/components/channels/ModelCatalogPanel.vue`
- `frontend/src/views/user/AvailableChannelsView.vue`
- `frontend/src/views/user/QuickStartView.vue`
- `frontend/src/components/user/quickstart/QuickStartExamples.vue`
- `frontend/src/views/user/DashboardView.vue`
- `frontend/src/components/user/dashboard/UserDashboardCommercialPath.vue`
- `frontend/src/views/admin/ChannelsView.vue`
- `frontend/src/utils/__tests__/modelCatalog.spec.ts`
- `frontend/src/components/channels/__tests__/ModelCatalogPanel.spec.ts`
- `frontend/src/views/user/__tests__/QuickStartView.spec.ts`
- `frontend/src/components/user/quickstart/__tests__/QuickStartExamples.spec.ts`
- `frontend/src/components/user/dashboard/__tests__/UserDashboardCommercialPath.spec.ts`
- `docs/user-onboarding-flow.md`
- `docs/api-client-examples.md`

## Migrations

None.

## Current Test Results

- P1-3 key Vitest command passed:
  `.\node_modules\.bin\vitest.cmd run src/utils/__tests__/modelCatalog.spec.ts src/components/channels/__tests__/ModelCatalogPanel.spec.ts src/components/user/quickstart/__tests__/QuickStartExamples.spec.ts src/views/user/__tests__/QuickStartView.spec.ts src/components/user/dashboard/__tests__/UserDashboardCommercialPath.spec.ts`
- Result: 5 files / 17 tests passed.
- Full frontend Vitest passed: 122 files / 757 tests.
- Frontend typecheck passed: `.\node_modules\.bin\vue-tsc.cmd --noEmit`.
- Frontend lint passed: `.\node_modules\.bin\eslint.cmd . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts`.
- Frontend build passed: `.\node_modules\.bin\vite.cmd build`.
- Backend tests passed: `go test ./...`.
- Backend build passed: `go build -o $env:TEMP\sub2api-p1-3-final.exe ./cmd/server`.
- `git diff --check` passed.
- `git check-ignore -v` confirmed the three docs files are ignored by `.gitignore:131 docs/*` and must be force-added if committed.

## Security Checks

- No real API key, upstream token, cookie, private key, or raw prompt is written to docs or tests.
- Tests use placeholders such as `sk-test-placeholder`.
- Quick Start example bodies use `<USER_MESSAGE>` rather than a real prompt body.
- Model catalog derivation copies only typed user-visible fields and ignores unknown extra payload fields.
- Historical API key plaintext is not restored or persisted.
- Quick Start regular context uses `<YOUR_API_KEY>`.

## Remaining Risks

- The user API cannot list disabled or configured-but-unavailable models; this is documented and not faked.
- Streaming support remains unknown because it is not present in the user-visible payload.
- Existing Vite dynamic-import and chunk-size warnings remain.
- `pnpm exec` remains blocked by ignored build scripts in this environment; node_modules `.cmd` equivalents are used.

## Recommendation

- Suggested commit message: `feat: improve model availability onboarding`
- The work is ready to commit after this report update.
- `origin/main` was already synchronized at task start (`0 0`), so no pre-development push was needed.
- Final push status: succeeded after the work commit was created and pushed.
- If the working tree is clean after commit and push, Trellis finish-work / archive is allowed.

## Next Round

- Add a safe backend model availability aggregate only if product wants disabled/unavailable model rows exposed to users.
- Consider user-visible model capability metadata only after deciding which fields are authoritative.
- Consider model health metrics only if real monitor data can be safely exposed without leaking internal accounts or credentials.
