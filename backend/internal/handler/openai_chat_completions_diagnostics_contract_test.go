package handler

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChatCompletionsCarriesRequestDiagnosticsIntoUsage(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath.Join(".", "openai_chat_completions.go"), nil, 0)
	require.NoError(t, err)

	requiredCalls := map[string]bool{
		"setOpenAIClientTransportHTTP":         false,
		"beginOpenAIRequestDiagnostics":        false,
		"readOpenAIRequestBodyWithDiagnostics": false,
		"SelectAccount":                        false,
		"snapshotOpenAIRequestDiagnostics":     false,
	}
	requiredUsageFields := map[string]bool{
		"ClientTransport":   false,
		"AuthLatencyMs":     false,
		"RoutingLatencyMs":  false,
		"UpstreamLatencyMs": false,
		"ResponseLatencyMs": false,
		"Diagnostics":       false,
	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.CallExpr:
			if name := calledFunctionName(typed.Fun); name != "" {
				if _, required := requiredCalls[name]; required {
					requiredCalls[name] = true
				}
			}
		case *ast.CompositeLit:
			if !isOpenAIRecordUsageInputLiteral(typed.Type) {
				return true
			}
			for field := range requiredUsageFields {
				if compositeLiteralHasKey(typed, field) {
					requiredUsageFields[field] = true
				}
			}
		}
		return true
	})

	for name, found := range requiredCalls {
		require.Truef(t, found, "Chat Completions must call %s", name)
	}
	for field, found := range requiredUsageFields {
		require.Truef(t, found, "Chat Completions usage input must carry %s", field)
	}
}

func calledFunctionName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return typed.Sel.Name
	default:
		return ""
	}
}
