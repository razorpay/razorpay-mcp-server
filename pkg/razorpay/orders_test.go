package razorpay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

	tests := []struct {
		name           string
		requestArgs    map[string]interface{}
		mockHttpClient func() (*http.Client, *httptest.Server)
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful order creation",
			requestArgs: map[string]interface{}{
				"amount":   float64(10000),
				"currency": "INR",
				"receipt":  "receipt-123",
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   createOrderPath,
						Method: "POST",
						Response: map[string]interface{}{
							"id":       "order_EKwxwAgItmmXdp",
							"amount":   float64(10000),
							"currency": "INR",
							"receipt":  "receipt-123",
							"status":   "created",
						},
					},
				)
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":       "order_EKwxwAgItmmXdp",
				"amount":   float64(10000),
				"currency": "INR",
				"receipt":  "receipt-123",
				"status":   "created",
			},
		},
		{
			name: "order with notes",
			requestArgs: map[string]interface{}{
				"amount":   float64(10000),
				"currency": "INR",
				"notes": map[string]interface{}{
					"customer_name": "test-customer",
					"product_name":  "test-product",
				},
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   createOrderPath,
						Method: "POST",
						Response: map[string]interface{}{
							"id":       "order_EKwxwAgItmmXdp",
							"amount":   float64(10000),
							"currency": "INR",
							"receipt":  "receipt-123",
							"notes": map[string]interface{}{
								"customer_name": "test-customer",
								"product_name":  "test-product",
							},
							"status": "created",
						},
					},
				)
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":       "order_EKwxwAgItmmXdp",
				"amount":   float64(10000),
				"currency": "INR",
				"receipt":  "receipt-123",
				"notes": map[string]interface{}{
					"customer_name": "test-customer",
					"product_name":  "test-product",
				},
				"status": "created",
			},
		},
		{
			name: "order with partial payment",
			requestArgs: map[string]interface{}{
				"amount":                   float64(10000),
				"currency":                 "INR",
				"partial_payment":          true,
				"first_payment_min_amount": float64(5000),
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   createOrderPath,
						Method: "POST",
						Response: map[string]interface{}{
							"id":                       "order_EKwxwAgItmmXdp",
							"amount":                   float64(10000),
							"currency":                 "INR",
							"partial_payment":          true,
							"first_payment_min_amount": float64(5000),
							"status":                   "created",
						},
					},
				)
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":                       "order_EKwxwAgItmmXdp",
				"amount":                   float64(10000),
				"currency":                 "INR",
				"partial_payment":          true,
				"first_payment_min_amount": float64(5000),
				"status":                   "created",
			},
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
						Path:   createOrderPath,
						Method: "POST",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "Razorpay API error: Bad request",
							},
						},
					},
				)
			},
			expectError:    true,
			expectedErrMsg: "creating order failed: Razorpay API error: Bad request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockrzpClient, mockServer := newRzpMockClient(tc.mockHttpClient)
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

			for key, expected := range tc.expectedResult {
				if key == "notes" ||
					key == "partial_payment" ||
					key == "first_payment_min_amount" {
					continue
				}
				assert.Equal(
					t,
					expected,
					returnedOrder[key],
					"Field %s doesn't match",
					key,
				)
			}

			if notes, ok := tc.expectedResult["notes"].(map[string]interface{}); ok {
				returnedNotes, hasNotes := returnedOrder["notes"].(map[string]interface{})
				require.True(t, hasNotes, "Expected notes in response")

				for noteKey, noteVal := range notes {
					assert.Equal(
						t,
						noteVal,
						returnedNotes[noteKey],
						"Note %s doesn't match",
						noteKey,
					)
				}
			}

			if pp, ok := tc.expectedResult["partial_payment"]; ok {
				assert.Equal(
					t,
					pp,
					returnedOrder["partial_payment"],
					"partial_payment field doesn't match",
				)
			}

			if minAmount, ok := tc.expectedResult["first_payment_min_amount"]; ok {
				assert.Equal(
					t,
					minAmount,
					returnedOrder["first_payment_min_amount"],
					"first_payment_min_amount field doesn't match",
				)
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
						Path:   fmt.Sprintf(fetchOrderPathFmt, "order_EKwxwAgItmmXdp"),
						Method: "GET",
						Response: map[string]interface{}{
							"id":       "order_EKwxwAgItmmXdp",
							"amount":   float64(10000),
							"currency": "INR",
							"receipt":  "receipt-123",
							"status":   "created",
						},
					},
				)
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":       "order_EKwxwAgItmmXdp",
				"amount":   float64(10000),
				"currency": "INR",
				"receipt":  "receipt-123",
				"status":   "created",
			},
		},
		{
			name: "order not found",
			requestArgs: map[string]interface{}{
				"order_id": "order_invalid",
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   fmt.Sprintf(fetchOrderPathFmt, "order_invalid"),
						Method: "GET",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "order not found",
							},
						},
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
			mockrzpClient, mockServer := newRzpMockClient(tc.mockHttpClient)
			if mockServer != nil {
				defer mockServer.Close()
			}

			log := CreateTestLogger()
			tool := FetchOrder(log, mockrzpClient)

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

			for key, expected := range tc.expectedResult {
				assert.Equal(t, expected, returnedOrder[key], "Field %s doesn't match", key)
			}
		})
	}
}
