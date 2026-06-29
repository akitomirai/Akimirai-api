# P1 Commercial User Flow Report

Date: 2026-06-29

## Scope

This round continues Akimirai-api / sub2api from "safe for trial production" toward a commercial user loop. Because the existing frontend full Vitest suite was failing, this round intentionally stopped at the highest-priority guarded phases:

- Phase 0: protect and verify the current uncommitted P0/P0-2 security hardening work.
- Phase 1: repair the existing full Vitest failures before adding new commercial features.

Phases 2-10 remain pending and are listed below as unfinished work.

## Completed Phases

### Phase 0: P0/P0-2 Protection And Pre-Commit Check

- Confirmed the working tree still contains the P0/P0-2 security hardening changes and is not archived.
- Confirmed these documents exist and are readable:
  - `docs/product-commercialization-audit.md`
  - `docs/production-readiness-checklist.md`
  - `docs/p0-2-production-readiness-report.md`
- Ran `git diff --check`; it passes with only the existing `.gitignore` LF/CRLF warning.
- No P0/P0-2 security implementation was reverted or overwritten.

Current status: there are still uncommitted changes, so Trellis finish-work archive must not be run yet.

### Phase 1: Existing Full Vitest Failures

Fresh full-suite verification confirms the previously recorded frontend failures are fixed in the current working tree:

- Updated platform quota expectations from 4 platforms to the current 5-platform product behavior including `grok`.
- Updated `AccountUsageCell` tests for the current `getUsage(id, source)` API shape.
- Fixed `AccountUsageCell` row-refresh behavior so OpenAI OAuth usage reloads bypass the stale cache when row data changes.
- Fixed chart/admin cost formatting to tolerate missing optional cost fields without `toFixed` crashes.
- Fixed persisted page-size precedence so injected system defaults override stale localStorage.
- Adjusted the pending OAuth account creation test to assert the real payload boundary without requiring unrelated affiliate data.

## Main Changed Files

- `frontend/src/composables/usePersistedPageSize.ts`
- `frontend/src/components/account/AccountUsageCell.vue`
- `frontend/src/components/charts/ModelDistributionChart.vue`
- `frontend/src/components/charts/GroupDistributionChart.vue`
- `frontend/src/views/admin/DashboardView.vue`
- `frontend/src/api/__tests__/settings.authSourceDefaults.spec.ts`
- `frontend/src/components/account/__tests__/AccountUsageCell.spec.ts`
- `frontend/src/components/admin/user/__tests__/UserPlatformQuotaModal.spec.ts`
- `frontend/src/views/admin/__tests__/SettingsView.spec.ts`
- `frontend/src/views/auth/__tests__/EmailVerifyView.spec.ts`

## Database Migrations

No new migration was added in this round.

The existing uncommitted P0 migration remains:

- `backend/migrations/158_add_api_key_hash_columns.sql`

This migration is additive and belongs to the P0 API Key hash storage work.

## User Main Path Status

Not implemented in this round. The user dashboard, guided API key creation path, Quick Start center, model marketplace, usage/billing drilldown, recharge entry, announcement entry, and ticket entry remain pending.

## Admin Main Path Status

Not implemented in this round. The admin operations overview, channel health drilldown, payment reconciliation surface, abnormal user/request review, announcement publishing flow, and ticket handling flow remain pending.

## Security And Privacy Notes

- No real API keys, upstream tokens, cookies, private keys, prompts, or order-sensitive data were added to docs or tests.
- The test changes use synthetic IDs and placeholder values only.
- The cost-formatting fix prevents optional API response omissions from becoming render-time exceptions.
- The OpenAI OAuth usage refresh fix avoids showing stale account usage after row data changes.

## Validation Results

| Command | Result | Notes |
| --- | --- | --- |
| `git status --short` | Completed | Working tree still has uncommitted P0/P0-2 and Phase 1 changes. |
| `git diff --stat` | Completed | Confirms scoped frontend Phase 1 changes plus existing P0/P0-2 files. |
| `git diff --check` | Passed | Only existing `.gitignore` LF/CRLF warning. |
| `pnpm run test:run` | Blocked | `ERR_PNPM_IGNORED_BUILDS` for `esbuild` and `vue-demi`; used `node_modules/.bin` commands instead. |
| `.\node_modules\.bin\vitest.CMD run src/components/admin/user/__tests__/UserPlatformQuotaModal.spec.ts src/components/account/__tests__/AccountUsageCell.spec.ts src/components/charts/__tests__/GroupDistributionChart.spec.ts src/components/charts/__tests__/ModelDistributionChart.spec.ts src/views/admin/__tests__/SettingsView.spec.ts src/views/auth/__tests__/EmailVerifyView.spec.ts src/api/__tests__/settings.authSourceDefaults.spec.ts` | Passed | 7 files, 67 tests. |
| `.\node_modules\.bin\vitest.CMD run` | Passed | 117 files, 739 tests. |
| `.\node_modules\.bin\vue-tsc.CMD --noEmit` | Passed | Type check passed. |
| `.\node_modules\.bin\eslint.CMD . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts` | Passed | First parallel run hit a transient missing `vitest.config.ts.timestamp-*.mjs` file while Vitest was running; single rerun passed. |
| `.\node_modules\.bin\vite.CMD build` | Passed | Existing chunk-size/dynamic-import warnings only. |
| `go test ./...` | Passed | Backend full test suite passed. |
| `go build -o "$env:TEMP\sub2api-server-p1-baseline.exe" ./cmd/server` | Passed | Backend server build passed. |

## Unfinished Items

- Phase 2: user dashboard commercial path enhancement.
- Phase 3: Quick Start / API client configuration center.
- Phase 4: model marketplace and model availability display.
- Phase 5: usage bill and balance ledger enhancement.
- Phase 6: announcement and minimum ticket entry.
- Phase 7: admin operations overview.
- Phase 8: privacy protection settings/info entry.
- Phase 9: user onboarding, admin operations, and API client example docs.
- Phase 10: final full-scope verification after the feature phases.

## Remaining Risks

- P0/P0-2 changes are still uncommitted; archive is not safe until they are committed or intentionally split.
- The current P1 commercial loop is not feature-complete; only the blocked test baseline has been restored.
- Existing build warnings about chunk size and mixed dynamic/static imports remain.
- Existing test stderr warnings from mocked error paths and unresolved test stubs remain, but the full Vitest suite now exits successfully.

## Commit Guidance

Recommendation: commit after reviewing the combined P0/P0-2 and Phase 1 diff.

Suggested commit message:

```text
chore: complete p0 production readiness hardening
```

Archive guidance: do not run Trellis archive while these changes remain uncommitted.

## Next Round Recommendation

Start Phase 2 only now that full Vitest is green. The next smallest useful product slice is the user dashboard commercial path: balance, today's usage, API key status, recent errors, model entry, quick-start entry, and recharge/subscription entry, all backed by real existing data or clearly marked empty states.

---

# P1-2 User Console Commercial Main Path

Date: 2026-06-29

## Round Goal

This round implements only the regular user console commercial main path. It does not include a full model marketplace redesign, billing-system rewrite, announcements/tickets, operations dashboard, or payment-core changes.

## Phase 0 Sync And Baseline

- `git status`: clean at start.
- Latest commit before this round: `ce8089b8 chore: complete p0 hardening and restore test baseline`.
- `main` was ahead of `origin/main` by 1 commit.
- `git push origin main`: passed; `origin/main` is now synchronized to `ce8089b8`.
- `git diff --check`: passed.
- `go test ./...` from `backend/`: passed.
- `pnpm exec vitest run`: blocked by `ERR_PNPM_IGNORED_BUILDS` for `esbuild@0.21.5` and `vue-demi@0.14.10`.
- `.\node_modules\.bin\vitest.cmd run`: passed, 117 files and 739 tests.
- `.\node_modules\.bin\vue-tsc.cmd --noEmit`: passed.
- `.\node_modules\.bin\eslint.cmd . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts`: passed.
- `.\node_modules\.bin\vite.cmd build`: passed with existing dynamic-import/chunk-size warnings.

## Phase 1 Audit Conclusions

### Existing User Capabilities

- Dashboard exists at `/dashboard` and already loads balance, dashboard stats, trends, model usage, recent usage, and platform quotas.
- API Key management exists at `/keys`, including create/edit/delete/status/group/quota/rate-limit operations.
- API Key create flow already preserves plaintext only when `key_visible_once === true` and blocks historical copy attempts.
- Integration examples already exist inside the API Key page/use-key modal, but they are not exposed as a clear user main-path Quick Start entry.
- Usage and request logs exist at `/usage`.
- User-facing redacted error list/detail exist through `/api/v1/usage/errors` and `/api/v1/usage/errors/:id`, gated by `allow_user_view_error_requests`.
- Model/channel entry exists at `/available-channels`; channel status exists at `/monitor`.
- Recharge, plans, and order records exist through `/purchase`, `/orders`, `/subscriptions`, and payment APIs.
- Public settings expose `api_base_url`, `custom_endpoints`, payment/channel flags, site identity, and the error-view flag.

### Current Gaps

- Dashboard first screen does not clearly connect the commercial path: create key, Quick Start, model list, usage/errors, recharge, orders, and subscriptions.
- Dashboard recent errors are not surfaced; users must discover the Usage error tab.
- Dashboard has no independent per-card error state for secondary data like recent errors/key last-used/public settings.
- API Key create success modal does not yet show a prominent one-time warning, key name/prefix, Base URL copy, Quick Start next step, and usage next step in one consolidated success area.
- There is no dedicated lightweight `/quick-start` route.
- The user-safe error response needs `request_id` in the frontend type and likely in the backend redacted DTO to satisfy the dashboard correlation requirement.

### Pages To Modify This Round

- `/dashboard`: add user onboarding/commercial action cards and recent error visibility.
- `/keys`: strengthen the create-success one-time copy path by enhancing the existing use-key modal.
- New `/quick-start`: minimal, reusable, user-facing Quick Start page.
- Sidebar/router: add Quick Start route and entry.

### Interfaces To Reuse

- `GET /api/v1/usage/dashboard/stats`
- `GET /api/v1/usage/dashboard/trend`
- `GET /api/v1/usage/dashboard/models`
- `GET /api/v1/usage`
- `GET /api/v1/usage/errors`
- `GET /api/v1/usage/errors/:id`
- `GET /api/v1/keys`
- `POST /api/v1/keys`
- `GET /api/v1/groups/available`
- `GET /api/v1/channels/available`
- `GET /api/v1/subscriptions`
- `GET /api/v1/subscriptions/summary`
- `GET /api/v1/payment/plans`
- `GET /api/v1/payment/orders/my`
- `GET /api/v1/settings/public`

### Not In This Round

- No destructive migration.
- No payment calculation/core deduction logic change.
- No full model marketplace rebuild.
- No billing ledger redesign.
- No announcements/tickets.
- No admin operations dashboard work.
- No fake operating metrics or hardcoded commercial data.

## P1-2 Implementation Update

### Completed Stages

- Phase 2: enhanced the user dashboard commercial main path.
- Phase 3: strengthened the API Key one-time copy success experience.
- Phase 4: added a lightweight Quick Start entry and reusable example block.
- Phase 5: connected navigation and route entries for the main user path.
- Phase 6: added/updated frontend coverage for dashboard cards, Quick Start examples, one-time API Key display, and sensitive-field non-rendering.
- Phase 7: updated this report with scope, files, tests, security notes, and remaining risks.

### Dashboard Main Path

`/dashboard` now keeps the existing charts and usage widgets, and adds a commercial path section backed by real interfaces:

- Current balance from the authenticated user profile, with an empty state when unavailable.
- Today usage from `GET /usage/dashboard/stats`, with amount, tokens, and request count.
- API Key status from dashboard stats plus `GET /keys`, including count, active count, recent use when present, and create/manage entries.
- New user access card with normalized Base URL, recommended model from real model usage stats when present, Quick Start entry, API Key entry, model entry, usage entry, and Base URL copy.
- Recent error card using `GET /usage/errors` only when the public setting allows user error viewing.
- Recharge/plans/orders/subscriptions entry using the existing payment and subscription routes, with no fabricated plan data.

The recent error card renders only user-safe fields: `error_code`, `explanation`, `suggestion`, `retryable`, `charged`, `request_id`, and time. It does not render prompt text, raw messages, `error_body`, tokens, cookies, private keys, upstream credentials, or full API keys.

### API Key One-Time Copy Experience

The existing `/keys` create flow already opens the use-key modal only when the backend returns a one-time visible key. This round keeps that behavior and adds:

- Clear one-time warning: copy and save immediately; closing the dialog prevents viewing the full key again.
- Key name and prefix/masked display.
- Copy API Key button.
- Copy Base URL button.
- "Next: Quick Start" button.
- "View Usage" button.
- OpenAI/Codex examples normalize Base URL to exactly one `/v1`.

Historical API keys still cannot recover plaintext. The list continues to use prefix/masked values only.

### Quick Start Entry

Added `/quick-start` and sidebar entry. The page provides:

- Base URL, normalized to include exactly one `/v1`.
- Recommended model from real model usage stats when present; otherwise uses `<MODEL_NAME>`.
- API key placeholder `<YOUR_API_KEY>` in normal page context.
- Copyable curl example.
- Copyable OpenAI SDK example.
- Copyable Codex example using `model_provider = "custom"` and `wire_api = "responses"`.
- Common error entry for 403, 429, 502, 503, and stream disconnected.

The shared Quick Start component can temporarily render the just-created API Key only when the key is passed from the one-time creation context. It does not load or recover historical plaintext keys.

### Navigation Entry Points

The user sidebar now includes `/quick-start` after `/keys`. The dashboard links now cover:

- `/keys`
- `/quick-start`
- `/available-channels`
- `/usage`
- `/usage?tab=errors`
- `/purchase` when payment is enabled
- `/orders` when payment is enabled
- `/subscriptions`
- `/redeem`

`/usage?tab=errors` now opens the existing user error tab when the page mounts.

### Backend/API Change

`UserErrorRequest` now includes a user-safe `request_id` JSON field. It is populated from the server request id and falls back to client request id when needed. This is an additive DTO change only; no database migration was added.

### Involved Files

- `backend/internal/service/ops_user_error.go`
- `backend/internal/service/ops_user_error_test.go`
- `frontend/src/utils/quickStart.ts`
- `frontend/src/components/user/quickstart/QuickStartExamples.vue`
- `frontend/src/views/user/QuickStartView.vue`
- `frontend/src/components/user/dashboard/UserDashboardCommercialPath.vue`
- `frontend/src/views/user/DashboardView.vue`
- `frontend/src/components/keys/UseKeyModal.vue`
- `frontend/src/views/user/KeysView.vue`
- `frontend/src/views/user/UsageView.vue`
- `frontend/src/components/user/UserErrorDetailModal.vue`
- `frontend/src/router/index.ts`
- `frontend/src/components/layout/AppSidebar.vue`
- `frontend/src/types/index.ts`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/zh.ts`
- `frontend/src/components/user/quickstart/__tests__/QuickStartExamples.spec.ts`
- `frontend/src/components/user/dashboard/__tests__/UserDashboardCommercialPath.spec.ts`
- `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts`

### Migrations

No new migration.

### Partial Validation Completed During Implementation

| Command | Result | Notes |
| --- | --- | --- |
| `.\node_modules\.bin\vitest.cmd run src/components/user/quickstart/__tests__/QuickStartExamples.spec.ts src/components/user/dashboard/__tests__/UserDashboardCommercialPath.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts` | Passed | 3 files, 11 tests. |
| `.\node_modules\.bin\vue-tsc.cmd --noEmit` | Passed | Type check passed after implementation. |
| `.\node_modules\.bin\eslint.cmd . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts` | Passed | Lint passed after implementation. |

Final full validation is recorded below after the final verification stage.

### Security Check Notes

- No plaintext historical API Key recovery was added.
- No localStorage/sessionStorage/URL persistence was added for plaintext API Keys.
- No prompt text, raw messages, tokens, cookies, private keys, upstream credentials, or full API Keys are rendered in the new dashboard error card.
- Quick Start regular page examples use `<YOUR_API_KEY>` and `<MODEL_NAME>` placeholders.
- Test keys use obvious placeholders such as `sk-test-placeholder`.
- No payment-core deduction logic was modified.
- No destructive or additive database migration was added.

### Remaining Risks Before Final Validation

- Final backend test/build and frontend full Vitest/build still need to be rerun after all edits.
- Existing Vite build warnings about chunks/dynamic imports may still appear.
- `/available-channels` remains the existing model/channel entry, not a full model marketplace redesign.
- Payment entry depends on `payment_enabled`; disabled environments show existing non-payment alternatives instead of fabricated plans.

### Current Commit Guidance

If final validation passes, commit message:

```text
feat: improve user onboarding dashboard flow
```

Trellis archive is allowed only after the commit is created and `git status` is clean.

## P1-2 Final Validation

### Push Status

- `git push origin main` was completed at the start of this round.
- `main` is up to date with `origin/main` before this round's local commit.

### Final Test Commands

| Command | Result | Notes |
| --- | --- | --- |
| `git diff --check` | Passed | Only LF/CRLF warning for this markdown report. |
| `go test ./internal/service -run TestToUserErrorRequest` | Passed | Targeted user error DTO regression coverage. |
| `go test ./...` from `backend/` | Passed | Backend full test suite passed. |
| `go build -o "$env:TEMP\sub2api-server-p1-2.exe" ./cmd/server` from `backend/` | Passed | Backend server build passed. |
| `.\node_modules\.bin\vitest.cmd run src/components/user/quickstart/__tests__/QuickStartExamples.spec.ts src/components/user/dashboard/__tests__/UserDashboardCommercialPath.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts` | Passed | 3 files, 11 tests. |
| `.\node_modules\.bin\vitest.cmd run src/views/user/__tests__/UsageView.spec.ts` | Passed | Rechecked `/usage?tab=errors` route-query compatibility. |
| `.\node_modules\.bin\vitest.cmd run` | Passed | 119 files, 746 tests. Existing stderr warnings are from pre-existing mocked error paths/router stubs. |
| `.\node_modules\.bin\vue-tsc.cmd --noEmit` | Passed | Type check passed. |
| `.\node_modules\.bin\eslint.cmd . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts` | Passed | Lint passed. |
| `.\node_modules\.bin\vite.cmd build` | Passed | Existing dynamic-import/chunk-size warnings only. |

`pnpm exec vitest run` remains blocked in this environment by ignored builds for `esbuild@0.21.5` and `vue-demi@0.14.10`; equivalent `node_modules/.bin` commands were used and recorded.

### Final Security Checks

- `git check-ignore -v docs/p1-commercial-user-flow-report.md frontend/src/views/user/QuickStartView.vue frontend/src/components/user/quickstart/QuickStartExamples.vue`: no matches, so these files are not ignored.
- `git diff --name-only -- backend/migrations frontend backend docs | rg "migrations/.*\.sql$"`: no matches, so no migration was added or modified.
- Diff scan for API-key/private-key/token/cookie/prompt/message leak patterns found only documentation safety wording. No real key, token, cookie, private key, prompt text, upstream credential, or full historical API Key was added.
- Hardcoded-data scan found only explicit test placeholders such as `sk-test-placeholder` and existing unrelated legacy comments/usages outside this P1-2 change. No new fake commercial metrics were added.

### Remaining Risks

- `/available-channels` remains the existing model/channel entry; the full model marketplace redesign is intentionally deferred.
- Payment plans/orders are only linked when `payment_enabled` is true; disabled deployments show safe alternatives instead of fabricated data.
- Existing Vite warnings about mixed dynamic/static imports and large chunks remain.
- Existing Vitest stderr warnings from deliberate mocked error paths remain, but the suite exits successfully.

### Final Recommendation

- Commit: yes.
- Commit message: `feat: improve user onboarding dashboard flow`.
- Trellis finish-work/archive: allowed only after the commit is created and `git status` is clean.

### Next Round

Recommended next slice: P1-3 model marketplace/accountability path. Keep it narrow: enrich `/available-channels` with real model availability, pricing/rate explanation, and model-to-Quick-Start entry points without changing routing or billing core.
