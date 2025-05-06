package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_FetchPayout(t *testing.T) {
	fetchPayoutPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.PAYOUT_URL,
	)

	successfulPayoutResp := map[string]interface{}{
		"id":     "pout_123",
		"entity": "payout",
		"fund_account": map[string]interface{}{
			"id":     "fa_123",
			"entity": "fund_account",
		},
		"amount":       float64(100000),
		"currency":     "INR",
		"notes":        map[string]interface{}{},
		"fees":         float64(0),
		"tax":          float64(0),
		"utr":          "123456789012345",
		"mode":         "IMPS",
		"purpose":      "payout",
		"processed_at": float64(1704067200),
		"created_at":   float64(1704067200),
		"updated_at":   float64(1704067200),
		"status":       "processed",
	}

	payoutNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "payout not found",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful fetch",
			Request: map[string]interface{}{
				"payout_id": "pout_123",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchPayoutPathFmt, "pout_123"),
						Method:   "GET",
						Response: successfulPayoutResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulPayoutResp,
		},
		{
			Name: "payout not found",
			Request: map[string]interface{}{
				"payout_id": "pout_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							fetchPayoutPathFmt,
							"pout_invalid",
						),
						Method:   "GET",
						Response: payoutNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching payout failed: payout not found",
		},
		{
			Name:           "missing payout_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payout_id",
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				// Missing payout_id parameter
				"non_existent_param": 12345, // Additional parameter
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payout_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchPayout, "Payout")
		})
	}
}

func Test_FetchAllPayouts(t *testing.T) {
	fetchAllPayoutsPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.PAYOUT_URL,
	)

	successfulPayoutsResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":     "pout_1",
				"entity": "payout",
				"fund_account": map[string]interface{}{
					"id":     "fa_1",
					"entity": "fund_account",
				},
				"amount":       float64(100000),
				"currency":     "INR",
				"notes":        map[string]interface{}{},
				"fees":         float64(0),
				"tax":          float64(0),
				"utr":          "123456789012345",
				"mode":         "IMPS",
				"purpose":      "payout",
				"processed_at": float64(1704067200),
				"created_at":   float64(1704067200),
				"updated_at":   float64(1704067200),
				"status":       "processed",
			},
			map[string]interface{}{
				"id":     "pout_2",
				"entity": "payout",
				"fund_account": map[string]interface{}{
					"id":     "fa_2",
					"entity": "fund_account",
				},
				"amount":       float64(200000),
				"currency":     "INR",
				"notes":        map[string]interface{}{},
				"fees":         float64(0),
				"tax":          float64(0),
				"utr":          "123456789012346",
				"mode":         "IMPS",
				"purpose":      "payout",
				"processed_at": float64(1704067200),
				"created_at":   float64(1704067200),
				"updated_at":   float64(1704067200),
				"status":       "pending",
			},
		},
	}

	invalidAccountErrorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Invalid account number",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful fetch with pagination",
			Request: map[string]interface{}{
				"account_number": "409002173420",
				"count":          float64(10),
				"skip":           float64(0),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllPayoutsPath,
						Method:   "GET",
						Response: successfulPayoutsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulPayoutsResp,
		},
		{
			Name: "successful fetch without pagination",
			Request: map[string]interface{}{
				"account_number": "409002173420",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllPayoutsPath,
						Method:   "GET",
						Response: successfulPayoutsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulPayoutsResp,
		},
		{
			Name: "invalid account number",
			Request: map[string]interface{}{
				"account_number": "invalid_account",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllPayoutsPath,
						Method:   "GET",
						Response: invalidAccountErrorResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching payouts failed: Invalid account number",
		},
		{
			Name: "missing account_number parameter",
			Request: map[string]interface{}{
				"count": float64(10),
				"skip":  float64(0),
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: account_number",
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				// Missing account_number parameter
				"count": "10", // Wrong type for count
				"skip":  "0",  // Wrong type for skip
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:\n- " +
				"missing required parameter: account_number\n- " +
				"invalid parameter type: count\n- " +
				"invalid parameter type: skip",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchAllPayouts, "Payouts")
		})
	}
}
