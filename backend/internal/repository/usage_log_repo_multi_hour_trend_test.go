package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestUsageTrendWithFiltersUsesAnchoredMultiHourBucket(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}
	start := time.Date(2026, 7, 10, 13, 25, 0, 0, time.FixedZone("CST", 8*60*60))
	end := start.Add(48 * time.Hour)

	mock.ExpectQuery(`(?s)TO_CHAR\(\(date_bin\(INTERVAL '2 hours', created_at, \$1\)\) AT TIME ZONE 'UTC'.*as date`).
		WithArgs(start, end).
		WillReturnRows(sqlmock.NewRows([]string{"date", "requests", "input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens", "total_tokens", "cost", "actual_cost"}))

	points, err := repo.GetUsageTrendWithFilters(context.Background(), start, end, "2h", 0, 0, 0, 0, "", nil, nil, nil)
	require.NoError(t, err)
	require.Empty(t, points)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserUsageTrendUsesAnchoredEightHourBucket(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(14 * 24 * time.Hour)

	mock.ExpectQuery(`(?s)TO_CHAR\(\(date_bin\(INTERVAL '8 hours', u\.created_at, \$1\)\) AT TIME ZONE 'UTC'.*as date`).
		WithArgs(start, end, 12, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"date", "user_id", "email", "username", "requests", "tokens", "cost", "actual_cost"}))

	points, err := repo.GetUserUsageTrend(context.Background(), start, end, "8h", 12)
	require.NoError(t, err)
	require.Empty(t, points)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageLogRepositoryReturnsUnambiguousRFC3339BucketsAcrossDSTRepeat(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}
	originalLocal := time.Local
	time.Local = time.FixedZone("server", 8*60*60)
	t.Cleanup(func() { time.Local = originalLocal })

	newYork, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	start := time.Date(2026, 11, 1, 0, 0, 0, 0, newYork)
	end := start.Add(4 * time.Hour)
	firstOneAM := time.Date(2026, 11, 1, 5, 0, 0, 0, time.UTC)
	secondOneAM := firstOneAM.Add(time.Hour)

	mock.ExpectQuery(`(?s)TO_CHAR\(\(date_bin\(INTERVAL '1 hour', created_at, \$1\)\) AT TIME ZONE 'UTC'.*as date`).
		WithArgs(start, end).
		WillReturnRows(sqlmock.NewRows([]string{"date", "requests", "input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens", "total_tokens", "cost", "actual_cost"}).
			AddRow(firstOneAM.Format(time.RFC3339), int64(1), int64(10), int64(0), int64(0), int64(0), int64(10), 0.1, 0.1).
			AddRow(secondOneAM.Format(time.RFC3339), int64(1), int64(20), int64(0), int64(0), int64(0), int64(20), 0.2, 0.2))

	points, err := repo.GetUsageTrendWithFilters(context.Background(), start, end, "hour", 0, 0, 0, 0, "", nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, []string{
		"2026-11-01T05:00:00Z",
		"2026-11-01T06:00:00Z",
	}, []string{points[0].Date, points[1].Date})
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPreaggregatedTrendRequiresAlignedRange(t *testing.T) {
	alignedStart := time.Date(2026, 7, 13, 15, 0, 0, 0, time.UTC)
	alignedEnd := alignedStart.Add(24 * time.Hour)
	require.True(t, shouldUsePreaggregatedTrend("hour", alignedStart, alignedEnd, 0, 0, 0, 0, "", nil, nil, nil, ""))
	require.False(t, shouldUsePreaggregatedTrend("hour", alignedStart.Add(25*time.Minute), alignedEnd, 0, 0, 0, 0, "", nil, nil, nil, ""))
	require.False(t, shouldUsePreaggregatedTrend("day", alignedStart, alignedEnd, 0, 0, 0, 0, "", nil, nil, nil, ""))

	dayStart := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	require.True(t, shouldUsePreaggregatedTrend("day", dayStart, dayStart.Add(24*time.Hour), 0, 0, 0, 0, "", nil, nil, nil, ""))

	halfHourZone := time.FixedZone("IST", 5*60*60+30*60)
	halfHourStart := time.Date(2026, 7, 13, 0, 0, 0, 0, halfHourZone)
	require.False(t, shouldUsePreaggregatedTrend("hour", halfHourStart, halfHourStart.Add(24*time.Hour), 0, 0, 0, 0, "", nil, nil, nil, ""))
	require.False(t, shouldUsePreaggregatedTrend("day", halfHourStart, halfHourStart.Add(24*time.Hour), 0, 0, 0, 0, "", nil, nil, nil, ""))
}
