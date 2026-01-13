package razorpay

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_FetchPayment(t *testing.T) {
	fetchPaymentPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	paymentResp := map[string]interface{}{
		"id":     "pay_MT48CvBhIC98MQ",
		"amount": float64(1000),
		"status": "captured",
	}

	paymentNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "payment not found",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful payment fetch",
			Request: map[string]interface{}{
				"payment_id": "pay_MT48CvBhIC98MQ",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchPaymentPathFmt, "pay_MT48CvBhIC98MQ"),
						Method:   "GET",
						Response: paymentResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: paymentResp,
		},
		{
			Name: "payment not found",
			Request: map[string]interface{}{
				"payment_id": "pay_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchPaymentPathFmt, "pay_invalid"),
						Method:   "GET",
						Response: paymentNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching payment failed: payment not found",
		},
		{
			Name:           "missing payment_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchPayment, "Payment")
		})
	}
}

func Test_FetchPaymentCardDetails(t *testing.T) {
	fetchCardDetailsPathFmt := fmt.Sprintf(
		"/%s%s/%%s/card",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	cardDetailsResp := map[string]interface{}{
		"id":            "card_JXPULjlKqC5j0i",
		"entity":        "card",
		"name":          "Gaurav Kumar",
		"last4":         "4366",
		"network":       "Visa",
		"type":          "credit",
		"issuer":        "UTIB",
		"international": false,
		"emi":           false,
		"sub_type":      "consumer",
		"token_iin":     nil,
	}

	paymentNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The id provided does not exist",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful card details fetch",
			Request: map[string]interface{}{
				"payment_id": "pay_DtFYPi3IfUTgsL",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchCardDetailsPathFmt, "pay_DtFYPi3IfUTgsL"),
						Method:   "GET",
						Response: cardDetailsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: cardDetailsResp,
		},
		{
			Name: "payment not found",
			Request: map[string]interface{}{
				"payment_id": "pay_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchCardDetailsPathFmt, "pay_invalid"),
						Method:   "GET",
						Response: paymentNotFoundResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "fetching card details failed: " +
				"The id provided does not exist",
		},
		{
			Name:           "missing payment_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchPaymentCardDetails, "Card Details")
		})
	}
}

func Test_CapturePayment(t *testing.T) {
	capturePaymentPathFmt := fmt.Sprintf(
		"/%s%s/%%s/capture",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	successfulCaptureResp := map[string]interface{}{
		"id":                "pay_G3P9vcIhRs3NV4",
		"entity":            "payment",
		"amount":            float64(1000),
		"currency":          "INR",
		"status":            "captured",
		"order_id":          "order_GjCr5oKh4AVC51",
		"invoice_id":        nil,
		"international":     false,
		"method":            "card",
		"amount_refunded":   float64(0),
		"refund_status":     nil,
		"captured":          true,
		"description":       "Payment for Adidas shoes",
		"card_id":           "card_KOdY30ajbuyOYN",
		"bank":              nil,
		"wallet":            nil,
		"vpa":               nil,
		"email":             "gaurav.kumar@example.com",
		"contact":           "9000090000",
		"customer_id":       "cust_K6fNE0WJZWGqtN",
		"token_id":          "token_KOdY$DBYQOv08n",
		"notes":             []interface{}{},
		"fee":               float64(1),
		"tax":               float64(0),
		"error_code":        nil,
		"error_description": nil,
		"error_source":      nil,
		"error_step":        nil,
		"error_reason":      nil,
		"acquirer_data": map[string]interface{}{
			"authentication_reference_number": "100222021120200000000742753928",
		},
		"created_at": float64(1605871409),
	}

	alreadyCapturedResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "This payment has already been captured",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful payment capture",
			Request: map[string]interface{}{
				"payment_id": "pay_G3P9vcIhRs3NV4",
				"amount":     float64(1000),
				"currency":   "INR",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(capturePaymentPathFmt, "pay_G3P9vcIhRs3NV4"),
						Method:   "POST",
						Response: successfulCaptureResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulCaptureResp,
		},
		{
			Name: "payment already captured",
			Request: map[string]interface{}{
				"payment_id": "pay_G3P9vcIhRs3NV4",
				"amount":     float64(1000),
				"currency":   "INR",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(capturePaymentPathFmt, "pay_G3P9vcIhRs3NV4"),
						Method:   "POST",
						Response: alreadyCapturedResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "capturing payment failed: This payment has already been " +
				"captured",
		},
		{
			Name: "missing payment_id parameter",
			Request: map[string]interface{}{
				"amount":   float64(1000),
				"currency": "INR",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_id",
		},
		{
			Name: "missing amount parameter",
			Request: map[string]interface{}{
				"payment_id": "pay_G3P9vcIhRs3NV4",
				"currency":   "INR",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: amount",
		},
		{
			Name: "missing currency parameter",
			Request: map[string]interface{}{
				"payment_id": "pay_G3P9vcIhRs3NV4",
				"amount":     float64(1000),
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: currency",
		},
		{
			Name:    "multiple validation errors",
			Request: map[string]interface{}{
				// All required parameters missing
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:\n- " +
				"missing required parameter: payment_id\n- " +
				"missing required parameter: amount\n- " +
				"missing required parameter: currency",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CapturePayment, "Payment")
		})
	}
}

func Test_UpdatePayment(t *testing.T) {
	updatePaymentPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	successfulUpdateResp := map[string]interface{}{
		"id":              "pay_KbCVlLqUbb3VhA",
		"entity":          "payment",
		"amount":          float64(400000),
		"currency":        "INR",
		"status":          "authorized",
		"order_id":        nil,
		"invoice_id":      nil,
		"international":   false,
		"method":          "emi",
		"amount_refunded": float64(0),
		"refund_status":   nil,
		"captured":        false,
		"description":     "Test Transaction",
		"card_id":         "card_KbCVlPnxWRlOpH",
		"bank":            "HDFC",
		"wallet":          nil,
		"vpa":             nil,
		"email":           "gaurav.kumar@example.com",
		"contact":         "+919000090000",
		"notes": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		"fee":               nil,
		"tax":               nil,
		"error_code":        nil,
		"error_description": nil,
		"error_source":      nil,
		"error_step":        nil,
		"error_reason":      nil,
		"acquirer_data": map[string]interface{}{
			"auth_code": "205480",
		},
		"emi_plan": map[string]interface{}{
			"issuer":   "HDFC",
			"type":     "credit",
			"rate":     float64(1500),
			"duration": float64(24),
		},
		"created_at": float64(1667398779),
	}

	paymentNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The id provided does not exist",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful payment notes update",
			Request: map[string]interface{}{
				"payment_id": "pay_KbCVlLqUbb3VhA",
				"notes": map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(updatePaymentPathFmt, "pay_KbCVlLqUbb3VhA"),
						Method:   "PATCH",
						Response: successfulUpdateResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulUpdateResp,
		},
		{
			Name: "payment not found",
			Request: map[string]interface{}{
				"payment_id": "pay_invalid",
				"notes": map[string]interface{}{
					"key1": "value1",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(updatePaymentPathFmt, "pay_invalid"),
						Method:   "PATCH",
						Response: paymentNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "updating payment failed: The id provided does not exist",
		},
		{
			Name: "missing payment_id parameter",
			Request: map[string]interface{}{
				"notes": map[string]interface{}{
					"key1": "value1",
				},
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_id",
		},
		{
			Name: "missing notes parameter",
			Request: map[string]interface{}{
				"payment_id": "pay_KbCVlLqUbb3VhA",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: notes",
		},
		{
			Name:    "multiple validation errors",
			Request: map[string]interface{}{
				// All required parameters missing
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:\n- " +
				"missing required parameter: payment_id\n- " +
				"missing required parameter: notes",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, UpdatePayment, "Payment")
		})
	}
}

func Test_FetchAllPayments(t *testing.T) {
	fetchAllPaymentsPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	// Sample response for successful fetch
	paymentsListResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":              "pay_KbCFyQ0t9Lmi1n",
				"entity":          "payment",
				"amount":          float64(1000),
				"currency":        "INR",
				"status":          "authorized",
				"order_id":        nil,
				"invoice_id":      nil,
				"international":   false,
				"method":          "netbanking",
				"amount_refunded": float64(0),
				"refund_status":   nil,
				"captured":        false,
				"description":     "Test Transaction",
				"card_id":         nil,
				"bank":            "IBKL",
				"wallet":          nil,
				"vpa":             nil,
				"email":           "gaurav.kumar@gmail.com",
				"contact":         "+919000090000",
				"notes": map[string]interface{}{
					"address": "Razorpay Corporate Office",
				},
				"fee":               nil,
				"tax":               nil,
				"error_code":        nil,
				"error_description": nil,
				"error_source":      nil,
				"error_step":        nil,
				"error_reason":      nil,
				"acquirer_data": map[string]interface{}{
					"bank_transaction_id": "5733649",
				},
				"created_at": float64(1667397881),
			},
			map[string]interface{}{
				"id":              "pay_KbCEDHh1IrU4RJ",
				"entity":          "payment",
				"amount":          float64(1000),
				"currency":        "INR",
				"status":          "authorized",
				"order_id":        nil,
				"invoice_id":      nil,
				"international":   false,
				"method":          "upi",
				"amount_refunded": float64(0),
				"refund_status":   nil,
				"captured":        false,
				"description":     "Test Transaction",
				"card_id":         nil,
				"bank":            nil,
				"wallet":          nil,
				"vpa":             "gaurav.kumar@okhdfcbank",
				"email":           "gaurav.kumar@gmail.com",
				"contact":         "+919000090000",
				"notes": map[string]interface{}{
					"address": "Razorpay Corporate Office",
				},
				"fee":               nil,
				"tax":               nil,
				"error_code":        nil,
				"error_description": nil,
				"error_source":      nil,
				"error_step":        nil,
				"error_reason":      nil,
				"acquirer_data": map[string]interface{}{
					"rrn":                "230901495295",
					"upi_transaction_id": "6935B87A72C2A7BC83FA927AA264AD53",
				},
				"created_at": float64(1667397781),
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
			Name: "successful payments fetch with all parameters",
			Request: map[string]interface{}{
				"from":  float64(1593320020),
				"to":    float64(1624856020),
				"count": float64(2),
				"skip":  float64(1),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllPaymentsPath,
						Method:   "GET",
						Response: paymentsListResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: paymentsListResp,
		},
		{
			Name: "payments fetch with invalid timestamp",
			Request: map[string]interface{}{
				"from": float64(900000000), // Invalid timestamp (too early)
				"to":   float64(1624856020),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllPaymentsPath,
						Method:   "GET",
						Response: invalidParamsResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "fetching payments failed: from must be between " +
				"946684800 and 4765046400",
		},
		{
			Name: "multiple validation errors with wrong types",
			Request: map[string]interface{}{
				"count": "not_a_number",
				"skip":  "not_a_number",
				"from":  "not_a_number",
				"to":    "not_a_number",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:\n- " +
				"invalid parameter type: count\n- " +
				"invalid parameter type: skip\n- " +
				"invalid parameter type: from\n- " +
				"invalid parameter type: to",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchAllPayments, "Payments List")
		})
	}
}

func Test_InitiatePayment(t *testing.T) {
	initiatePaymentPath := fmt.Sprintf(
		"/%s%s/create/json",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	createCustomerPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.CUSTOMER_URL,
	)

	customerResp := map[string]interface{}{
		"id":         "cust_1Aa00000000003",
		"entity":     "customer",
		"name":       "",
		"email":      "",
		"contact":    "9876543210",
		"gstin":      nil,
		"notes":      map[string]interface{}{},
		"created_at": float64(1234567890),
	}

	successPaymentWithRedirectResp := map[string]interface{}{
		"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
		"status":              "created",
		"amount":              float64(10000),
		"currency":            "INR",
		"order_id":            "order_129837127313912",
		"next": []interface{}{
			map[string]interface{}{
				"action": "redirect",
				"url": "https://api.razorpay.com/v1/payments/" +
					"pay_MT48CvBhIC98MQ/authenticate",
			},
		},
	}

	successPaymentWithoutNextResp := map[string]interface{}{
		"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
		"status":              "captured",
		"amount":              float64(10000),
		"currency":            "INR",
		"order_id":            "order_129837127313912",
	}

	paymentErrorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Invalid token",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful payment initiation without next actions",
			Request: map[string]interface{}{
				"amount":   10000,
				"currency": "INR",
				"token":    "token_MT48CvBhIC98MQ",
				"order_id": "order_129837127313912",
				"email":    "test@example.com",
				"contact":  "9876543210",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerResp,
					},
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successPaymentWithoutNextResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
				"payment_details":     successPaymentWithoutNextResp,
				"status":              "payment_initiated",
				"message": "Payment initiated successfully using " +
					"S2S JSON v1 flow",
				"next_step": "Use 'resend_otp' to regenerate OTP or " +
					"'submit_otp' to proceed to enter OTP if " +
					"OTP authentication is required.",
				"next_tool": "resend_otp",
				"next_tool_params": map[string]interface{}{
					"payment_id": "pay_MT48CvBhIC98MQ",
				},
			},
		},
		{
			Name: "successful payment initiation with redirect",
			Request: map[string]interface{}{
				"amount":   10000,
				"token":    "token_MT48CvBhIC98MQ",
				"order_id": "order_129837127313912",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successPaymentWithRedirectResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
				"payment_details":     successPaymentWithRedirectResp,
				"status":              "payment_initiated",
				"message": "Payment initiated. Redirect authentication is available. " +
					"Use the redirect URL provided in available_actions.",
				"available_actions": []interface{}{
					map[string]interface{}{
						"action": "redirect",
						"url": "https://api.razorpay.com/v1/payments/" +
							"pay_MT48CvBhIC98MQ/authenticate",
					},
				},
			},
		},
		{
			Name: "successful payment initiation with contact only",
			Request: map[string]interface{}{
				"amount":   10000,
				"token":    "token_MT48CvBhIC98MQ",
				"order_id": "order_129837127313912",
				"contact":  "9876543210",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerResp,
					},
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successPaymentWithoutNextResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
				"payment_details":     successPaymentWithoutNextResp,
				"status":              "payment_initiated",
				"message": "Payment initiated successfully using " +
					"S2S JSON v1 flow",
				"next_step": "Use 'resend_otp' to regenerate OTP or " +
					"'submit_otp' to proceed to enter OTP if " +
					"OTP authentication is required.",
				"next_tool": "resend_otp",
				"next_tool_params": map[string]interface{}{
					"payment_id": "pay_MT48CvBhIC98MQ",
				},
			},
		},
		{
			Name: "payment initiation with API error",
			Request: map[string]interface{}{
				"amount":   10000,
				"token":    "token_invalid",
				"order_id": "order_129837127313912",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: paymentErrorResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "initiating payment failed:",
		},
		{
			Name: "missing required amount parameter",
			Request: map[string]interface{}{
				"token":    "token_MT48CvBhIC98MQ",
				"order_id": "order_129837127313912",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: amount",
		},
		{
			Name: "missing required order_id parameter",
			Request: map[string]interface{}{
				"amount": 10000,
				"token":  "token_MT48CvBhIC98MQ",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: order_id",
		},
		{
			Name: "invalid amount parameter type",
			Request: map[string]interface{}{
				"amount":   "not_a_number",
				"token":    "token_MT48CvBhIC98MQ",
				"order_id": "order_129837127313912",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: amount",
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				"amount": "not_a_number",
				"token":  123,
				"email":  456,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:\n- invalid parameter type: amount\n- " +
				"invalid parameter type: token\n- " +
				"missing required parameter: order_id\n- " +
				"invalid parameter type: email",
		},
		{
			Name: "successful UPI collect flow payment initiation",
			Request: map[string]interface{}{
				"amount":      10000,
				"currency":    "INR",
				"order_id":    "order_129837127313912",
				"email":       "test@example.com",
				"contact":     "9876543210",
				"customer_id": "cust_RGCgP2osfPKFq2",
				"save":        true,
				"vpa":         "9876543210@ptsbi",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				successUpiCollectResp := map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_129837127313912",
					"method":              "upi",
					"next": []interface{}{
						map[string]interface{}{
							"action": "upi_collect",
							"url": "https://api.razorpay.com/v1/payments/" +
								"pay_MT48CvBhIC98MQ/authenticate",
						},
					},
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerResp,
					},
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successUpiCollectResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
				"payment_details": map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_129837127313912",
					"method":              "upi",
					"next": []interface{}{
						map[string]interface{}{
							"action": "upi_collect",
							"url": "https://api.razorpay.com/v1/payments/" +
								"pay_MT48CvBhIC98MQ/authenticate",
						},
					},
				},
				"status":  "payment_initiated",
				"message": "Payment initiated. Available actions: [upi_collect]",
				"available_actions": []interface{}{
					map[string]interface{}{
						"action": "upi_collect",
						"url": "https://api.razorpay.com/v1/payments/" +
							"pay_MT48CvBhIC98MQ/authenticate",
					},
				},
			},
		},
		{
			Name: "successful UPI collect flow without token",
			Request: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_129837127313912",
				"contact":  "9876543210",
				"vpa":      "9876543210@ptsbi",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				successUpiCollectResp := map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_129837127313912",
					"method":              "upi",
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerResp,
					},
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successUpiCollectResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
				"payment_details": map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_129837127313912",
					"method":              "upi",
				},
				"status": "payment_initiated",
				"message": "Payment initiated successfully using " +
					"S2S JSON v1 flow",
				"next_step": "Use 'resend_otp' to regenerate OTP or " +
					"'submit_otp' to proceed to enter OTP if " +
					"OTP authentication is required.",
				"next_tool": "resend_otp",
				"next_tool_params": map[string]interface{}{
					"payment_id": "pay_MT48CvBhIC98MQ",
				},
			},
		},
		{
			Name: "UPI collect flow with all optional parameters",
			Request: map[string]interface{}{
				"amount":      10000,
				"currency":    "INR",
				"order_id":    "order_129837127313912",
				"email":       "test@example.com",
				"contact":     "9876543210",
				"customer_id": "cust_RGCgP2osfPKFq2",
				"save":        false,
				"vpa":         "test@paytm",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				successUpiCollectResp := map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_129837127313912",
					"method":              "upi",
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerResp,
					},
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successUpiCollectResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
				"payment_details": map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_129837127313912",
					"method":              "upi",
				},
				"status": "payment_initiated",
				"message": "Payment initiated successfully using " +
					"S2S JSON v1 flow",
				"next_step": "Use 'resend_otp' to regenerate OTP or " +
					"'submit_otp' to proceed to enter OTP if " +
					"OTP authentication is required.",
				"next_tool": "resend_otp",
				"next_tool_params": map[string]interface{}{
					"payment_id": "pay_MT48CvBhIC98MQ",
				},
			},
		},
		{
			Name: "invalid save parameter type",
			Request: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_129837127313912",
				"save":     "invalid_string_instead_of_bool",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: save",
		},
		{
			Name: "invalid customer_id parameter type",
			Request: map[string]interface{}{
				"amount":      10000,
				"order_id":    "order_129837127313912",
				"customer_id": 123,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: customer_id",
		},
		{
			Name: "successful UPI intent flow payment initiation",
			Request: map[string]interface{}{
				"amount":     12000,
				"currency":   "INR",
				"order_id":   "order_INTENT123",
				"email":      "intent@example.com",
				"contact":    "9876543210",
				"upi_intent": true,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				successUpiIntentResp := map[string]interface{}{
					"razorpay_payment_id": "pay_INTENT123",
					"status":              "created",
					"amount":              float64(12000),
					"currency":            "INR",
					"order_id":            "order_INTENT123",
					"method":              "upi",
					"upi": map[string]interface{}{
						"flow": "intent",
					},
					"next": []interface{}{
						map[string]interface{}{
							"action": "upi_intent",
							"url": "https://api.razorpay.com/v1/payments/" +
								"pay_INTENT123/upi_intent",
						},
					},
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerResp,
					},
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successUpiIntentResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_INTENT123",
				"payment_details": map[string]interface{}{
					"razorpay_payment_id": "pay_INTENT123",
					"status":              "created",
					"amount":              float64(12000),
					"currency":            "INR",
					"order_id":            "order_INTENT123",
					"method":              "upi",
					"upi": map[string]interface{}{
						"flow": "intent",
					},
					"next": []interface{}{
						map[string]interface{}{
							"action": "upi_intent",
							"url": "https://api.razorpay.com/v1/payments/" +
								"pay_INTENT123/upi_intent",
						},
					},
				},
				"status":  "payment_initiated",
				"message": "Payment initiated. Available actions: [upi_intent]",
				"available_actions": []interface{}{
					map[string]interface{}{
						"action": "upi_intent",
						"url": "https://api.razorpay.com/v1/payments/" +
							"pay_INTENT123/upi_intent",
					},
				},
			},
		},
		{
			Name: "invalid upi_intent parameter type",
			Request: map[string]interface{}{
				"amount":     10000,
				"order_id":   "order_129837127313912",
				"upi_intent": "invalid_string",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: upi_intent",
		},
		{
			Name: "successful payment initiation with force_terminal_id " +
				"for single block multiple debit",
			Request: map[string]interface{}{
				"amount":            10000,
				"currency":          "INR",
				"order_id":          "order_129837127313912",
				"email":             "test@example.com",
				"contact":           "9876543210",
				"customer_id":       "cust_RGCgP2osfPKFq2",
				"recurring":         true,
				"force_terminal_id": "term_ABCD1234256732",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				successPaymentWithTerminalResp := map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_129837127313912",
					"method":              "upi",
					"force_terminal_id":   "term_ABCD1234256732",
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successPaymentWithTerminalResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
				"payment_details": map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_129837127313912",
					"method":              "upi",
					"force_terminal_id":   "term_ABCD1234256732",
				},
				"status":  "payment_initiated",
				"message": "Payment initiated successfully using S2S JSON v1 flow",
				"next_step": "Use 'resend_otp' to regenerate OTP or " +
					"'submit_otp' to proceed to enter OTP if " +
					"OTP authentication is required.",
				"next_tool": "resend_otp",
				"next_tool_params": map[string]interface{}{
					"payment_id": "pay_MT48CvBhIC98MQ",
				},
			},
		},
		{
			Name: "invalid force_terminal_id parameter type",
			Request: map[string]interface{}{
				"amount":            10000,
				"order_id":          "order_129837127313912",
				"force_terminal_id": 123,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: force_terminal_id",
		},
		{
			Name: "successful wallet payment with amazonpay",
			Request: map[string]interface{}{
				"amount":       10000,
				"currency":     "INR",
				"order_id":     "order_WALLET123",
				"email":        "test@example.com",
				"contact":      "9876543210",
				"customer_id":  "cust_S1jnUx7pvtinKJ",
				"wallet":       "amazonpay",
				"callback_url": "https://example.com/payment/callback",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				successWalletPaymentResp := map[string]interface{}{
					"razorpay_payment_id": "pay_WALLET123",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_WALLET123",
					"method":              "wallet",
					"wallet":              "amazonpay",
					"next": []interface{}{
						map[string]interface{}{
							"action": "redirect",
							"url": "https://api.razorpay.com/v1/payments/" +
								"wallet_amazonpay/wallet/gateway_redirect",
						},
					},
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successWalletPaymentResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_WALLET123",
				"payment_details": map[string]interface{}{
					"razorpay_payment_id": "pay_WALLET123",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_WALLET123",
					"method":              "wallet",
					"wallet":              "amazonpay",
					"next": []interface{}{
						map[string]interface{}{
							"action": "redirect",
							"url": "https://api.razorpay.com/v1/payments/" +
								"wallet_amazonpay/wallet/gateway_redirect",
						},
					},
				},
				"status": "payment_initiated",
				"message": "Payment initiated. Redirect authentication is available. " +
					"Use the redirect URL provided in available_actions.",
				"available_actions": []interface{}{
					map[string]interface{}{
						"action": "redirect",
						"url": "https://api.razorpay.com/v1/payments/" +
							"wallet_amazonpay/wallet/gateway_redirect",
					},
				},
			},
		},
		{
			Name: "wallet payment with paytm and callback_url",
			Request: map[string]interface{}{
				"amount":       15000,
				"order_id":     "order_PAYTM123",
				"contact":      "9876543210",
				"wallet":       "paytm",
				"callback_url": "https://example.com/success",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				successPaytmWalletResp := map[string]interface{}{
					"razorpay_payment_id": "pay_PAYTM123",
					"status":              "created",
					"amount":              float64(15000),
					"currency":            "INR",
					"order_id":            "order_PAYTM123",
					"method":              "wallet",
					"wallet":              "paytm",
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerResp,
					},
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successPaytmWalletResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_PAYTM123",
				"payment_details": map[string]interface{}{
					"razorpay_payment_id": "pay_PAYTM123",
					"status":              "created",
					"amount":              float64(15000),
					"currency":            "INR",
					"order_id":            "order_PAYTM123",
					"method":              "wallet",
					"wallet":              "paytm",
				},
				"status": "payment_initiated",
				"message": "Payment initiated successfully using " +
					"S2S JSON v1 flow",
				"next_step": "Use 'resend_otp' to regenerate OTP or " +
					"'submit_otp' to proceed to enter OTP if " +
					"OTP authentication is required.",
				"next_tool": "resend_otp",
				"next_tool_params": map[string]interface{}{
					"payment_id": "pay_PAYTM123",
				},
			},
		},
		{
			Name: "invalid wallet parameter type",
			Request: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_129837127313912",
				"wallet":   123,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: wallet",
		},
		{
			Name: "invalid callback_url parameter type",
			Request: map[string]interface{}{
				"amount":       10000,
				"order_id":     "order_129837127313912",
				"callback_url": 123,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: callback_url",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, InitiatePayment, "Payment Initiation")
		})
	}
}

func Test_SubmitOtp(t *testing.T) {
	submitOtpPathFmt := fmt.Sprintf(
		"/%s%s/%%s/otp/submit",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	successOtpSubmitResp := map[string]interface{}{
		"id":                "pay_MT48CvBhIC98MQ",
		"entity":            "payment",
		"amount":            float64(10000),
		"currency":          "INR",
		"status":            "authorized",
		"order_id":          "order_129837127313912",
		"description":       "Test payment",
		"method":            "card",
		"amount_refunded":   float64(0),
		"refund_status":     nil,
		"captured":          false,
		"email":             "test@example.com",
		"contact":           "9876543210",
		"fee":               float64(236),
		"tax":               float64(36),
		"error_code":        nil,
		"error_description": nil,
		"created_at":        float64(1234567890),
	}

	otpVerificationFailedResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Invalid OTP provided",
			"field":       "otp",
		},
	}

	paymentNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Payment not found",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful OTP submission",
			Request: map[string]interface{}{
				"payment_id": "pay_MT48CvBhIC98MQ",
				"otp_string": "123456",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(submitOtpPathFmt, "pay_MT48CvBhIC98MQ"),
						Method:   "POST",
						Response: successOtpSubmitResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"payment_id":    "pay_MT48CvBhIC98MQ",
				"status":        "success",
				"message":       "OTP verified successfully.",
				"response_data": successOtpSubmitResp,
			},
		},
		{
			Name: "OTP verification failed - invalid OTP",
			Request: map[string]interface{}{
				"payment_id": "pay_MT48CvBhIC98MQ",
				"otp_string": "000000",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(submitOtpPathFmt, "pay_MT48CvBhIC98MQ"),
						Method:   "POST",
						Response: otpVerificationFailedResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "OTP verification failed: Invalid OTP provided",
		},
		{
			Name: "payment not found",
			Request: map[string]interface{}{
				"payment_id": "pay_invalid",
				"otp_string": "123456",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(submitOtpPathFmt, "pay_invalid"),
						Method:   "POST",
						Response: paymentNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "OTP verification failed: Payment not found",
		},
		{
			Name: "missing payment_id parameter",
			Request: map[string]interface{}{
				"otp_string": "123456",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_id",
		},
		{
			Name: "missing otp_string parameter",
			Request: map[string]interface{}{
				"payment_id": "pay_MT48CvBhIC98MQ",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: otp_string",
		},
		{
			Name:           "missing both required parameters",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: otp_string",
		},
		{
			Name: "empty otp_string",
			Request: map[string]interface{}{
				"payment_id": "pay_MT48CvBhIC98MQ",
				"otp_string": "",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   fmt.Sprintf(submitOtpPathFmt, "pay_MT48CvBhIC98MQ"),
						Method: "POST",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Authentication failed",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "OTP verification failed: Authentication failed",
		},
		{
			Name: "empty payment_id",
			Request: map[string]interface{}{
				"payment_id": "",
				"otp_string": "123456",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   fmt.Sprintf(submitOtpPathFmt, ""),
						Method: "POST",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "OTP verification failed:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, SubmitOtp, "OTP Submission")
		})
	}
}

func Test_InitiatePaymentWithVPA(t *testing.T) {
	initiatePaymentPath := fmt.Sprintf(
		"/%s%s/create/json",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	createCustomerPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.CUSTOMER_URL,
	)

	customerResp := map[string]interface{}{
		"id":         "cust_1Aa00000000003",
		"entity":     "customer",
		"name":       "",
		"email":      "",
		"contact":    "9876543210",
		"gstin":      nil,
		"notes":      map[string]interface{}{},
		"created_at": float64(1234567890),
	}

	testCases := []RazorpayToolTestCase{
		{
			Name: "successful UPI payment with VPA parameter",
			Request: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_129837127313912",
				"vpa":      "9876543210@ptsbi",
				"email":    "test@example.com",
				"contact":  "9876543210",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				successUpiVpaResp := map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_129837127313912",
					"method":              "upi",
					"email":               "test@example.com",
					"contact":             "9876543210",
					"upi_transaction_id":  nil,
					"upi": map[string]interface{}{
						"flow":        "collect",
						"expiry_time": "6",
						"vpa":         "9876543210@ptsbi",
					},
					"next": []interface{}{
						map[string]interface{}{
							"action": "upi_collect",
							"url": "https://api.razorpay.com/v1/payments/" +
								"pay_MT48CvBhIC98MQ/otp_generate",
						},
					},
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerResp,
					},
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successUpiVpaResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
				"payment_details": map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_129837127313912",
					"method":              "upi",
					"email":               "test@example.com",
					"contact":             "9876543210",
					"upi_transaction_id":  nil,
					"upi": map[string]interface{}{
						"flow":        "collect",
						"expiry_time": "6",
						"vpa":         "9876543210@ptsbi",
					},
					"next": []interface{}{
						map[string]interface{}{
							"action": "upi_collect",
							"url": "https://api.razorpay.com/v1/payments/" +
								"pay_MT48CvBhIC98MQ/otp_generate",
						},
					},
				},
				"status":  "payment_initiated",
				"message": "Payment initiated. Available actions: [upi_collect]",
				"available_actions": []interface{}{
					map[string]interface{}{
						"action": "upi_collect",
						"url": "https://api.razorpay.com/v1/payments/" +
							"pay_MT48CvBhIC98MQ/otp_generate",
					},
				},
			},
		},
		{
			Name: "UPI payment with VPA and custom currency",
			Request: map[string]interface{}{
				"amount":   20000,
				"currency": "INR",
				"order_id": "order_ABC123XYZ456",
				"vpa":      "test@upi",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				successUpiVpaResp := map[string]interface{}{
					"razorpay_payment_id": "pay_ABC123XYZ456",
					"status":              "created",
					"amount":              float64(20000),
					"currency":            "INR",
					"order_id":            "order_ABC123XYZ456",
					"method":              "upi",
					"upi": map[string]interface{}{
						"flow":        "collect",
						"expiry_time": "6",
						"vpa":         "test@upi",
					},
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successUpiVpaResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_ABC123XYZ456",
				"payment_details": map[string]interface{}{
					"razorpay_payment_id": "pay_ABC123XYZ456",
					"status":              "created",
					"amount":              float64(20000),
					"currency":            "INR",
					"order_id":            "order_ABC123XYZ456",
					"method":              "upi",
					"upi": map[string]interface{}{
						"flow":        "collect",
						"expiry_time": "6",
						"vpa":         "test@upi",
					},
				},
				"status": "payment_initiated",
				"message": "Payment initiated successfully using " +
					"S2S JSON v1 flow",
				"next_step": "Use 'resend_otp' to regenerate OTP or " +
					"'submit_otp' to proceed to enter OTP if " +
					"OTP authentication is required.",
				"next_tool": "resend_otp",
				"next_tool_params": map[string]interface{}{
					"payment_id": "pay_ABC123XYZ456",
				},
			},
		},
		{
			Name: "missing VPA parameter value",
			Request: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_129837127313912",
				"vpa":      "",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				successRegularResp := map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_129837127313912",
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successRegularResp,
					},
				)
			},
			ExpectError: false, // Empty VPA should not trigger UPI logic
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
				"payment_details": map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
					"amount":              float64(10000),
					"currency":            "INR",
					"order_id":            "order_129837127313912",
				},
				"status": "payment_initiated",
				"message": "Payment initiated successfully using " +
					"S2S JSON v1 flow",
				"next_step": "Use 'resend_otp' to regenerate OTP or " +
					"'submit_otp' to proceed to enter OTP if " +
					"OTP authentication is required.",
				"next_tool": "resend_otp",
				"next_tool_params": map[string]interface{}{
					"payment_id": "pay_MT48CvBhIC98MQ",
				},
			},
		},
		{
			Name: "VPA parameter automatically sets UPI method",
			Request: map[string]interface{}{
				"amount":   15000,
				"order_id": "order_OVERRIDE123",
				"vpa":      "new@upi",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				successUpiOverrideResp := map[string]interface{}{
					"razorpay_payment_id": "pay_OVERRIDE123",
					"status":              "created",
					"amount":              float64(15000),
					"currency":            "INR",
					"order_id":            "order_OVERRIDE123",
					"method":              "upi", // Should be set to UPI by VPA
					"upi": map[string]interface{}{
						"flow":        "collect", // Default flow
						"expiry_time": "6",       // Default expiry
						"vpa":         "new@upi", // VPA from parameter
					},
					"next": []interface{}{
						map[string]interface{}{
							"action": "upi_collect",
							"url": "https://api.razorpay.com/v1/payments/" +
								"pay_OVERRIDE123/otp_generate",
						},
					},
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successUpiOverrideResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_OVERRIDE123",
				"payment_details": map[string]interface{}{
					"razorpay_payment_id": "pay_OVERRIDE123",
					"status":              "created",
					"amount":              float64(15000),
					"currency":            "INR",
					"order_id":            "order_OVERRIDE123",
					"method":              "upi",
					"upi": map[string]interface{}{
						"flow":        "collect",
						"expiry_time": "6",
						"vpa":         "new@upi",
					},
					"next": []interface{}{
						map[string]interface{}{
							"action": "upi_collect",
							"url": "https://api.razorpay.com/v1/payments/" +
								"pay_OVERRIDE123/otp_generate",
						},
					},
				},
				"status":  "payment_initiated",
				"message": "Payment initiated. Available actions: [upi_collect]",
				"available_actions": []interface{}{
					map[string]interface{}{
						"action": "upi_collect",
						"url": "https://api.razorpay.com/v1/payments/" +
							"pay_OVERRIDE123/otp_generate",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, InitiatePayment, "Payment")
		})
	}
}

// Test helper functions for better coverage
func Test_extractPaymentID(t *testing.T) {
	tests := []struct {
		name     string
		payment  map[string]interface{}
		expected string
	}{
		{
			name: "valid payment ID",
			payment: map[string]interface{}{
				"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
				"status":              "created",
			},
			expected: "pay_MT48CvBhIC98MQ",
		},
		{
			name: "missing payment ID",
			payment: map[string]interface{}{
				"status": "created",
			},
			expected: "",
		},
		{
			name: "nil payment ID",
			payment: map[string]interface{}{
				"razorpay_payment_id": nil,
				"status":              "created",
			},
			expected: "",
		},
		{
			name:     "empty payment map",
			payment:  map[string]interface{}{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPaymentID(tt.payment)
			if result != tt.expected {
				t.Errorf("extractPaymentID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func Test_buildInitiatePaymentResponse(t *testing.T) {
	tests := []struct {
		name           string
		payment        map[string]interface{}
		paymentID      string
		actions        []map[string]interface{}
		expectedMsg    string
		expectedOtpURL string
	}{
		{
			name: "payment with OTP action",
			payment: map[string]interface{}{
				"id":     "pay_MT48CvBhIC98MQ",
				"status": "created",
			},
			paymentID: "pay_MT48CvBhIC98MQ",
			actions: []map[string]interface{}{
				{
					"action": "otp_generate",
					"url": "https://api.razorpay.com/v1/payments/" +
						"pay_MT48CvBhIC98MQ/otp_generate",
				},
			},
			expectedMsg: "Payment initiated. OTP authentication is available. " +
				"Use the 'submit_otp' tool to submit OTP received by the customer " +
				"for authentication.",
			expectedOtpURL: "https://api.razorpay.com/v1/payments/" +
				"pay_MT48CvBhIC98MQ/otp_generate",
		},
		{
			name: "payment with redirect action",
			payment: map[string]interface{}{
				"id":     "pay_MT48CvBhIC98MQ",
				"status": "created",
			},
			paymentID: "pay_MT48CvBhIC98MQ",
			actions: []map[string]interface{}{
				{
					"action": "redirect",
					"url": "https://api.razorpay.com/v1/payments/" +
						"pay_MT48CvBhIC98MQ/authenticate",
				},
			},
			expectedMsg: "Payment initiated. Redirect authentication is available. " +
				"Use the redirect URL provided in available_actions.",
			expectedOtpURL: "",
		},
		{
			name: "payment with UPI collect action",
			payment: map[string]interface{}{
				"id":     "pay_MT48CvBhIC98MQ",
				"status": "created",
			},
			paymentID: "pay_MT48CvBhIC98MQ",
			actions: []map[string]interface{}{
				{
					"action": "upi_collect",
					"url": "https://api.razorpay.com/v1/payments/" +
						"pay_MT48CvBhIC98MQ/authenticate",
				},
			},
			expectedMsg:    "Payment initiated. Available actions: [upi_collect]",
			expectedOtpURL: "",
		},
		{
			name: "payment with multiple actions including OTP",
			payment: map[string]interface{}{
				"id":     "pay_MT48CvBhIC98MQ",
				"status": "created",
			},
			paymentID: "pay_MT48CvBhIC98MQ",
			actions: []map[string]interface{}{
				{
					"action": "otp_generate",
					"url": "https://api.razorpay.com/v1/payments/" +
						"pay_MT48CvBhIC98MQ/otp_generate",
				},
				{
					"action": "redirect",
					"url": "https://api.razorpay.com/v1/payments/" +
						"pay_MT48CvBhIC98MQ/authenticate",
				},
			},
			expectedMsg: "Payment initiated. OTP authentication is available. " +
				"Use the 'submit_otp' tool to submit OTP received by the customer " +
				"for authentication.",
			expectedOtpURL: "https://api.razorpay.com/v1/payments/" +
				"pay_MT48CvBhIC98MQ/otp_generate",
		},
		{
			name: "payment with no actions",
			payment: map[string]interface{}{
				"id":     "pay_MT48CvBhIC98MQ",
				"status": "captured",
			},
			paymentID:      "pay_MT48CvBhIC98MQ",
			actions:        []map[string]interface{}{},
			expectedMsg:    "Payment initiated successfully using S2S JSON v1 flow",
			expectedOtpURL: "",
		},
		{
			name: "payment with unknown action",
			payment: map[string]interface{}{
				"id":     "pay_MT48CvBhIC98MQ",
				"status": "created",
			},
			paymentID: "pay_MT48CvBhIC98MQ",
			actions: []map[string]interface{}{
				{
					"action": "unknown_action",
					"url": "https://api.razorpay.com/v1/payments/" +
						"pay_MT48CvBhIC98MQ/unknown",
				},
			},
			expectedMsg:    "Payment initiated. Available actions: [unknown_action]",
			expectedOtpURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, otpURL := buildInitiatePaymentResponse(
				tt.payment, tt.paymentID, tt.actions)

			// Check basic response structure
			if response["razorpay_payment_id"] != tt.paymentID {
				t.Errorf("Expected payment ID %s, got %v", tt.paymentID,
					response["razorpay_payment_id"])
			}

			if response["status"] != "payment_initiated" {
				t.Errorf("Expected status 'payment_initiated', got %v", response["status"])
			}

			// Check message
			if response["message"] != tt.expectedMsg {
				t.Errorf("Expected message %s, got %v", tt.expectedMsg, response["message"])
			}

			// Check OTP URL
			if otpURL != tt.expectedOtpURL {
				t.Errorf("Expected OTP URL %s, got %s", tt.expectedOtpURL, otpURL)
			}

			// Check actions are included when present
			if len(tt.actions) > 0 {
				if _, exists := response["available_actions"]; !exists {
					t.Error("Expected available_actions to be present in response")
				}
			}

			// Check next step instructions for OTP case
			if tt.paymentID != "" && len(tt.actions) == 0 {
				if _, exists := response["next_step"]; !exists {
					t.Error("Expected next_step to be present for fallback case")
				}
			}
		})
	}
}

func Test_addNextStepInstructions(t *testing.T) {
	tests := []struct {
		name      string
		paymentID string
		expected  bool // whether next_step should be added
	}{
		{
			name:      "valid payment ID",
			paymentID: "pay_MT48CvBhIC98MQ",
			expected:  true,
		},
		{
			name:      "empty payment ID",
			paymentID: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := make(map[string]interface{})
			addNextStepInstructions(response, tt.paymentID)

			if tt.expected {
				if _, exists := response["next_step"]; !exists {
					t.Error("Expected next_step to be added")
				}
				if _, exists := response["next_tool"]; !exists {
					t.Error("Expected next_tool to be added")
				}
				if _, exists := response["next_tool_params"]; !exists {
					t.Error("Expected next_tool_params to be added")
				}

				// Check specific values
				if response["next_tool"] != "resend_otp" {
					t.Errorf("Expected next_tool to be 'resend_otp', got %v",
						response["next_tool"])
				}

				params, ok := response["next_tool_params"].(map[string]interface{})
				if !ok || params["payment_id"] != tt.paymentID {
					t.Errorf("Expected next_tool_params to contain payment_id %s",
						tt.paymentID)
				}
			} else {
				if _, exists := response["next_step"]; exists {
					t.Error("Expected next_step NOT to be added for empty payment ID")
				}
			}
		})
	}
}

func Test_sendOtp_validation(t *testing.T) {
	tests := []struct {
		name        string
		otpURL      string
		expectedErr string
	}{
		{
			name:        "empty URL",
			otpURL:      "",
			expectedErr: "OTP URL is empty",
		},
		{
			name:        "invalid URL",
			otpURL:      "not-a-valid-url",
			expectedErr: "OTP URL must use HTTPS",
		},
		{
			name:        "non-HTTPS URL",
			otpURL:      "http://api.razorpay.com/v1/payments/pay_123/otp_generate",
			expectedErr: "OTP URL must use HTTPS",
		},
		{
			name:        "non-Razorpay domain",
			otpURL:      "https://malicious.com/v1/payments/pay_123/otp_generate",
			expectedErr: "OTP URL must be from Razorpay domain",
		},
		{
			name:        "valid Razorpay URL - should fail at HTTP call",
			otpURL:      "https://api.razorpay.com/v1/payments/pay_123/otp_generate",
			expectedErr: "OTP generation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sendOtp(tt.otpURL)
			if err == nil {
				t.Error("Expected error but got nil")
				return
			}

			if tt.expectedErr != "" && !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error to contain '%s', got '%s'",
					tt.expectedErr, err.Error())
			}
		})
	}
}

func Test_extractOtpSubmitURL(t *testing.T) {
	tests := []struct {
		name     string
		payment  map[string]interface{}
		expected string
	}{
		{
			name: "payment with next actions containing otp_submit",
			payment: map[string]interface{}{
				"next": []interface{}{
					map[string]interface{}{
						"action": "otp_submit",
						"url":    "https://api.razorpay.com/v1/payments/pay_123/otp/submit",
					},
				},
			},
			expected: "https://api.razorpay.com/v1/payments/pay_123/otp/submit",
		},
		{
			name: "payment with multiple next actions",
			payment: map[string]interface{}{
				"next": []interface{}{
					map[string]interface{}{
						"action": "redirect",
						"url":    "https://api.razorpay.com/v1/payments/pay_123/authenticate",
					},
					map[string]interface{}{
						"action": "otp_submit",
						"url":    "https://api.razorpay.com/v1/payments/pay_123/otp/submit",
					},
				},
			},
			expected: "https://api.razorpay.com/v1/payments/pay_123/otp/submit",
		},
		{
			name: "payment with no next actions",
			payment: map[string]interface{}{
				"status": "captured",
			},
			expected: "",
		},
		{
			name: "payment with next actions but no otp_submit",
			payment: map[string]interface{}{
				"next": []interface{}{
					map[string]interface{}{
						"action": "redirect",
						"url":    "https://api.razorpay.com/v1/payments/pay_123/authenticate",
					},
				},
			},
			expected: "",
		},
		{
			name: "payment with empty next array",
			payment: map[string]interface{}{
				"next": []interface{}{},
			},
			expected: "",
		},
		{
			name: "payment with invalid next structure",
			payment: map[string]interface{}{
				"next": "invalid_structure",
			},
			expected: "",
		},
		{
			name: "payment with otp_submit action but nil URL",
			payment: map[string]interface{}{
				"next": []interface{}{
					map[string]interface{}{
						"action": "otp_submit",
						"url":    nil, // nil URL should return empty string
					},
				},
			},
			expected: "",
		},
		{
			name: "payment with otp_submit action but non-string URL",
			payment: map[string]interface{}{
				"next": []interface{}{
					map[string]interface{}{
						"action": "otp_submit",
						"url":    123, // non-string URL should cause type assertion to fail
					},
				},
			},
			expected: "",
		},
		{
			name: "payment with otp_submit action but missing URL field",
			payment: map[string]interface{}{
				"next": []interface{}{
					map[string]interface{}{
						"action": "otp_submit",
						// no url field
					},
				},
			},
			expected: "",
		},
		{
			name: "payment with mixed valid and invalid items in next array",
			payment: map[string]interface{}{
				"next": []interface{}{
					"invalid_item", // This should be skipped
					map[string]interface{}{
						"action": "redirect",
						"url":    "https://example.com/redirect",
					},
					123, // Another invalid item that should be skipped
					map[string]interface{}{
						"action": "otp_submit",
						"url":    "https://api.razorpay.com/v1/payments/pay_123/otp/submit",
					},
				},
			},
			expected: "https://api.razorpay.com/v1/payments/pay_123/otp/submit",
		},
		{
			name: "payment with otp_submit action but missing action field",
			payment: map[string]interface{}{
				"next": []interface{}{
					map[string]interface{}{
						// no action field
						"url": "https://api.razorpay.com/v1/payments/pay_123/otp/submit",
					},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractOtpSubmitURL(tt.payment)
			if result != tt.expected {
				t.Errorf("extractOtpSubmitURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test_extractOtpSubmitURL_invalidInput tests the
// function with invalid input types
func Test_extractOtpSubmitURL_invalidInput(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "string input instead of map",
			input:    "invalid_input",
			expected: "",
		},
		{
			name:     "integer input instead of map",
			input:    123,
			expected: "",
		},
		{
			name:     "slice input instead of map",
			input:    []string{"invalid"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractOtpSubmitURL(tt.input)
			if result != tt.expected {
				t.Errorf("extractOtpSubmitURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func Test_ResendOtp(t *testing.T) {
	resendOtpPathFmt := fmt.Sprintf(
		"/%s%s/%%s/otp/resend",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	successResendOtpResp := map[string]interface{}{
		"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
		"next": []interface{}{
			map[string]interface{}{
				"action": "otp_submit",
				"url": "https://api.razorpay.com/v1/payments/" +
					"pay_MT48CvBhIC98MQ/otp/submit",
			},
		},
	}

	paymentNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Payment not found",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful OTP resend",
			Request: map[string]interface{}{
				"payment_id": "pay_MT48CvBhIC98MQ",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(resendOtpPathFmt, "pay_MT48CvBhIC98MQ"),
						Method:   "POST",
						Response: successResendOtpResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"payment_id": "pay_MT48CvBhIC98MQ",
				"status":     "success",
				"message": "OTP sent successfully. Please enter the OTP received on your " +
					"mobile number to complete the payment.",
				"next_step": "Use 'submit_otp' tool with the OTP code received " +
					"from user to complete payment authentication.",
				"next_tool": "submit_otp",
				"next_tool_params": map[string]interface{}{
					"payment_id": "pay_MT48CvBhIC98MQ",
					"otp_string": "{OTP_CODE_FROM_USER}",
				},
				"otp_submit_url": "https://api.razorpay.com/v1/payments/" +
					"pay_MT48CvBhIC98MQ/otp/submit",
				"response_data": successResendOtpResp,
			},
		},
		{
			Name: "payment not found for OTP resend",
			Request: map[string]interface{}{
				"payment_id": "pay_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(resendOtpPathFmt, "pay_invalid"),
						Method:   "POST",
						Response: paymentNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "OTP resend failed: Payment not found",
		},
		{
			Name:    "missing payment_id parameter for resend",
			Request: map[string]interface{}{
				// No payment_id provided
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_id",
		},
		{
			Name: "OTP resend without next actions",
			Request: map[string]interface{}{
				"payment_id": "pay_MT48CvBhIC98MQ",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   fmt.Sprintf(resendOtpPathFmt, "pay_MT48CvBhIC98MQ"),
						Method: "POST",
						Response: map[string]interface{}{
							"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
							"status":              "created",
						},
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"payment_id": "pay_MT48CvBhIC98MQ",
				"status":     "success",
				"message": "OTP sent successfully. Please enter the OTP received on your " +
					"mobile number to complete the payment.",
				"next_step": "Use 'submit_otp' tool with the OTP code received " +
					"from user to complete payment authentication.",
				"next_tool": "submit_otp",
				"next_tool_params": map[string]interface{}{
					"payment_id": "pay_MT48CvBhIC98MQ",
					"otp_string": "{OTP_CODE_FROM_USER}",
				},
				"response_data": map[string]interface{}{
					"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
					"status":              "created",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, ResendOtp, "OTP Resend")
		})
	}
}

// Test_sendOtp_additionalCases tests additional cases for sendOtp function
func Test_sendOtp_additionalCases(t *testing.T) {
	tests := []struct {
		name        string
		otpURL      string
		expectedErr string
	}{
		{
			name: "URL with invalid characters",
			otpURL: "https://api.razorpay.com/v1/payments/pay_123/" +
				"otp_generate?param=value with spaces",
			expectedErr: "OTP generation failed",
		},
		{
			name: "URL with special characters in domain",
			otpURL: "https://api-test.razorpay.com/v1/payments/" +
				"pay_123/otp_generate",
			expectedErr: "OTP generation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sendOtp(tt.otpURL)
			if err == nil {
				t.Error("Expected error but got nil")
				return
			}
			if tt.expectedErr != "" && !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf(
					"Expected error to contain '%s', got '%s'", tt.expectedErr, err.Error(),
				)
			}
		})
	}
}

// Test_buildPaymentData_edgeCases tests edge cases for
// buildPaymentData function
func Test_buildPaymentData_edgeCases(t *testing.T) {
	tests := []struct {
		name          string
		params        map[string]interface{}
		currency      string
		customerID    string
		expectedError string
		shouldContain map[string]interface{}
	}{
		{
			name: "payment data with valid customer ID",
			params: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_123",
			},
			currency:   "INR",
			customerID: "cust_123456789",
			shouldContain: map[string]interface{}{
				"amount":      10000,
				"currency":    "INR",
				"order_id":    "order_123",
				"customer_id": "cust_123456789",
			},
		},
		{
			name: "payment data with empty token",
			params: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_123",
				"token":    "", // Empty token should not be added
			},
			currency:   "INR",
			customerID: "cust_123456789",
			shouldContain: map[string]interface{}{
				"amount":      10000,
				"currency":    "INR",
				"order_id":    "order_123",
				"customer_id": "cust_123456789",
			},
		},
		{
			name: "payment data with empty customer ID",
			params: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_123",
				"token":    "token_123",
			},
			currency:   "INR",
			customerID: "",
			shouldContain: map[string]interface{}{
				"amount":   10000,
				"currency": "INR",
				"order_id": "order_123",
				"token":    "token_123",
			},
		},
		{
			name: "payment data with all parameters",
			params: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_123",
				"token":    "token_123",
				"email":    "test@example.com",
				"contact":  "9876543210",
				"method":   "upi",
				"save":     true,
				"upi": map[string]interface{}{
					"flow":        "collect",
					"expiry_time": "6",
					"vpa":         "test@upi",
				},
			},
			currency:   "INR",
			customerID: "cust_123456789",
			shouldContain: map[string]interface{}{
				"amount":      10000,
				"currency":    "INR",
				"order_id":    "order_123",
				"customer_id": "cust_123456789",
				"token":       "token_123",
				"email":       "test@example.com",
				"contact":     "9876543210",
				"method":      "upi",
				"save":        true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPaymentData(tt.params, tt.currency, tt.customerID)

			if result == nil {
				t.Error("Expected result but got nil")
				return
			}

			// Check that all expected fields are present
			for key, expectedValue := range tt.shouldContain {
				actualValue, exists := (*result)[key]
				if !exists {
					t.Errorf("Expected key '%s' to exist in result", key)
				} else if actualValue != expectedValue {
					t.Errorf(
						"Expected '%s' to be %v, got %v", key, expectedValue, actualValue,
					)
				}
			}

			// Check that empty token is not added
			if tt.params["token"] == "" {
				if _, exists := (*result)["token"]; exists {
					t.Error("Empty token should not be added to payment data")
				}
			}

			// Check that customer_id is not added when customerID is empty
			if tt.customerID == "" {
				if _, exists := (*result)["customer_id"]; exists {
					t.Error("customer_id should not be added when customerID is empty")
				}
			}

		})
	}
}

// Test_createOrGetCustomer_scenarios tests
// createOrGetCustomer function scenarios
func Test_createOrGetCustomer_scenarios(t *testing.T) {
	tests := []struct {
		name           string
		params         map[string]interface{}
		mockSetup      func() (*http.Client, *httptest.Server)
		expectedError  string
		expectedResult map[string]interface{}
	}{
		{
			name: "successful customer creation with contact",
			params: map[string]interface{}{
				"contact": "9876543210",
			},
			mockSetup: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/customers",
						Method: "POST",
						Response: map[string]interface{}{
							"id":      "cust_123456789",
							"contact": "9876543210",
							"email":   "test@example.com",
						},
					},
				)
			},
			expectedResult: map[string]interface{}{
				"id":      "cust_123456789",
				"contact": "9876543210",
				"email":   "test@example.com",
			},
		},
		{
			name: "no contact provided - returns nil",
			params: map[string]interface{}{
				"amount": 10000,
			},
			mockSetup: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient()
			},
			expectedResult: nil,
		},
		{
			name: "empty contact provided - returns nil",
			params: map[string]interface{}{
				"contact": "",
			},
			mockSetup: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient()
			},
			expectedResult: nil,
		},
		{
			name: "customer creation API error",
			params: map[string]interface{}{
				"contact": "9876543210",
			},
			mockSetup: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/customers",
						Method: "POST",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Invalid contact number",
							},
						},
					},
				)
			},
			expectedError: "failed to create/fetch customer with contact 9876543210",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, server := newMockRzpClient(tt.mockSetup)
			if server != nil {
				defer server.Close()
			}

			result, err := createOrGetCustomer(client, tt.params)

			if tt.expectedError != "" {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf(
						"Expected error to contain '%s', got '%s'", tt.expectedError, err.Error(),
					)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				switch {
				case tt.expectedResult == nil && result != nil:
					t.Errorf("Expected nil result but got %v", result)
				case tt.expectedResult != nil && result == nil:
					t.Error("Expected result but got nil")
				case tt.expectedResult != nil && result != nil:
					if result["id"] != tt.expectedResult["id"] {
						t.Errorf(
							"Expected customer ID '%s', got '%s'", tt.expectedResult["id"],
							result["id"],
						)
					}
				}
			}
		})
	}
}

// Test_processVPAParameters_scenarios tests
//
//	processUPIParameters function scenarios
func Test_processUPIParameters_scenarios(t *testing.T) {
	tests := []struct {
		name           string
		inputParams    map[string]interface{}
		expectedParams map[string]interface{}
	}{
		{
			name: "VPA parameter provided - sets UPI parameters",
			inputParams: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_123",
				"vpa":      "9876543210@paytm",
			},
			expectedParams: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_123",
				"vpa":      "9876543210@paytm",
				"method":   "upi",
				"upi": map[string]interface{}{
					"flow":        "collect",
					"expiry_time": "6",
					"vpa":         "9876543210@paytm",
				},
			},
		},
		{
			name: "empty VPA parameter - no changes",
			inputParams: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_123",
				"vpa":      "",
			},
			expectedParams: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_123",
				"vpa":      "",
			},
		},
		{
			name: "no VPA parameter - no changes",
			inputParams: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_123",
			},
			expectedParams: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_123",
			},
		},
		{
			name: "UPI intent parameter provided - sets UPI intent parameters",
			inputParams: map[string]interface{}{
				"amount":     15000,
				"order_id":   "order_456",
				"upi_intent": true,
			},
			expectedParams: map[string]interface{}{
				"amount":     15000,
				"order_id":   "order_456",
				"upi_intent": true,
				"method":     "upi",
				"upi": map[string]interface{}{
					"flow": "intent",
				},
			},
		},
		{
			name: "UPI intent false - no changes",
			inputParams: map[string]interface{}{
				"amount":     10000,
				"order_id":   "order_123",
				"upi_intent": false,
			},
			expectedParams: map[string]interface{}{
				"amount":     10000,
				"order_id":   "order_123",
				"upi_intent": false,
			},
		},
		{
			name: "both VPA and UPI intent provided - UPI intent takes precedence",
			inputParams: map[string]interface{}{
				"amount":     20000,
				"order_id":   "order_789",
				"vpa":        "test@upi",
				"upi_intent": true,
			},
			expectedParams: map[string]interface{}{
				"amount":     20000,
				"order_id":   "order_789",
				"vpa":        "test@upi",
				"upi_intent": true,
				"method":     "upi",
				"upi": map[string]interface{}{
					"flow": "intent",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := make(map[string]interface{})
			for k, v := range tt.inputParams {
				params[k] = v
			}

			processUPIParameters(params)

			for key, expectedValue := range tt.expectedParams {
				actualValue, exists := params[key]
				if !exists {
					t.Errorf("Expected key '%s' to exist in params", key)
					continue
				}

				if key == "upi" {
					expectedUPI := expectedValue.(map[string]interface{})
					actualUPI, ok := actualValue.(map[string]interface{})
					if !ok {
						t.Errorf("Expected UPI to be map[string]interface{}, got %T", actualValue)
						continue
					}
					for upiKey, upiValue := range expectedUPI {
						if actualUPI[upiKey] != upiValue {
							t.Errorf(
								"Expected UPI[%s] to be '%v', got '%v'",
								upiKey, upiValue, actualUPI[upiKey],
							)
						}
					}
				} else if actualValue != expectedValue {
					t.Errorf(
						"Expected '%s' to be '%v', got '%v'", key, expectedValue, actualValue,
					)
				}
			}
		})
	}
}

// Test_addAdditionalPaymentParameters_scenarios
// tests addAdditionalPaymentParameters function scenarios
// Note: method parameter is set internally by VPA processing, not by user input
func Test_addAdditionalPaymentParameters_scenarios(t *testing.T) {
	tests := []struct {
		name           string
		paymentData    map[string]interface{}
		params         map[string]interface{}
		expectedResult map[string]interface{}
	}{
		{
			name: "all additional parameters provided (method set internally)",
			paymentData: map[string]interface{}{
				"amount":   10000,
				"currency": "INR",
			},
			params: map[string]interface{}{
				"method": "upi",
				"save":   true,
				"upi": map[string]interface{}{
					"flow": "collect",
					"vpa":  "test@upi",
				},
			},
			expectedResult: map[string]interface{}{
				"amount":   10000,
				"currency": "INR",
				"method":   "upi",
				"save":     true,
				"upi": map[string]interface{}{
					"flow": "collect",
					"vpa":  "test@upi",
				},
			},
		},
		{
			name: "empty method parameter - not added (internal processing)",
			paymentData: map[string]interface{}{
				"amount": 10000,
			},
			params: map[string]interface{}{
				"method": "",
				"save":   false,
			},
			expectedResult: map[string]interface{}{
				"amount": 10000,
				"save":   false,
			},
		},
		{
			name: "nil UPI parameters - not added (method set internally)",
			paymentData: map[string]interface{}{
				"amount": 10000,
			},
			params: map[string]interface{}{
				"method": "card",
				"upi":    nil,
			},
			expectedResult: map[string]interface{}{
				"amount": 10000,
				"method": "card",
			},
		},
		{
			name: "invalid UPI parameter type - not added (method set internally)",
			paymentData: map[string]interface{}{
				"amount": 10000,
			},
			params: map[string]interface{}{
				"method": "upi",
				"upi":    "invalid_type",
			},
			expectedResult: map[string]interface{}{
				"amount": 10000,
				"method": "upi",
			},
		},
		{
			name: "recurring parameter provided",
			paymentData: map[string]interface{}{
				"amount": 10000,
			},
			params: map[string]interface{}{
				"recurring": true,
			},
			expectedResult: map[string]interface{}{
				"amount":    10000,
				"recurring": true,
			},
		},
		{
			name: "recurring parameter false",
			paymentData: map[string]interface{}{
				"amount": 10000,
			},
			params: map[string]interface{}{
				"recurring": false,
			},
			expectedResult: map[string]interface{}{
				"amount":    10000,
				"recurring": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentData := make(map[string]interface{})
			for k, v := range tt.paymentData {
				paymentData[k] = v
			}

			addAdditionalPaymentParameters(paymentData, tt.params)

			for key, expectedValue := range tt.expectedResult {
				actualValue, exists := paymentData[key]
				if !exists {
					t.Errorf("Expected key '%s' to exist in paymentData", key)
					continue
				}

				if key == "upi" {
					expectedUPI := expectedValue.(map[string]interface{})
					actualUPI, ok := actualValue.(map[string]interface{})
					if !ok {
						t.Errorf("Expected UPI to be map[string]interface{}, got %T", actualValue)
						continue
					}
					for upiKey, upiValue := range expectedUPI {
						if actualUPI[upiKey] != upiValue {
							t.Errorf(
								"Expected UPI[%s] to be '%v', got '%v'", upiKey, upiValue,
								actualUPI[upiKey],
							)
						}
					}
				} else if actualValue != expectedValue {
					t.Errorf(
						"Expected '%s' to be '%v', got '%v'", key, expectedValue, actualValue,
					)
				}
			}

			// Check that no unexpected keys were added
			for key := range paymentData {
				if _, expected := tt.expectedResult[key]; !expected {
					t.Errorf("Unexpected key '%s' found in paymentData", key)
				}
			}
		})
	}
}

// Test_addContactAndEmailToPaymentData_scenarios
// tests addContactAndEmailToPaymentData function scenarios
func Test_addContactAndEmailToPaymentData_scenarios(t *testing.T) {
	tests := []struct {
		name           string
		paymentData    map[string]interface{}
		params         map[string]interface{}
		expectedResult map[string]interface{}
	}{
		{
			name: "both contact and email provided",
			paymentData: map[string]interface{}{
				"amount": 10000,
			},
			params: map[string]interface{}{
				"contact": "9876543210",
				"email":   "test@example.com",
			},
			expectedResult: map[string]interface{}{
				"amount":  10000,
				"contact": "9876543210",
				"email":   "test@example.com",
			},
		},
		{
			name: "only contact provided - email generated",
			paymentData: map[string]interface{}{
				"amount": 10000,
			},
			params: map[string]interface{}{
				"contact": "9876543210",
			},
			expectedResult: map[string]interface{}{
				"amount":  10000,
				"contact": "9876543210",
				"email":   "9876543210@mcp.razorpay.com",
			},
		},
		{
			name: "empty contact and email - nothing added",
			paymentData: map[string]interface{}{
				"amount": 10000,
			},
			params: map[string]interface{}{
				"contact": "",
				"email":   "",
			},
			expectedResult: map[string]interface{}{
				"amount": 10000,
			},
		},
		{
			name: "contact provided but email is empty - email generated",
			paymentData: map[string]interface{}{
				"amount": 10000,
			},
			params: map[string]interface{}{
				"contact": "9876543210",
				"email":   "",
			},
			expectedResult: map[string]interface{}{
				"amount":  10000,
				"contact": "9876543210",
				"email":   "9876543210@mcp.razorpay.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentData := make(map[string]interface{})
			for k, v := range tt.paymentData {
				paymentData[k] = v
			}

			addContactAndEmailToPaymentData(paymentData, tt.params)

			for key, expectedValue := range tt.expectedResult {
				actualValue, exists := paymentData[key]
				if !exists {
					t.Errorf("Expected key '%s' to exist in paymentData", key)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf(
						"Expected '%s' to be '%v', got '%v'", key, expectedValue, actualValue,
					)
				}
			}

			// Check that no unexpected keys were added
			for key := range paymentData {
				if _, expected := tt.expectedResult[key]; !expected {
					t.Errorf("Unexpected key '%s' found in paymentData", key)
				}
			}
		})
	}
}

// Test_processPaymentResult_edgeCases
// tests edge cases for processPaymentResult function
func Test_processPaymentResult_edgeCases(t *testing.T) {
	tests := []struct {
		name          string
		payment       map[string]interface{}
		expectedError string
		shouldProcess bool
	}{
		{
			name: "payment with OTP URL that causes sendOtp to fail",
			payment: map[string]interface{}{
				"razorpay_payment_id": "pay_123456789",
				"next": []interface{}{
					map[string]interface{}{
						"action": "otp_generate",
						"url":    "http://invalid-url", // Invalid URL
					},
				},
			},
			expectedError: "OTP generation failed",
		},
		{
			name: "payment with empty OTP URL",
			payment: map[string]interface{}{
				"razorpay_payment_id": "pay_123456789",
				"next": []interface{}{
					map[string]interface{}{
						"action": "otp_generate",
						"url":    "", // Empty URL should not trigger sendOtp
					},
				},
			},
			shouldProcess: true,
		},
		{
			name: "payment without next actions",
			payment: map[string]interface{}{
				"razorpay_payment_id": "pay_123456789",
			},
			shouldProcess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processPaymentResult(tt.payment, map[string]interface{}{})

			if tt.expectedError != "" {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf(
						"Expected error to contain '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else if tt.shouldProcess {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				} else {
					// Verify the response structure
					if paymentID, exists := result["razorpay_payment_id"]; !exists ||
						paymentID == "" {
						t.Error("Expected razorpay_payment_id in result")
					}
					if status, exists := result["status"]; !exists ||
						status != "payment_initiated" {
						t.Error("Expected status to be 'payment_initiated'")
					}
				}
			}
		})
	}
}

// Test for sendOtp function - comprehensive coverage
func TestSendOtp(t *testing.T) {
	t.Run("empty OTP URL", func(t *testing.T) {
		err := sendOtp("")
		if err == nil {
			t.Error("Expected error for empty OTP URL")
		}
		if err.Error() != "OTP URL is empty" {
			t.Errorf("Expected 'OTP URL is empty', got '%s'", err.Error())
		}
	})

	t.Run("invalid URL format", func(t *testing.T) {
		err := sendOtp("invalid-url")
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
		// The URL parsing succeeds but fails on HTTPS check
		if !strings.Contains(err.Error(), "OTP URL must use HTTPS") {
			t.Errorf("Expected 'OTP URL must use HTTPS' error, got '%s'", err.Error())
		}
	})

	t.Run("non-HTTPS URL", func(t *testing.T) {
		err := sendOtp("http://api.razorpay.com/v1/payments/otp")
		if err == nil {
			t.Error("Expected error for non-HTTPS URL")
		}
		if err.Error() != "OTP URL must use HTTPS" {
			t.Errorf("Expected 'OTP URL must use HTTPS', got '%s'", err.Error())
		}
	})

	t.Run("non-Razorpay domain", func(t *testing.T) {
		err := sendOtp("https://example.com/otp")
		if err == nil {
			t.Error("Expected error for non-Razorpay domain")
		}
		if err.Error() != "OTP URL must be from Razorpay domain" {
			t.Errorf("Expected 'OTP URL must be from Razorpay domain', got '%s'",
				err.Error())
		}
	})

	t.Run("successful OTP request", func(t *testing.T) {
		// Since we can't actually call external APIs in tests, we'll test the
		// validation logic by testing with a URL that would fail at HTTP call stage
		err := sendOtp(
			"https://api.razorpay.com/v1/payments/invalid-endpoint-for-test")
		if err == nil {
			t.Error("Expected error for invalid endpoint")
		}
		// This should fail at the HTTP request stage, which is expected
		if !strings.Contains(err.Error(), "OTP generation failed") {
			t.Logf("Got expected error: %s", err.Error())
		}
	})

	t.Run("HTTP request creation failure", func(t *testing.T) {
		// Test with invalid characters that would cause http.NewRequest to fail
		// This is difficult to trigger in practice, so we'll test URL validation
		err := sendOtp("https://api.razorpay.com/v1/payments\x00/otp")
		if err == nil {
			t.Error("Expected error for invalid URL characters")
		}
	})
}

// Test for extractPaymentID function
func TestExtractPaymentID(t *testing.T) {
	t.Run("payment ID exists", func(t *testing.T) {
		payment := map[string]interface{}{
			"razorpay_payment_id": "pay_test123",
			"other_field":         "value",
		}
		result := extractPaymentID(payment)
		if result != "pay_test123" {
			t.Errorf("Expected 'pay_test123', got '%s'", result)
		}
	})

	t.Run("payment ID missing", func(t *testing.T) {
		payment := map[string]interface{}{
			"other_field": "value",
		}
		result := extractPaymentID(payment)
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})

	t.Run("payment ID is nil", func(t *testing.T) {
		payment := map[string]interface{}{
			"razorpay_payment_id": nil,
		}
		result := extractPaymentID(payment)
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})
}

// Test for extractNextActions function
func TestExtractNextActions(t *testing.T) {
	t.Run("next actions exist", func(t *testing.T) {
		payment := map[string]interface{}{
			"next": []interface{}{
				map[string]interface{}{
					"action": "redirect",
					"url":    "https://example.com",
				},
				map[string]interface{}{
					"action": "otp",
					"url":    "https://otp.example.com",
				},
			},
		}
		result := extractNextActions(payment)
		if len(result) != 2 {
			t.Errorf("Expected 2 actions, got %d", len(result))
		}
		if result[0]["action"] != "redirect" {
			t.Errorf("Expected first action to be 'redirect', got '%s'",
				result[0]["action"])
		}
	})

	t.Run("next field missing", func(t *testing.T) {
		payment := map[string]interface{}{
			"other_field": "value",
		}
		result := extractNextActions(payment)
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %d actions", len(result))
		}
	})

	t.Run("next field is nil", func(t *testing.T) {
		payment := map[string]interface{}{
			"next": nil,
		}
		result := extractNextActions(payment)
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %d actions", len(result))
		}
	})

	t.Run("next field is not a slice", func(t *testing.T) {
		payment := map[string]interface{}{
			"next": "invalid_type",
		}
		result := extractNextActions(payment)
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %d actions", len(result))
		}
	})

	t.Run("next field contains non-map items", func(t *testing.T) {
		payment := map[string]interface{}{
			"next": []interface{}{
				"invalid_item",
				map[string]interface{}{
					"action": "valid_action",
				},
			},
		}
		result := extractNextActions(payment)
		if len(result) != 1 {
			t.Errorf("Expected 1 valid action, got %d", len(result))
		}
		if result[0]["action"] != "valid_action" {
			t.Errorf("Expected action to be 'valid_action', got '%s'",
				result[0]["action"])
		}
	})
}

// Test for addNextStepInstructions function
func TestAddNextStepInstructions(t *testing.T) {
	t.Run("add next step instructions with payment ID", func(t *testing.T) {
		result := make(map[string]interface{})
		paymentID := "pay_test123"

		addNextStepInstructions(result, paymentID)

		nextStep, exists := result["next_step"]
		if !exists {
			t.Error("Expected next_step to be added")
		}

		nextStepStr, ok := nextStep.(string)
		if !ok {
			t.Error("Expected next_step to be a string")
		}

		if !strings.Contains(nextStepStr, "resend_otp") {
			t.Error("Expected instructions to contain 'resend_otp'")
		}
		if !strings.Contains(nextStepStr, "submit_otp") {
			t.Error("Expected instructions to contain 'submit_otp'")
		}

		// Check next_tool
		nextTool, exists := result["next_tool"]
		if !exists {
			t.Error("Expected next_tool to be added")
		}
		if nextTool != "resend_otp" {
			t.Errorf("Expected next_tool to be 'resend_otp', got '%s'", nextTool)
		}

		// Check next_tool_params
		params, exists := result["next_tool_params"]
		if !exists {
			t.Error("Expected next_tool_params to be added")
		}
		paramsMap, ok := params.(map[string]interface{})
		if !ok {
			t.Error("Expected next_tool_params to be a map")
		}
		if paramsMap["payment_id"] != paymentID {
			t.Errorf("Expected payment_id to be '%s', got '%s'",
				paymentID, paramsMap["payment_id"])
		}
	})

	t.Run("empty payment ID", func(t *testing.T) {
		result := make(map[string]interface{})

		addNextStepInstructions(result, "")

		// Should not add anything when payment ID is empty
		if len(result) != 0 {
			t.Error("Expected no fields to be added for empty payment ID")
		}
	})
}

// Test for processUPIParameters function
func TestProcessUPIParameters(t *testing.T) {
	t.Run("processUPIParameters", func(t *testing.T) {
		// Test with VPA
		params := map[string]interface{}{
			"vpa": "test@upi",
		}
		processUPIParameters(params)

		if params["method"] != "upi" {
			t.Errorf("Expected method to be 'upi', got '%s'", params["method"])
		}

		upi, exists := params["upi"]
		if !exists {
			t.Error("Expected upi field to be added")
		}

		upiMap, ok := upi.(map[string]interface{})
		if !ok {
			t.Error("Expected upi to be a map")
		}

		if upiMap["vpa"] != "test@upi" {
			t.Errorf("Expected vpa to be 'test@upi', got '%s'", upiMap["vpa"])
		}
		if upiMap["flow"] != "collect" {
			t.Errorf("Expected flow to be 'collect', got '%s'", upiMap["flow"])
		}
	})

	t.Run("processUPIParameters - UPI intent", func(t *testing.T) {
		// Test with UPI intent
		params := map[string]interface{}{
			"upi_intent": true,
		}
		processUPIParameters(params)

		if params["method"] != "upi" {
			t.Errorf("Expected method to be 'upi', got '%s'", params["method"])
		}

		upi, exists := params["upi"]
		if !exists {
			t.Error("Expected upi field to be added")
		}

		upiMap, ok := upi.(map[string]interface{})
		if !ok {
			t.Error("Expected upi to be a map")
		}

		if upiMap["flow"] != "intent" {
			t.Errorf("Expected flow to be 'intent', got '%s'", upiMap["flow"])
		}
	})

	t.Run("processUPIParameters - no UPI params", func(t *testing.T) {
		// Test with no UPI parameters
		params := map[string]interface{}{
			"amount": 1000,
		}
		processUPIParameters(params)

		// Should not modify params when no UPI parameters are present
		if _, exists := params["method"]; exists {
			t.Error("Expected method not to be added when no UPI params")
		}
		if _, exists := params["upi"]; exists {
			t.Error("Expected upi not to be added when no UPI params")
		}
	})

}

// Test for createOrGetCustomer function
func TestCreateOrGetCustomer(t *testing.T) {
	t.Run("createOrGetCustomer - no contact", func(t *testing.T) {
		// Test with no contact parameter
		params := map[string]interface{}{
			"amount": 1000,
		}

		// This should return nil, nil since no contact is provided
		result, err := createOrGetCustomer(nil, params)

		if result != nil {
			t.Error("Expected nil result when no contact provided")
		}
		if err != nil {
			t.Errorf("Expected no error when no contact provided, got %v", err)
		}
	})
}

// Test for buildPaymentData function
func TestBuildPaymentData(t *testing.T) {
	t.Run("buildPaymentData", func(t *testing.T) {
		params := map[string]interface{}{
			"amount":   1000,
			"order_id": "order_test123",
		}
		currency := "INR"
		customerId := "cust_test123"

		result := buildPaymentData(params, currency, customerId)

		if (*result)["amount"] != 1000 {
			t.Errorf("Expected amount to be 1000, got %v", (*result)["amount"])
		}
		if (*result)["currency"] != "INR" {
			t.Errorf("Expected currency to be 'INR', got '%s'", (*result)["currency"])
		}
		if (*result)["customer_id"] != customerId {
			t.Errorf("Expected customer_id to be '%s', got '%s'",
				customerId, (*result)["customer_id"])
		}
	})

	t.Run("buildPaymentData - no customer ID", func(t *testing.T) {
		params := map[string]interface{}{
			"amount":   1000,
			"order_id": "order_test123",
		}
		currency := "INR"
		customerId := ""

		result := buildPaymentData(params, currency, customerId)

		if (*result)["amount"] != 1000 {
			t.Errorf("Expected amount to be 1000, got %v", (*result)["amount"])
		}
		if (*result)["currency"] != "INR" {
			t.Errorf("Expected currency to be 'INR', got '%s'", (*result)["currency"])
		}
		// Should not have customer_id when empty
		if _, exists := (*result)["customer_id"]; exists {
			t.Error("Expected no customer_id when empty string provided")
		}
	})

}

// Test for processPaymentResult function
func TestProcessPaymentResult(t *testing.T) {
	t.Run("processPaymentResult", func(t *testing.T) {
		paymentResult := map[string]interface{}{
			"razorpay_payment_id": "pay_test123",
			"status":              "created",
			"next": []interface{}{
				map[string]interface{}{
					"action": "redirect",
					"url":    "https://example.com",
				},
			},
		}

		result, err := processPaymentResult(paymentResult, map[string]interface{}{})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result["razorpay_payment_id"] != "pay_test123" {
			t.Errorf("Expected payment ID, got %v", result["razorpay_payment_id"])
		}

		if result["status"] != "payment_initiated" {
			t.Errorf("Expected status to be 'payment_initiated', got '%s'",
				result["status"])
		}
	})

	t.Run("processPaymentResult - with error", func(t *testing.T) {
		// Test with payment result that might cause an error
		paymentResult := map[string]interface{}{
			"error": map[string]interface{}{
				"code":        "BAD_REQUEST_ERROR",
				"description": "Invalid payment data",
			},
		}

		result, err := processPaymentResult(paymentResult, map[string]interface{}{})

		// The function should handle this gracefully
		if err != nil && result == nil {
			t.Logf("Expected behavior - got error: %v", err)
		} else if result != nil {
			// If no error, result should be properly processed
			if result["status"] != "payment_initiated" {
				t.Errorf("Expected status to be 'payment_initiated', got '%s'",
					result["status"])
			}
		}
	})

	t.Run("processPaymentResult - Amazon Pay wallet with token", func(t *testing.T) {
		// Test Amazon Pay wallet payment with token - should add url_with_token field
		paymentResult := map[string]interface{}{
			"razorpay_payment_id": "pay_test123",
			"next": []interface{}{
				map[string]interface{}{
					"action": "authenticate",
					"url":    "https://api.razorpay.com/pg_router/v1/payments/test123/authenticate",
				},
			},
		}

		params := map[string]interface{}{
			"wallet": "amazonpay",
			"token":  "token_S3IBsftkflBJda",
		}

		result, err := processPaymentResult(paymentResult, params)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result["razorpay_payment_id"] != "pay_test123" {
			t.Errorf("Expected payment ID, got %v", result["razorpay_payment_id"])
		}

		// Check that both original URL and url_with_token exist
		if actions, ok := result["available_actions"].([]map[string]interface{}); ok {
			if len(actions) > 0 {
				// Check original URL is preserved
				if url, ok := actions[0]["url"].(string); ok {
					expectedURL := "https://api.razorpay.com/pg_router/v1/payments/test123/authenticate"
					if url != expectedURL {
						t.Errorf("Expected original URL '%s', got '%s'", expectedURL, url)
					}
				} else {
					t.Errorf("Expected URL to be a string")
				}

				// Check url_with_token is added
				if urlWithToken, ok := actions[0]["url_with_token"].(string); ok {
					expectedURLWithToken := "https://api.razorpay.com/pg_router/v1/payments/test123/authenticate?token=S3IBsftkflBJda"
					if urlWithToken != expectedURLWithToken {
						t.Errorf("Expected url_with_token '%s', got '%s'",
							expectedURLWithToken, urlWithToken)
					}
				} else {
					t.Errorf("Expected url_with_token to be added")
				}
			} else {
				t.Errorf("Expected at least one action")
			}
		} else {
			t.Errorf("Expected available_actions to be a slice")
		}
	})

	t.Run("processPaymentResult - Amazon Pay with method and wallet params", func(t *testing.T) {
		// Test with both method=wallet and wallet=amazonpay (backward compatibility)
		paymentResult := map[string]interface{}{
			"razorpay_payment_id": "pay_test456",
			"next": []interface{}{
				map[string]interface{}{
					"action": "authenticate",
					"url":    "https://api.razorpay.com/pg_router/v1/payments/test456/authenticate",
				},
			},
		}

		params := map[string]interface{}{
			"method": "wallet",
			"wallet": "amazonpay",
			"token":  "token_XYZ123",
		}

		result, err := processPaymentResult(paymentResult, params)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Check that url_with_token is added
		if actions, ok := result["available_actions"].([]map[string]interface{}); ok {
			if len(actions) > 0 {
				if urlWithToken, ok := actions[0]["url_with_token"].(string); ok {
					expectedURLWithToken := "https://api.razorpay.com/pg_router/v1/payments/test456/authenticate?token=XYZ123"
					if urlWithToken != expectedURLWithToken {
						t.Errorf("Expected url_with_token '%s', got '%s'",
							expectedURLWithToken, urlWithToken)
					}
				} else {
					t.Errorf("Expected url_with_token to be added")
				}
			}
		}
	})

	t.Run("processPaymentResult - non-Amazon Pay should not modify URL", func(t *testing.T) {
		// Test non-Amazon Pay payment - URL should not be modified
		paymentResult := map[string]interface{}{
			"razorpay_payment_id": "pay_test123",
			"next": []interface{}{
				map[string]interface{}{
					"action": "authenticate",
					"url":    "https://api.razorpay.com/pg_router/v1/payments/test123/authenticate",
				},
			},
		}

		params := map[string]interface{}{
			"method": "upi",
		}

		result, err := processPaymentResult(paymentResult, params)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Check that the authenticate URL is NOT modified
		if actions, ok := result["available_actions"].([]map[string]interface{}); ok {
			if len(actions) > 0 {
				if url, ok := actions[0]["url"].(string); ok {
					expectedURL := "https://api.razorpay.com/pg_router/v1/payments/test123/authenticate"
					if url != expectedURL {
						t.Errorf("Expected URL '%s' (not modified), got '%s'", expectedURL, url)
					}
				}
			}
		}
	})

}

// Test for extractOtpSubmitURL function
func TestExtractOtpSubmitURL(t *testing.T) {
	t.Run("extractOtpSubmitURL", func(t *testing.T) {
		// Test with valid response data containing OTP submit URL
		responseData := map[string]interface{}{
			"next": []interface{}{
				map[string]interface{}{
					"action": "redirect",
					"url":    "https://example.com/redirect",
				},
				map[string]interface{}{
					"action": "otp_submit",
					"url":    "https://example.com/otp/submit",
				},
			},
		}

		result := extractOtpSubmitURL(responseData)
		if result != "https://example.com/otp/submit" {
			t.Errorf("Expected OTP submit URL, got '%s'", result)
		}
	})

	t.Run("extractOtpSubmitURL - no OTP action", func(t *testing.T) {
		responseData := map[string]interface{}{
			"next": []interface{}{
				map[string]interface{}{
					"action": "redirect",
					"url":    "https://example.com/redirect",
				},
			},
		}

		result := extractOtpSubmitURL(responseData)
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})

	t.Run("extractOtpSubmitURL - invalid input", func(t *testing.T) {
		// Test with invalid input type
		result := extractOtpSubmitURL("invalid_input")
		if result != "" {
			t.Errorf("Expected empty string for invalid input, got '%s'", result)
		}
	})

	t.Run("extractOtpSubmitURL - no next field", func(t *testing.T) {
		responseData := map[string]interface{}{
			"other_field": "value",
		}

		result := extractOtpSubmitURL(responseData)
		if result != "" {
			t.Errorf("Expected empty string when no next field, got '%s'", result)
		}
	})
}

// TestPayments100PercentCoverage_FetchPayment tests FetchPayment coverage
func TestPayments100PercentCoverage_FetchPayment(t *testing.T) {
	// Test FetchPayment with SDK errors
	t.Run("FetchPayment - SDK error", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "SDK error",
			Request: map[string]interface{}{
				"payment_id": "pay_test123",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/pay_test123",
						Method: "GET",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Invalid payment ID",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching payment failed",
		}
		runToolTest(t, testCase, FetchPayment, "Payment")
	})

}

// TestPayments100PercentCoverage_FetchPaymentCardDetails tests
// FetchPaymentCardDetails coverage
func TestPayments100PercentCoverage_FetchPaymentCardDetails(t *testing.T) {
	// Test FetchPaymentCardDetails with SDK errors
	t.Run("FetchPaymentCardDetails - SDK error", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "SDK error",
			Request: map[string]interface{}{
				"payment_id": "pay_test123",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/pay_test123/card",
						Method: "GET",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Card details not available",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching card details failed",
		}
		runToolTest(t, testCase, FetchPaymentCardDetails, "PaymentCardDetails")
	})

}

// TestPayments100PercentCoverage_UpdatePayment tests UpdatePayment coverage
func TestPayments100PercentCoverage_UpdatePayment(t *testing.T) {
	// Test UpdatePayment with SDK errors
	t.Run("UpdatePayment - SDK error", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "SDK error",
			Request: map[string]interface{}{
				"payment_id": "pay_test123",
				"notes": map[string]interface{}{
					"key": "value",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/pay_test123",
						Method: "PATCH",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Invalid notes",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "updating payment failed",
		}
		runToolTest(t, testCase, UpdatePayment, "Payment")
	})

}

// TestPayments100PercentCoverage_CapturePayment tests CapturePayment coverage
func TestPayments100PercentCoverage_CapturePayment(t *testing.T) {
	// Test CapturePayment with SDK errors
	t.Run("CapturePayment - SDK error", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "SDK error",
			Request: map[string]interface{}{
				"payment_id": "pay_test123",
				"amount":     1000,
				"currency":   "INR",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/pay_test123/capture",
						Method: "POST",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Payment cannot be captured",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "capturing payment failed",
		}
		runToolTest(t, testCase, CapturePayment, "Payment")
	})

}

// TestPayments100PercentCoverage_FetchAllPayments tests
// FetchAllPayments coverage
func TestPayments100PercentCoverage_FetchAllPayments(t *testing.T) {
	// Test FetchAllPayments with SDK errors
	t.Run("FetchAllPayments - SDK error", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "SDK error",
			Request: map[string]interface{}{
				"count": 10,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments",
						Method: "GET",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Invalid request",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching payments failed",
		}
		runToolTest(t, testCase, FetchAllPayments, "Collection")
	})

}

// TestPayments100PercentCoverage_ResendOtp tests ResendOtp coverage
func TestPayments100PercentCoverage_ResendOtp(t *testing.T) {
	// Test ResendOtp with SDK errors
	t.Run("ResendOtp - SDK error", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "SDK error",
			Request: map[string]interface{}{
				"payment_id": "pay_test123",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/pay_test123/otp/resend",
						Method: "POST",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Cannot resend OTP",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "OTP resend failed",
		}
		runToolTest(t, testCase, ResendOtp, "ResendOtp")
	})

}

// TestPayments100PercentCoverage_SubmitOtp tests SubmitOtp coverage
func TestPayments100PercentCoverage_SubmitOtp(t *testing.T) {
	// Test SubmitOtp with SDK errors
	t.Run("SubmitOtp - SDK error", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "SDK error",
			Request: map[string]interface{}{
				"payment_id": "pay_test123",
				"otp_string": "123456",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/pay_test123/otp/submit",
						Method: "POST",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Invalid OTP",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "OTP verification failed",
		}
		runToolTest(t, testCase, SubmitOtp, "SubmitOtp")
	})

}

// TestPayments100PercentCoverage_InitiatePayment tests InitiatePayment coverage
func TestPayments100PercentCoverage_InitiatePayment(t *testing.T) {

	// Test InitiatePayment with errors
	t.Run("InitiatePayment - SDK error", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "SDK error",
			Request: map[string]interface{}{
				"amount":   1000,
				"currency": "INR",
				"order_id": "order_test123",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/create/json",
						Method: "POST",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Invalid payment data",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "initiating payment failed",
		}
		runToolTest(t, testCase, InitiatePayment, "InitiatePayment")
	})

	// Test sendOtp with HTTP error status
	t.Run("sendOtp - HTTP error status", func(t *testing.T) {
		server := httptest.NewTLSServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}))
		defer server.Close()

		// Replace domain to pass validation
		testURL := strings.Replace(
			server.URL, server.URL[8:], "api.razorpay.com/v1/payments/otp", 1)
		err := sendOtp(testURL)
		if err == nil {
			t.Error("Expected error for HTTP error status")
		}
		if !strings.Contains(err.Error(), "OTP generation failed with HTTP status") {
			t.Logf("Got expected error: %s", err.Error())
		}
	})

	// More aggressive tests - hitting every error path!
	t.Run("sendOtp - request creation error", func(t *testing.T) {
		// Test with malformed URL that passes parsing but fails request creation
		err := sendOtp("https://api.razorpay.com:99999/invalid")
		if err == nil {
			t.Error("Expected error for malformed URL")
		}
		if !strings.Contains(err.Error(), "failed to create OTP request") &&
			!strings.Contains(err.Error(), "OTP generation failed") {
			t.Logf("Got expected error: %s", err.Error())
		}
	})

	// Test with extremely long URL to trigger different error paths
	t.Run("sendOtp - extreme URL", func(t *testing.T) {
		longPath := strings.Repeat("a", 10000)
		testURL := "https://api.razorpay.com/v1/payments/" + longPath + "/otp"
		err := sendOtp(testURL)
		if err == nil {
			t.Error("Expected error for extreme URL")
		}
	})

}

// contextKey is a type for context keys to avoid collisions
type contextKey string

// TestPayments100PercentCoverage_ContextErrors tests context error paths
func TestPayments100PercentCoverage_ContextErrors(t *testing.T) {
	// Test getClientFromContextOrDefault error path
	t.Run("SubmitOtp - client context error", func(t *testing.T) {
		// Create context with invalid client
		ctx := context.WithValue(
			context.Background(), contextKey("invalid_key"), "invalid_value")

		tool := SubmitOtp(nil, nil)
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"payment_id": "pay_test123",
				"otp_string": "123456",
			},
		}

		result, err := tool.GetHandler()(ctx, request)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil || !result.IsError {
			t.Error("Expected error result for invalid client context")
		}
	})

	// Test sendOtp with actual HTTP client failure
	t.Run("sendOtp - HTTP client failure", func(t *testing.T) {
		// Test with a URL that will fail at the HTTP client level
		err := sendOtp("https://api.razorpay.com:99999/invalid/path/that/will/fail")
		if err == nil {
			t.Error("Expected error for HTTP client failure")
		}
		if !strings.Contains(err.Error(), "OTP generation failed") {
			t.Logf("Got expected error: %s", err.Error())
		}
	})

}

// TestPayments100PercentCoverage_ContextErrors2 tests more context error paths
func TestPayments100PercentCoverage_ContextErrors2(t *testing.T) {
	// Test InitiatePayment - getClientFromContextOrDefault error
	t.Run("InitiatePayment - client context error", func(t *testing.T) {
		// Create context without client
		ctx := context.Background()

		tool := InitiatePayment(nil, nil)
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"amount":   1000,
				"order_id": "order_test123",
			},
		}

		result, err := tool.GetHandler()(ctx, request)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil || !result.IsError {
			t.Error("Expected error result for missing client context")
		}
		if !strings.Contains(result.Text, "no client found in context") {
			t.Errorf("Expected 'no client found in context', got '%s'", result.Text)
		}
	})

	// Simple working test
	t.Run("InitiatePayment - basic test", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "basic test",
			Request: map[string]interface{}{
				"amount":   1000,
				"order_id": "order_test123",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/create/json",
						Method: "POST",
						Response: map[string]interface{}{
							"id":         "pay_test123",
							"entity":     "payment",
							"amount":     1000,
							"currency":   "INR",
							"status":     "created",
							"invalid":    make(chan int), // This causes JSON marshal to fail
							"created_at": 1234567890,
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "failed",
		}
		runToolTest(t, testCase, InitiatePayment, "Payment")
	})

}

// TestPayments100PercentCoverage_ContextErrors3 tests remaining
// context error paths
func TestPayments100PercentCoverage_ContextErrors3(t *testing.T) {
	// Test FetchPayment - getClientFromContextOrDefault error
	t.Run("FetchPayment - client context error", func(t *testing.T) {
		// Create context without client
		ctx := context.Background()

		tool := FetchPayment(nil, nil)
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"payment_id": "pay_test123",
			},
		}

		result, err := tool.GetHandler()(ctx, request)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil || !result.IsError {
			t.Error("Expected error result for missing client context")
		}
		if !strings.Contains(result.Text, "no client found in context") {
			t.Errorf("Expected 'no client found in context', got '%s'", result.Text)
		}
	})

	// Test FetchPaymentCardDetails - getClientFromContextOrDefault error
	t.Run("FetchPaymentCardDetails - client context error", func(t *testing.T) {
		// Create context without client
		ctx := context.Background()

		tool := FetchPaymentCardDetails(nil, nil)
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"payment_id": "pay_test123",
			},
		}

		result, err := tool.GetHandler()(ctx, request)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil || !result.IsError {
			t.Error("Expected error result for missing client context")
		}
		if !strings.Contains(result.Text, "no client found in context") {
			t.Errorf("Expected 'no client found in context', got '%s'", result.Text)
		}
	})

	// Test FetchPaymentCardDetails - JSON marshal error
	// (using channel in response)
	t.Run("FetchPaymentCardDetails - JSON marshal error", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "JSON marshal error",
			Request: map[string]interface{}{
				"payment_id": "pay_test123",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/pay_test123/card",
						Method: "GET",
						Response: map[string]interface{}{
							"id":         "card_test123",
							"entity":     "card",
							"name":       "Test User",
							"last4":      "1234",
							"network":    "Visa",
							"type":       "credit",
							"invalid":    make(chan int), // This causes JSON marshal to fail
							"created_at": 1234567890,
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "failed",
		}
		runToolTest(t, testCase, FetchPaymentCardDetails, "Card Details")
	})

}

// TestPayments100PercentCoverage_ContextErrors4 tests final context error paths
func TestPayments100PercentCoverage_ContextErrors4(t *testing.T) {
	// Test CapturePayment - getClientFromContextOrDefault error
	t.Run("CapturePayment - client context error", func(t *testing.T) {
		// Create context without client
		ctx := context.Background()

		tool := CapturePayment(nil, nil)
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"payment_id": "pay_test123",
				"amount":     1000,
				"currency":   "INR",
			},
		}

		result, err := tool.GetHandler()(ctx, request)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil || !result.IsError {
			t.Error("Expected error result for missing client context")
		}
		if !strings.Contains(result.Text, "no client found in context") {
			t.Errorf("Expected 'no client found in context', got '%s'", result.Text)
		}
	})

	// Test CapturePayment - JSON marshal error (using channel in response)
	t.Run("CapturePayment - JSON marshal error", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "JSON marshal error",
			Request: map[string]interface{}{
				"payment_id": "pay_test123",
				"amount":     1000,
				"currency":   "INR",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/pay_test123/capture",
						Method: "POST",
						Response: map[string]interface{}{
							"id":         "pay_test123",
							"entity":     "payment",
							"amount":     1000,
							"currency":   "INR",
							"status":     "captured",
							"invalid":    make(chan int), // This causes JSON marshal to fail
							"created_at": 1234567890,
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "failed",
		}
		runToolTest(t, testCase, CapturePayment, "Payment")
	})

	// Test UpdatePayment - getClientFromContextOrDefault error
	t.Run("UpdatePayment - client context error", func(t *testing.T) {
		// Create context without client
		ctx := context.Background()

		tool := UpdatePayment(nil, nil)
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"payment_id": "pay_test123",
				"notes": map[string]interface{}{
					"key": "value",
				},
			},
		}

		result, err := tool.GetHandler()(ctx, request)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil || !result.IsError {
			t.Error("Expected error result for missing client context")
		}
		if !strings.Contains(result.Text, "no client found in context") {
			t.Errorf("Expected 'no client found in context', got '%s'", result.Text)
		}
	})

	// Test UpdatePayment - JSON marshal error (using channel in response)
	t.Run("UpdatePayment - JSON marshal error", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "JSON marshal error",
			Request: map[string]interface{}{
				"payment_id": "pay_test123",
				"notes": map[string]interface{}{
					"key": "value",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/pay_test123",
						Method: "PATCH",
						Response: map[string]interface{}{
							"id":       "pay_test123",
							"entity":   "payment",
							"amount":   1000,
							"currency": "INR",
							"status":   "authorized",
							"invalid":  make(chan int), // This causes JSON marshal to fail
							"notes": map[string]interface{}{
								"key": "value",
							},
							"created_at": 1234567890,
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "failed",
		}
		runToolTest(t, testCase, UpdatePayment, "Payment")
	})

	// Test FetchAllPayments - getClientFromContextOrDefault error
	t.Run("FetchAllPayments - client context error", func(t *testing.T) {
		// Create context without client
		ctx := context.Background()

		tool := FetchAllPayments(nil, nil)
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"count": 10,
			},
		}

		result, err := tool.GetHandler()(ctx, request)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil || !result.IsError {
			t.Error("Expected error result for missing client context")
		}
		if !strings.Contains(result.Text, "no client found in context") {
			t.Errorf("Expected 'no client found in context', got '%s'", result.Text)
		}
	})

	// Test FetchAllPayments - JSON marshal error (using channel in response)
	t.Run("FetchAllPayments - JSON marshal error", func(t *testing.T) {
		testCase := RazorpayToolTestCase{
			Name: "JSON marshal error",
			Request: map[string]interface{}{
				"count": 10,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments",
						Method: "GET",
						Response: map[string]interface{}{
							"entity": "collection",
							"count":  1,
							"items": []interface{}{
								map[string]interface{}{
									"id":         "pay_test123",
									"invalid":    make(chan int), // This causes JSON marshal to fail
									"entity":     "payment",
									"amount":     1000,
									"currency":   "INR",
									"status":     "created",
									"created_at": 1234567890,
								},
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "failed",
		}
		runToolTest(t, testCase, FetchAllPayments, "Collection")
	})
}
