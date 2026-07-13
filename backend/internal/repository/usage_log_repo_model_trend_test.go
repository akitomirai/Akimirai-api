package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUsageLogRepositoryGetModelUsageTrendWithUsageFilters(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}

	start := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 7)
	requestType := int16(service.RequestTypeStream)
	filters := usagestats.UsageLogFilters{
		UserID:      7,
		Model:       "gpt-5.6-sol",
		RequestType: &requestType,
		BillingType: func() *int8 { value := int8(service.BillingTypeBalance); return &value }(),
		BillingMode: string(service.BillingModeToken),
	}

	mock.ExpectQuery(`(?s)WITH filtered AS .*requested_model.*ranked_models AS .*ROW_NUMBER\(\).*LIMIT \$8.*bucketed AS \(.*COALESCE\(rm\.model, ''\) AS model.*\(rm\.model IS NULL\) AS is_other.*GROUP BY f\.date, rm\.model, rm\.model_rank.*ORDER BY model_rank ASC, date ASC`).
		WithArgs(start, end, int64(7), "gpt-5.6-sol", requestType, int16(service.BillingTypeBalance), string(service.BillingModeToken), 12).
		WillReturnRows(sqlmock.NewRows([]string{"date", "model", "is_other", "requests", "total_tokens", "cost", "actual_cost"}).
			AddRow("2026-07-06T00:00:00Z", "gpt-5.6-sol", false, int64(42), int64(12000), 1.25, 1.1).
			AddRow("2026-07-06T00:00:00Z", "", true, int64(3), int64(900), 0.15, 0.12))

	points, err := repo.GetModelUsageTrendWithUsageFilters(context.Background(), start, end, "day", filters, 99)
	require.NoError(t, err)
	require.Equal(t, []ModelUsageTrendPoint{
		{Date: "2026-07-06T00:00:00Z", Model: "gpt-5.6-sol", Requests: 42, TotalTokens: 12000, Cost: 1.25, ActualCost: 1.1},
		{Date: "2026-07-06T00:00:00Z", IsOther: true, Requests: 3, TotalTokens: 900, Cost: 0.15, ActualCost: 0.12},
	}, points)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageLogRepositoryGetModelUsageTrendWithUsageFiltersEmpty(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}

	start := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	mock.ExpectQuery(`(?s)WITH filtered AS .*LIMIT \$3.*COALESCE\(rm\.model_rank, 9\).*ORDER BY model_rank ASC, date ASC`).
		WithArgs(start, end, 8).
		WillReturnRows(sqlmock.NewRows([]string{"date", "model", "is_other", "requests", "total_tokens", "cost", "actual_cost"}))

	points, err := repo.GetModelUsageTrendWithUsageFilters(context.Background(), start, end, "hour", usagestats.UsageLogFilters{}, 0)
	require.NoError(t, err)
	require.Empty(t, points)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageLogRepositoryAnchorsMultiHourBucketsAtRangeStart(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}

	start := time.Date(2026, 7, 10, 13, 25, 0, 0, time.FixedZone("CST", 8*60*60))
	end := start.Add(48 * time.Hour)
	mock.ExpectQuery(`(?s)TO_CHAR\(\(date_bin\(INTERVAL '2 hours', ul\.created_at, \$1\)\) AT TIME ZONE 'UTC'.*AS date`).
		WithArgs(start, end, 8).
		WillReturnRows(sqlmock.NewRows([]string{"date", "model", "is_other", "requests", "total_tokens", "cost", "actual_cost"}))

	points, err := repo.GetModelUsageTrendWithUsageFilters(context.Background(), start, end, "2h", usagestats.UsageLogFilters{}, 8)
	require.NoError(t, err)
	require.Empty(t, points)
	require.NoError(t, mock.ExpectationsWereMet())
}
