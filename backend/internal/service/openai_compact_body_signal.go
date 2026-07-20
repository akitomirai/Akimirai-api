package service

import (
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const openAICompactBodySignalKey = "openai_compact_body_signal"

// MarkOpenAICompactBodySignal preserves the original remote-compact v2
// request. The handler may normalize the request for internal scheduling, but
// compatible API-key passthrough upstreams must receive the original
// /responses body-signal protocol.
func MarkOpenAICompactBodySignal(c *gin.Context, body []byte) {
	if c == nil || len(body) == 0 {
		return
	}
	c.Set(openAICompactBodySignalKey, append([]byte(nil), body...))
}

func GetOpenAICompactBodySignal(c *gin.Context) ([]byte, bool) {
	if c == nil {
		return nil, false
	}
	value, ok := c.Get(openAICompactBodySignalKey)
	if !ok {
		return nil, false
	}
	body, ok := value.([]byte)
	if !ok || len(body) == 0 {
		return nil, false
	}
	return append([]byte(nil), body...), true
}

// HasCompactionTriggerInInput detects the Codex remote compact v2 body signal:
// an input item with type "compaction_trigger". When the client sends this
// inside a normal POST /v1/responses (instead of POST /v1/responses/compact),
// the request must still be treated as a compact request — otherwise the
// upstream path, model mapping, and body normalization are all wrong, causing
// Codex to receive a non-compact response and fail with:
//
//	"remote compaction v2 expected exactly one compaction output item, got 0"
//
// The gateway handler promotes such requests for internal scheduling while
// preserving the original payload for API-key passthrough upstreams that
// implement remote compact v2 directly on /responses.
func HasCompactionTriggerInInput(body []byte) bool {
	if len(body) == 0 {
		return false
	}
	input := gjson.GetBytes(body, "input")
	if !input.IsArray() {
		return false
	}
	found := false
	input.ForEach(func(_, item gjson.Result) bool {
		if item.Get("type").String() == "compaction_trigger" {
			found = true
			return false
		}
		return true
	})
	return found
}
