package razorpay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

	tests := []struct {
		name           string
		requestArgs    map[string]interface{}
		mockHttpClient func() (*http.Client, *httptest.Server)
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful order creation with all parameters",
			requestArgs: map[string]interface{}{
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
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createOrderPath,
						Method:   "POST",
						Response: orderWithAllParamsResp,
					},
				)
			},
			expectError:    false,
			expectedResult: orderWithAllParamsResp,
		},
		{
			name: "successful order creation with required params only",
			requestArgs: map[string]interface{}{
				"amount":   float64(10000),
				"currency": "INR",
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createOrderPath,
						Method:   "POST",
						Response: orderWithRequiredParamsResp,
					},
				)
			},
			expectError:    false,
			expectedResult: orderWithRequiredParamsResp,
		},
		{
			name: "missing required parameters",
			requestArgs: map[string]interface{}{
				"amount": float64(10000),
				// Missing currency
			},
			mockHttpClient: nil, // No HTTP client needed for validation error
			expectError:    true,
			expectedErrMsg: "missing required parameter: currency",
		},
		{
			name: "order creation fails",
			requestArgs: map[string]interface{}{
				"amount":   float64(10000),
				"currency": "INR",
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createOrderPath,
						Method:   "POST",
						Response: errorResp,
					},
				)
			},
			expectError:    true,
			expectedErrMsg: "creating order failed: Razorpay API error: Bad request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockrzpClient, mockServer := newMockRzpClient(tc.mockHttpClient)
			if mockServer != nil {
				defer mockServer.Close()
			}

			log := CreateTestLogger()
			tool := CreateOrder(log, mockrzpClient)

			request := createMCPRequest(tc.requestArgs)

			result, err := tool.GetHandler()(context.Background(), request)

			if tc.expectError {
				require.NotNil(t, result)
				assert.Contains(t, result.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			var returnedOrder map[string]interface{}
			err = json.Unmarshal([]byte(result.Text), &returnedOrder)
			require.NoError(t, err)

			if diff := deep.Equal(tc.expectedResult, returnedOrder); diff != nil {
				t.Errorf("Order mismatch: %s", diff)
			}
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

	tests := []struct {
		name           string
		requestArgs    map[string]interface{}
		mockHttpClient func() (*http.Client, *httptest.Server)
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful order fetch",
			requestArgs: map[string]interface{}{
				"order_id": "order_EKwxwAgItmmXdp",
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchOrderPathFmt, "order_EKwxwAgItmmXdp"),
						Method:   "GET",
						Response: orderResp,
					},
				)
			},
			expectError:    false,
			expectedResult: orderResp,
		},
		{
			name: "order not found",
			requestArgs: map[string]interface{}{
				"order_id": "order_invalid",
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchOrderPathFmt, "order_invalid"),
						Method:   "GET",
						Response: orderNotFoundResp,
					},
				)
			},
			expectError:    true,
			expectedErrMsg: "fetching order failed: order not found",
		},
		{
			name:           "missing order_id parameter",
			requestArgs:    map[string]interface{}{},
			mockHttpClient: nil, // No HTTP client needed for validation error
			expectError:    true,
			expectedErrMsg: "missing required parameter: order_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRzpClient, mockServer := newMockRzpClient(tc.mockHttpClient)
			if mockServer != nil {
				defer mockServer.Close()
			}

			log := CreateTestLogger()
			tool := FetchOrder(log, mockRzpClient)

			request := createMCPRequest(tc.requestArgs)

			result, err := tool.GetHandler()(context.Background(), request)

			if tc.expectError {
				require.NotNil(t, result)
				assert.Contains(t, result.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			var returnedOrder map[string]interface{}
			err = json.Unmarshal([]byte(result.Text), &returnedOrder)
			require.NoError(t, err)

			if diff := deep.Equal(tc.expectedResult, returnedOrder); diff != nil {
				t.Errorf("Order mismatch: %s", diff)
			}
		})
	}
}
