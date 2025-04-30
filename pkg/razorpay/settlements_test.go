package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_FetchSettlement(t *testing.T) {
	fetchSettlementPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.SETTLEMENT_URL,
	)

	settlementResp := map[string]interface{}{
		"id":         "setl_FNj7g2YS5J67Rz",
		"entity":     "settlement",
		"amount":     float64(9973635),
		"status":     "processed",
		"fees":       float64(471),
		"tax":        float64(72),
		"utr":        "1568176198",
		"created_at": float64(1568176198),
	}

	settlementNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "settlement not found",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful settlement fetch",
			Request: map[string]interface{}{
				"settlement_id": "setl_FNj7g2YS5J67Rz",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchSettlementPathFmt, "setl_FNj7g2YS5J67Rz"),
						Method:   "GET",
						Response: settlementResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: settlementResp,
		},
		{
			Name: "settlement not found",
			Request: map[string]interface{}{
				"settlement_id": "setl_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchSettlementPathFmt, "setl_invalid"),
						Method:   "GET",
						Response: settlementNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching settlement failed: settlement not found",
		},
		{
			Name:           "missing settlement_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: settlement_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchSettlement, "Settlement")
		})
	}
}
