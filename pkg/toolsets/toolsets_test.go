package toolsets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAddToolset tests adding a toolset to a group
func TestAddToolset(t *testing.T) {
	tg := NewToolsetGroup(false)
	ts := NewToolset("test", "description")

	tg.AddToolset(ts)

	assert.Contains(t, tg.Toolsets, "test")
	assert.Equal(t, ts, tg.Toolsets["test"])
	assert.False(t, ts.readOnly)

	// Test readOnly propagation
	tgReadOnly := NewToolsetGroup(true)
	tsForReadOnly := NewToolset("test2", "description")

	tgReadOnly.AddToolset(tsForReadOnly)

	assert.True(t, tsForReadOnly.readOnly)
}

// TestEnableToolset tests enabling a specific toolset
func TestEnableToolset(t *testing.T) {
	tg := NewToolsetGroup(false)
	ts := NewToolset("test", "description")
	tg.AddToolset(ts)

	// Test enabling existing toolset
	err := tg.EnableToolset("test")

	assert.NoError(t, err)
	assert.True(t, ts.Enabled)

	// Test enabling non-existent toolset
	err = tg.EnableToolset("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

// TestEnableToolsets tests enabling multiple toolsets
func TestEnableToolsets(t *testing.T) {
	tests := []struct {
		name          string
		toolsets      []string
		toEnable      []string
		expectError   bool
		expectEnabled map[string]bool
	}{
		{
			name:          "Empty list enables all",
			toolsets:      []string{"ts1", "ts2", "ts3"},
			toEnable:      []string{},
			expectError:   false,
			expectEnabled: map[string]bool{"ts1": true, "ts2": true, "ts3": true},
		},
		{
			name:          "Enable specific toolsets",
			toolsets:      []string{"ts1", "ts2", "ts3"},
			toEnable:      []string{"ts1", "ts3"},
			expectError:   false,
			expectEnabled: map[string]bool{"ts1": true, "ts2": false, "ts3": true},
		},
		{
			name:          "Error on non-existent toolset",
			toolsets:      []string{"ts1", "ts2"},
			toEnable:      []string{"ts1", "nonexistent"},
			expectError:   true,
			expectEnabled: map[string]bool{"ts1": true, "ts2": false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tg := NewToolsetGroup(false)

			// Create and add toolsets
			for _, name := range tc.toolsets {
				ts := NewToolset(name, "description")
				tg.AddToolset(ts)
			}

			// Enable toolsets
			err := tg.EnableToolsets(tc.toEnable)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify enabled state
			for name, shouldBeEnabled := range tc.expectEnabled {
				assert.Equal(t, shouldBeEnabled, tg.Toolsets[name].Enabled)
			}
		})
	}
}
