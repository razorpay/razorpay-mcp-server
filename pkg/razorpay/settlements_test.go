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

func Test_FetchSettlementRecon(t *testing.T) {
	fetchSettlementReconPath := fmt.Sprintf(
		"/%s%s/recon/combined",
		constants.VERSION_V1,
		constants.SETTLEMENT_URL,
	)

	// Sample response for successful fetch
	reconReportResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(1),
		"items": []interface{}{
			map[string]interface{}{
				"entity_id":     "setl_FNj7g2YS5J67Rz",
				"type":          "settlement",
				"debit":         float64(0),
				"credit":        float64(9973635),
				"amount":        float64(9973635),
				"fee":           float64(471),
				"tax":           float64(72),
				"settlement_id": "setl_FNj7g2YS5J67Rz",
				"created_at":    float64(1568176198),
			},
		},
	}

	// Error response when required parameters are missing
	invalidParamsResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The year, month parameters are required",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful reconciliation report fetch with required params",
			Request: map[string]interface{}{
				"year":  "2023",
				"month": "09",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchSettlementReconPath,
						Method:   "GET",
						Response: reconReportResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: reconReportResp,
		},
		{
			Name: "successful reconciliation report fetch with all params",
			Request: map[string]interface{}{
				"year":  "2023",
				"month": "09",
				"day":   "15",
				"count": "20",
				"skip":  "0",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchSettlementReconPath,
						Method:   "GET",
						Response: reconReportResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: reconReportResp,
		},
		{
			Name: "missing required parameters",
			Request: map[string]interface{}{
				"day": "15", // Missing year and month
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchSettlementReconPath,
						Method:   "GET",
						Response: invalidParamsResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: year",
		},
		{
			Name:           "empty request",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: year",
		},
		{
			Name: "invalid year format",
			Request: map[string]interface{}{
				"year":  "20", // Invalid year format (not 4 digits)
				"month": "09",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchSettlementReconPath,
						Method:   "GET",
						Response: invalidParamsResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching settlement reconciliation report failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchSettlementRecon, "Settlement Reconciliation Report")
		})
	}
}
