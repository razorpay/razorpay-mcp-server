package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_FetchToken(t *testing.T) {
	fetchTokenPathFmt := fmt.Sprintf(
		"/%s%s/%%s/tokens/%%s",
		constants.VERSION_V1,
		constants.CUSTOMER_URL,
	)

	tokenResp := map[string]interface{}{
		"id":     "token_KOdY30ajbuyOYN",
		"entity": "token",
		"token":  "KOdY30ajbuyOYN",
		"bank":   "HDFC",
		"wallet": nil,
		"method": "card",
		"card": map[string]interface{}{
			"entity":        "card",
			"name":          "Gaurav Kumar",
			"last4":         "4366",
			"network":       "Visa",
			"type":          "credit",
			"issuer":        "UTIB",
			"international": false,
			"emi":           false,
		},
		"vpa":                                    nil,
		"recurring":                              true,
		"recurring_details":                      map[string]interface{}{},
		"auth_type":                              nil,
		"mrn":                                    nil,
		"used_at":                                float64(1629779657),
		"created_at":                             float64(1629779657),
		"max_amount":                             float64(900000),
		"expired_at":                             float64(1772851657),
		"dcc_enabled":                            false,
		"notes":                                  []interface{}{},
		"compliant_with_tokenisation_guidelines": true,
	}

	tokenNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The id provided does not exist",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful token fetch",
			Request: map[string]interface{}{
				"customer_id": "cust_KZrR0wFY0lERs2",
				"token_id":    "token_KOdY30ajbuyOYN",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchTokenPathFmt, "cust_KZrR0wFY0lERs2", "token_KOdY30ajbuyOYN"),
						Method:   "GET",
						Response: tokenResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: tokenResp,
		},
		{
			Name: "token not found",
			Request: map[string]interface{}{
				"customer_id": "cust_invalid",
				"token_id":    "token_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchTokenPathFmt, "cust_invalid", "token_invalid"),
						Method:   "GET",
						Response: tokenNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching token failed: The id provided does not exist",
		},
		{
			Name: "missing customer_id parameter",
			Request: map[string]interface{}{
				"token_id": "token_KOdY30ajbuyOYN",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_id",
		},
		{
			Name: "missing token_id parameter",
			Request: map[string]interface{}{
				"customer_id": "cust_KZrR0wFY0lERs2",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: token_id",
		},
		{
			Name:           "missing both parameters",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchToken, "Token")
		})
	}
}

func Test_FetchAllTokens(t *testing.T) {
	fetchAllTokensPathFmt := fmt.Sprintf(
		"/%s%s/%%s/tokens",
		constants.VERSION_V1,
		constants.CUSTOMER_URL,
	)

	allTokensResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":     "token_KOdY30ajbuyOYN",
				"entity": "token",
				"token":  "KOdY30ajbuyOYN",
				"bank":   "HDFC",
				"wallet": nil,
				"method": "card",
				"card": map[string]interface{}{
					"entity":        "card",
					"name":          "Gaurav Kumar",
					"last4":         "4366",
					"network":       "Visa",
					"type":          "credit",
					"issuer":        "UTIB",
					"international": false,
					"emi":           false,
				},
				"vpa":                                    nil,
				"recurring":                              true,
				"recurring_details":                      map[string]interface{}{},
				"auth_type":                              nil,
				"mrn":                                    nil,
				"used_at":                                float64(1629779657),
				"created_at":                             float64(1629779657),
				"max_amount":                             float64(900000),
				"expired_at":                             float64(1772851657),
				"dcc_enabled":                            false,
				"notes":                                  []interface{}{},
				"compliant_with_tokenisation_guidelines": true,
			},
			map[string]interface{}{
				"id":     "token_KOdY30ajbuyOZ1",
				"entity": "token",
				"token":  "KOdY30ajbuyOZ1",
				"bank":   "HDFC",
				"wallet": nil,
				"method": "card",
				"card": map[string]interface{}{
					"entity":        "card",
					"name":          "Gaurav Kumar",
					"last4":         "1111",
					"network":       "Visa",
					"type":          "debit",
					"issuer":        "HDFC",
					"international": false,
					"emi":           false,
				},
				"vpa":                                    nil,
				"recurring":                              true,
				"recurring_details":                      map[string]interface{}{},
				"auth_type":                              nil,
				"mrn":                                    nil,
				"used_at":                                float64(1629779700),
				"created_at":                             float64(1629779700),
				"max_amount":                             float64(500000),
				"expired_at":                             float64(1772851700),
				"dcc_enabled":                            false,
				"notes":                                  []interface{}{},
				"compliant_with_tokenisation_guidelines": true,
			},
		},
	}

	customerNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The customer id provided does not exist",
		},
	}

	emptyTokensResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(0),
		"items":  []interface{}{},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful tokens fetch",
			Request: map[string]interface{}{
				"customer_id": "cust_KZrR0wFY0lERs2",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchAllTokensPathFmt, "cust_KZrR0wFY0lERs2"),
						Method:   "GET",
						Response: allTokensResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: allTokensResp,
		},
		{
			Name: "customer with no tokens",
			Request: map[string]interface{}{
				"customer_id": "cust_NoTokensCustomer",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchAllTokensPathFmt, "cust_NoTokensCustomer"),
						Method:   "GET",
						Response: emptyTokensResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: emptyTokensResp,
		},
		{
			Name: "customer not found",
			Request: map[string]interface{}{
				"customer_id": "cust_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchAllTokensPathFmt, "cust_invalid"),
						Method:   "GET",
						Response: customerNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching tokens failed: The customer id provided does not exist",
		},
		{
			Name:           "missing customer_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchAllTokens, "Tokens")
		})
	}
}
