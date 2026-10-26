package toolsets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

func namedTool(name string) mcpgo.Tool {
	return mcpgo.NewTool(name, name, nil,
		func(context.Context, mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
			return mcpgo.NewToolResultText("ok"), nil
		})
}

func newFilterGroup(readOnly bool) *ToolsetGroup {
	tg := NewToolsetGroup(readOnly)
	tg.AddToolset(NewToolset("payments", "Payments").
		AddReadTools(namedTool("fetch_payment")).
		AddWriteTools(namedTool("capture_payment")))
	tg.AddToolset(NewToolset("payouts", "Payouts").
		AddReadTools(namedTool("fetch_all_payouts")))
	return tg
}

func notPayouts(name string) bool { return name != "fetch_all_payouts" }

func registeredNames(s *mockServer) []string {
	names := []string{}
	for _, tool := range s.GetTools() {
		names = append(names, tool.(interface{ GetName() string }).GetName())
	}
	return names
}

func TestToolsetGroup_SetToolFilter(t *testing.T) {
	t.Run("filter skips rejected tools", func(t *testing.T) {
		tg := newFilterGroup(false)
		tg.SetToolFilter(notPayouts)
		require.NoError(t, tg.EnableToolsets(nil))

		s := &mockServer{}
		tg.RegisterTools(s)

		assert.ElementsMatch(t,
			[]string{"fetch_payment", "capture_payment"}, registeredNames(s))
	})

	t.Run("no filter registers every tool", func(t *testing.T) {
		tg := newFilterGroup(false)
		require.NoError(t, tg.EnableToolsets(nil))

		s := &mockServer{}
		tg.RegisterTools(s)

		assert.Len(t, s.GetTools(), 3)
	})
}

func TestToolsetGroup_AllowedToolCount(t *testing.T) {
	t.Run("counts tools the filter allows", func(t *testing.T) {
		tg := newFilterGroup(false)
		tg.SetToolFilter(notPayouts)

		assert.Equal(t, 2, tg.AllowedToolCount("payments"))
		assert.Equal(t, 0, tg.AllowedToolCount("payouts"))
	})

	t.Run("read-only mode excludes write tools", func(t *testing.T) {
		tg := newFilterGroup(true)
		assert.Equal(t, 1, tg.AllowedToolCount("payments"))
	})

	t.Run("unknown toolset counts zero", func(t *testing.T) {
		assert.Equal(t, 0, newFilterGroup(false).AllowedToolCount("nope"))
	})
}
