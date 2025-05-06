package razorpay

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func TestFetchPayoutByID(t *testing.T) {
	testCases := []RazorpayToolTestCase{
		{
			Name: "successful fetch",
			Request: map[string]interface{}{
				"payout_id": "pout_123",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     "/v1/payouts/pout_123",
						Method:   "GET",
						Response: map[string]interface{}{"id": "pout_123", "status": "processed"},
					},
				)
			},
			ExpectedResult: map[string]interface{}{
				"id":     "pout_123",
				"status": "processed",
			},
		},
		{
			Name:    "missing payout_id",
			Request: map[string]interface{}{
				// No payout_id provided
			},
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payout_id",
		},
		{
			Name: "invalid payout_id type",
			Request: map[string]interface{}{
				"payout_id": 123, // Should be string
			},
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: payout_id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, func(log *slog.Logger, client *rzpsdk.Client) mcpgo.Tool {
				return FetchPayoutByID(log, client)
			}, "payout")
		})
	}
}

func TestFetchAllPayouts(t *testing.T) {
	testCases := []RazorpayToolTestCase{
		{
			Name: "successful fetch with pagination",
			Request: map[string]interface{}{
				"account_number": "acc_123",
				"count":          10,
				"skip":           0,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payouts",
						Method: "GET",
						Response: map[string]interface{}{
							"items": []interface{}{
								map[string]interface{}{"id": "pout_1", "status": "processed"},
								map[string]interface{}{"id": "pout_2", "status": "pending"},
							},
							"count": 2,
						},
					},
				)
			},
			ExpectedResult: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"id": "pout_1", "status": "processed"},
					map[string]interface{}{"id": "pout_2", "status": "pending"},
				},
				"count": float64(2),
			},
		},
		{
			Name: "successful fetch without pagination",
			Request: map[string]interface{}{
				"account_number": "acc_123",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payouts",
						Method: "GET",
						Response: map[string]interface{}{
							"items": []interface{}{
								map[string]interface{}{"id": "pout_1", "status": "processed"},
							},
							"count": 1,
						},
					},
				)
			},
			ExpectedResult: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"id": "pout_1", "status": "processed"},
				},
				"count": float64(1),
			},
		},
		{
			Name: "missing account_number",
			Request: map[string]interface{}{
				"count": 10,
				"skip":  0,
			},
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: account_number",
		},
		{
			Name: "invalid count type",
			Request: map[string]interface{}{
				"account_number": "acc_123",
				"count":          "10", // Should be number
			},
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: count",
		},
		{
			Name: "invalid skip type",
			Request: map[string]interface{}{
				"account_number": "acc_123",
				"skip":           "0", // Should be number
			},
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: skip",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, func(log *slog.Logger, client *rzpsdk.Client) mcpgo.Tool {
				return FetchAllPayouts(log, client)
			}, "payouts")
		})
	}
}
