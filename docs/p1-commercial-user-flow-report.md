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
