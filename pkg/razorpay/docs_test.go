package razorpay

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchDocumentation(t *testing.T) {
	obs := CreateTestObservability()
	tool := SearchDocumentation(obs)

	t.Run("returns results for a matching query", func(t *testing.T) {
		req := createMCPRequest(map[string]interface{}{
			"query": "webhook signature verification",
		})

		result, err := tool.GetHandler()(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)

		var out searchDocumentationOutput
		require.NoError(t, json.Unmarshal([]byte(result.Text), &out))
		assert.NotEmpty(t, out.Results)
		assert.NotEmpty(t, out.CodeExamples)
		assert.LessOrEqual(t, len(out.Results), 3)
		assert.Equal(t, "node", out.CodeExamples[0].Language)
	})

	t.Run("filters by topic", func(t *testing.T) {
		req := createMCPRequest(map[string]interface{}{
			"query": "integration",
			"topic": "upi-autopay",
		})

		result, err := tool.GetHandler()(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		var out searchDocumentationOutput
		require.NoError(t, json.Unmarshal([]byte(result.Text), &out))
		require.NotEmpty(t, out.Results)
	})

	t.Run("uses requested language note for non-node language", func(t *testing.T) {
		req := createMCPRequest(map[string]interface{}{
			"query":    "webhook signature verification",
			"language": "python",
		})

		result, err := tool.GetHandler()(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		var out searchDocumentationOutput
		require.NoError(t, json.Unmarshal([]byte(result.Text), &out))
		require.NotEmpty(t, out.CodeExamples)
		assert.Equal(t, "python", out.CodeExamples[0].Language)
		assert.Contains(t, out.CodeExamples[0].Code, "Example shown in Node.js")
	})

	t.Run("returns a helpful message when nothing matches", func(t *testing.T) {
		req := createMCPRequest(map[string]interface{}{
			"query": "zzzznonexistentqueryzzzz",
		})

		result, err := tool.GetHandler()(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		var out searchDocumentationOutput
		require.NoError(t, json.Unmarshal([]byte(result.Text), &out))
		assert.Empty(t, out.Results)
		assert.Contains(t, out.Message, "No documentation found")
	})

	t.Run("errors on missing query", func(t *testing.T) {
		req := createMCPRequest(map[string]interface{}{})

		result, err := tool.GetHandler()(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("errors on invalid argument type", func(t *testing.T) {
		req := createMCPRequest("not-a-map")

		result, err := tool.GetHandler()(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

func TestExplainError(t *testing.T) {
	obs := CreateTestObservability()
	tool := ExplainError(obs)

	t.Run("matches a known error code and sub-description", func(t *testing.T) {
		require.NotEmpty(t, errorRegistryData)
		known := errorRegistryData[0]

		req := createMCPRequest(map[string]interface{}{
			"error_code":      known.Code,
			"sub_description": known.SubDescription,
		})

		result, err := tool.GetHandler()(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)

		var out explainErrorOutput
		require.NoError(t, json.Unmarshal([]byte(result.Text), &out))
		assert.Equal(t, known.Code, out.ErrorCode)
		assert.Equal(t, known.Title, out.Title)
		assert.Equal(t, known.IsRetriable, out.IsRetriable)
	})

	t.Run("falls back to code-only match when description does not match", func(t *testing.T) {
		require.NotEmpty(t, errorRegistryData)
		known := errorRegistryData[0]

		req := createMCPRequest(map[string]interface{}{
			"error_code":      known.Code,
			"sub_description": "some unrelated description that will not match",
		})

		result, err := tool.GetHandler()(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		var out explainErrorOutput
		require.NoError(t, json.Unmarshal([]byte(result.Text), &out))
		assert.Equal(t, known.Code, out.ErrorCode)
	})

	t.Run("returns not-found fallback for unknown code and description", func(t *testing.T) {
		req := createMCPRequest(map[string]interface{}{
			"error_code":        "TOTALLY_UNKNOWN_ERROR_CODE",
			"error_description": "zzzznonexistentzzzz",
		})

		result, err := tool.GetHandler()(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		var out explainErrorOutput
		require.NoError(t, json.Unmarshal([]byte(result.Text), &out))
		assert.Equal(t, "Error Not Found in Registry", out.Title)
		assert.False(t, out.IsRetriable)
		assert.Nil(t, out.GuardrailRef)
		assert.Equal(t, "https://razorpay.com/docs/api/errors/", out.DocURL)
	})

	t.Run("errors on missing error_code", func(t *testing.T) {
		req := createMCPRequest(map[string]interface{}{})

		result, err := tool.GetHandler()(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("errors on invalid argument type", func(t *testing.T) {
		req := createMCPRequest(42)

		result, err := tool.GetHandler()(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}
