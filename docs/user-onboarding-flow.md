# User Onboarding Flow

## Choose A Model

Users can open `/available-channels` to see the model catalog derived from real available channel data.

Each model row shows:

- model id to use in the `model` parameter
- platform
- user-visible availability
- effective multiplier range from visible groups
- visible channel/group source
- copy model action
- Quick Start action

No fake model, fake health, fake channel count, or fake availability is shown. If the platform has no visible model data, the page shows an empty state.

## Quick Start From A Model

The model catalog links to:

```text
/quick-start?model=<MODEL_NAME>
```

Quick Start validates the query model against the real available model list:

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
