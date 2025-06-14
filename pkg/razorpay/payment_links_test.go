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
				"currency":                 "INR",
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
			ExpectedErrMsg: "missing required parameter: currency",
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

func Test_UpdatePaymentLink(t *testing.T) {
	updatePaymentLinkPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.PaymentLink_URL,
	)

	updatedPaymentLinkResp := map[string]interface{}{
		"id":              "plink_FL5HCrWEO112OW",
		"amount":          float64(1000),
		"currency":        "INR",
		"status":          "created",
		"reference_id":    "TS35",
		"expire_by":       float64(1612092283),
		"reminder_enable": false,
		"notes": []interface{}{
			map[string]interface{}{
				"key":   "policy_name",
				"value": "Jeevan Saral",
			},
		},
	}

	invalidStateResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "update can only be made in created or partially paid state",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful update with multiple fields",
			Request: map[string]interface{}{
				"payment_link_id": "plink_FL5HCrWEO112OW",
				"reference_id":    "TS35",
				"expire_by":       float64(1612092283),
				"reminder_enable": false,
				"accept_partial":  true,
				"notes": map[string]interface{}{
					"policy_name": "Jeevan Saral",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							updatePaymentLinkPathFmt,
							"plink_FL5HCrWEO112OW",
						),
						Method:   "PATCH",
						Response: updatedPaymentLinkResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: updatedPaymentLinkResp,
		},
		{
			Name: "successful update with single field",
			Request: map[string]interface{}{
				"payment_link_id": "plink_FL5HCrWEO112OW",
				"reference_id":    "TS35",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							updatePaymentLinkPathFmt,
							"plink_FL5HCrWEO112OW",
						),
						Method:   "PATCH",
						Response: updatedPaymentLinkResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: updatedPaymentLinkResp,
		},
		{
			Name: "missing payment_link_id parameter",
			Request: map[string]interface{}{
				"reference_id": "TS35",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_link_id",
		},
		{
			Name: "no update fields provided",
			Request: map[string]interface{}{
				"payment_link_id": "plink_FL5HCrWEO112OW",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "at least one field to update must be provided",
		},
		{
			Name: "payment link in invalid state",
			Request: map[string]interface{}{
				"payment_link_id": "plink_Paid",
				"reference_id":    "TS35",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							updatePaymentLinkPathFmt,
							"plink_Paid",
						),
						Method:   "PATCH",
						Response: invalidStateResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "updating payment link failed: update can only be made in " +
				"created or partially paid state",
		},
		{
			Name: "update with explicit false value",
			Request: map[string]interface{}{
				"payment_link_id": "plink_FL5HCrWEO112OW",
				"reminder_enable": false, // Explicitly set to false
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							updatePaymentLinkPathFmt,
							"plink_FL5HCrWEO112OW",
						),
						Method:   "PATCH",
						Response: updatedPaymentLinkResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: updatedPaymentLinkResp,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			toolFunc := UpdatePaymentLink
			runToolTest(t, tc, toolFunc, "Payment Link Update")
		})
	}
}

func Test_FetchAllPaymentLinks(t *testing.T) {
	fetchAllPaymentLinksPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.PaymentLink_URL,
	)

	allPaymentLinksResp := map[string]interface{}{
		"payment_links": []interface{}{
			map[string]interface{}{
				"id":           "plink_KBnb7I424Rc1R9",
				"amount":       float64(10000),
				"currency":     "INR",
				"status":       "paid",
				"description":  "Grocery",
				"reference_id": "111",
				"short_url":    "https://rzp.io/i/alaBxs0i",
				"upi_link":     false,
			},
			map[string]interface{}{
				"id":           "plink_JP6yOUDCuHgcrl",
				"amount":       float64(10000),
				"currency":     "INR",
				"status":       "paid",
				"description":  "Online Tutoring - 1 Month",
				"reference_id": "11212",
				"short_url":    "https://rzp.io/i/0ioYuawFu",
				"upi_link":     false,
			},
		},
	}

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The api key/secret provided is invalid",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name:    "fetch all payment links",
			Request: map[string]interface{}{},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllPaymentLinksPath,
						Method:   "GET",
						Response: allPaymentLinksResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: allPaymentLinksResp,
		},
		{
			Name:    "api error",
			Request: map[string]interface{}{},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllPaymentLinksPath,
						Method:   "GET",
						Response: errorResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching payment links failed: The api key/secret provided is invalid", // nolint:lll
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			toolFunc := FetchAllPaymentLinks
			runToolTest(t, tc, toolFunc, "Payment Links")
		})
	}
}
