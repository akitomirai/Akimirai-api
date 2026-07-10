package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

func TestUsageLogRepositoryGetAPIKeyStatsWithUsageFilters(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}

	start := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	filters := usagestats.UsageLogFilters{
		UserID:            42,
		GroupID:           7,
		Model:             "gpt-5.6-sol",
		ModelFilterSource: usagestats.ModelSourceRequested,
	}

	mock.ExpectQuery(`(?s)FROM usage_logs ul.*LEFT JOIN api_keys k.*AND ul.user_id = \$3.*AND ul.group_id = \$4.*COALESCE\(NULLIF\(TRIM\(ul.requested_model\), ''\), ul.model\) = \$5.*GROUP BY ul.api_key_id, k.name`).
		WithArgs(start, end, int64(42), int64(7), "gpt-5.6-sol").
		WillReturnRows(sqlmock.NewRows([]string{
			"api_key_id", "api_key_name", "requests", "total_tokens", "cost", "actual_cost",
		}).AddRow(int64(9), "production", int64(3), int64(1200), 0.25, 0.12))

	stats, err := repo.GetAPIKeyStatsWithUsageFilters(context.Background(), start, end, filters)
	require.NoError(t, err)
	require.Equal(t, []usagestats.APIKeyStat{{
		APIKeyID:    9,
		APIKeyName:  "production",
		Requests:    3,
		TotalTokens: 1200,
		Cost:        0.25,
		ActualCost:  0.12,
	}}, stats)
	require.NoError(t, mock.ExpectationsWereMet())
}
