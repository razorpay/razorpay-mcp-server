package razorpay

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/stretchr/testify/assert"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// Test_PathIDsRejectTraversal checks that an id which would change the REST
// path built by the SDK is rejected before any upstream request is sent.
func Test_PathIDsRejectTraversal(t *testing.T) {
	notes := map[string]interface{}{"k": "v"}

	tools := []struct {
		name    string
		creator func(*observability.Observability, *rzpsdk.Client) mcpgo.Tool
		idParam string
		prefix  string
		extra   map[string]interface{}
	}{
		{"fetch_order", FetchOrder, "order_id", "order_", nil},
		{"fetch_order_payments", FetchOrderPayments, "order_id", "order_", nil},
		{"update_order", UpdateOrder, "order_id", "order_",
			map[string]interface{}{"notes": notes}},
		{"fetch_payment_link", FetchPaymentLink, "payment_link_id", "plink_", nil},
		{"send_payment_link", ResendPaymentLinkNotification, "payment_link_id",
			"plink_", map[string]interface{}{"medium": "sms"}},
		{"update_payment_link", UpdatePaymentLink, "payment_link_id", "plink_",
			map[string]interface{}{"notes": notes}},
		{"fetch_payout", FetchPayout, "payout_id", "pout_", nil},
	}

	payloads := []string{
		"X/cancel#",
		"X/cancel",
		"X?expand[]=payments",
		"X/../../payouts",
		"X%2Fcancel",
		"X\n",
		"X ",
	}

	for _, tool := range tools {
		for _, suffix := range payloads {
			id := tool.prefix + suffix
			t.Run(tool.name+"/"+id, func(t *testing.T) {
				var hits int32
				server := httptest.NewServer(http.HandlerFunc(
					func(w http.ResponseWriter, _ *http.Request) {
						atomic.AddInt32(&hits, 1)
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte(`{}`))
					}))
				defer server.Close()

				client := rzpsdk.NewClient("sample_key", "sample_secret")
				client.Order.Request.BaseURL = server.URL
				client.Order.Request.HTTPClient = server.Client()

				args := map[string]interface{}{tool.idParam: id}
				for k, v := range tool.extra {
					args[k] = v
				}

				handler := tool.creator(CreateTestObservability(), client).
					GetHandler()
				result, err := handler(
					context.Background(), createMCPRequest(args))

				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.IsError)
				assert.Contains(t, result.Text,
					tool.idParam+" has an invalid format")
				assert.Zero(t, atomic.LoadInt32(&hits),
					"no request may reach the API")
			})
		}
	}
}
