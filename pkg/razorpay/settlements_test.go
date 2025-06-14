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

	settlementReconResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(1),
		"items": []interface{}{
			map[string]interface{}{
				"entity":            "settlement",
				"settlement_id":     "setl_FNj7g2YS5J67Rz",
				"settlement_utr":    "1568176198",
				"amount":            float64(9973635),
				"settlement_type":   "regular",
				"settlement_status": "processed",
				"created_at":        float64(1568176198),
			},
		},
	}

	invalidParamsResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "missing required parameters",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful settlement reconciliation fetch",
			Request: map[string]interface{}{
				"year":  float64(2022),
				"month": float64(10),
				"day":   float64(15),
				"count": float64(10),
				"skip":  float64(0),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchSettlementReconPath,
						Method:   "GET",
						Response: settlementReconResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: settlementReconResp,
		},
		{
			Name: "settlement reconciliation with required params only",
			Request: map[string]interface{}{
				"year":  float64(2022),
				"month": float64(10),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchSettlementReconPath,
						Method:   "GET",
						Response: settlementReconResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: settlementReconResp,
		},
		{
			Name: "settlement reconciliation with invalid params",
			Request: map[string]interface{}{
				"year": float64(2022),
				// missing month parameter
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
			ExpectedErrMsg: "missing required parameter: month",
		},
		{
			Name:           "missing required parameters",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: year",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchSettlementRecon, "Settlement Reconciliation")
		})
	}
}

func Test_FetchAllSettlements(t *testing.T) {
	fetchAllSettlementsPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.SETTLEMENT_URL,
	)

	// Define the sample response for all settlements
	settlementsResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":     "setl_FNj7g2YS5J67Rz",
				"entity": "settlement",
				"amount": float64(9973635),
				"status": "processed",
			},
			map[string]interface{}{
				"id":     "setl_FJOp0jOWlalIvt",
				"entity": "settlement",
				"amount": float64(299114),
				"status": "processed",
			},
		},
	}

	invalidParamsResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "from must be between 946684800 and 4765046400",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name:    "successful settlements fetch with no parameters",
			Request: map[string]interface{}{},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllSettlementsPath,
						Method:   "GET",
						Response: settlementsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: settlementsResp,
		},
		{
			Name: "successful settlements fetch with pagination",
			Request: map[string]interface{}{
				"count": float64(10),
				"skip":  float64(0),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllSettlementsPath,
						Method:   "GET",
						Response: settlementsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: settlementsResp,
		},
		{
			Name: "successful settlements fetch with date range",
			Request: map[string]interface{}{
				"from": float64(1609459200), // 2021-01-01
				"to":   float64(1640995199), // 2021-12-31
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllSettlementsPath,
						Method:   "GET",
						Response: settlementsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: settlementsResp,
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
			ExpectedErrMsg: "fetching settlements failed: from must be " +
				"between 946684800 and 4765046400",
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
			Name: "settlement creation with required parameters only",
			Request: map[string]interface{}{
				"amount": float64(200000),
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

	// Sample response for successful fetch without expanded payouts
	basicSettlementListResp := map[string]interface{}{
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

	// Sample response with expanded payouts
	expandedSettlementListResp := map[string]interface{}{
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
				"ondemand_payouts": []interface{}{
					map[string]interface{}{
						"id":     "pout_FNj7g2YS5J67Rz",
						"entity": "payout",
						"amount": float64(199410),
						"status": "processed",
					},
				},
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
				"ondemand_payouts": []interface{}{
					map[string]interface{}{
						"id":     "pout_FJOp0jOWlalIvt",
						"entity": "payout",
						"amount": float64(299114),
						"status": "processed",
					},
				},
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
			Name:    "successful instant settlements fetch with no parameters",
			Request: map[string]interface{}{},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllInstantSettlementsPath,
						Method:   "GET",
						Response: basicSettlementListResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: basicSettlementListResp,
		},
		{
			Name: "instant settlements fetch with pagination",
			Request: map[string]interface{}{
				"count": float64(10),
				"skip":  float64(0),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllInstantSettlementsPath,
						Method:   "GET",
						Response: basicSettlementListResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: basicSettlementListResp,
		},
		{
			Name: "instant settlements fetch with expanded payouts",
			Request: map[string]interface{}{
				"expand": []interface{}{"ondemand_payouts"},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllInstantSettlementsPath,
						Method:   "GET",
						Response: expandedSettlementListResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: expandedSettlementListResp,
		},
		{
			Name: "instant settlements fetch with date range",
			Request: map[string]interface{}{
				"from": float64(1609459200), // 2021-01-01
				"to":   float64(1640995199), // 2021-12-31
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllInstantSettlementsPath,
						Method:   "GET",
						Response: basicSettlementListResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: basicSettlementListResp,
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
			ExpectedErrMsg: "fetching instant settlements failed: from must be " +
				"between 946684800 and 4765046400",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchAllInstantSettlements, "Instant Settlements List")
		})
	}
}

func Test_FetchInstantSettlement(t *testing.T) {
	fetchInstantSettlementPathFmt := fmt.Sprintf(
		"/%s%s/ondemand/%%s",
		constants.VERSION_V1,
		constants.SETTLEMENT_URL,
	)

	instantSettlementResp := map[string]interface{}{
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
	}

	instantSettlementNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "instant settlement not found",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful instant settlement fetch",
			Request: map[string]interface{}{
				"settlement_id": "setlod_FNj7g2YS5J67Rz",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(fetchInstantSettlementPathFmt,
							"setlod_FNj7g2YS5J67Rz"),
						Method:   "GET",
						Response: instantSettlementResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: instantSettlementResp,
		},
		{
			Name: "instant settlement not found",
			Request: map[string]interface{}{
				"settlement_id": "setlod_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchInstantSettlementPathFmt, "setlod_invalid"),
						Method:   "GET",
						Response: instantSettlementNotFoundResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "fetching instant settlement failed: " +
				"instant settlement not found",
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
			runToolTest(t, tc, FetchInstantSettlement, "Instant Settlement")
		})
	}
}
