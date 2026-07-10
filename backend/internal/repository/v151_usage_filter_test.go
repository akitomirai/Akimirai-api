package repository

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestV151RequestTypeFilterSupportsQualifiedAlias(t *testing.T) {
	condition, args := buildRequestTypeFilterConditionWithAlias(3, int16(service.RequestTypeStream), "ul")
	require.Equal(t, "(ul.request_type = $3 OR (ul.request_type = 0 AND ul.stream = TRUE AND ul.openai_ws_mode = FALSE))", condition)
	require.Equal(t, []any{int16(service.RequestTypeStream)}, args)
}
