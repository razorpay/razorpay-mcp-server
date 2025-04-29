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
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchPaymentLink, "Payment Link")
		})
	}
}

func Test_CreateUpiPaymentLink(t *testing.T) {
	createPaymentLinkPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.PaymentLink_URL,
	)

	upiPaymentLinkWithAllParamsResp := map[string]interface{}{
		"id":              "plink_UpiAllParamsExjpAUN3gVHrPJ",
		"amount":          float64(50000),
		"currency":        "INR",
		"description":     "Test UPI payment with all params",
		"reference_id":    "REF12345",
		"accept_partial":  true,
		"expire_by":       float64(1718196584),
		"reminder_enable": true,
		"status":          "created",
		"short_url":       "https://rzp.io/i/upiAllParams123",
		"upi_link":        true,
		"customer": map[string]interface{}{
			"name":    "Test Customer",
			"email":   "test@example.com",
			"contact": "+919876543210",
		},
		"notes": map[string]interface{}{
			"policy_name": "Test Policy",
			"user_id":     "usr_123",
		},
	}

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "API error: Something went wrong",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "UPI payment link with all parameters",
			Request: map[string]interface{}{
				"amount":                   float64(50000),
				"description":              "Test UPI payment with all params",
				"reference_id":             "REF12345",
				"accept_partial":           true,
				"first_min_partial_amount": float64(10000),
				"expire_by":                float64(1718196584),
				"customer_name":            "Test Customer",
				"customer_email":           "test@example.com",
				"customer_contact":         "+919876543210",
				"notify_sms":               true,
				"notify_email":             true,
				"reminder_enable":          true,
				"notes": map[string]interface{}{
					"policy_name": "Test Policy",
					"user_id":     "usr_123",
				},
				"callback_url":    "https://example.com/callback",
				"callback_method": "get",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createPaymentLinkPath,
						Method:   "POST",
						Response: upiPaymentLinkWithAllParamsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: upiPaymentLinkWithAllParamsResp,
		},
		{
			Name:           "missing amount parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: amount",
		},
		{
			Name: "UPI payment link creation fails",
			Request: map[string]interface{}{
				"amount": float64(50000),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createPaymentLinkPath,
						Method:   "POST",
						Response: errorResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "upi pl create failed: API error: Something went wrong",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CreateUpiPaymentLink, "UPI Payment Link")
		})
	}
}

func Test_ResendPaymentLinkNotification(t *testing.T) {
	notifyPaymentLinkPathFmt := fmt.Sprintf(
		"/%s%s/%%s/notify_by/%%s",
		constants.VERSION_V1,
		constants.PaymentLink_URL,
	)

	successResponse := map[string]interface{}{
		"success": true,
	}

	invalidMediumErrorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "not a valid notification medium",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful SMS notification",
			Request: map[string]interface{}{
				"payment_link_id": "plink_ExjpAUN3gVHrPJ",
				"medium":          "sms",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							notifyPaymentLinkPathFmt,
							"plink_ExjpAUN3gVHrPJ",
							"sms",
						),
						Method:   "POST",
						Response: successResponse,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name: "invalid medium",
			Request: map[string]interface{}{
				"payment_link_id": "plink_ExjpAUN3gVHrPJ",
				"medium":          "invalid",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "medium must be either 'sms' or 'email'",
		},
		{
			Name: "missing payment_link_id parameter",
			Request: map[string]interface{}{
				"medium": "sms",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_link_id",
		},
		{
			Name: "missing medium parameter",
			Request: map[string]interface{}{
				"payment_link_id": "plink_ExjpAUN3gVHrPJ",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: medium",
		},
		{
			Name: "API error response",
			Request: map[string]interface{}{
				"payment_link_id": "plink_Invalid",
				"medium":          "sms", // Using valid medium so it passes validation
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							notifyPaymentLinkPathFmt,
							"plink_Invalid",
							"sms",
						),
						Method:   "POST",
						Response: invalidMediumErrorResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "sending notification failed: " +
				"not a valid notification medium",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			toolFunc := ResendPaymentLinkNotification
			runToolTest(t, tc, toolFunc, "Payment Link Notification")
		})
	}
}
