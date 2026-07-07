package service

import (
	"context"
	"reflect"
	"testing"
)

func TestAPIKeyService_RejectsV10AuthSnapshotWithoutModelsListConfig(t *testing.T) {
	groupID := int64(9)
	svc := &APIKeyService{}

	apiKey, ok, err := svc.applyAuthCacheEntry("k-legacy-models-list", &APIKeyAuthCacheEntry{
		Snapshot: &APIKeyAuthSnapshot{
			Version:  10,
			APIKeyID: 1,
			UserID:   2,
			GroupID:  &groupID,
			Status:   StatusActive,
			User: APIKeyAuthUserSnapshot{
				ID:          2,
				Status:      StatusActive,
				Role:        RoleUser,
				Balance:     10,
				Concurrency: 3,
			},
			Group: &APIKeyAuthGroupSnapshot{
				ID:               groupID,
				Name:             "openai",
				Platform:         PlatformOpenAI,
				Status:           StatusActive,
				SubscriptionType: SubscriptionTypeStandard,
				RateMultiplier:   1,
			},
		},
	})

	if err != nil {
		t.Fatalf("expected stale snapshot to be ignored without error, got %v", err)
	}
	if ok {
		t.Fatalf("expected v10 auth snapshot to be rejected after models_list_config was added")
	}
	if apiKey != nil {
		t.Fatalf("expected no API key from stale snapshot, got %#v", apiKey)
	}
}

func TestAPIKeyService_RejectsV13AuthSnapshotWithoutPrivacyFilterConfig(t *testing.T) {
	groupID := int64(9)
	svc := &APIKeyService{}

	apiKey, ok, err := svc.applyAuthCacheEntry("k-legacy-privacy-filter", &APIKeyAuthCacheEntry{
		Snapshot: &APIKeyAuthSnapshot{
			Version:  13,
			APIKeyID: 1,
			UserID:   2,
			GroupID:  &groupID,
			Status:   StatusActive,
			User: APIKeyAuthUserSnapshot{
				ID:          2,
				Status:      StatusActive,
				Role:        RoleUser,
				Balance:     10,
				Concurrency: 3,
			},
			Group: &APIKeyAuthGroupSnapshot{
				ID:               groupID,
				Name:             "openai",
				Platform:         PlatformOpenAI,
				Status:           StatusActive,
				SubscriptionType: SubscriptionTypeStandard,
				RateMultiplier:   1,
			},
		},
	})

	if err != nil {
		t.Fatalf("expected stale snapshot to be ignored without error, got %v", err)
	}
	if ok {
		t.Fatalf("expected v13 auth snapshot to be rejected after privacy_filter_config was added")
	}
	if apiKey != nil {
		t.Fatalf("expected no API key from stale snapshot, got %#v", apiKey)
	}
}

func TestAPIKeyService_AuthSnapshotPreservesUserPrivacyFilterConfig(t *testing.T) {
	svc := &APIKeyService{}
	apiKey := &APIKey{
		ID:     1,
		UserID: 2,
		Key:    "k-privacy",
		Status: StatusActive,
		User: &User{
			ID:          2,
			Status:      StatusActive,
			Role:        RoleUser,
			Balance:     10,
			Concurrency: 3,
			PrivacyFilterConfig: PrivacyFilterConfig{
				Enabled: true,
				Types:   []string{"token", "email", "not-real", "token"},
			},
		},
	}

	snapshot := svc.snapshotFromAPIKey(context.Background(), apiKey)
	roundTrip := svc.snapshotToAPIKey(apiKey.Key, snapshot)

	want := PrivacyFilterConfig{Enabled: true, Types: []string{"token", "email"}}
	if !reflect.DeepEqual(snapshot.User.PrivacyFilterConfig, want) {
		t.Fatalf("snapshot privacy config mismatch: want %#v got %#v", want, snapshot.User.PrivacyFilterConfig)
	}
	if roundTrip == nil || roundTrip.User == nil {
		t.Fatalf("expected round-trip API key user")
	}
	if !reflect.DeepEqual(roundTrip.User.PrivacyFilterConfig, want) {
		t.Fatalf("round-trip privacy config mismatch: want %#v got %#v", want, roundTrip.User.PrivacyFilterConfig)
	}
}
