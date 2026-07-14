package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestOpsRepositoryListUserRequestLogs_UnifiesAndPaginatesAfterDedup(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &opsRepository{db: db}

	start := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	filter := &service.UserRequestLogFilter{
		StartTime:   start,
		EndTime:     end,
		UserID:      42,
		Kind:        service.UserRequestLogKindAll,
		SortBy:      "created_at",
		SortOrder:   "desc",
		Page:        1,
		PageSize:    20,
		AllowErrors: true,
	}

	mock.ExpectQuery(`(?s)WITH usage_rows AS .*ranked_errors AS .*FULL OUTER JOIN error_rows.*SELECT COUNT\(1\) FROM combined`).
		WithArgs(start, end, int64(42), true).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	columns := []string{
		"id", "kind", "created_at", "request_id",
		"api_key_id", "api_key_name", "api_key_deleted",
		"group_id", "group_name", "rate_multiplier",
		"model", "reasoning_effort", "first_token_ms", "duration_ms",
		"input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens", "total_tokens",
		"input_cost", "output_cost", "cache_creation_cost", "cache_read_cost", "total_cost", "actual_cost", "status_code",
		"error_phase", "error_type", "error_message", "error_stream",
	}
	rows := sqlmock.NewRows(columns).
		AddRow(
			int64(91), "error", start.Add(2*time.Hour), "req-error",
			int64(7), "Baka1", false,
			int64(9), "Pro号池", 1.2,
			"gpt-5.6-sol", "high", int64(3000), int64(4000),
			int64(1000), int64(196), int64(2000), int64(14000), int64(17196),
			0.001, 0.002, 0.003, 0.004, 0.010, 0.011508, int64(429),
			"upstream", "rate_limit_error", "rate limited", true,
		).
		AddRow(
			int64(90), "consumption", start.Add(time.Hour), "req-ok",
			int64(7), "Baka1", false,
			int64(9), "Pro号池", 1.0,
			"gpt-5.4-mini", nil, int64(3500), int64(6000),
			int64(5000), int64(740), int64(0), int64(1000), int64(6740),
			0.001, 0.002, 0.0, 0.001, 0.004, 0.005888, nil,
			"", "", "", false,
		)
	mock.ExpectQuery(`(?s)FROM combined.*ORDER BY created_at DESC, kind DESC, id DESC.*LIMIT \$5 OFFSET \$6`).
		WithArgs(start, end, int64(42), true, 20, 0).
		WillReturnRows(rows)

	items, total, err := repo.ListUserRequestLogs(context.Background(), filter)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, items, 2)
	require.Equal(t, service.UserRequestLogKindError, items[0].Kind)
	require.Equal(t, "Pro号池", items[0].GroupName)
	require.NotNil(t, items[0].RateMultiplier)
	require.InDelta(t, 1.2, *items[0].RateMultiplier, 0.0001)
	require.NotNil(t, items[0].ReasoningEffort)
	require.Equal(t, "high", *items[0].ReasoningEffort)
	require.NotNil(t, items[0].StatusCode)
	require.Equal(t, 429, *items[0].StatusCode)
	require.NotNil(t, items[0].TotalTokens)
	require.Equal(t, int64(17196), *items[0].TotalTokens)
	require.NotNil(t, items[0].InputTokens)
	require.Equal(t, int64(1000), *items[0].InputTokens)
	require.NotNil(t, items[0].OutputTokens)
	require.Equal(t, int64(196), *items[0].OutputTokens)
	require.NotNil(t, items[0].CacheCreationTokens)
	require.Equal(t, int64(2000), *items[0].CacheCreationTokens)
	require.NotNil(t, items[0].CacheReadTokens)
	require.Equal(t, int64(14000), *items[0].CacheReadTokens)
	require.NotNil(t, items[0].TotalCost)
	require.InDelta(t, 0.010, *items[0].TotalCost, 0.000001)
	require.NotNil(t, items[0].ActualCost)
	require.InDelta(t, 0.011508, *items[0].ActualCost, 0.000001)
	require.Equal(t, service.UserRequestLogKindConsumption, items[1].Kind)
	require.Nil(t, items[1].StatusCode)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOpsRepositoryListUserRequestLogs_PreservesUnknownStandaloneErrorMetrics(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &opsRepository{db: db}
	start := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	filter := &service.UserRequestLogFilter{
		StartTime: start, EndTime: end, UserID: 7,
		Kind: service.UserRequestLogKindError, Page: 1, PageSize: 10, AllowErrors: true,
	}

	mock.ExpectQuery(`(?s)SELECT COUNT\(1\) FROM combined.*WHERE kind = \$5`).
		WithArgs(start, end, int64(7), true, "error").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`(?s)FROM combined.*WHERE kind = \$5.*LIMIT \$6 OFFSET \$7`).
		WithArgs(start, end, int64(7), true, "error", 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "kind", "created_at", "request_id",
			"api_key_id", "api_key_name", "api_key_deleted",
			"group_id", "group_name", "rate_multiplier",
			"model", "reasoning_effort", "first_token_ms", "duration_ms",
			"input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens", "total_tokens",
			"input_cost", "output_cost", "cache_creation_cost", "cache_read_cost", "total_cost", "actual_cost", "status_code",
			"error_phase", "error_type", "error_message", "error_stream",
		}).AddRow(
			int64(99), "error", start, "req-error-only",
			int64(2), "key", false,
			nil, "", nil,
			"gpt-5", nil, nil, int64(1200),
			nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, int64(500),
			"upstream", "upstream_error", "failed", false,
		))

	items, total, err := repo.ListUserRequestLogs(context.Background(), filter)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Nil(t, items[0].TotalTokens)
	require.Nil(t, items[0].InputTokens)
	require.Nil(t, items[0].CacheReadTokens)
	require.Nil(t, items[0].TotalCost)
	require.Nil(t, items[0].ActualCost)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOpsRepositoryListUserRequestLogs_DisablesErrorLedgerBeforeCorrelation(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &opsRepository{db: db}
	start := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	filter := &service.UserRequestLogFilter{
		StartTime:   start,
		EndTime:     end,
		UserID:      7,
		Kind:        service.UserRequestLogKindConsumption,
		Page:        1,
		PageSize:    10,
		AllowErrors: false,
	}

	mock.ExpectQuery(`(?s)WHERE \$4::BOOLEAN.*o\.is_count_tokens.*FULL OUTER JOIN error_rows.*WHERE kind = \$5`).
		WithArgs(start, end, int64(7), false, "consumption").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`(?s)FROM combined.*WHERE kind = \$5.*LIMIT \$6 OFFSET \$7`).
		WithArgs(start, end, int64(7), false, "consumption", 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "kind", "created_at", "request_id",
			"api_key_id", "api_key_name", "api_key_deleted",
			"group_id", "group_name", "rate_multiplier",
			"model", "reasoning_effort", "first_token_ms", "duration_ms",
			"input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens", "total_tokens",
			"input_cost", "output_cost", "cache_creation_cost", "cache_read_cost", "total_cost", "actual_cost", "status_code",
			"error_phase", "error_type", "error_message", "error_stream",
		}))

	items, total, err := repo.ListUserRequestLogs(context.Background(), filter)
	require.NoError(t, err)
	require.Empty(t, items)
	require.Zero(t, total)
	require.NoError(t, mock.ExpectationsWereMet())
}
