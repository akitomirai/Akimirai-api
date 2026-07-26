package service

import (
	"context"
	"log/slog"
	"sort"
	"strings"
	"time"
)

const configuredAIPlatformLabelsCacheTTL = 15 * time.Second

type ConfiguredAIPlatformSource struct {
	Platform            string
	ModelMapping        map[string]string
	CompactModelMapping map[string]string
}

type configuredAIPlatformSourceRepository interface {
	ListSchedulableAIPlatformSources(ctx context.Context) ([]ConfiguredAIPlatformSource, error)
}

type cachedConfiguredAIPlatformLabels struct {
	labels    []string
	expiresAt time.Time
}

// GetConfiguredAIPlatformLabels returns user-facing model-family labels derived
// from currently schedulable accounts and their model mappings.
func (s *SettingService) GetConfiguredAIPlatformLabels(ctx context.Context) []string {
	if s == nil || s.accountRepo == nil {
		return []string{}
	}
	if cached := s.configuredAIPlatformsCache.Load(); cached != nil && time.Now().Before(cached.expiresAt) {
		return append([]string(nil), cached.labels...)
	}
	value, _, _ := s.configuredAIPlatformsSF.Do("configured-ai-platform-labels", func() (any, error) {
		if cached := s.configuredAIPlatformsCache.Load(); cached != nil && time.Now().Before(cached.expiresAt) {
			return cached.labels, nil
		}
		labels := s.loadConfiguredAIPlatformLabels(ctx)
		s.configuredAIPlatformsCache.Store(&cachedConfiguredAIPlatformLabels{
			labels:    append([]string(nil), labels...),
			expiresAt: time.Now().Add(configuredAIPlatformLabelsCacheTTL),
		})
		return labels, nil
	})
	labels, _ := value.([]string)
	return append([]string(nil), labels...)
}

func (s *SettingService) loadConfiguredAIPlatformLabels(ctx context.Context) []string {
	if repo, ok := s.accountRepo.(configuredAIPlatformSourceRepository); ok {
		sources, err := repo.ListSchedulableAIPlatformSources(ctx)
		if err != nil {
			slog.Warn("failed to list schedulable account platform projections", "error", err)
			return []string{}
		}
		return configuredAIPlatformLabelsFromSources(sources)
	}

	accounts, err := s.accountRepo.ListSchedulable(ctx)
	if err != nil {
		slog.Warn("failed to list schedulable accounts for public platform summary", "error", err)
		return []string{}
	}
	return configuredAIPlatformLabelsFromAccounts(accounts)
}

func configuredAIPlatformLabelsFromAccounts(accounts []Account) []string {
	sources := make([]ConfiguredAIPlatformSource, 0, len(accounts))
	for i := range accounts {
		sources = append(sources, ConfiguredAIPlatformSource{
			Platform:            accounts[i].Platform,
			ModelMapping:        accounts[i].GetModelMapping(),
			CompactModelMapping: accounts[i].GetCompactModelMapping(),
		})
	}
	return configuredAIPlatformLabelsFromSources(sources)
}

func configuredAIPlatformLabelsFromSources(sources []ConfiguredAIPlatformSource) []string {
	seen := make(map[string]struct{})
	for i := range sources {
		source := &sources[i]
		if label := modelFamilyLabelForPlatform(source.Platform); label != "" {
			seen[label] = struct{}{}
		}
		for requested, mapped := range source.ModelMapping {
			if label := modelFamilyLabelForModel(requested); label != "" {
				seen[label] = struct{}{}
			}
			if label := modelFamilyLabelForModel(mapped); label != "" {
				seen[label] = struct{}{}
			}
		}
		for requested, mapped := range source.CompactModelMapping {
			if label := modelFamilyLabelForModel(requested); label != "" {
				seen[label] = struct{}{}
			}
			if label := modelFamilyLabelForModel(mapped); label != "" {
				seen[label] = struct{}{}
			}
		}
	}

	if len(seen) == 0 {
		return []string{}
	}
	labels := make([]string, 0, len(seen))
	for label := range seen {
		labels = append(labels, label)
	}
	sort.Slice(labels, func(i, j int) bool {
		left, right := modelFamilyLabelPriority(labels[i]), modelFamilyLabelPriority(labels[j])
		if left != right {
			return left < right
		}
		return labels[i] < labels[j]
	})
	return labels
}

func modelFamilyLabelForPlatform(platform string) string {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case PlatformOpenAI, "openai-compatible":
		return "GPT"
	case PlatformAnthropic, "claude":
		return "Claude"
	case PlatformGemini, "google":
		return "Gemini"
	case PlatformGrok, "xai", "x-ai":
		return "Grok"
	case PlatformAntigravity:
		return "Claude"
	case "zhipu", "glm", "bigmodel":
		return "GLM"
	case "deepseek":
		return "DeepSeek"
	case "qwen", "dashscope", "aliyun":
		return "Qwen"
	case "moonshot", "kimi":
		return "Kimi"
	default:
		return ""
	}
}

func modelFamilyLabelForModel(model string) string {
	normalized := strings.ToLower(strings.TrimSpace(model))
	normalized = strings.TrimPrefix(normalized, "models/")
	if normalized == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(normalized, "gpt-"),
		strings.HasPrefix(normalized, "o1"),
		strings.HasPrefix(normalized, "o3"),
		strings.HasPrefix(normalized, "o4"),
		strings.HasPrefix(normalized, "o5"),
		strings.HasPrefix(normalized, "codex"):
		return "GPT"
	case strings.HasPrefix(normalized, "claude-"),
		strings.Contains(normalized, ".claude-"):
		return "Claude"
	case strings.HasPrefix(normalized, "gemini-"):
		return "Gemini"
	case strings.HasPrefix(normalized, "glm-"),
		strings.HasPrefix(normalized, "chatglm"),
		strings.HasPrefix(normalized, "cogview"),
		strings.HasPrefix(normalized, "cogvideo"):
		return "GLM"
	case strings.HasPrefix(normalized, "deepseek-"):
		return "DeepSeek"
	case strings.HasPrefix(normalized, "grok-"):
		return "Grok"
	case strings.HasPrefix(normalized, "qwen"),
		strings.HasPrefix(normalized, "qwq-"):
		return "Qwen"
	case strings.HasPrefix(normalized, "kimi-"),
		strings.HasPrefix(normalized, "moonshot-"):
		return "Kimi"
	case strings.HasPrefix(normalized, "doubao-"):
		return "Doubao"
	case strings.HasPrefix(normalized, "llama-"),
		strings.HasPrefix(normalized, "codellama-"):
		return "Llama"
	case strings.HasPrefix(normalized, "mistral-"),
		strings.HasPrefix(normalized, "codestral-"),
		strings.HasPrefix(normalized, "pixtral-"),
		strings.HasPrefix(normalized, "open-mistral-"),
		strings.HasPrefix(normalized, "open-mixtral-"):
		return "Mistral"
	case strings.HasPrefix(normalized, "yi-"):
		return "Yi"
	case strings.HasPrefix(normalized, "abab"):
		return "MiniMax"
	default:
		return ""
	}
}

func modelFamilyLabelPriority(label string) int {
	switch label {
	case "GPT":
		return 0
	case "Claude":
		return 1
	case "Gemini":
		return 2
	case "GLM":
		return 3
	case "DeepSeek":
		return 4
	case "Grok":
		return 5
	case "Qwen":
		return 6
	case "Kimi":
		return 7
	case "Doubao":
		return 8
	case "Llama":
		return 9
	case "Mistral":
		return 10
	case "Yi":
		return 11
	case "MiniMax":
		return 12
	default:
		return 100
	}
}
