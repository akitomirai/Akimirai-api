# Admin Operations Guide

## User Model Catalog Status

The user model catalog is exposed through `GET /api/v1/user/models/catalog`. It is derived from real channel, group, pricing, and user visibility data.

### Make A Model Available

- Keep at least one related channel active.
- Attach the channel to a group visible to the user.
- Configure model pricing or model mapping so the model can be resolved as a concrete `model_id`.
- Keep the group multiplier accurate; user-facing multiplier text is based on real group rate data.

### Why A Model Shows Maintenance

A model can show `maintenance` when the user can see a related model configuration, but the related channel path is not active. Check channel status first, then check whether the model pricing/mapping still belongs to a visible group.

### Why A Model Shows Unavailable

A model can show `unavailable` only when there is real configured model evidence but no currently available user-visible path. Do not add placeholder channels or fake status rows to force this state.

### Why A Model Shows Unknown

`unknown` means the system does not have enough reliable data to claim availability. Add authoritative channel/model configuration instead of guessing from a model name.

### Troubleshooting A User Report

1. Confirm the user can access at least one group for the model provider.
2. Confirm a channel for that group is active.
3. Confirm the requested `model` appears in pricing or mapping as a concrete model id.
4. Confirm any user-specific group rate is expected.
5. Ask the user to retry with a model copied from `/available-channels`.

Never copy upstream API keys, access tokens, refresh tokens, cookies, private keys, service account JSON, prompt originals, or internal account names into tickets, docs, logs, or screenshots.
