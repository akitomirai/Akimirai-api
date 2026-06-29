# API Client Examples

## Base URL

Use the OpenAI-compatible Base URL with exactly one `/v1` suffix:

```text
<BASE_URL>/v1
```

Avoid both missing `/v1` and duplicate `/v1/v1`.

## curl

```bash
curl <BASE_URL>/v1/chat/completions \
  -H "Authorization: Bearer <YOUR_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "<MODEL_NAME>",
    "messages": [
      { "role": "user", "content": "<USER_MESSAGE>" }
    ]
  }'
```

## OpenAI SDK

```ts
import OpenAI from 'openai'

const client = new OpenAI({
  apiKey: '<YOUR_API_KEY>',
  baseURL: '<BASE_URL>/v1'
})

const response = await client.chat.completions.create({
  model: '<MODEL_NAME>',
  messages: [{ role: 'user', content: '<USER_MESSAGE>' }]
})
```

## Codex

```toml
model_provider = "custom"
model = "<MODEL_NAME>"

[model_providers.custom]
name = "Akimirai"
base_url = "<BASE_URL>/v1"
wire_api = "responses"
env_key = "OPENAI_API_KEY"
```

Set `OPENAI_API_KEY` outside the config to your API key.

## Common Model Errors

- `403`: insufficient balance or subscription unavailable.
- `429`: rate limited.
- `502`: upstream failed.
- `503`: no available channel.
- `MODEL_DISABLED`: selected model is disabled.
- `NO_AVAILABLE_CHANNEL`: no user-visible channel can serve the selected model.
- `UPSTREAM_RATE_LIMITED`: upstream provider is rate limited.
- `UPSTREAM_5XX`: upstream provider returned a server error.
- `REQUEST_FORMAT_INVALID`: request body or model parameter is invalid.
- `stream disconnected`: streaming ended before completion.
