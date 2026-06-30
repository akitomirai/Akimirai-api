# User Onboarding Flow

## Choose A Model

Users can open `/available-channels` to see the model catalog returned by the authenticated backend catalog endpoint.

Each model row shows:

- model id to use in the `model` parameter
- platform
- user-visible availability and status reason
- effective multiplier from visible groups
- visible channel/group source
- copy model action
- Quick Start action

No fake model, fake health, fake channel count, or fake availability is shown. If the catalog endpoint has no visible model data, the page shows an empty state.

## Model Status

- `available`: this model currently has at least one user-visible active channel.
- `maintenance`: related configuration is visible, but the related path is not active.
- `unavailable`: real configuration exists, but there is no currently available user-visible path.
- `unknown`: the system does not have enough reliable data to claim availability.

When a model is not available, copy another `model` value from `/available-channels` and retry the request. Do not guess model availability from a static list.

## Quick Start From A Model

The model catalog links to:

```text
/quick-start?model=<MODEL_NAME>
```

Quick Start validates the query model against `/api/v1/user/models/catalog`:

- If available, examples use that model.
- If missing or unavailable, examples fall back to a recommended available model and show a warning.
- If no model is available, examples use `<MODEL_NAME>`.

## API Key Safety

Normal Quick Start pages use:

```text
<YOUR_API_KEY>
```

Only the API Key creation success flow may temporarily pass a just-created plaintext key into examples. Historical keys are never restored in plaintext.

## Switching Models After Errors

If a recent dashboard error has `MODEL_DISABLED` or `NO_AVAILABLE_CHANNEL`, the dashboard links users back to `/available-channels` so they can copy a currently available model and retry.
