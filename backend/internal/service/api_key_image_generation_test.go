package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsImageGenerationAPIKeyEligible(t *testing.T) {
	groupID := int64(7)

	base := func() *APIKey {
		return &APIKey{
			ID:      1,
			UserID:  42,
			GroupID: &groupID,
			Status:  StatusActive,
			User: &User{
				ID:     42,
				Role:   RoleUser,
				Status: StatusActive,
			},
			Group: &Group{
				ID:                   groupID,
				Platform:             PlatformOpenAI,
				Status:               StatusActive,
				SubscriptionType:     SubscriptionTypeStandard,
				AllowImageGeneration: true,
			},
		}
	}

	t.Run("allows active user key for image-enabled openai group", func(t *testing.T) {
		require.True(t, isImageGenerationAPIKeyEligible(base(), 42))
	})

	t.Run("allows active user key for image platform group", func(t *testing.T) {
		key := base()
		key.Group.Platform = PlatformImage
		require.True(t, isImageGenerationAPIKeyEligible(key, 42))
	})

	t.Run("allows active user key for image-enabled grok group", func(t *testing.T) {
		key := base()
		key.Group.Platform = PlatformGrok
		require.True(t, isImageGenerationAPIKeyEligible(key, 42))
	})

	t.Run("rejects another user's key", func(t *testing.T) {
		require.False(t, isImageGenerationAPIKeyEligible(base(), 99))
	})

	t.Run("rejects group without image generation permission", func(t *testing.T) {
		key := base()
		key.Group.AllowImageGeneration = false
		require.False(t, isImageGenerationAPIKeyEligible(key, 42))
	})

	t.Run("rejects exclusive standard group not bound to user", func(t *testing.T) {
		key := base()
		key.Group.IsExclusive = true
		require.False(t, isImageGenerationAPIKeyEligible(key, 42))
	})

	t.Run("allows exclusive standard group bound to user", func(t *testing.T) {
		key := base()
		key.Group.IsExclusive = true
		key.User.AllowedGroups = []int64{groupID}
		require.True(t, isImageGenerationAPIKeyEligible(key, 42))
	})

	t.Run("allows subscription group without allowed-groups binding", func(t *testing.T) {
		key := base()
		key.Group.IsExclusive = true
		key.Group.SubscriptionType = SubscriptionTypeSubscription
		require.True(t, isImageGenerationAPIKeyEligible(key, 42))
	})

	t.Run("rejects unsupported platform", func(t *testing.T) {
		key := base()
		key.Group.Platform = PlatformAnthropic
		require.False(t, isImageGenerationAPIKeyEligible(key, 42))
	})
}
