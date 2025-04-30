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

func Test_FetchAllSettlements(t *testing.T) {
	fetchAllSettlementsPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.SETTLEMENT_URL,
	)

	// Sample response for successful fetch
	settlementListResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":         "setl_DGlQ1Rj8os78Ec",
				"entity":     "settlement",
				"amount":     float64(9973635),
				"status":     "processed",
				"fees":       float64(0),
				"tax":        float64(0),
				"utr":        "1568176960vxp0rj",
				"created_at": float64(1568176960),
			},
			map[string]interface{}{
				"id":         "setl_4xbSwsPABDJ8oK",
				"entity":     "settlement",
				"amount":     float64(50000),
				"status":     "processed",
				"fees":       float64(0),
				"tax":        float64(0),
				"utr":        "RZRP173069230702",
				"created_at": float64(1509622306),
			},
		},
	}

	// Error response when parameters are invalid
	invalidParamsResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "from must be between 946684800 and 4765046400",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful settlements fetch with all parameters",
			Request: map[string]interface{}{
				"from":  float64(1500000000),
				"to":    float64(1600000000),
				"count": float64(20),
				"skip":  float64(0),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllSettlementsPath,
						Method:   "GET",
						Response: settlementListResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: settlementListResp,
		},
		{
			Name: "settlements fetch with invalid timestamp",
			Request: map[string]interface{}{
				"from": float64(900000000), // Invalid timestamp (too early)
				"to":   float64(1600000000),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllSettlementsPath,
						Method:   "GET",
						Response: invalidParamsResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "fetching settlements failed: from must be between " +
				"946684800 and 4765046400",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchAllSettlements, "Settlements List")
		})
	}
}

func Test_CreateInstantSettlement(t *testing.T) {
	createInstantSettlementPath := fmt.Sprintf(
		"/%s%s/ondemand",
		constants.VERSION_V1,
		constants.SETTLEMENT_URL,
	)

	// Successful response with all parameters
	successfulSettlementResp := map[string]interface{}{
		"id":                  "setlod_FNj7g2YS5J67Rz",
		"entity":              "settlement.ondemand",
		"amount_requested":    float64(200000),
		"amount_settled":      float64(0),
		"amount_pending":      float64(199410),
		"amount_reversed":     float64(0),
		"fees":                float64(590),
		"tax":                 float64(90),
		"currency":            "INR",
		"settle_full_balance": false,
		"status":              "initiated",
		"description":         "Need this to make vendor payments.",
		"notes": map[string]interface{}{
			"notes_key_1": "Tea, Earl Grey, Hot",
			"notes_key_2": "Tea, Earl Grey… decaf.",
		},
		"created_at": float64(1596771429),
	}

	// Error response for insufficient amount
	insufficientAmountResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Minimum amount that can be settled is ₹ 1.",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful settlement creation with all parameters",
			Request: map[string]interface{}{
				"amount":              float64(200000),
				"settle_full_balance": false,
				"description":         "Need this to make vendor payments.",
				"notes": map[string]interface{}{
					"notes_key_1": "Tea, Earl Grey, Hot",
					"notes_key_2": "Tea, Earl Grey… decaf.",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createInstantSettlementPath,
						Method:   "POST",
						Response: successfulSettlementResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulSettlementResp,
		},
		{
			Name: "settlement creation with insufficient amount",
			Request: map[string]interface{}{
				"amount": float64(10), // Less than minimum
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createInstantSettlementPath,
						Method:   "POST",
						Response: insufficientAmountResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "creating instant settlement failed: Minimum amount that " +
				"can be settled is ₹ 1.",
		},
		{
			Name:           "missing amount parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: amount",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CreateInstantSettlement, "Instant Settlement")
		})
	}
}

func Test_FetchAllInstantSettlements(t *testing.T) {
	fetchAllInstantSettlementsPath := fmt.Sprintf(
		"/%s%s/ondemand",
		constants.VERSION_V1,
		constants.SETTLEMENT_URL,
	)

	// Sample response for successful fetch
	instantSettlementListResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":                  "setlod_FNj7g2YS5J67Rz",
				"entity":              "settlement.ondemand",
				"amount_requested":    float64(200000),
				"amount_settled":      float64(199410),
				"amount_pending":      float64(0),
				"amount_reversed":     float64(0),
				"fees":                float64(590),
				"tax":                 float64(90),
				"currency":            "INR",
				"settle_full_balance": false,
				"status":              "processed",
				"description":         "Need this to make vendor payments.",
				"notes": map[string]interface{}{
					"notes_key_1": "Tea, Earl Grey, Hot",
					"notes_key_2": "Tea, Earl Grey… decaf.",
				},
				"created_at": float64(1596771429),
			},
			map[string]interface{}{
				"id":                  "setlod_FJOp0jOWlalIvt",
				"entity":              "settlement.ondemand",
				"amount_requested":    float64(300000),
				"amount_settled":      float64(299114),
				"amount_pending":      float64(0),
				"amount_reversed":     float64(0),
				"fees":                float64(886),
				"tax":                 float64(136),
				"currency":            "INR",
				"settle_full_balance": false,
				"status":              "processed",
				"description":         "Need this to buy stock.",
				"notes": map[string]interface{}{
					"notes_key_1": "Tea, Earl Grey, Hot",
					"notes_key_2": "Tea, Earl Grey… decaf.",
				},
				"created_at": float64(1595826576),
			},
		},
	}

	// Error response when parameters are invalid
	invalidParamsResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "from must be between 946684800 and 4765046400",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name:    "successful instant settlements fetch with all parameters",
			Request: map[string]interface{}{},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllInstantSettlementsPath,
						Method:   "GET",
						Response: instantSettlementListResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: instantSettlementListResp,
		},
		{
			Name: "instant settlements fetch with invalid timestamp",
			Request: map[string]interface{}{
				"from": float64(900000000), // Invalid timestamp (too early)
				"to":   float64(1600000000),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllInstantSettlementsPath,
						Method:   "GET",
						Response: invalidParamsResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "fetching instant settlements failed: from must be between " +
				"946684800 and 4765046400",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchAllInstantSettlements, "Instant Settlements List")
		})
	}
}
