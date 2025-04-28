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
			Name: "missing required parameters",
			Request: map[string]interface{}{
				"amount": float64(10000),
				// Missing currency
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: currency",
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
