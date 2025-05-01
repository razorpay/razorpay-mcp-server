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
