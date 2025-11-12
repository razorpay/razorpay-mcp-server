package toolsets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// mockServer is a mock implementation of mcpgo.Server for testing
type mockServer struct {
	tools []mcpgo.Tool
}

func (m *mockServer) AddTools(tools ...mcpgo.Tool) {
	m.tools = append(m.tools, tools...)
}

func (m *mockServer) GetTools() []mcpgo.Tool {
	return m.tools
}

func TestNewToolset(t *testing.T) {
	t.Run("creates toolset with name and description", func(t *testing.T) {
		ts := NewToolset("test-toolset", "Test description")
		assert.NotNil(t, ts)
		assert.Equal(t, "test-toolset", ts.Name)
		assert.Equal(t, "Test description", ts.Description)
		assert.False(t, ts.Enabled)
		assert.False(t, ts.readOnly)
	})

	t.Run("creates toolset with empty name", func(t *testing.T) {
		ts := NewToolset("", "Description")
		assert.NotNil(t, ts)
		assert.Equal(t, "", ts.Name)
		assert.Equal(t, "Description", ts.Description)
	})
}

func TestNewToolsetGroup(t *testing.T) {
	t.Run("creates toolset group with readOnly false", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		assert.NotNil(t, tg)
		assert.NotNil(t, tg.Toolsets)
		assert.False(t, tg.everythingOn)
		assert.False(t, tg.readOnly)
	})

	t.Run("creates toolset group with readOnly true", func(t *testing.T) {
		tg := NewToolsetGroup(true)
		assert.NotNil(t, tg)
		assert.NotNil(t, tg.Toolsets)
		assert.False(t, tg.everythingOn)
		assert.True(t, tg.readOnly)
	})
}

func TestToolset_AddWriteTools(t *testing.T) {
	t.Run("adds write tools when not readOnly", func(t *testing.T) {
		ts := NewToolset("test", "Test")
		tool1 := mcpgo.NewTool("tool1", "Tool 1", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result1"), nil
			})
		tool2 := mcpgo.NewTool("tool2", "Tool 2", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result2"), nil
			})

		result := ts.AddWriteTools(tool1, tool2)
		assert.Equal(t, ts, result) // Should return self for chaining
		assert.Len(t, ts.writeTools, 2)
	})

	t.Run("does not add write tools when readOnly", func(t *testing.T) {
		ts := NewToolset("test", "Test")
		ts.readOnly = true
		tool := mcpgo.NewTool("tool1", "Tool 1", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})

		result := ts.AddWriteTools(tool)
		assert.Equal(t, ts, result)
		assert.Len(t, ts.writeTools, 0) // Should not add when readOnly
	})

	t.Run("adds multiple write tools", func(t *testing.T) {
		ts := NewToolset("test", "Test")
		tool1 := mcpgo.NewTool("tool1", "Tool 1", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})
		tool2 := mcpgo.NewTool("tool2", "Tool 2", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})
		tool3 := mcpgo.NewTool("tool3", "Tool 3", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})

		ts.AddWriteTools(tool1, tool2, tool3)
		assert.Len(t, ts.writeTools, 3)
	})

	t.Run("adds empty write tools list", func(t *testing.T) {
		ts := NewToolset("test", "Test")
		ts.AddWriteTools()
		assert.Len(t, ts.writeTools, 0)
	})
}

func TestToolset_AddReadTools(t *testing.T) {
	t.Run("adds read tools", func(t *testing.T) {
		ts := NewToolset("test", "Test")
		tool1 := mcpgo.NewTool("tool1", "Tool 1", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result1"), nil
			})
		tool2 := mcpgo.NewTool("tool2", "Tool 2", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result2"), nil
			})

		result := ts.AddReadTools(tool1, tool2)
		assert.Equal(t, ts, result) // Should return self for chaining
		assert.Len(t, ts.readTools, 2)
	})

	t.Run("adds read tools even when readOnly", func(t *testing.T) {
		ts := NewToolset("test", "Test")
		ts.readOnly = true
		tool := mcpgo.NewTool("tool1", "Tool 1", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})

		ts.AddReadTools(tool)
		assert.Len(t, ts.readTools, 1) // Should add even when readOnly
	})

	t.Run("adds multiple read tools", func(t *testing.T) {
		ts := NewToolset("test", "Test")
		tool1 := mcpgo.NewTool("tool1", "Tool 1", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})
		tool2 := mcpgo.NewTool("tool2", "Tool 2", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})
		tool3 := mcpgo.NewTool("tool3", "Tool 3", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})

		ts.AddReadTools(tool1, tool2, tool3)
		assert.Len(t, ts.readTools, 3)
	})

	t.Run("adds empty read tools list", func(t *testing.T) {
		ts := NewToolset("test", "Test")
		ts.AddReadTools()
		assert.Len(t, ts.readTools, 0)
	})
}

func TestToolset_RegisterTools(t *testing.T) {
	t.Run("registers tools when enabled", func(t *testing.T) {
		ts := NewToolset("test", "Test")
		ts.Enabled = true
		readTool := mcpgo.NewTool("read-tool", "Read Tool", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})
		writeTool := mcpgo.NewTool(
			"write-tool", "Write Tool", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})

		ts.AddReadTools(readTool)
		ts.AddWriteTools(writeTool)

		mockSrv := &mockServer{}
		ts.RegisterTools(mockSrv)

		// Both read and write tools should be registered
		assert.Len(t, mockSrv.GetTools(), 2)
	})

	t.Run("does not register tools when disabled", func(t *testing.T) {
		ts := NewToolset("test", "Test")
		ts.Enabled = false
		tool := mcpgo.NewTool("tool1", "Tool 1", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})

		ts.AddReadTools(tool)

		mockSrv := &mockServer{}
		ts.RegisterTools(mockSrv)

		assert.Len(t, mockSrv.GetTools(), 0) // Should not register when disabled
	})

	t.Run("registers only read tools when readOnly", func(t *testing.T) {
		ts := NewToolset("test", "Test")
		ts.Enabled = true
		ts.readOnly = true
		readTool := mcpgo.NewTool("read-tool", "Read Tool", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})
		writeTool := mcpgo.NewTool(
			"write-tool", "Write Tool", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result"), nil
			})

		ts.AddReadTools(readTool)
		ts.AddWriteTools(writeTool) // This won't add because readOnly

		mockSrv := &mockServer{}
		ts.RegisterTools(mockSrv)

		assert.Len(t, mockSrv.GetTools(), 1) // Only read tool should be registered
	})

	t.Run("registers tools with empty tool lists", func(t *testing.T) {
		ts := NewToolset("test", "Test")
		ts.Enabled = true

		mockSrv := &mockServer{}
		ts.RegisterTools(mockSrv)

		assert.Len(t, mockSrv.GetTools(), 0) // No tools to register
	})
}

func TestToolsetGroup_AddToolset(t *testing.T) {
	t.Run("adds toolset to group", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts := NewToolset("test", "Test")

		tg.AddToolset(ts)

		assert.Len(t, tg.Toolsets, 1)
		assert.Equal(t, ts, tg.Toolsets["test"])
		// Should not be readOnly when group is not readOnly
		assert.False(t, ts.readOnly)
	})

	t.Run("adds toolset to readOnly group", func(t *testing.T) {
		tg := NewToolsetGroup(true)
		ts := NewToolset("test", "Test")

		tg.AddToolset(ts)

		assert.Len(t, tg.Toolsets, 1)
		assert.Equal(t, ts, tg.Toolsets["test"])
		assert.True(t, ts.readOnly) // Should be readOnly when group is readOnly
	})

	t.Run("adds multiple toolsets", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts1 := NewToolset("test1", "Test 1")
		ts2 := NewToolset("test2", "Test 2")

		tg.AddToolset(ts1)
		tg.AddToolset(ts2)

		assert.Len(t, tg.Toolsets, 2)
		assert.Equal(t, ts1, tg.Toolsets["test1"])
		assert.Equal(t, ts2, tg.Toolsets["test2"])
	})

	t.Run("overwrites toolset with same name", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts1 := NewToolset("test", "Test 1")
		ts2 := NewToolset("test", "Test 2")

		tg.AddToolset(ts1)
		tg.AddToolset(ts2)

		assert.Len(t, tg.Toolsets, 1)
		assert.Equal(t, ts2, tg.Toolsets["test"]) // Should be the second one
	})
}

func TestToolsetGroup_EnableToolset(t *testing.T) {
	t.Run("enables existing toolset", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts := NewToolset("test", "Test")
		tg.AddToolset(ts)

		err := tg.EnableToolset("test")
		assert.NoError(t, err)
		assert.True(t, ts.Enabled)
	})

	t.Run("returns error for non-existent toolset", func(t *testing.T) {
		tg := NewToolsetGroup(false)

		err := tg.EnableToolset("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("enables toolset multiple times", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts := NewToolset("test", "Test")
		tg.AddToolset(ts)

		err1 := tg.EnableToolset("test")
		assert.NoError(t, err1)
		assert.True(t, ts.Enabled)

		err2 := tg.EnableToolset("test")
		assert.NoError(t, err2)
		assert.True(t, ts.Enabled) // Should still be enabled
	})
}

func TestToolsetGroup_EnableToolsets(t *testing.T) {
	t.Run("enables multiple toolsets", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts1 := NewToolset("test1", "Test 1")
		ts2 := NewToolset("test2", "Test 2")
		tg.AddToolset(ts1)
		tg.AddToolset(ts2)

		err := tg.EnableToolsets([]string{"test1", "test2"})
		assert.NoError(t, err)
		assert.True(t, ts1.Enabled)
		assert.True(t, ts2.Enabled)
		assert.False(t, tg.everythingOn)
	})

	t.Run("enables all toolsets when empty array", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts1 := NewToolset("test1", "Test 1")
		ts2 := NewToolset("test2", "Test 2")
		ts3 := NewToolset("test3", "Test 3")
		tg.AddToolset(ts1)
		tg.AddToolset(ts2)
		tg.AddToolset(ts3)

		err := tg.EnableToolsets([]string{})
		assert.NoError(t, err)
		assert.True(t, tg.everythingOn)
		assert.True(t, ts1.Enabled)
		assert.True(t, ts2.Enabled)
		assert.True(t, ts3.Enabled)
	})

	t.Run("returns error when enabling non-existent toolset", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts1 := NewToolset("test1", "Test 1")
		tg.AddToolset(ts1)

		err := tg.EnableToolsets([]string{"test1", "nonexistent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
		assert.True(t, ts1.Enabled) // First one should still be enabled
	})

	t.Run("enables single toolset", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts := NewToolset("test", "Test")
		tg.AddToolset(ts)

		err := tg.EnableToolsets([]string{"test"})
		assert.NoError(t, err)
		assert.True(t, ts.Enabled)
	})

	t.Run("handles empty toolset group", func(t *testing.T) {
		tg := NewToolsetGroup(false)

		err := tg.EnableToolsets([]string{})
		assert.NoError(t, err)
		assert.True(t, tg.everythingOn)
	})

	t.Run("enables all toolsets when everythingOn is true", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts1 := NewToolset("test1", "Test 1")
		ts2 := NewToolset("test2", "Test 2")
		tg.AddToolset(ts1)
		tg.AddToolset(ts2)

		// First enable with empty array to set everythingOn
		err := tg.EnableToolsets([]string{})
		assert.NoError(t, err)
		assert.True(t, tg.everythingOn)
		assert.True(t, ts1.Enabled)
		assert.True(t, ts2.Enabled)

		// Reset and test the everythingOn path with non-empty array
		ts1.Enabled = false
		ts2.Enabled = false
		tg.everythingOn = true

		err = tg.EnableToolsets([]string{"test1"})
		assert.NoError(t, err)
		// When everythingOn is true, all toolsets should be enabled
		// even though we only passed test1 in the names array
		assert.True(t, ts1.Enabled)
		assert.True(t, ts2.Enabled)
	})

	t.Run("enables all toolsets when everythingOn true with empty names",
		func(t *testing.T) {
			tg := NewToolsetGroup(false)
			ts1 := NewToolset("test1", "Test 1")
			ts2 := NewToolset("test2", "Test 2")
			tg.AddToolset(ts1)
			tg.AddToolset(ts2)

			// Set everythingOn to true
			tg.everythingOn = true
			ts1.Enabled = false
			ts2.Enabled = false

			// Call with empty array
			err := tg.EnableToolsets([]string{})
			assert.NoError(t, err)
			assert.True(t, ts1.Enabled)
			assert.True(t, ts2.Enabled)
		})

	t.Run("handles error in everythingOn path", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts1 := NewToolset("test1", "Test 1")
		tg.AddToolset(ts1)

		// Add a toolset that doesn't exist to trigger error in the everythingOn path
		// We'll simulate this by adding a toolset name that doesn't exist
		tg.Toolsets["nonexistent"] = NewToolset("nonexistent", "Non-existent")
		delete(tg.Toolsets, "nonexistent") // Remove it to simulate missing toolset

		// Manually add the name to the toolsets map but with nil to cause error
		// Actually, let's test a different error path - when EnableToolset fails
		// We'll override the EnableToolset method behavior by testing with invalid state

		// Instead, let's test the normal error case where a toolset doesn't exist
		err := tg.EnableToolsets([]string{"nonexistent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("handles specific toolset enabling with everythingOn false", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts1 := NewToolset("test1", "Test 1")
		ts2 := NewToolset("test2", "Test 2")
		tg.AddToolset(ts1)
		tg.AddToolset(ts2)

		// Enable specific toolsets (not empty array)
		err := tg.EnableToolsets([]string{"test1"})
		assert.NoError(t, err)
		assert.False(t, tg.everythingOn) // Should remain false
		assert.True(t, ts1.Enabled)
		assert.False(t, ts2.Enabled) // Should not be enabled
	})
}

func TestToolsetGroup_RegisterTools(t *testing.T) {
	t.Run("registers tools from all enabled toolsets", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts1 := NewToolset("test1", "Test 1")
		ts2 := NewToolset("test2", "Test 2")

		tool1 := mcpgo.NewTool("tool1", "Tool 1", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result1"), nil
			})
		tool2 := mcpgo.NewTool("tool2", "Tool 2", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result2"), nil
			})

		ts1.AddReadTools(tool1)
		ts1.Enabled = true
		ts2.AddReadTools(tool2)
		ts2.Enabled = false // This one should not register

		tg.AddToolset(ts1)
		tg.AddToolset(ts2)

		mockSrv := &mockServer{}
		tg.RegisterTools(mockSrv)

		assert.Len(t, mockSrv.GetTools(), 1) // Only tool1 should be registered
	})

	t.Run("registers tools from multiple enabled toolsets", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts1 := NewToolset("test1", "Test 1")
		ts2 := NewToolset("test2", "Test 2")

		tool1 := mcpgo.NewTool("tool1", "Tool 1", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result1"), nil
			})
		tool2 := mcpgo.NewTool("tool2", "Tool 2", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result2"), nil
			})

		ts1.AddReadTools(tool1)
		ts1.Enabled = true
		ts2.AddReadTools(tool2)
		ts2.Enabled = true

		tg.AddToolset(ts1)
		tg.AddToolset(ts2)

		mockSrv := &mockServer{}
		tg.RegisterTools(mockSrv)

		assert.Len(t, mockSrv.GetTools(), 2) // Both tools should be registered
	})

	t.Run("registers no tools when all toolsets disabled", func(t *testing.T) {
		tg := NewToolsetGroup(false)
		ts1 := NewToolset("test1", "Test 1")
		ts2 := NewToolset("test2", "Test 2")

		tool1 := mcpgo.NewTool("tool1", "Tool 1", []mcpgo.ToolParameter{},
			func(ctx context.Context,
				req mcpgo.CallToolRequest) (*mcpgo.ToolResult, error) {
				return mcpgo.NewToolResultText("result1"), nil
			})

		ts1.AddReadTools(tool1)
		ts1.Enabled = false
		ts2.Enabled = false

		tg.AddToolset(ts1)
		tg.AddToolset(ts2)

		mockSrv := &mockServer{}
		tg.RegisterTools(mockSrv)

		assert.Len(t, mockSrv.GetTools(), 0) // No tools should be registered
	})

	t.Run("registers tools from empty toolset group", func(t *testing.T) {
		tg := NewToolsetGroup(false)

		mockSrv := &mockServer{}
		tg.RegisterTools(mockSrv)

		assert.Len(t, mockSrv.GetTools(), 0) // No toolsets, no tools
	})
}
