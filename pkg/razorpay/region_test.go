package razorpay

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/region"
)

func sgProfile(t *testing.T) region.Profile {
	t.Helper()
	profiles, err := region.Load("")
	require.NoError(t, err)
	return profiles["SG"]
}

// listedTools returns the tools/list payload per tool name, as JSON text.
func listedTools(t *testing.T, srv mcpgo.Server) map[string]string {
	t.Helper()
	impl, ok := srv.(*mcpgo.Mark3labsImpl)
	require.True(t, ok)

	out := map[string]string{}
	for name, st := range impl.McpServer.ListTools() {
		data, err := json.Marshal(st.Tool)
		require.NoError(t, err)
		out[name] = string(data)
	}
	return out
}

func toolNames(tools map[string]string) []string {
	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func TestNewRzpMcpServerForRegion(t *testing.T) {
	obs := CreateTestObservability()
	client := rzpsdk.NewClient("key", "secret")

	t.Run("IN registers every tool", func(t *testing.T) {
		srv, err := NewRzpMcpServerForRegion(
			obs, client, region.India(), nil, false)
		require.NoError(t, err)
		assert.Len(t, listedTools(t, srv), 45)
	})

	t.Run("SG registers exactly its profile tools", func(t *testing.T) {
		sg := sgProfile(t)
		srv, err := NewRzpMcpServerForRegion(obs, client, sg, nil, false)
		require.NoError(t, err)

		want := append([]string{}, sg.Tools...)
		sort.Strings(want)
		assert.Equal(t, want, toolNames(listedTools(t, srv)))
	})

	t.Run("no placeholder is left for IN, SG or the default server",
		func(t *testing.T) {
			in, err := NewRzpMcpServerForRegion(
				obs, client, region.India(), nil, false)
			require.NoError(t, err)
			sg, err := NewRzpMcpServerForRegion(
				obs, client, sgProfile(t), nil, false)
			require.NoError(t, err)
			def, err := NewRzpMcpServer(obs, client, nil, false)
			require.NoError(t, err)

			for _, srv := range []mcpgo.Server{in, sg, def} {
				for name, tool := range listedTools(t, srv) {
					assert.NotContains(t, tool, "{{", "tool %s", name)
				}
			}
		})

	t.Run("requesting a toolset empty in the region fails", func(t *testing.T) {
		_, err := NewRzpMcpServerForRegion(
			obs, client, sgProfile(t), []string{"payouts"}, false)
		assert.ErrorContains(t, err,
			`toolset "payouts" has no tools available in region SG`)
	})

	t.Run("profile naming an unknown tool fails", func(t *testing.T) {
		bad := region.Profile{
			CountryCode: "SG",
			APIBaseURL:  "https://api-sg.razorpay.com",
			Tools:       []string{"fetch_payment", "fetch_paymnt"},
		}
		_, err := NewRzpMcpServerForRegion(obs, client, bad, nil, false)
		assert.ErrorContains(t, err, `region SG: unknown tool "fetch_paymnt"`)
	})
}
