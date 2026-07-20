package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type userRequestLogRepoStub struct {
	OpsRepository
	filter *UserRequestLogFilter
	items  []*UserRequestLog
	total  int64
}

func (s *userRequestLogRepoStub) ListUserRequestLogs(_ context.Context, filter *UserRequestLogFilter) ([]*UserRequestLog, int64, error) {
	copy := *filter
	s.filter = &copy
	return s.items, s.total, nil
}

func TestOpsServiceListUserRequestLogs_ForcesScopeAndClassifiesErrors(t *testing.T) {
	status := 429
	repo := &userRequestLogRepoStub{
		items: []*UserRequestLog{{
			ID:              3,
			Kind:            UserRequestLogKindError,
			StatusCode:      &status,
			RawErrorPhase:   "upstream",
			RawErrorType:    "rate_limit_error",
			RawErrorMessage: "rate limited",
		}},
		total: 1,
	}
	svc := &OpsService{opsRepo: repo}
	filter := &UserRequestLogFilter{
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now(),
		Kind:      UserRequestLogKindAll,
		PageSize:  500,
	}

	result, err := svc.ListUserRequestLogs(context.Background(), 42, true, filter)
	require.NoError(t, err)
	require.Equal(t, int64(42), repo.filter.UserID)
	require.True(t, repo.filter.AllowErrors)
	require.Equal(t, 100, repo.filter.PageSize)
	require.Equal(t, UserRequestLogKindAll, filter.Kind, "caller filter must not be mutated")
	require.Len(t, result.Items, 1)
	require.NotNil(t, result.Items[0].ErrorCode)
	require.Equal(t, string(UserErrorCodeUpstreamRateLimited), *result.Items[0].ErrorCode)
	require.NotNil(t, result.Items[0].ErrorMessage)
}

func TestOpsServiceListUserRequestLogs_HidesErrorsWhenDisabled(t *testing.T) {
	repo := &userRequestLogRepoStub{}
	svc := &OpsService{opsRepo: repo}

	result, err := svc.ListUserRequestLogs(context.Background(), 5, false, &UserRequestLogFilter{
		Kind: UserRequestLogKindAll,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, UserRequestLogKindConsumption, repo.filter.Kind)
	require.False(t, repo.filter.AllowErrors)

	_, err = svc.ListUserRequestLogs(context.Background(), 5, false, &UserRequestLogFilter{
		Kind: UserRequestLogKindError,
	})
	require.Error(t, err)
}
