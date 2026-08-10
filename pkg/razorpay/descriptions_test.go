package razorpay

import (
	"testing"

	"github.com/mark3labs/mcp-go/server"
	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// knownGrammarIssues are common description mistakes to guard against.
var knownGrammarIssues = []string{
	"it's ID",
	"in paisa",
}

// s2DescriptionChecks verify grammar and spelling in affected tool
// descriptions.
var s2DescriptionChecks = []struct {
	toolName string
	want     string
	notWant  string
}{
	{
		toolName: "fetch_qr_code",
		want:     "using its ID",
		notWant:  "it's ID",
	},
	{
		toolName: "fetch_payment_link",
		want:     "using its ID",
		notWant:  "it's ID",
	},
	{
		toolName: "fetch_payment",
		want:     "in paise",
		notWant:  "paisa",
	},
}

func TestToolDescriptions_NoKnownGrammarIssues(t *testing.T) {
	for _, desc := range collectToolDescriptions(t) {
		for _, bad := range knownGrammarIssues {
			assert.NotContains(
				t,
				desc,
				bad,
				"description contains known grammar issue %q: %s",
				bad,
				desc,
			)
		}
	}
}

func TestToolDescriptions_S2GrammarAndSpelling(t *testing.T) {
	tools := listRegisteredToolsForDescriptions(t)

	for _, check := range s2DescriptionChecks {
		desc := toolDescriptionFromRegistry(tools, check.toolName, "")
		require.NotEmpty(t, desc, "missing description for %s", check.toolName)
		assert.Contains(t, desc, check.want)
		assert.NotContains(t, desc, check.notWant)
	}
}

func collectToolDescriptions(t *testing.T) []string {
	tools := listRegisteredToolsForDescriptions(t)
	descs := make([]string, 0, len(tools))

	for _, st := range tools {
		descs = append(descs, st.Tool.Description)
	}

	return descs
}

func listRegisteredToolsForDescriptions(
	t *testing.T,
) map[string]*server.ServerTool {
	t.Helper()

	obs := CreateTestObservability()
	client := &rzpsdk.Client{}
	group, err := NewToolSets(obs, client, []string{}, false)
	require.NoError(t, err)

	srv := mcpgo.NewMcpServer(
		"test",
		"1.0.0",
		mcpgo.WithToolCapabilities(true),
	)
	group.RegisterTools(srv)

	tools := srv.McpServer.ListTools()
	require.NotEmpty(t, tools)

	return tools
}

func toolDescriptionFromRegistry(
	tools map[string]*server.ServerTool,
	toolName,
	paramName string,
) string {
	st, ok := tools[toolName]
	if !ok {
		return ""
	}

	if paramName == "" {
		return st.Tool.Description
	}

	prop, ok := st.Tool.InputSchema.Properties[paramName]
	if !ok {
		return ""
	}

	propMap, ok := prop.(map[string]any)
	if !ok {
		return ""
	}

	desc, _ := propMap["description"].(string)
	return desc
}
