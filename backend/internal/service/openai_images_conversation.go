package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const conversationAPIURL = "https://chatgpt.com/backend-api/conversation"

// conversationRequest is the payload sent to the ChatGPT web conversation API.
type conversationRequest struct {
	Action                     string                `json:"action"`
	Messages                   []conversationMessage `json:"messages"`
	Model                      string                `json:"model"`
	ParentMessageID            string                `json:"parent_message_id"`
	HistoryAndTrainingDisabled bool                  `json:"history_and_training_disabled"`
	ConversationMode           map[string]string     `json:"conversation_mode,omitempty"`
	TimezoneOffsetMin          int                   `json:"timezone_offset_min"`
	Suggestions                []string              `json:"suggestions"`
	ArkoseToken                any                   `json:"arkose_token,omitempty"`
}

type conversationMessage struct {
	ID      string                     `json:"id"`
	Author  conversationAuthor         `json:"author"`
	Content conversationMessageContent `json:"content"`
}

type conversationAuthor struct {
	Role string `json:"role"`
}

type conversationMessageContent struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

// forwardOpenAIImagesConversation routes image generation through the ChatGPT web conversation API.
// This is used for free-tier accounts that lack Codex API access but have valid session tokens.
func (s *OpenAIGatewayService) forwardOpenAIImagesConversation(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	parsed *OpenAIImagesRequest,
	channelMappedModel string,
) (*OpenAIForwardResult, error) {
	startTime := time.Now()

	prompt := strings.TrimSpace(parsed.Prompt)
	if prompt == "" {
		return nil, fmt.Errorf("image generation requires a prompt")
	}
	imagePrompt := fmt.Sprintf("Generate an image: %s", prompt)

	logger.LegacyPrintf(
		"service.openai_gateway",
		"[OpenAI] Conversation images request account=%s prompt_len=%d",
		account.Name,
		len(prompt),
	)

	upstreamCtx, releaseUpstreamCtx := detachUpstreamContext(ctx)
	defer releaseUpstreamCtx()

	token, _, err := s.GetAccessToken(upstreamCtx, account)
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	msgID := generateRandomUUID()
	parentID := generateRandomUUID()
	convReq := conversationRequest{
		Action: "next",
		Messages: []conversationMessage{{
			ID:     msgID,
			Author: conversationAuthor{Role: "user"},
			Content: conversationMessageContent{
				ContentType: "text",
				Parts:       []string{imagePrompt},
			},
		}},
		Model:                      "auto",
		ParentMessageID:            parentID,
		HistoryAndTrainingDisabled: true,
		TimezoneOffsetMin:          0,
		Suggestions:                []string{},
		ConversationMode:           map[string]string{"kind": "primary_assistant"},
	}

	bodyBytes, err := json.Marshal(convReq)
	if err != nil {
		return nil, fmt.Errorf("marshal conversation request: %w", err)
	}

	upstreamReq, err := http.NewRequestWithContext(upstreamCtx, "POST", conversationAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create upstream request: %w", err)
	}
	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("Accept", "text/event-stream")
	upstreamReq.Header.Set("Cookie", "__Secure-next-auth.session-token="+token)
	upstreamReq.Host = "chatgpt.com"
	upstreamReq.Header.Set("User-Agent", openAIImageBackendUserAgent)
upstreamReq.Header.Set("Origin", "https://chatgpt.com")
	upstreamReq.Header.Set("Referer", "https://chatgpt.com/")

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	upstreamStart := time.Now()
	resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
	SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
	if err != nil {
		safeErr := sanitizeUpstreamErrorMessage(err.Error())
		return nil, fmt.Errorf("conversation upstream request failed: %s", safeErr)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		respBody := s.readUpstreamErrorBody(resp)
		return nil, &UpstreamFailoverError{
			StatusCode:             resp.StatusCode,
			ResponseBody:           respBody,
			RetryableOnSameAccount: false,
		}
	}

	imageResult, err := s.parseConversationImageSSE(c, resp, account, startTime, parsed)
	if err != nil {
		return nil, err
	}

	return imageResult, nil
}

// parseConversationImageSSE reads the conversation API SSE stream and extracts generated images.
func (s *OpenAIGatewayService) parseConversationImageSSE(
	c *gin.Context,
	resp *http.Response,
	account *Account,
	startTime time.Time,
	parsed *OpenAIImagesRequest,
) (*OpenAIForwardResult, error) {
	var (
		conversationID string
		imagePointers  []openAIImagePointerInfo
	)

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		raw := strings.TrimPrefix(line, "data: ")
		if raw == "[DONE]" {
			break
		}

		// Extract conversation_id
		if conversationID == "" {
			cid := strings.TrimSpace(gjson.Get(raw, "conversation_id").String())
			if cid != "" {
				conversationID = cid
			}
		}

		// Look for asset pointers in message parts
		gjson.Get(raw, "message.content.parts").ForEach(func(_, part gjson.Result) bool {
			assetPointer := strings.TrimSpace(part.Get("asset_pointer").String())
			if assetPointer != "" {
				pointer := openAIImagePointerInfo{
					Pointer: assetPointer,
				}
				b64 := strings.TrimSpace(part.Get("b64_json").String())
				if b64 != "" {
					pointer.B64JSON = b64
				}
				imagePointers = append(imagePointers, pointer)
			}
			return true
		})

		// Also look for image_generation content_type parts
		gjson.Get(raw, "message.content.parts").ForEach(func(_, part gjson.Result) bool {
			ct := strings.TrimSpace(part.Get("content_type").String())
			if ct == "image_generation" {
				resultRaw := part.Get("result")
				assetPointer := strings.TrimSpace(resultRaw.Get("asset_pointer").String())
				if assetPointer != "" {
					for _, existing := range imagePointers {
						if existing.Pointer == assetPointer {
							return true
						}
					}
					imagePointers = append(imagePointers, openAIImagePointerInfo{
						Pointer: assetPointer,
					})
				}
			}
			return true
		})

		// Check for top-level error
		errMsg := strings.TrimSpace(gjson.Get(raw, "error").String())
		if errMsg != "" {
			return nil, fmt.Errorf("conversation API error: %s", errMsg)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("conversation SSE scan error: %w", err)
	}

	if len(imagePointers) == 0 {
		return nil, fmt.Errorf("no images generated in conversation response (conversation_id=%s)", conversationID)
	}

	imageCount := 0
	for i, ptr := range imagePointers {
		if i >= parsed.N {
			break
		}
		headers := resp.Header.Clone()
		_, err := resolveOpenAIImageBytes(context.Background(), nil, headers, conversationID, ptr, openAIImageMaxDownloadBytes)
		if err != nil {
			logger.LegacyPrintf("service.openai_gateway",
				"[OpenAI] Conversation image download failed: pointer=%s err=%v", ptr.Pointer, err)
			continue
		}
		imageCount++
	}

	if imageCount == 0 {
		return nil, fmt.Errorf("failed to download any conversation images")
	}

	modelID := firstNonEmptyString(parsed.Model, "gpt-image-2")

	return &OpenAIForwardResult{
		Model:           modelID,
		UpstreamModel:   "auto",
		Stream:          parsed.Stream,
		Duration:        time.Since(startTime),
		ImageCount:      imageCount,
		ImageSize:       parsed.SizeTier,
		ResponseHeaders: resp.Header.Clone(),
	}, nil
}
