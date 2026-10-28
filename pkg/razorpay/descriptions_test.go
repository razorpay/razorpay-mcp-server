package razorpay

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/server"
	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// knownBrokenSubstrings are regression targets from missing-space
// concatenation.
var knownBrokenSubstrings = []string{
	"retrievedThe",
	"fetched(ID",
	"linksIf",
	"done.For",
	"10.Maximum",
	"pagination,in",
	"fetch.ID",
}

// s1FixedDescriptionChecks verify S1 tools render with proper word boundaries.
var s1FixedDescriptionChecks = []struct {
	toolName  string
	paramName string // empty for tool-level description
	want      string
}{
	{
		toolName:  "fetch_qr_code",
		paramName: "qr_code_id",
		want:      "retrieved The",
	},
	{
		toolName:  "fetch_qr_codes_by_payment_id",
		paramName: "payment_id",
		want:      "payment The",
	},
	{
		toolName:  "fetch_payments_for_qr_code",
		paramName: "qr_code_id",
		want:      "for The",
	},
	{
		toolName:  "close_qr_code",
		paramName: "qr_code_id",
		want:      "closed The",
	},
	{
		toolName:  "fetch_payment_link",
		paramName: "payment_link_id",
		want:      "fetched (",
	},
	{
		toolName:  "fetch_all_payment_links",
		paramName: "upi_link",
		want:      "standard links If",
	},
	{
		toolName: "fetch_all_payment_links",
		want:     "reference ID. You",
	},
	{
		toolName:  "fetch_all_payouts",
		paramName: "account_number",
		want:      "done. For",
	},
	{
		toolName:  "fetch_all_payouts",
		paramName: "count",
		want:      "10. Maximum",
	},
	{
		toolName:  "fetch_all_payouts",
		paramName: "count",
		want:      "pagination, in",
	},
	{
		toolName:  "fetch_all_payouts",
		paramName: "skip",
		want:      "0. This",
	},
	{
		toolName:  "fetch_settlement_with_id",
		paramName: "settlement_id",
		want:      "fetch. Starts",
	},
}

// descriptionConcatFiles are the S1 scope files checked for trailing spaces.
var descriptionConcatFiles = []string{
	"qr_codes.go",
	"payment_links.go",
	"payouts.go",
	"settlements.go",
}

// missingConcatSpacePattern matches `"segment"+` without a trailing
// space in segment.
var missingConcatSpacePattern = regexp.MustCompile(`"[^"]*[^ +\\]"\\s*\\+`)

func TestToolDescriptions_NoBrokenConcatenation(t *testing.T) {
	for _, desc := range collectRegisteredDescriptions(t) {
		for _, bad := range knownBrokenSubstrings {
			assert.NotContains(
				t,
				desc,
				bad,
				"description contains known broken substring %q: %s",
				bad,
				desc,
			)
		}
	}
}

func TestToolDescriptions_S1FixedPhrases(t *testing.T) {
	tools := listRegisteredTools(t)

	for _, check := range s1FixedDescriptionChecks {
		desc := toolDescription(tools, check.toolName, check.paramName)
		require.NotEmpty(t, desc, "missing description for %s/%s",
			check.toolName, check.paramName)
		assert.Contains(t, desc, check.want)
	}
}

func TestDescriptionSource_ConcatenationHasTrailingSpaces(t *testing.T) {
	for _, fileName := range descriptionConcatFiles {
		lines, err := readSourceLines(fileName)
		require.NoError(t, err, "read %s", fileName)

		for lineNum, line := range lines {
			if !strings.Contains(line, "Description(") &&
				!strings.Contains(line, "NewTool(") {
				continue
			}
			if !missingConcatSpacePattern.MatchString(line) {
				continue
			}

			t.Errorf(
				"%s:%d: description segment must end with a space before '+': %s",
				fileName,
				lineNum+1,
				strings.TrimSpace(line),
			)
		}
	}
}

func collectRegisteredDescriptions(t *testing.T) []string {
	tools := listRegisteredTools(t)
	descs := make([]string, 0, len(tools)*4)

	for _, st := range tools {
		descs = append(descs, st.Tool.Description)
		descs = append(descs, paramDescriptions(st)...)
	}

	return descs
}

func listRegisteredTools(t *testing.T) map[string]*server.ServerTool {
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

func toolDescription(
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

func paramDescriptions(st *server.ServerTool) []string {
	descs := make([]string, 0, len(st.Tool.InputSchema.Properties))
	for _, prop := range st.Tool.InputSchema.Properties {
		propMap, ok := prop.(map[string]any)
		if !ok {
			continue
		}

		desc, ok := propMap["description"].(string)
		if !ok || desc == "" {
			continue
		}

		descs = append(descs, desc)
	}

	return descs
}

func readSourceLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}
