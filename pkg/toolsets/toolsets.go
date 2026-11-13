package toolsets

import (
	"fmt"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// Toolset represents a group of related tools
type Toolset struct {
	Name        string
	Description string
	Enabled     bool
	readOnly    bool
	writeTools  []mcpgo.Tool
	readTools   []mcpgo.Tool
	allow       func(toolName string) bool
}

// ToolsetGroup manages multiple toolsets
type ToolsetGroup struct {
	Toolsets     map[string]*Toolset
	everythingOn bool
	readOnly     bool
	allow        func(toolName string) bool
}

// NewToolset creates a new toolset with the given name and description
func NewToolset(name string, description string) *Toolset {
	return &Toolset{
		Name:        name,
		Description: description,
		Enabled:     false,
		readOnly:    false,
	}
}

// NewToolsetGroup creates a new toolset group
func NewToolsetGroup(readOnly bool) *ToolsetGroup {
	return &ToolsetGroup{
		Toolsets:     make(map[string]*Toolset),
		everythingOn: false,
		readOnly:     readOnly,
	}
}

// AddWriteTools adds write tools to the toolset
func (t *Toolset) AddWriteTools(tools ...mcpgo.Tool) *Toolset {
	if !t.readOnly {
		t.writeTools = append(t.writeTools, tools...)
	}
	return t
}

// AddReadTools adds read tools to the toolset
func (t *Toolset) AddReadTools(tools ...mcpgo.Tool) *Toolset {
	t.readTools = append(t.readTools, tools...)
	return t
}

// RegisterTools registers all active tools with the server
func (t *Toolset) RegisterTools(s mcpgo.Server) {
	if !t.Enabled {
		return
	}
	t.forEachRegistrable(func(tool mcpgo.Tool, readOnly bool) {
		tool.SetReadOnly(readOnly)
		s.AddTools(tool)
	})
}

// forEachRegistrable visits each tool RegisterTools would add, enabled or not.
func (t *Toolset) forEachRegistrable(fn func(tool mcpgo.Tool, readOnly bool)) {
	for _, tool := range t.readTools {
		if t.allows(tool) {
			fn(tool, true)
		}
	}
	if t.readOnly {
		return
	}
	for _, tool := range t.writeTools {
		if t.allows(tool) {
			fn(tool, false)
		}
	}
}

// allows reports whether the tool passes the filter.
func (t *Toolset) allows(tool mcpgo.Tool) bool {
	return t.allow == nil || t.allow(tool.GetName())
}

// AddToolset adds a toolset to the group
func (tg *ToolsetGroup) AddToolset(ts *Toolset) {
	if tg.readOnly {
		ts.readOnly = true
	}
	ts.allow = tg.allow
	tg.Toolsets[ts.Name] = ts
}

// SetToolFilter registers only the tools for which allow returns true.
func (tg *ToolsetGroup) SetToolFilter(allow func(toolName string) bool) {
	tg.allow = allow
	for _, ts := range tg.Toolsets {
		ts.allow = allow
	}
}

// AllowedToolCount counts tools the filter allows, whatever the read-only mode.
func (tg *ToolsetGroup) AllowedToolCount(name string) int {
	ts, ok := tg.Toolsets[name]
	if !ok {
		return 0
	}
	count := 0
	for _, tools := range [][]mcpgo.Tool{ts.readTools, ts.writeTools} {
		for _, tool := range tools {
			if ts.allows(tool) {
				count++
			}
		}
	}
	return count
}

// EnableToolset enables a specific toolset
func (tg *ToolsetGroup) EnableToolset(name string) error {
	toolset, exists := tg.Toolsets[name]
	if !exists {
		return fmt.Errorf("toolset %s does not exist", name)
	}
	toolset.Enabled = true
	return nil
}

// EnableToolsets enables multiple toolsets
func (tg *ToolsetGroup) EnableToolsets(names []string) error {
	if len(names) == 0 {
		tg.everythingOn = true
	}

	for _, name := range names {
		err := tg.EnableToolset(name)
		if err != nil {
			return err
		}
	}

	if tg.everythingOn {
		for name := range tg.Toolsets {
			err := tg.EnableToolset(name)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}

// RegisterTools registers all active toolsets with the server
func (tg *ToolsetGroup) RegisterTools(s mcpgo.Server) {
	for _, toolset := range tg.Toolsets {
		toolset.RegisterTools(s)
	}
}

// HiddenTools lists tools in the toolset that the filter removes.
func (tg *ToolsetGroup) HiddenTools(name string) []string {
	ts, ok := tg.Toolsets[name]
	if !ok {
		return nil
	}
	hidden := make([]string, 0, len(ts.readTools)+len(ts.writeTools))
	for _, tools := range [][]mcpgo.Tool{ts.readTools, ts.writeTools} {
		for _, tool := range tools {
			if !ts.allows(tool) {
				hidden = append(hidden, tool.GetName())
			}
		}
	}
	return hidden
}

// AllTools returns every tool in the group, ignoring enablement and filter.
func (tg *ToolsetGroup) AllTools() []mcpgo.Tool {
	n := 0
	for _, ts := range tg.Toolsets {
		n += len(ts.readTools) + len(ts.writeTools)
	}
	tools := make([]mcpgo.Tool, 0, n)
	for _, ts := range tg.Toolsets {
		tools = append(tools, ts.readTools...)
		tools = append(tools, ts.writeTools...)
	}
	return tools
}
