package handler

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AvailableChannelHandler 处理用户侧「可用渠道」查询。
//
// 用户侧接口委托 ChannelService.ListAvailable，并在返回前做三层过滤：
//  1. 行过滤：只保留状态为 Active 且与当前用户可访问分组有交集的渠道；
//  2. 分组过滤：渠道的 Groups 只保留用户可访问的那些；
//  3. 平台过滤：渠道的 SupportedModels 只保留平台在用户可见 Groups 中出现过的模型，
//     防止"渠道同时挂在 antigravity / anthropic 两个平台的分组上，用户只访问
//     antigravity，却看到 anthropic 模型"这类跨平台信息泄漏；
//  4. 字段白名单：仅返回用户需要的字段（省略 BillingModelSource / RestrictModels
//     / 内部 ID / Status 等管理字段）。
type AvailableChannelHandler struct {
	channelService *service.ChannelService
	apiKeyService  *service.APIKeyService
	pricingService *service.PricingService
	settingService *service.SettingService
}

// NewAvailableChannelHandler 创建用户侧可用渠道 handler。
func NewAvailableChannelHandler(
	channelService *service.ChannelService,
	apiKeyService *service.APIKeyService,
	pricingService *service.PricingService,
	settingService *service.SettingService,
) *AvailableChannelHandler {
	return &AvailableChannelHandler{
		channelService: channelService,
		apiKeyService:  apiKeyService,
		pricingService: pricingService,
		settingService: settingService,
	}
}

// featureEnabled 返回 available-channels 开关是否启用。默认关闭（opt-in）。
func (h *AvailableChannelHandler) featureEnabled(c *gin.Context) bool {
	if h.settingService == nil {
		return false
	}
	return h.settingService.GetAvailableChannelsRuntime(c.Request.Context()).Enabled
}

// userAvailableGroup 用户可见的分组概要（白名单字段）。
//
// 前端据此区分专属 vs 公开分组（IsExclusive）、订阅 vs 标准分组（SubscriptionType，
// 订阅视觉加深），并用 RateMultiplier 作为默认倍率；用户专属倍率前端走
// /groups/rates，和 API 密钥页面保持一致。
type userAvailableGroup struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Platform         string  `json:"platform"`
	SubscriptionType string  `json:"subscription_type"`
	RateMultiplier   float64 `json:"rate_multiplier"`
	IsExclusive      bool    `json:"is_exclusive"`
}

// userSupportedModelPricing 用户可见的定价字段白名单。
type userSupportedModelPricing struct {
	BillingMode      string                   `json:"billing_mode"`
	InputPrice       *float64                 `json:"input_price"`
	OutputPrice      *float64                 `json:"output_price"`
	CacheWritePrice  *float64                 `json:"cache_write_price"`
	CacheReadPrice   *float64                 `json:"cache_read_price"`
	ImageOutputPrice *float64                 `json:"image_output_price"`
	PerRequestPrice  *float64                 `json:"per_request_price"`
	Intervals        []userPricingIntervalDTO `json:"intervals"`
}

// userPricingIntervalDTO 定价区间白名单（去掉内部 ID、SortOrder 等前端不渲染的字段）。
type userPricingIntervalDTO struct {
	MinTokens       int      `json:"min_tokens"`
	MaxTokens       *int     `json:"max_tokens"`
	TierLabel       string   `json:"tier_label,omitempty"`
	InputPrice      *float64 `json:"input_price"`
	OutputPrice     *float64 `json:"output_price"`
	CacheWritePrice *float64 `json:"cache_write_price"`
	CacheReadPrice  *float64 `json:"cache_read_price"`
	PerRequestPrice *float64 `json:"per_request_price"`
}

// userSupportedModel 用户可见的支持模型条目。
type userSupportedModel struct {
	Name     string                     `json:"name"`
	Platform string                     `json:"platform"`
	Pricing  *userSupportedModelPricing `json:"pricing"`
}

// userChannelPlatformSection 单渠道内某个平台的子视图：用户可见的分组 + 该平台
// 支持的模型。按 platform 聚合后让前端可以把渠道名作为 row-group 一次渲染，
// 后面的平台行按 sections 顺序铺开。
type userChannelPlatformSection struct {
	Platform        string               `json:"platform"`
	Groups          []userAvailableGroup `json:"groups"`
	SupportedModels []userSupportedModel `json:"supported_models"`
}

// userAvailableChannel 用户可见的渠道条目（白名单字段）。
//
// 每个渠道聚合为一条记录，内嵌 platforms 子数组：每个 section 对应一个平台，
// 包含该平台的 groups 和 supported_models。
type userAvailableChannel struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Platforms   []userChannelPlatformSection `json:"platforms"`
}

// List 列出当前用户可见的「可用渠道」。
// GET /api/v1/channels/available
func (h *AvailableChannelHandler) List(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Feature 未启用时返回空数组（不暴露渠道信息）。检查放在认证之后，
	// 保持与未开关前的 401 行为一致：未登录先 401，登录后再按开关决定。
	if !h.featureEnabled(c) {
		response.Success(c, []userAvailableChannel{})
		return
	}

	userGroups, err := h.apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	allowedGroupIDs := make(map[int64]struct{}, len(userGroups))
	for i := range userGroups {
		allowedGroupIDs[userGroups[i].ID] = struct{}{}
	}

	channels, err := h.channelService.ListAvailable(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]userAvailableChannel, 0, len(channels))
	for _, ch := range channels {
		if ch.Status != service.StatusActive {
			continue
		}
		visibleGroups := filterUserVisibleGroups(ch.Groups, allowedGroupIDs)
		if len(visibleGroups) == 0 {
			continue
		}
		sections := buildPlatformSections(ch, visibleGroups)
		if len(sections) == 0 {
			continue
		}
		out = append(out, userAvailableChannel{
			Name:        ch.Name,
			Description: ch.Description,
			Platforms:   sections,
		})
	}

	response.Success(c, out)
}

// buildPlatformSections 把一个渠道按 visibleGroups 的平台集合拆成有序的 section 列表：
// 每个 section 对应一个平台，只包含该平台的 groups 和 supported_models。
// 输出按 platform 字母序稳定排序，便于前端等效比较与回归测试。
func buildPlatformSections(
	ch service.AvailableChannel,
	visibleGroups []userAvailableGroup,
) []userChannelPlatformSection {
	groupsByPlatform := make(map[string][]userAvailableGroup, 4)
	for _, g := range visibleGroups {
		if g.Platform == "" {
			continue
		}
		groupsByPlatform[g.Platform] = append(groupsByPlatform[g.Platform], g)
	}
	if len(groupsByPlatform) == 0 {
		return nil
	}

	platforms := make([]string, 0, len(groupsByPlatform))
	for p := range groupsByPlatform {
		platforms = append(platforms, p)
	}
	sort.Strings(platforms)

	sections := make([]userChannelPlatformSection, 0, len(platforms))
	for _, platform := range platforms {
		platformSet := map[string]struct{}{platform: {}}
		sections = append(sections, userChannelPlatformSection{
			Platform:        platform,
			Groups:          groupsByPlatform[platform],
			SupportedModels: toUserSupportedModels(ch.SupportedModels, platformSet),
		})
	}
	return sections
}

// filterUserVisibleGroups 仅保留用户可访问的分组。
func filterUserVisibleGroups(
	groups []service.AvailableGroupRef,
	allowed map[int64]struct{},
) []userAvailableGroup {
	visible := make([]userAvailableGroup, 0, len(groups))
	for _, g := range groups {
		if _, ok := allowed[g.ID]; !ok {
			continue
		}
		visible = append(visible, userAvailableGroup{
			ID:               g.ID,
			Name:             g.Name,
			Platform:         g.Platform,
			SubscriptionType: g.SubscriptionType,
			RateMultiplier:   g.RateMultiplier,
			IsExclusive:      g.IsExclusive,
		})
	}
	return visible
}

// toUserSupportedModels 将 service 层支持模型转换为用户 DTO（字段白名单）。
// 仅保留平台在 allowedPlatforms 中的条目，防止跨平台模型信息泄漏。
// allowedPlatforms 为 nil 时不做平台过滤（保留全部，供测试或明确无过滤场景使用）。
func toUserSupportedModels(
	src []service.SupportedModel,
	allowedPlatforms map[string]struct{},
) []userSupportedModel {
	out := make([]userSupportedModel, 0, len(src))
	for i := range src {
		m := src[i]
		if allowedPlatforms != nil {
			if _, ok := allowedPlatforms[m.Platform]; !ok {
				continue
			}
		}
		out = append(out, userSupportedModel{
			Name:     m.Name,
			Platform: m.Platform,
			Pricing:  toUserPricing(m.Pricing),
		})
	}
	return out
}

// toUserPricing 将 service 层定价转换为用户 DTO；入参为 nil 时返回 nil。
func toUserPricing(p *service.ChannelModelPricing) *userSupportedModelPricing {
	if p == nil {
		return nil
	}
	intervals := make([]userPricingIntervalDTO, 0, len(p.Intervals))
	for _, iv := range p.Intervals {
		intervals = append(intervals, userPricingIntervalDTO{
			MinTokens:       iv.MinTokens,
			MaxTokens:       iv.MaxTokens,
			TierLabel:       iv.TierLabel,
			InputPrice:      iv.InputPrice,
			OutputPrice:     iv.OutputPrice,
			CacheWritePrice: iv.CacheWritePrice,
			CacheReadPrice:  iv.CacheReadPrice,
			PerRequestPrice: iv.PerRequestPrice,
		})
	}
	billingMode := string(p.BillingMode)
	if billingMode == "" {
		billingMode = string(service.BillingModeToken)
	}
	return &userSupportedModelPricing{
		BillingMode:      billingMode,
		InputPrice:       p.InputPrice,
		OutputPrice:      p.OutputPrice,
		CacheWritePrice:  p.CacheWritePrice,
		CacheReadPrice:   p.CacheReadPrice,
		ImageOutputPrice: p.ImageOutputPrice,
		PerRequestPrice:  p.PerRequestPrice,
		Intervals:        intervals,
	}
}

type userModelCatalogItem struct {
	ID                    string                     `json:"id"`
	DisplayName           string                     `json:"display_name"`
	ModelID               string                     `json:"model_id"`
	Provider              string                     `json:"provider"`
	Family                *string                    `json:"family"`
	Status                string                     `json:"status"`
	StatusReason          string                     `json:"status_reason"`
	BillingMultiplier     *float64                   `json:"billing_multiplier"`
	BillingDescription    string                     `json:"billing_description"`
	SupportsStreaming     *bool                      `json:"supports_streaming"`
	SupportsVision        *bool                      `json:"supports_vision"`
	SupportsTools         *bool                      `json:"supports_tools"`
	SupportsJSON          *bool                      `json:"supports_json"`
	ContextWindow         *int                       `json:"context_window"`
	RecommendedUse        *string                    `json:"recommended_use"`
	AvailableChannelCount int                        `json:"available_channel_count"`
	QuickStartURL         string                     `json:"quick_start_url"`
	UpdatedAt             *time.Time                 `json:"updated_at"`
	Channels              []string                   `json:"channels"`
	Groups                []userAvailableGroup       `json:"groups"`
	Pricing               *userSupportedModelPricing `json:"pricing"`
}

type catalogAccumulator struct {
	item                  userModelCatalogItem
	channelNames          map[string]struct{}
	groupIDs              map[int64]struct{}
	multipliers           []float64
	activeVisibleCount    int
	nonActiveVisibleCount int
	configuredOnlyCount   int
	updatedAt             *time.Time
}

// Catalog lists the current user's model catalog as an aggregated, user-safe DTO.
// GET /api/v1/user/models/catalog
func (h *AvailableChannelHandler) Catalog(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if !h.featureEnabled(c) {
		response.Success(c, []userModelCatalogItem{})
		return
	}

	userGroups, err := h.apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	userGroupRates, err := h.apiKeyService.GetUserGroupRates(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	groupRefs := toAvailableGroupRefs(userGroups)
	allowedGroupIDs := make(map[int64]struct{}, len(groupRefs))
	for _, group := range groupRefs {
		allowedGroupIDs[group.ID] = struct{}{}
	}

	channels, err := h.channelService.ListAll(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	items := h.buildModelCatalog(channels, groupRefs, allowedGroupIDs, userGroupRates)
	response.Success(c, items)
}

func (h *AvailableChannelHandler) buildModelCatalog(
	channels []service.Channel,
	groupRefs []service.AvailableGroupRef,
	allowedGroupIDs map[int64]struct{},
	userGroupRates map[int64]float64,
) []userModelCatalogItem {
	byGroupID := make(map[int64]service.AvailableGroupRef, len(groupRefs))
	for _, group := range groupRefs {
		byGroupID[group.ID] = group
	}

	accs := make(map[string]*catalogAccumulator)
	for i := range channels {
		ch := channels[i]
		visibleGroups := visibleGroupsForChannel(ch.GroupIDs, byGroupID, allowedGroupIDs)
		if len(visibleGroups) == 0 {
			continue
		}
		visiblePlatforms := platformsForGroups(visibleGroups)
		if len(visiblePlatforms) == 0 {
			continue
		}

		supportedModels := ch.SupportedModels()
		for j := range supportedModels {
			model := supportedModels[j]
			if _, ok := visiblePlatforms[model.Platform]; !ok {
				continue
			}
			acc := h.catalogAccumulatorFor(accs, model.Platform, model.Name)
			if ch.Name != "" {
				acc.channelNames[ch.Name] = struct{}{}
			}
			if acc.item.Pricing == nil && model.Pricing != nil {
				acc.item.Pricing = toUserPricing(model.Pricing)
			}
			acc.addGroups(visibleGroups, model.Platform, userGroupRates)
			acc.updatedAt = latestTime(acc.updatedAt, ch.UpdatedAt)
			switch ch.Status {
			case service.StatusActive:
				acc.activeVisibleCount++
			default:
				acc.nonActiveVisibleCount++
			}
			h.applyCatalogMetadata(&acc.item, model.Name)
		}

		for _, model := range configuredModelsWithoutSupported(ch, supportedModels, visiblePlatforms) {
			acc := h.catalogAccumulatorFor(accs, model.Platform, model.Name)
			acc.configuredOnlyCount++
			acc.addGroups(visibleGroups, model.Platform, userGroupRates)
			acc.updatedAt = latestTime(acc.updatedAt, ch.UpdatedAt)
			h.applyCatalogMetadata(&acc.item, model.Name)
		}
	}

	out := make([]userModelCatalogItem, 0, len(accs))
	for _, acc := range accs {
		acc.finalize()
		out = append(out, acc.item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Provider != out[j].Provider {
			return out[i].Provider < out[j].Provider
		}
		return strings.ToLower(out[i].ModelID) < strings.ToLower(out[j].ModelID)
	})
	return out
}

func (h *AvailableChannelHandler) catalogAccumulatorFor(
	accs map[string]*catalogAccumulator,
	provider string,
	modelID string,
) *catalogAccumulator {
	provider = strings.TrimSpace(provider)
	modelID = strings.TrimSpace(modelID)
	key := strings.ToLower(provider) + "\x00" + strings.ToLower(modelID)
	if acc, ok := accs[key]; ok {
		return acc
	}
	item := userModelCatalogItem{
		ID:            provider + ":" + modelID,
		DisplayName:   modelID,
		ModelID:       modelID,
		Provider:      provider,
		Status:        "unknown",
		StatusReason:  "数据不足，暂无法判断",
		QuickStartURL: "/quick-start?model=" + url.QueryEscape(modelID),
		Channels:      []string{},
		Groups:        []userAvailableGroup{},
	}
	if family := modelFamilyLabel(provider, modelID); family != "" {
		item.Family = &family
	}
	acc := &catalogAccumulator{
		item:         item,
		channelNames: make(map[string]struct{}),
		groupIDs:     make(map[int64]struct{}),
	}
	accs[key] = acc
	return acc
}

func (h *AvailableChannelHandler) applyCatalogMetadata(item *userModelCatalogItem, modelID string) {
	if item == nil || h.pricingService == nil {
		return
	}
	pricing := h.pricingService.GetModelCapabilityMetadata(modelID)
	if pricing == nil {
		return
	}
	if item.ContextWindow == nil && pricing.MaxInputTokens != nil {
		item.ContextWindow = pricing.MaxInputTokens
	}
	if item.SupportsStreaming == nil && pricing.SupportsNativeStreaming != nil {
		item.SupportsStreaming = pricing.SupportsNativeStreaming
	}
	if item.SupportsVision == nil && pricing.SupportsVision != nil {
		item.SupportsVision = pricing.SupportsVision
	}
	if item.SupportsTools == nil {
		item.SupportsTools = boolOr(pricing.SupportsFunctionCalling, pricing.SupportsToolChoice)
	}
	if item.SupportsJSON == nil && pricing.SupportsResponseSchema != nil {
		item.SupportsJSON = pricing.SupportsResponseSchema
	}
}

func (a *catalogAccumulator) addGroups(
	groups []service.AvailableGroupRef,
	platform string,
	userGroupRates map[int64]float64,
) {
	for _, group := range groups {
		if group.Platform != platform {
			continue
		}
		if _, ok := a.groupIDs[group.ID]; ok {
			continue
		}
		a.groupIDs[group.ID] = struct{}{}
		rate := group.RateMultiplier
		if override, ok := userGroupRates[group.ID]; ok {
			rate = override
		}
		if rate >= 0 {
			a.multipliers = append(a.multipliers, rate)
		}
		a.item.Groups = append(a.item.Groups, userAvailableGroup{
			ID:               group.ID,
			Name:             group.Name,
			Platform:         group.Platform,
			SubscriptionType: group.SubscriptionType,
			RateMultiplier:   rate,
			IsExclusive:      group.IsExclusive,
		})
	}
}

func (a *catalogAccumulator) finalize() {
	a.item.AvailableChannelCount = a.activeVisibleCount
	a.item.Channels = sortedStringKeys(a.channelNames)
	a.item.UpdatedAt = a.updatedAt
	a.item.BillingMultiplier = minFloatPtr(a.multipliers)
	a.item.BillingDescription = billingDescription(a.multipliers)
	switch {
	case a.activeVisibleCount > 0:
		a.item.Status = "available"
		a.item.StatusReason = "当前有可用渠道"
	case a.nonActiveVisibleCount > 0:
		a.item.Status = "maintenance"
		a.item.StatusReason = "相关渠道处于维护或停用状态"
	case a.configuredOnlyCount > 0:
		a.item.Status = "unavailable"
		a.item.StatusReason = "当前没有可用渠道"
	default:
		a.item.Status = "unknown"
		a.item.StatusReason = "数据不足，暂无法判断"
	}
}

func toAvailableGroupRefs(groups []service.Group) []service.AvailableGroupRef {
	out := make([]service.AvailableGroupRef, 0, len(groups))
	for _, group := range groups {
		out = append(out, service.AvailableGroupRef{
			ID:               group.ID,
			Name:             group.Name,
			Platform:         group.Platform,
			SubscriptionType: group.SubscriptionType,
			RateMultiplier:   group.RateMultiplier,
			IsExclusive:      group.IsExclusive,
		})
	}
	return out
}

func visibleGroupsForChannel(
	groupIDs []int64,
	byGroupID map[int64]service.AvailableGroupRef,
	allowed map[int64]struct{},
) []service.AvailableGroupRef {
	out := make([]service.AvailableGroupRef, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		if _, ok := allowed[groupID]; !ok {
			continue
		}
		if group, ok := byGroupID[groupID]; ok {
			out = append(out, group)
		}
	}
	return out
}

func platformsForGroups(groups []service.AvailableGroupRef) map[string]struct{} {
	out := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		if group.Platform != "" {
			out[group.Platform] = struct{}{}
		}
	}
	return out
}

func configuredModelsWithoutSupported(
	ch service.Channel,
	supported []service.SupportedModel,
	visiblePlatforms map[string]struct{},
) []service.SupportedModel {
	seen := make(map[string]struct{}, len(supported))
	for _, model := range supported {
		seen[strings.ToLower(model.Platform)+"\x00"+strings.ToLower(model.Name)] = struct{}{}
	}
	out := make([]service.SupportedModel, 0)
	for _, pricing := range ch.ModelPricing {
		if _, ok := visiblePlatforms[pricing.Platform]; !ok {
			continue
		}
		for _, modelName := range pricing.Models {
			if _, wildcard := splitUserCatalogWildcard(modelName); wildcard {
				continue
			}
			key := strings.ToLower(pricing.Platform) + "\x00" + strings.ToLower(modelName)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			pricingCopy := pricing.Clone()
			out = append(out, service.SupportedModel{Name: modelName, Platform: pricing.Platform, Pricing: &pricingCopy})
		}
	}
	return out
}

func splitUserCatalogWildcard(model string) (string, bool) {
	return strings.TrimSuffix(model, "*"), strings.HasSuffix(model, "*")
}

func modelFamilyLabel(provider, modelID string) string {
	normalized := strings.ToLower(strings.TrimSpace(modelID))
	normalized = strings.TrimPrefix(normalized, "models/")
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
	}
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case service.PlatformOpenAI, "openai-compatible":
		return "GPT"
	case service.PlatformAnthropic, "claude", service.PlatformAntigravity:
		return "Claude"
	case service.PlatformGemini, "google":
		return "Gemini"
	case service.PlatformGrok, "xai", "x-ai":
		return "Grok"
	default:
		return ""
	}
}

func boolOr(values ...*bool) *bool {
	hasValue := false
	out := false
	for _, value := range values {
		if value == nil {
			continue
		}
		hasValue = true
		out = out || *value
	}
	if !hasValue {
		return nil
	}
	return &out
}

func sortedStringKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func latestTime(current *time.Time, next time.Time) *time.Time {
	if next.IsZero() {
		return current
	}
	if current == nil || next.After(*current) {
		t := next
		return &t
	}
	return current
}

func minFloatPtr(values []float64) *float64 {
	if len(values) == 0 {
		return nil
	}
	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	return &min
}

func billingDescription(values []float64) string {
	if len(values) == 0 {
		return "暂无可展示的倍率数据"
	}
	min, max := values[0], values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	if min == max {
		return fmt.Sprintf("当前可用路径倍率为 %sx", formatCatalogMultiplier(min))
	}
	return fmt.Sprintf("当前可用路径倍率范围为 %sx - %sx", formatCatalogMultiplier(min), formatCatalogMultiplier(max))
}

func formatCatalogMultiplier(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}
