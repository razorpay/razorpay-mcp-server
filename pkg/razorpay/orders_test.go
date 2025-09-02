package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_CreateOrder(t *testing.T) {
	createOrderPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.ORDER_URL,
	)

	// Define common response maps to be reused
	orderWithAllParamsResp := map[string]interface{}{
		"id":                       "order_EKwxwAgItmmXdp",
		"amount":                   float64(10000),
		"currency":                 "INR",
		"receipt":                  "receipt-123",
		"partial_payment":          true,
		"first_payment_min_amount": float64(5000),
		"notes": map[string]interface{}{
			"customer_name": "test-customer",
			"product_name":  "test-product",
		},
		"status": "created",
	}

	orderWithRequiredParamsResp := map[string]interface{}{
		"id":       "order_EKwxwAgItmmXdp",
		"amount":   float64(10000),
		"currency": "INR",
		"status":   "created",
	}

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Razorpay API error: Bad request",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful order creation with all parameters",
			Request: map[string]interface{}{
				"amount":                   float64(10000),
				"currency":                 "INR",
				"receipt":                  "receipt-123",
				"partial_payment":          true,
				"first_payment_min_amount": float64(5000),
				"notes": map[string]interface{}{
					"customer_name": "test-customer",
					"product_name":  "test-product",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createOrderPath,
						Method:   "POST",
						Response: orderWithAllParamsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: orderWithAllParamsResp,
		},
		{
			Name: "successful order creation with required params only",
			Request: map[string]interface{}{
				"amount":   float64(10000),
				"currency": "INR",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createOrderPath,
						Method:   "POST",
						Response: orderWithRequiredParamsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: orderWithRequiredParamsResp,
		},
		{
			Name: "successful order creation with transfers",
			Request: map[string]interface{}{
				"amount":   float64(100000),
				"currency": "INR",
				"transfers": []interface{}{
					map[string]interface{}{
						"account":  "acc_KD088uBnQf0XeK",
						"amount":   float64(80000),
						"currency": "INR",
						"notes": map[string]interface{}{
							"account": "Seller Account",
						},
						"linked_account_notes": []interface{}{"account"},
					},
					map[string]interface{}{
						"account":  "acc_KD097GTiIju5WG",
						"amount":   float64(20000),
						"currency": "INR",
						"on_hold":  true,
						"notes": map[string]interface{}{
							"account": "Platform Fee Account",
						},
						"linked_account_notes": []interface{}{"account"},
					},
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				orderWithTransfersResp := map[string]interface{}{
					"id":       "order_EKwxwAgItmmXdp",
					"amount":   float64(100000),
					"currency": "INR",
					"status":   "created",
					"transfers": []interface{}{
						map[string]interface{}{
							"account":  "acc_KD088uBnQf0XeK",
							"amount":   float64(80000),
							"currency": "INR",
						},
						map[string]interface{}{
							"account":  "acc_KD097GTiIju5WG",
							"amount":   float64(20000),
							"currency": "INR",
							"on_hold":  true,
						},
					},
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createOrderPath,
						Method:   "POST",
						Response: orderWithTransfersResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"id":       "order_EKwxwAgItmmXdp",
				"amount":   float64(100000),
				"currency": "INR",
				"status":   "created",
				"transfers": []interface{}{
					map[string]interface{}{
						"account":  "acc_KD088uBnQf0XeK",
						"amount":   float64(80000),
						"currency": "INR",
					},
					map[string]interface{}{
						"account":  "acc_KD097GTiIju5WG",
						"amount":   float64(20000),
						"currency": "INR",
						"on_hold":  true,
					},
				},
			},
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				// Missing both amount and currency (required parameters)
				"partial_payment":          "invalid_boolean", // Wrong type for boolean
				"first_payment_min_amount": "invalid_number",  // Wrong type for number
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:\n- " +
				"missing required parameter: amount\n- " +
				"missing required parameter: currency\n- " +
				"invalid parameter type: partial_payment",
		},
		{
			Name: "first_payment_min_amount validation when partial_payment is true",
			Request: map[string]interface{}{
				"amount":                   float64(10000),
				"currency":                 "INR",
				"partial_payment":          true,
				"first_payment_min_amount": "invalid_number",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:\n- " +
				"invalid parameter type: first_payment_min_amount",
		},
		{
			Name: "order creation fails",
			Request: map[string]interface{}{
				"amount":   float64(10000),
				"currency": "INR",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createOrderPath,
						Method:   "POST",
						Response: errorResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "creating order failed: Razorpay API error: Bad request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CreateOrder, "Order")
		})
	}
}

func Test_FetchOrder(t *testing.T) {
	fetchOrderPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.ORDER_URL,
	)

	orderResp := map[string]interface{}{
		"id":       "order_EKwxwAgItmmXdp",
		"amount":   float64(10000),
		"currency": "INR",
		"receipt":  "receipt-123",
		"status":   "created",
	}

	orderNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "order not found",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful order fetch",
			Request: map[string]interface{}{
				"order_id": "order_EKwxwAgItmmXdp",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchOrderPathFmt, "order_EKwxwAgItmmXdp"),
						Method:   "GET",
						Response: orderResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: orderResp,
		},
		{
			Name: "order not found",
			Request: map[string]interface{}{
				"order_id": "order_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchOrderPathFmt, "order_invalid"),
						Method:   "GET",
						Response: orderNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching order failed: order not found",
		},
		{
			Name:           "missing order_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: order_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchOrder, "Order")
		})
	}
}

func Test_FetchAllOrders(t *testing.T) {
	fetchAllOrdersPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.ORDER_URL,
	)

	// Define the sample response for all orders
	ordersResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":          "order_EKzX2WiEWbMxmx",
				"entity":      "order",
				"amount":      float64(1234),
				"amount_paid": float64(0),
				"amount_due":  float64(1234),
				"currency":    "INR",
				"receipt":     "Receipt No. 1",
				"offer_id":    nil,
				"status":      "created",
				"attempts":    float64(0),
				"notes":       []interface{}{},
				"created_at":  float64(1582637108),
			},
			map[string]interface{}{
				"id":          "order_EAI5nRfThga2TU",
				"entity":      "order",
				"amount":      float64(100),
				"amount_paid": float64(0),
				"amount_due":  float64(100),
				"currency":    "INR",
				"receipt":     "Receipt No. 1",
				"offer_id":    nil,
				"status":      "created",
				"attempts":    float64(0),
				"notes":       []interface{}{},
				"created_at":  float64(1580300731),
			},
		},
	}

	// Define error response
	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Razorpay API error: Bad request",
		},
	}

	// Define the test cases
	tests := []RazorpayToolTestCase{
		{
			Name:    "successful fetch all orders with no parameters",
			Request: map[string]interface{}{},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllOrdersPath,
						Method:   "GET",
						Response: ordersResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: ordersResp,
		},
		{
			Name: "successful fetch all orders with pagination",
			Request: map[string]interface{}{
				"count": 2,
				"skip":  1,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllOrdersPath,
						Method:   "GET",
						Response: ordersResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: ordersResp,
		},
		{
			Name: "successful fetch all orders with time range",
			Request: map[string]interface{}{
				"from": 1580000000,
				"to":   1590000000,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllOrdersPath,
						Method:   "GET",
						Response: ordersResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: ordersResp,
		},
		{
			Name: "successful fetch all orders with filtering",
			Request: map[string]interface{}{
				"authorized": 1,
				"receipt":    "Receipt No. 1",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllOrdersPath,
						Method:   "GET",
						Response: ordersResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: ordersResp,
		},
		{
			Name: "successful fetch all orders with expand",
			Request: map[string]interface{}{
				"expand": []interface{}{"payments"},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllOrdersPath,
						Method:   "GET",
						Response: ordersResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: ordersResp,
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				"count":  "not-a-number",
				"skip":   "not-a-number",
				"from":   "not-a-number",
				"to":     "not-a-number",
				"expand": "not-an-array",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:\n- " +
				"invalid parameter type: count\n- " +
				"invalid parameter type: skip\n- " +
				"invalid parameter type: from\n- " +
				"invalid parameter type: to\n- " +
				"invalid parameter type: expand",
		},
		{
			Name: "fetch all orders fails",
			Request: map[string]interface{}{
				"count": 100,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fetchAllOrdersPath,
						Method:   "GET",
						Response: errorResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching orders failed: Razorpay API error: Bad request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchAllOrders, "Order")
		})
	}
}

func Test_FetchOrderPayments(t *testing.T) {
	fetchOrderPaymentsPathFmt := fmt.Sprintf(
		"/%s%s/%%s/payments",
		constants.VERSION_V1,
		constants.ORDER_URL,
	)

	// Define the sample response for order payments
	paymentsResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":              "pay_N8FUmetkCE2hZP",
				"entity":          "payment",
				"amount":          float64(100),
				"currency":        "INR",
				"status":          "failed",
				"order_id":        "order_N8FRN5zTm5S3wx",
				"invoice_id":      nil,
				"international":   false,
				"method":          "upi",
				"amount_refunded": float64(0),
				"refund_status":   nil,
				"captured":        false,
				"description":     nil,
				"card_id":         nil,
				"bank":            nil,
				"wallet":          nil,
				"vpa":             "failure@razorpay",
				"email":           "void@razorpay.com",
				"contact":         "+919999999999",
				"notes": map[string]interface{}{
					"notes_key_1": "Tea, Earl Grey, Hot",
					"notes_key_2": "Tea, Earl Grey… decaf.",
				},
				"fee":               nil,
				"tax":               nil,
				"error_code":        "BAD_REQUEST_ERROR",
				"error_description": "Payment was unsuccessful due to a temporary issue.",
				"error_source":      "gateway",
				"error_step":        "payment_response",
				"error_reason":      "payment_failed",
				"acquirer_data": map[string]interface{}{
					"rrn": nil,
				},
				"created_at": float64(1701688684),
				"upi": map[string]interface{}{
					"vpa": "failure@razorpay",
				},
			},
			map[string]interface{}{
				"id":              "pay_N8FVRD1DzYzBh1",
				"entity":          "payment",
				"amount":          float64(100),
				"currency":        "INR",
				"status":          "captured",
				"order_id":        "order_N8FRN5zTm5S3wx",
				"invoice_id":      nil,
				"international":   false,
				"method":          "upi",
				"amount_refunded": float64(0),
				"refund_status":   nil,
				"captured":        true,
				"description":     nil,
				"card_id":         nil,
				"bank":            nil,
				"wallet":          nil,
				"vpa":             "success@razorpay",
				"email":           "void@razorpay.com",
				"contact":         "+919999999999",
				"notes": map[string]interface{}{
					"notes_key_1": "Tea, Earl Grey, Hot",
					"notes_key_2": "Tea, Earl Grey… decaf.",
				},
				"fee":               float64(2),
				"tax":               float64(0),
				"error_code":        nil,
				"error_description": nil,
				"error_source":      nil,
				"error_step":        nil,
				"error_reason":      nil,
				"acquirer_data": map[string]interface{}{
					"rrn":                "267567962619",
					"upi_transaction_id": "F5B66C7C07CA6FEAD77E956DC2FC7ABE",
				},
				"created_at": float64(1701688721),
				"upi": map[string]interface{}{
					"vpa": "success@razorpay",
				},
			},
		},
	}

	orderNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "order not found",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful fetch of order payments",
			Request: map[string]interface{}{
				"order_id": "order_N8FRN5zTm5S3wx",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							fetchOrderPaymentsPathFmt,
							"order_N8FRN5zTm5S3wx",
						),
						Method:   "GET",
						Response: paymentsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: paymentsResp,
		},
		{
			Name: "order not found",
			Request: map[string]interface{}{
				"order_id": "order_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							fetchOrderPaymentsPathFmt,
							"order_invalid",
						),
						Method:   "GET",
						Response: orderNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching payments for order failed: order not found",
		},
		{
			Name:           "missing order_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: order_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchOrderPayments, "Order")
		})
	}
}

func Test_UpdateOrder(t *testing.T) {
	updateOrderPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.ORDER_URL,
	)

	updatedOrderResp := map[string]interface{}{
		"id":         "order_EKwxwAgItmmXdp",
		"entity":     "order",
		"amount":     float64(10000),
		"currency":   "INR",
		"receipt":    "receipt-123",
		"status":     "created",
		"attempts":   float64(0),
		"created_at": float64(1572505143),
		"notes": map[string]interface{}{
			"customer_name": "updated-customer",
			"product_name":  "updated-product",
		},
	}

	orderNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "order not found",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful order update",
			Request: map[string]interface{}{
				"order_id": "order_EKwxwAgItmmXdp",
				"notes": map[string]interface{}{
					"customer_name": "updated-customer",
					"product_name":  "updated-product",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							updateOrderPathFmt, "order_EKwxwAgItmmXdp"),
						Method:   "PATCH",
						Response: updatedOrderResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: updatedOrderResp,
		},
		{
			Name: "missing required parameters - order_id",
			Request: map[string]interface{}{
				// Missing order_id
				"notes": map[string]interface{}{
					"customer_name": "updated-customer",
					"product_name":  "updated-product",
				},
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: order_id",
		},
		{
			Name: "missing required parameters - notes",
			Request: map[string]interface{}{
				"order_id": "order_EKwxwAgItmmXdp",
				// Missing notes
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: notes",
		},
		{
			Name: "order not found",
			Request: map[string]interface{}{
				"order_id": "order_invalid_id",
				"notes": map[string]interface{}{
					"customer_name": "updated-customer",
					"product_name":  "updated-product",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(updateOrderPathFmt, "order_invalid_id"),
						Method:   "PATCH",
						Response: orderNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "updating order failed: order not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, UpdateOrder, "Order")
		})
	}
}
