package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_CreatePaymentLink(t *testing.T) {
	createPaymentLinkPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.PaymentLink_URL,
	)

	successfulPaymentLinkResp := map[string]interface{}{
		"id":          "plink_ExjpAUN3gVHrPJ",
		"amount":      float64(50000),
		"currency":    "INR",
		"description": "Test payment",
		"status":      "created",
		"short_url":   "https://rzp.io/i/nxrHnLJ",
	}

	paymentLinkWithoutDescResp := map[string]interface{}{
		"id":        "plink_ExjpAUN3gVHrPJ",
		"amount":    float64(50000),
		"currency":  "INR",
		"status":    "created",
		"short_url": "https://rzp.io/i/nxrHnLJ",
	}

	invalidCurrencyErrorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "API error: Invalid currency",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful payment link creation",
			Request: map[string]interface{}{
				"amount":      float64(50000),
				"currency":    "INR",
				"description": "Test payment",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createPaymentLinkPath,
						Method:   "POST",
						Response: successfulPaymentLinkResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulPaymentLinkResp,
		},
		{
			Name: "payment link without description",
			Request: map[string]interface{}{
				"amount":   float64(50000),
				"currency": "INR",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createPaymentLinkPath,
						Method:   "POST",
						Response: paymentLinkWithoutDescResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: paymentLinkWithoutDescResp,
		},
		{
			Name: "missing amount parameter",
			Request: map[string]interface{}{
				"currency": "INR",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: amount",
		},
		{
			Name: "missing currency parameter",
			Request: map[string]interface{}{
				"amount": float64(50000),
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: currency",
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				// Missing both amount and currency (required parameters)
				"description": 12345, // Wrong type for description
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:\n- " +
				"missing required parameter: amount\n- " +
				"missing required parameter: currency\n- " +
				"invalid parameter type: description",
		},
		{
			Name: "payment link creation fails",
			Request: map[string]interface{}{
				"amount":   float64(50000),
				"currency": "XYZ", // Invalid currency
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createPaymentLinkPath,
						Method:   "POST",
						Response: invalidCurrencyErrorResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "creating payment link failed: API error: Invalid currency",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CreatePaymentLink, "Payment Link")
		})
	}
}

func Test_FetchPaymentLink(t *testing.T) {
	fetchPaymentLinkPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.PaymentLink_URL,
	)

	// Define common response maps to be reused
	paymentLinkResp := map[string]interface{}{
		"id":          "plink_ExjpAUN3gVHrPJ",
		"amount":      float64(50000),
		"currency":    "INR",
		"description": "Test payment",
		"status":      "paid",
		"short_url":   "https://rzp.io/i/nxrHnLJ",
	}

	paymentLinkNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "payment link not found",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful payment link fetch",
			Request: map[string]interface{}{
				"payment_link_id": "plink_ExjpAUN3gVHrPJ",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchPaymentLinkPathFmt, "plink_ExjpAUN3gVHrPJ"),
						Method:   "GET",
						Response: paymentLinkResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: paymentLinkResp,
		},
		{
			Name: "payment link not found",
			Request: map[string]interface{}{
				"payment_link_id": "plink_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchPaymentLinkPathFmt, "plink_invalid"),
						Method:   "GET",
						Response: paymentLinkNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching payment link failed: payment link not found",
		},
		{
			Name:           "missing payment_link_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_link_id",
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				// Missing payment_link_id parameter
				"non_existent_param": 12345, // Additional parameter that doesn't exist
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_link_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchPaymentLink, "Payment Link")
		})
	}
}
