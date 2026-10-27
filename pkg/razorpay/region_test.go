package razorpay

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/region"
)

func builtinProfile(t *testing.T, code string) region.Profile {
	t.Helper()
	profiles, err := region.Load("")
	require.NoError(t, err)
	p, ok := profiles[code]
	require.True(t, ok, "no built-in profile %s", code)
	return p
}

func sgProfile(t *testing.T) region.Profile {
	return builtinProfile(t, "SG")
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

	for _, code := range []string{"SG", "MY", "US"} {
		t.Run(code+" registers exactly its profile tools", func(t *testing.T) {
			p := builtinProfile(t, code)
			srv, err := NewRzpMcpServerForRegion(obs, client, p, nil, false)
			require.NoError(t, err)

			want := append([]string{}, p.Tools...)
			sort.Strings(want)
			assert.Equal(t, want, toolNames(listedTools(t, srv)))
		})

		t.Run(code+" tool list never mentions paise, rupee or INR",
			func(t *testing.T) {
				srv, err := NewRzpMcpServerForRegion(
					obs, client, builtinProfile(t, code), nil, false)
				require.NoError(t, err)

				for name, tool := range listedTools(t, srv) {
					for _, banned := range []string{"paise", "₹", "INR"} {
						assert.NotContains(t, tool, banned, "tool %s", name)
					}
				}
			})
	}

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

func TestNewRzpMcpServerForRegion_AmountMinimum(t *testing.T) {
	srv, err := NewRzpMcpServerForRegion(CreateTestObservability(),
		rzpsdk.NewClient("key", "secret"), sgProfile(t), nil, false)
	require.NoError(t, err)

	var order struct {
		InputSchema struct {
			Properties map[string]struct {
				Minimum float64 `json:"minimum"`
			} `json:"properties"`
		} `json:"inputSchema"`
	}
	require.NoError(t,
		json.Unmarshal([]byte(listedTools(t, srv)["create_order"]), &order))

	// SGD minimum is 50, so the schema must not impose INR's 100.
	assert.Equal(t, float64(1), order.InputSchema.Properties["amount"].Minimum)
}

// recordPaymentCreate serves the payment create endpoints and records the
// path and body of the last request.
func recordPaymentCreate(t *testing.T) (*rzpsdk.Client, *string, *string) {
	t.Helper()
	var gotPath, gotBody string
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			gotPath, gotBody = r.URL.Path, string(body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"razorpay_payment_id": "pay_123"}`))
		}))
	t.Cleanup(server.Close)

	client := rzpsdk.NewClient("key", "secret")
	client.Request.BaseURL = server.URL
	return client, &gotPath, &gotBody
}

func callInitiatePayment(
	t *testing.T,
	ctx context.Context,
	client *rzpsdk.Client,
	args map[string]interface{},
) {
	t.Helper()
	tool := InitiatePayment(CreateTestObservability(), client)
	_, err := tool.GetHandler()(ctx, createMCPRequest(args))
	require.NoError(t, err)
}

func TestInitiatePayment_DefaultCurrency(t *testing.T) {
	args := map[string]interface{}{"amount": 5000, "order_id": "order_123"}

	t.Run("no profile in context defaults to INR", func(t *testing.T) {
		client, _, body := recordPaymentCreate(t)
		callInitiatePayment(t, context.Background(), client, args)
		assert.Contains(t, *body, `"currency":"INR"`)
	})

	t.Run("SG profile defaults to SGD", func(t *testing.T) {
		client, _, body := recordPaymentCreate(t)
		ctx := region.WithProfile(context.Background(), sgProfile(t))
		callInitiatePayment(t, ctx, client, args)
		assert.Contains(t, *body, `"currency":"SGD"`)
	})

	t.Run("explicit currency wins over the region default", func(t *testing.T) {
		client, _, body := recordPaymentCreate(t)
		ctx := region.WithProfile(context.Background(), sgProfile(t))
		callInitiatePayment(t, ctx, client, map[string]interface{}{
			"amount": 5000, "order_id": "order_123", "currency": "USD",
		})
		assert.Contains(t, *body, `"currency":"USD"`)
	})
}

func TestInitiatePayment_RecurringSwitchUnchanged(t *testing.T) {
	recurring := map[string]interface{}{
		"amount": 5000, "order_id": "order_123",
		"token": "token_123", "recurring": true,
	}

	t.Run("INR recurring uses the JSON API", func(t *testing.T) {
		client, path, _ := recordPaymentCreate(t)
		callInitiatePayment(t, context.Background(), client, recurring)
		assert.Equal(t, "/v1/payments/create/json", *path)
	})

	t.Run("SGD recurring uses the recurring API", func(t *testing.T) {
		client, path, _ := recordPaymentCreate(t)
		ctx := region.WithProfile(context.Background(), sgProfile(t))
		callInitiatePayment(t, ctx, client, recurring)
		assert.Equal(t, "/v1/payments/create/recurring", *path)
	})
}

func TestRegionContactExample(t *testing.T) {
	dial := map[string]string{"IN": "+91", "SG": "+65", "MY": "+60", "US": "+1"}
	for code, want := range dial {
		t.Run(code+" shows only its own dial code", func(t *testing.T) {
			srv, err := NewRzpMcpServerForRegion(CreateTestObservability(),
				rzpsdk.NewClient("key", "secret"), builtinProfile(t, code), nil, false)
			require.NoError(t, err)

			var tool struct {
				InputSchema struct {
					Properties map[string]struct {
						Description string `json:"description"`
					} `json:"properties"`
				} `json:"inputSchema"`
			}
			require.NoError(t, json.Unmarshal(
				[]byte(listedTools(t, srv)["fetch_tokens"]), &tool))
			desc := tool.InputSchema.Properties["contact"].Description

			assert.Contains(t, desc, want)
			for other, d := range dial {
				if other != code {
					assert.NotContains(t, desc, d)
				}
			}
		})
	}
}

func TestIndiaToolTextHasNoPaise(t *testing.T) {
	srv, err := NewRzpMcpServerForRegion(CreateTestObservability(),
		rzpsdk.NewClient("key", "secret"), region.India(), nil, false)
	require.NoError(t, err)

	for name, tool := range listedTools(t, srv) {
		assert.NotContains(t, tool, "paise", "tool %s", name)
	}
}
