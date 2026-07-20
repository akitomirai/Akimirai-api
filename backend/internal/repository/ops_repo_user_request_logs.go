package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *opsRepository) ListUserRequestLogs(
	ctx context.Context,
	filter *service.UserRequestLogFilter,
) ([]*service.UserRequestLog, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("nil ops repository")
	}
	if filter == nil || filter.UserID <= 0 {
		return nil, 0, fmt.Errorf("invalid user request log filter")
	}

	args := []any{filter.StartTime.UTC(), filter.EndTime.UTC(), filter.UserID, filter.AllowErrors}
	conditions := make([]string, 0, 5)
	addCondition := func(format string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(format, len(args)))
	}

	if filter.Kind != "" && filter.Kind != service.UserRequestLogKindAll {
		addCondition("kind = $%d", string(filter.Kind))
	}
	if filter.APIKeyID != nil && *filter.APIKeyID > 0 {
		addCondition("api_key_id = $%d", *filter.APIKeyID)
	}
	if filter.GroupID != nil && *filter.GroupID > 0 {
		addCondition("group_id = $%d", *filter.GroupID)
	}
	if model := strings.TrimSpace(filter.Model); model != "" {
		addCondition("model = $%d", model)
	}
	if requestID := strings.TrimSpace(filter.RequestID); requestID != "" {
		addCondition("request_id = $%d", requestID)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	cte := `
WITH usage_rows AS (
  SELECT
    ul.id,
    ul.created_at,
    COALESCE(ul.request_id, '') AS request_id,
    ul.api_key_id,
    COALESCE(k.name, '') AS api_key_name,
    (k.deleted_at IS NOT NULL) AS api_key_deleted,
    ul.group_id,
    COALESCE(g.name, '') AS group_name,
    ul.rate_multiplier::DOUBLE PRECISION AS rate_multiplier,
    COALESCE(NULLIF(TRIM(ul.requested_model), ''), ul.model, '') AS model,
    NULLIF(TRIM(ul.reasoning_effort), '') AS reasoning_effort,
    ul.first_token_ms,
    ul.duration_ms,
	ul.input_tokens::BIGINT AS input_tokens,
	ul.output_tokens::BIGINT AS output_tokens,
	ul.cache_creation_tokens::BIGINT AS cache_creation_tokens,
	ul.cache_read_tokens::BIGINT AS cache_read_tokens,
    (ul.input_tokens + ul.output_tokens + ul.cache_creation_tokens + ul.cache_read_tokens)::BIGINT AS total_tokens,
	ul.input_cost::DOUBLE PRECISION AS input_cost,
	ul.output_cost::DOUBLE PRECISION AS output_cost,
	ul.cache_creation_cost::DOUBLE PRECISION AS cache_creation_cost,
	ul.cache_read_cost::DOUBLE PRECISION AS cache_read_cost,
	ul.total_cost::DOUBLE PRECISION AS total_cost,
    ul.actual_cost::DOUBLE PRECISION AS actual_cost
  FROM usage_logs ul
  LEFT JOIN api_keys k ON k.id = ul.api_key_id
  LEFT JOIN groups g ON g.id = ul.group_id
  WHERE ul.created_at >= $1 AND ul.created_at < $2
    AND ul.user_id = $3
),
error_base AS (
  SELECT
    o.id,
    o.created_at,
    COALESCE(NULLIF(TRIM(o.request_id), ''), NULLIF(TRIM(o.client_request_id), ''), '') AS request_id,
    o.api_key_id,
    COALESCE(NULLIF(k.name, ''), NULLIF(o.deleted_key_name, ''), '') AS api_key_name,
    (k.deleted_at IS NOT NULL OR (o.api_key_id IS NULL AND NULLIF(o.deleted_key_name, '') IS NOT NULL)) AS api_key_deleted,
    o.group_id,
    COALESCE(g.name, '') AS group_name,
    COALESCE(NULLIF(TRIM(o.requested_model), ''), o.model, '') AS model,
    o.time_to_first_token_ms,
    o.duration_ms,
    o.status_code,
    COALESCE(o.error_phase, '') AS error_phase,
    COALESCE(o.error_type, '') AS error_type,
    COALESCE(o.error_message, '') AS error_message,
    o.stream,
    o.is_business_limited
  FROM ops_error_logs o
  LEFT JOIN api_keys k ON k.id = o.api_key_id
  LEFT JOIN groups g ON g.id = o.group_id
  WHERE $4::BOOLEAN
    AND o.created_at >= $1 AND o.created_at < $2
    AND (o.user_id = $3 OR o.deleted_key_owner_user_id = $3)
    AND NOT COALESCE(o.is_count_tokens, FALSE)
    AND (COALESCE(o.status_code, 0) >= 400 OR COALESCE(o.error_type, '') LIKE 'cyber_policy%')
),
ranked_errors AS (
  SELECT
    error_base.*,
    ROW_NUMBER() OVER (
      PARTITION BY
        CASE WHEN request_id <> '' THEN request_id ELSE '__error__' || id::TEXT END,
        COALESCE(api_key_id, 0)
      ORDER BY created_at DESC, id DESC
    ) AS row_number
  FROM error_base
),
error_rows AS (
  SELECT * FROM ranked_errors WHERE row_number = 1
),
combined AS (
  SELECT
    COALESCE(e.id, u.id) AS id,
    CASE WHEN e.id IS NOT NULL THEN 'error' ELSE 'consumption' END AS kind,
    CASE WHEN e.id IS NOT NULL THEN e.created_at ELSE u.created_at END AS created_at,
    COALESCE(NULLIF(e.request_id, ''), u.request_id, '') AS request_id,
    COALESCE(e.api_key_id, u.api_key_id) AS api_key_id,
    COALESCE(NULLIF(e.api_key_name, ''), u.api_key_name, '') AS api_key_name,
    COALESCE(e.api_key_deleted, u.api_key_deleted, FALSE) AS api_key_deleted,
    COALESCE(e.group_id, u.group_id) AS group_id,
    COALESCE(NULLIF(e.group_name, ''), u.group_name, '') AS group_name,
    u.rate_multiplier AS rate_multiplier,
    COALESCE(NULLIF(e.model, ''), u.model, '') AS model,
    u.reasoning_effort AS reasoning_effort,
    COALESCE(e.time_to_first_token_ms, u.first_token_ms) AS first_token_ms,
    COALESCE(e.duration_ms, u.duration_ms) AS duration_ms,
	u.input_tokens AS input_tokens,
	u.output_tokens AS output_tokens,
	u.cache_creation_tokens AS cache_creation_tokens,
	u.cache_read_tokens AS cache_read_tokens,
    u.total_tokens AS total_tokens,
	u.input_cost AS input_cost,
	u.output_cost AS output_cost,
	u.cache_creation_cost AS cache_creation_cost,
	u.cache_read_cost AS cache_read_cost,
	u.total_cost AS total_cost,
    u.actual_cost AS actual_cost,
    e.status_code,
    COALESCE(e.error_phase, '') AS error_phase,
    COALESCE(e.error_type, '') AS error_type,
    COALESCE(e.error_message, '') AS error_message,
    COALESCE(e.stream, FALSE) AS error_stream
  FROM usage_rows u
  FULL OUTER JOIN error_rows e
    ON e.request_id <> ''
   AND u.request_id = e.request_id
   AND u.api_key_id = e.api_key_id
)
`

	countQuery := fmt.Sprintf("%s SELECT COUNT(1) FROM combined %s", cte, where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortColumn := "created_at"
	if filter.SortBy == "duration_ms" {
		sortColumn = "duration_ms"
	}
	sortOrder := "DESC"
	if strings.EqualFold(filter.SortOrder, "asc") {
		sortOrder = "ASC"
	}
	orderBy := fmt.Sprintf("ORDER BY %s %s NULLS LAST, created_at DESC, kind DESC, id DESC", sortColumn, sortOrder)
	if sortColumn == "created_at" {
		orderBy = fmt.Sprintf("ORDER BY created_at %s, kind DESC, id DESC", sortOrder)
	}

	limitPosition := len(args) + 1
	offsetPosition := len(args) + 2
	listQuery := fmt.Sprintf(`
%s
SELECT
  id, kind, created_at, request_id,
  api_key_id, api_key_name, api_key_deleted,
  group_id, group_name, rate_multiplier,
  model, reasoning_effort, first_token_ms, duration_ms,
	input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens, total_tokens,
	input_cost, output_cost, cache_creation_cost, cache_read_cost, total_cost, actual_cost, status_code,
  error_phase, error_type, error_message, error_stream
FROM combined
%s
%s
LIMIT $%d OFFSET $%d
`, cte, where, orderBy, limitPosition, offsetPosition)

	listArgs := append(append([]any{}, args...), filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]*service.UserRequestLog, 0, filter.PageSize)
	for rows.Next() {
		var (
			item                service.UserRequestLog
			kind                string
			apiKeyID            sql.NullInt64
			groupID             sql.NullInt64
			rateMultiplier      sql.NullFloat64
			reasoningEffort     sql.NullString
			firstTokenMs        sql.NullInt64
			durationMs          sql.NullInt64
			inputTokens         sql.NullInt64
			outputTokens        sql.NullInt64
			cacheCreationTokens sql.NullInt64
			cacheReadTokens     sql.NullInt64
			totalTokens         sql.NullInt64
			inputCost           sql.NullFloat64
			outputCost          sql.NullFloat64
			cacheCreationCost   sql.NullFloat64
			cacheReadCost       sql.NullFloat64
			totalCost           sql.NullFloat64
			actualCost          sql.NullFloat64
			statusCode          sql.NullInt64
		)
		if err := rows.Scan(
			&item.ID, &kind, &item.CreatedAt, &item.RequestID,
			&apiKeyID, &item.APIKeyName, &item.APIKeyDeleted,
			&groupID, &item.GroupName, &rateMultiplier,
			&item.Model, &reasoningEffort, &firstTokenMs, &durationMs,
			&inputTokens, &outputTokens, &cacheCreationTokens, &cacheReadTokens, &totalTokens,
			&inputCost, &outputCost, &cacheCreationCost, &cacheReadCost, &totalCost, &actualCost, &statusCode,
			&item.RawErrorPhase, &item.RawErrorType, &item.RawErrorMessage,
			&item.RawErrorStream,
		); err != nil {
			return nil, 0, err
		}
		item.Kind = service.UserRequestLogKind(kind)
		item.APIKeyID = nullInt64Pointer(apiKeyID)
		item.GroupID = nullInt64Pointer(groupID)
		item.RateMultiplier = nullFloat64Pointer(rateMultiplier)
		item.ReasoningEffort = nullStringPointer(reasoningEffort)
		item.FirstTokenMs = nullIntPointer(firstTokenMs)
		item.DurationMs = nullIntPointer(durationMs)
		item.InputTokens = nullInt64Pointer(inputTokens)
		item.OutputTokens = nullInt64Pointer(outputTokens)
		item.CacheCreationTokens = nullInt64Pointer(cacheCreationTokens)
		item.CacheReadTokens = nullInt64Pointer(cacheReadTokens)
		item.TotalTokens = nullInt64Pointer(totalTokens)
		item.InputCost = nullFloat64Pointer(inputCost)
		item.OutputCost = nullFloat64Pointer(outputCost)
		item.CacheCreationCost = nullFloat64Pointer(cacheCreationCost)
		item.CacheReadCost = nullFloat64Pointer(cacheReadCost)
		item.TotalCost = nullFloat64Pointer(totalCost)
		item.ActualCost = nullFloat64Pointer(actualCost)
		item.StatusCode = nullIntPointer(statusCode)
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func nullInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func nullIntPointer(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	result := int(value.Int64)
	return &result
}

func nullFloat64Pointer(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	result := value.Float64
	return &result
}

func nullStringPointer(value sql.NullString) *string {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}
	result := strings.TrimSpace(value.String)
	return &result
}
