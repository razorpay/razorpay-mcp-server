package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

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
				"The id provided does not exist", //nolint:lll
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

	successPaymentWithRedirectResp := map[string]interface{}{
		"razorpay_payment_id": "pay_MT48CvBhIC98MQ",
		"status":              "created",
		"amount":              float64(10000),
		"currency":            "INR",
		"order_id":            "order_MT48CvBhIC98MQ",
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
		"order_id":            "order_MT48CvBhIC98MQ",
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
				"order_id": "order_MT48CvBhIC98MQ",
				"email":    "test@example.com",
				"contact":  "9876543210",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
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
			Name: "successful payment initiation with CVV",
			Request: map[string]interface{}{
				"amount":   10000,
				"currency": "INR",
				"token":    "token_MT48CvBhIC98MQ",
				"order_id": "order_MT48CvBhIC98MQ",
				"email":    "test@example.com",
				"contact":  "9876543210",
				"cvv":      "123",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
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
			Name: "successful payment initiation without CVV (empty string)",
			Request: map[string]interface{}{
				"amount":   10000,
				"currency": "INR",
				"token":    "token_MT48CvBhIC98MQ",
				"order_id": "order_MT48CvBhIC98MQ",
				"email":    "test@example.com",
				"contact":  "9876543210",
				"cvv":      "",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
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
				"order_id": "order_MT48CvBhIC98MQ",
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
				"order_id": "order_MT48CvBhIC98MQ",
				"contact":  "9876543210",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
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
				"order_id": "order_MT48CvBhIC98MQ",
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
				"order_id": "order_MT48CvBhIC98MQ",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: amount",
		},
		{
			Name: "missing required token parameter",
			Request: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_MT48CvBhIC98MQ",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: token",
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
				"order_id": "order_MT48CvBhIC98MQ",
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
		"order_id":          "order_MT48CvBhIC98MQ",
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
