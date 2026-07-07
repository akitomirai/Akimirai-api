package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type privacyFilterSettingRepoStub struct {
	values        map[string]string
	getValueCalls int
}

func (s *privacyFilterSettingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *privacyFilterSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	s.getValueCalls++
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *privacyFilterSettingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *privacyFilterSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *privacyFilterSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *privacyFilterSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *privacyFilterSettingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingService_GetPrivacyFilterConfig(t *testing.T) {
	t.Run("missing setting uses default and caches", func(t *testing.T) {
		repo := &privacyFilterSettingRepoStub{values: map[string]string{}}
		svc := NewSettingService(repo, nil)

		require.Equal(t, DefaultPrivacyFilterConfig(), svc.GetPrivacyFilterConfig(context.Background()))
		require.Equal(t, DefaultPrivacyFilterConfig(), svc.GetPrivacyFilterConfig(context.Background()))
		require.Equal(t, 1, repo.getValueCalls)
	})

	t.Run("configured setting is normalized", func(t *testing.T) {
		repo := &privacyFilterSettingRepoStub{
			values: map[string]string{
				SettingKeyPrivacyFilterConfig: `{"enabled":true,"types":["email","token","unknown","email"]}`,
			},
		}
		svc := NewSettingService(repo, nil)

		require.Equal(t, NormalizePrivacyFilterConfig(PrivacyFilterConfig{
			Enabled: true,
			Types:   []string{"email", "token"},
		}), svc.GetPrivacyFilterConfig(context.Background()))
	})
}

func TestSettingService_GetPrivacyFilterConfigErrorFallback(t *testing.T) {
	repo := &privacyFilterSettingRepoStubWithError{err: errors.New("db unavailable")}
	svc := NewSettingService(repo, nil)

	require.Equal(t, DefaultPrivacyFilterConfig(), svc.GetPrivacyFilterConfig(context.Background()))
}

type privacyFilterSettingRepoStubWithError struct {
	err error
}

func (s *privacyFilterSettingRepoStubWithError) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *privacyFilterSettingRepoStubWithError) GetValue(ctx context.Context, key string) (string, error) {
	return "", s.err
}

func (s *privacyFilterSettingRepoStubWithError) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *privacyFilterSettingRepoStubWithError) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *privacyFilterSettingRepoStubWithError) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *privacyFilterSettingRepoStubWithError) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *privacyFilterSettingRepoStubWithError) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}
