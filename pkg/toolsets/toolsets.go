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
	for _, tool := range t.readTools {
		if t.allows(tool) {
			tool.SetReadOnly(true)
			s.AddTools(tool)
		}
	}
	if !t.readOnly {
		for _, tool := range t.writeTools {
			if t.allows(tool) {
				tool.SetReadOnly(false)
				s.AddTools(tool)
			}
		}
	}
}

// allows reports whether the tool passes the toolset's filter. Tools are
// allowed when no filter is set or when the tool has no name to filter on.
func (t *Toolset) allows(tool mcpgo.Tool) bool {
	if t.allow == nil {
		return true
	}
	named, ok := tool.(interface{ GetName() string })
	return !ok || t.allow(named.GetName())
}

// allowedToolCount returns how many tools RegisterTools would register,
// ignoring whether the toolset is enabled.
func (t *Toolset) allowedToolCount() int {
	count := 0
	for _, tool := range t.readTools {
		if t.allows(tool) {
			count++
		}
	}
	if !t.readOnly {
		for _, tool := range t.writeTools {
			if t.allows(tool) {
				count++
			}
		}
	}
	return count
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

// AllowedToolCount returns how many tools the named toolset would register
// after the read-only setting and the tool filter are applied.
func (tg *ToolsetGroup) AllowedToolCount(name string) int {
	ts, ok := tg.Toolsets[name]
	if !ok {
		return 0
	}
	return ts.allowedToolCount()
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
