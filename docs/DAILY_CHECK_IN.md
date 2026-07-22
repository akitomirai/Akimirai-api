# Daily Check-In

The daily check-in feature grants a small balance reward exactly once per user and service day. The durable ledger is `daily_check_ins`; neither browser state nor cache state is authoritative.

## Service Day

- Time zone: `Asia/Shanghai` (`UTC+08:00`).
- Reset time: `02:00` local time.
- A request at `01:59:59` belongs to the previous calendar date; a request at `02:00:00` starts the new service date.

## Reward Policy

The service uses `crypto/rand` to select one of 100 equally likely buckets:

| Bucket | Reward | Probability |
| --- | ---: | ---: |
| `0-49` | `1` | `50%` |
| `50-79` | `2` | `30%` |
| `80-99` | `3` | `20%` |

Only `users.balance` is increased. The reward does not change `total_recharged` or any payment accounting field.

## Exact-Once Claim

`POST /api/v1/user/check-in` performs the ledger insert and balance update in one database transaction. The unique key on `(user_id, service_date)` prevents duplicate rewards under retries or concurrent requests. A replay returns the original persisted reward and does not invalidate balance caches again.

`GET /api/v1/user/check-in` returns the current service-day status without creating a reward.

## Administrator Records

`GET /api/v1/admin/daily-check-ins` is admin-only and reads the immutable ledger joined with user identity.

Supported query parameters:

- `page`, `page_size`: page size defaults to 20 and is capped at 200.
- `q`: user ID, email, or username keyword.
- `service_date`: exact `YYYY-MM-DD` service date.
- `all=true`: explicitly query all service dates. It cannot be combined with `service_date`.

When neither date parameter is present, the backend applies the current Shanghai 02:00 service date. Results are ordered by `checked_in_at DESC, id DESC`.

## Schema and Indexes

- `192_daily_check_ins.sql` creates the immutable exact-once ledger and its reward/balance constraints.
- `193_daily_check_in_admin_indexes.sql` adds current-day and all-history newest-first indexes for administrator pagination.

Reward distribution is enforced in the service and tested across all 100 buckets. The database constraint remains the final guard that accepted rewards are one of `1`, `2`, or `3`.
