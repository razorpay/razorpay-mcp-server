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
		"/%s%s",
		constants.VERSION_V1,
		constants.PAYMENT_URL+"/create/json",
	)

	successResponse := map[string]interface{}{
		"razorpay_payment_id": "pay_29QQoUBi66xm2f",
		"status":              "created",
		"amount":              float64(10000),
		"currency":            "INR",
		"method":              "card",
		"order_id":            "order_9A33XWu170gUtm",
		"next": []interface{}{
			map[string]interface{}{
				"action": "otp_generate",
				"url": "https://api.razorpay.com/v1/payments/" +
					"pay_29QQoUBi66xm2f/otp_generate",
			},
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful initiate payment with all parameters",
			Request: map[string]interface{}{
				"amount":      10000,
				"currency":    "INR",
				"method":      "card",
				"token_id":    "token_KWbKaXtNPdYHWJ",
				"order_id":    "order_9A33XWu170gUtm",
				"email":       "gaurav.kumar@example.com",
				"contact":     "9090909090",
				"cvv":         "123",
				"ip":          "192.168.0.103",
				"user_agent":  "Mozilla/5.0",
				"description": "Test payment",
				"notes": map[string]interface{}{
					"note_key": "value1",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successResponse,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_29QQoUBi66xm2f",
				"payment_details":     successResponse,
				"status":              "payment_initiated",
				"action":              "otp_generate",
				"url": "https://api.razorpay.com/v1/payments/" +
					"pay_29QQoUBi66xm2f/otp_generate",
				"message": "Payment initiated. Next action: otp_generate. " +
					"Use the provided URL for next step.",
				"next_step_instruction": "Call the 'send_otp' tool next " +
					"in the process to complete the payment authentication.",
			},
		},
		{
			Name: "successful initiate payment with defaults",
			Request: map[string]interface{}{
				"amount":   10000,
				"token_id": "token_KWbKaXtNPdYHWJ",
				"order_id": "order_9A33XWu170gUtm",
				"email":    "test@example.com",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successResponse,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_29QQoUBi66xm2f",
				"payment_details":     successResponse,
				"status":              "payment_initiated",
				"action":              "otp_generate",
				"url": "https://api.razorpay.com/v1/payments/" +
					"pay_29QQoUBi66xm2f/otp_generate",
				"message": "Payment initiated. Next action: otp_generate. " +
					"Use the provided URL for next step.",
				"next_step_instruction": "Call the 'send_otp' tool next " +
					"in the process to complete the payment authentication.",
			},
		},
		{
			Name: "missing amount parameter",
			Request: map[string]interface{}{
				"token_id": "token_KWbKaXtNPdYHWJ",
				"order_id": "order_9A33XWu170gUtm",
				"email":    "test@example.com",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient()
			},
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: amount",
		},
		{
			Name: "missing email parameter - generates dummy email",
			Request: map[string]interface{}{
				"amount":   10000,
				"token_id": "token_KWbKaXtNPdYHWJ",
				"order_id": "order_9A33XWu170gUtm",
				"contact":  "9090909090",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successResponse,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_29QQoUBi66xm2f",
				"payment_details":     successResponse,
				"status":              "payment_initiated",
				"action":              "otp_generate",
				"url": "https://api.razorpay.com/v1/payments/" +
					"pay_29QQoUBi66xm2f/otp_generate",
				"message": "Payment initiated. Next action: otp_generate. " +
					"Use the provided URL for next step.",
				"next_step_instruction": "Call the 'send_otp' tool next " +
					"in the process to complete the payment authentication.",
			},
		},
		{
			Name: "missing email and contact - generates default dummy email",
			Request: map[string]interface{}{
				"amount":   10000,
				"token_id": "token_KWbKaXtNPdYHWJ",
				"order_id": "order_9A33XWu170gUtm",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     initiatePaymentPath,
						Method:   "POST",
						Response: successResponse,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"razorpay_payment_id": "pay_29QQoUBi66xm2f",
				"payment_details":     successResponse,
				"status":              "payment_initiated",
				"action":              "otp_generate",
				"url": "https://api.razorpay.com/v1/payments/" +
					"pay_29QQoUBi66xm2f/otp_generate",
				"message": "Payment initiated. Next action: otp_generate. " +
					"Use the provided URL for next step.",
				"next_step_instruction": "Call the 'send_otp' tool next " +
					"in the process to complete the payment authentication.",
			},
		},
		{
			Name: "missing token_id parameter",
			Request: map[string]interface{}{
				"amount":   10000,
				"order_id": "order_9A33XWu170gUtm",
				"email":    "test@example.com",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient()
			},
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: token_id",
		},
		{
			Name: "missing order_id parameter",
			Request: map[string]interface{}{
				"amount":   10000,
				"token_id": "token_KWbKaXtNPdYHWJ",
				"email":    "test@example.com",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient()
			},
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: order_id",
		},
		{
			Name: "API error - invalid amount",
			Request: map[string]interface{}{
				"amount":   50,
				"token_id": "token_KWbKaXtNPdYHWJ",
				"order_id": "order_9A33XWu170gUtm",
				"email":    "test@example.com",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   initiatePaymentPath,
						Method: "POST",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Amount must be at least 100",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "initiating payment failed",
		},
		{
			Name: "invalid parameter types",
			Request: map[string]interface{}{
				"amount":   "invalid_amount",
				"currency": 123,
				"method":   456,
				"token_id": 789,
				"order_id": 101,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient()
			},
			ExpectError: true,
			ExpectedErrMsg: "invalid parameter type: amount\n- " +
				"invalid parameter type: currency\n- " +
				"invalid parameter type: method\n- " +
				"invalid parameter type: token_id\n- " +
				"invalid parameter type: order_id",
		},
		{
			Name: "API error response",
			Request: map[string]interface{}{
				"amount":   10000,
				"token_id": "invalid_token",
				"order_id": "order_9A33XWu170gUtm",
				"email":    "test@example.com",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   initiatePaymentPath,
						Method: "POST",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Invalid token_id provided",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "initiating payment failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, InitiatePayment, "Initiate Payment")
		})
	}
}

func Test_SendOtp(t *testing.T) {
	successResponse := map[string]interface{}{
		"status": "success",
		"next": []interface{}{
			map[string]interface{}{
				"action": "otp_submit",
				"url": "https://api.razorpay.com/v1/payments/" +
					"pay_29QQoUBi66xm2f/otp_submit",
			},
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful OTP generation",
			Request: map[string]interface{}{
				"url": "https://api.razorpay.com/v1/payments/" +
					"pay_29QQoUBi66xm2f/otp_generate",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     "/v1/payments/pay_29QQoUBi66xm2f/otp_generate",
						Method:   "POST",
						Response: successResponse,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"url":            "https://api.razorpay.com/v1/payments/pay_29QQoUBi66xm2f/otp_generate",
				"status_code":    200,
				"response_data":  successResponse,
				"content_type":   "application/json",
				"status":         "success",
				"message":        "OTP sent successfully. Please enter the OTP received on your mobile number to complete the payment.",
				"otp_submit_url": "https://api.razorpay.com/v1/payments/pay_29QQoUBi66xm2f/otp_submit",
				"next_step":      "Use 'verify_otp' tool with the OTP code received from user and url https://api.razorpay.com/v1/payments/pay_29QQoUBi66xm2f/otp_submit to complete payment.",
				"next_tool":      "verify_otp",
				"next_tool_params": map[string]interface{}{
					"url":        "https://api.razorpay.com/v1/payments/pay_29QQoUBi66xm2f/otp_submit",
					"otp_string": "{OTP_CODE_FROM_USER}",
				},
			},
		},
		{
			Name: "OTP generation without next step",
			Request: map[string]interface{}{
				"url": "https://api.razorpay.com/v1/payments/pay_29QQoUBi66xm2f/otp_generate",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     "/v1/payments/pay_29QQoUBi66xm2f/otp_generate",
						Method:   "POST",
						Response: map[string]interface{}{"status": "success"},
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"url":           "https://api.razorpay.com/v1/payments/pay_29QQoUBi66xm2f/otp_generate",
				"status_code":   200,
				"response_data": map[string]interface{}{"status": "success"},
				"content_type":  "application/json",
				"status":        "success",
				"message":       "OTP sent successfully. Please enter the OTP received on your mobile number to complete the payment.",
			},
		},
		{
			Name:    "missing URL parameter",
			Request: map[string]interface{}{
				// Missing url parameter
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: url",
		},
		{
			Name: "API error response",
			Request: map[string]interface{}{
				"url": "https://api.razorpay.com/v1/payments/pay_invalid/otp_generate",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/pay_invalid/otp_generate",
						Method: "POST",
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
			ExpectedErrMsg: "OTP generation failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, SendOtp, "Send OTP")
		})
	}
}

func Test_VerifyOtp(t *testing.T) {
	successResponse := map[string]interface{}{
		"razorpay_payment_id": "pay_29QQoUBi66xm2f",
		"razorpay_order_id":   "order_9A33XWu170gUtm",
		"razorpay_signature":  "signature_abc123",
		"status":              "captured",
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful OTP verification",
			Request: map[string]interface{}{
				"url":        "https://api.razorpay.com/v1/payments/pay_29QQoUBi66xm2f/otp_submit",
				"otp_string": "123456",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     "/v1/payments/pay_29QQoUBi66xm2f/otp_submit",
						Method:   "POST",
						Response: successResponse,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"url":           "https://api.razorpay.com/v1/payments/pay_29QQoUBi66xm2f/otp_submit",
				"otp_sent":      "123456",
				"status_code":   200,
				"response_data": successResponse,
				"content_type":  "application/json",
				"request_body": map[string]interface{}{
					"otp": "123456",
				},
				"status":  "success",
				"message": "OTP verification completed successfully. Payment has been processed.",
			},
		},
		{
			Name: "missing URL parameter",
			Request: map[string]interface{}{
				"otp_string": "123456",
				// Missing url parameter
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: url",
		},
		{
			Name: "missing OTP string parameter",
			Request: map[string]interface{}{
				"url": "https://api.razorpay.com/v1/payments/pay_29QQoUBi66xm2f/otp_submit",
				// Missing otp_string parameter
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: otp_string",
		},
		{
			Name: "invalid OTP error",
			Request: map[string]interface{}{
				"url":        "https://api.razorpay.com/v1/payments/pay_29QQoUBi66xm2f/otp_submit",
				"otp_string": "999999",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   "/v1/payments/pay_29QQoUBi66xm2f/otp_submit",
						Method: "POST",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Invalid OTP",
								"source":      "business",
								"step":        "payment_authentication",
								"reason":      "invalid_otp",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "OTP verification failed",
		},
		{
			Name: "invalid parameter types",
			Request: map[string]interface{}{
				"url":        123, // Invalid type
				"otp_string": 456, // Invalid type
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: url\n- invalid parameter type: otp_string",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, VerifyOtp, "Verify OTP")
		})
	}
}
