package mcpgo

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/razorpay/razorpay-mcp-server/pkg/region"
)

func noopHandler(context.Context, CallToolRequest) (*ToolResult, error) {
	return NewToolResultText("ok"), nil
}

func TestMark3labsToolImpl_GetName(t *testing.T) {
	tool := NewTool("fetch_payment", "Fetch", nil, noopHandler)
	assert.Equal(t, "fetch_payment", tool.GetName())
}

func TestMark3labsToolImpl_Localize(t *testing.T) {
	fill := func(s string) string {
		return strings.ReplaceAll(s, "{{currency}}", "SGD")
	}

	t.Run("rewrites the tool description", func(t *testing.T) {
		tool := NewTool("t", "Charge in {{currency}}", nil, noopHandler)
		tool.Localize(fill)
		assert.Equal(t, "Charge in SGD", tool.description)
	})

	t.Run("rewrites parameter descriptions", func(t *testing.T) {
		tool := NewTool("t", "d", []ToolParameter{
			WithNumber("amount", Description("Amount in {{currency}}")),
		}, noopHandler)
		tool.Localize(fill)
		assert.Equal(t, "Amount in SGD",
			tool.parameters[0].Schema["description"])
	})

	t.Run("rewrites descriptions nested in properties and items",
		func(t *testing.T) {
			tool := NewTool("t", "d", []ToolParameter{
				WithArray("transfers", Items(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"amount": map[string]interface{}{
							"type":        "integer",
							"description": "Transfer in {{currency}}",
						},
					},
				})),
			}, noopHandler)
			tool.Localize(fill)

			items := tool.parameters[0].Schema["items"].(map[string]interface{})
			props := items["properties"].(map[string]interface{})
			amount := props["amount"].(map[string]interface{})
			assert.Equal(t, "Transfer in SGD", amount["description"])
		})
}

func TestMark3labsToolImpl_ApplyProfile(t *testing.T) {
	sg := region.Profile{CountryCode: "SG"}
	us := region.Profile{CountryCode: "US"}
	countryHandler := func(
		ctx context.Context, _ CallToolRequest,
	) (*ToolResult, error) {
		return NewToolResultText(region.FromContext(ctx).CountryCode), nil
	}
	call := func(t *testing.T, tool Tool, ctx context.Context) string {
		t.Helper()
		res, err := tool.GetHandler()(ctx, CallToolRequest{})
		require.NoError(t, err)
		return res.Text
	}

	t.Run("localizes the description", func(t *testing.T) {
		tool := NewTool("t", "Charge in {{currency}}", nil, noopHandler)
		tool.ApplyProfile(sg)
		assert.Equal(t, "Charge in SGD", tool.description)
	})

	t.Run("handler gets the profile when ctx has none", func(t *testing.T) {
		tool := NewTool("t", "d", nil, countryHandler)
		tool.ApplyProfile(sg)
		assert.Equal(t, "SG", call(t, tool, context.Background()))
	})

	t.Run("profile in ctx overrides the applied one", func(t *testing.T) {
		tool := NewTool("t", "d", nil, countryHandler)
		tool.ApplyProfile(sg)
		ctx := region.WithProfile(context.Background(), us)
		assert.Equal(t, "US", call(t, tool, ctx))
	})
}
