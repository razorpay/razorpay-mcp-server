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

	"github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mocks"
)

func Test_CreateOrder(t *testing.T) {
	createOrderPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.ORDER_URL,
	)

	tests := []struct {
		name           string
		mockHttpClient func() (*http.Client, *httptest.Server)
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful order creation",
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mocks.NewMockedHTTPClient(
					mocks.MockEndpoint{
						Path:   createOrderPath,
						Method: "POST",
						Response: map[string]interface{}{
							"id":       "order_123456789",
							"amount":   float64(10000),
							"currency": "INR",
							"receipt":  "receipt-123",
							"status":   "created",
						},
					},
				)
			},
			requestArgs: map[string]interface{}{
				"amount":   float64(10000),
				"currency": "INR",
				"receipt":  "receipt-123",
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":       "order_123456789",
				"amount":   float64(10000),
				"currency": "INR",
				"receipt":  "receipt-123",
				"status":   "created",
			},
		},
		{
			name: "order with notes",
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mocks.NewMockedHTTPClient(
					mocks.MockEndpoint{
						Path:   createOrderPath,
						Method: "POST",
						Response: map[string]interface{}{
							"id":       "order_123456789",
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
			requestArgs: map[string]interface{}{
				"amount":   float64(10000),
				"currency": "INR",
				"notes": map[string]interface{}{
					"customer_name": "test-customer",
					"product_name":  "test-product",
				},
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":       "order_123456789",
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
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mocks.NewMockedHTTPClient(
					mocks.MockEndpoint{
						Path:   createOrderPath,
						Method: "POST",
						Response: map[string]interface{}{
							"id":                       "order_123456789",
							"amount":                   float64(10000),
							"currency":                 "INR",
							"partial_payment":          true,
							"first_payment_min_amount": float64(5000),
							"status":                   "created",
						},
					},
				)
			},
			requestArgs: map[string]interface{}{
				"amount":                   float64(10000),
				"currency":                 "INR",
				"partial_payment":          true,
				"first_payment_min_amount": float64(5000),
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":                       "order_123456789",
				"amount":                   float64(10000),
				"currency":                 "INR",
				"partial_payment":          true,
				"first_payment_min_amount": float64(5000),
				"status":                   "created",
			},
		},
		{
			name:           "missing required parameters",
			mockHttpClient: nil, // No HTTP client needed for validation error
			requestArgs: map[string]interface{}{
				"amount": float64(10000),
				// Missing currency
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: currency",
		},
		{
			name: "order creation fails",
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mocks.NewMockedHTTPClient(
					mocks.MockEndpoint{
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
			requestArgs: map[string]interface{}{
				"amount":   float64(10000),
				"currency": "INR",
			},
			expectError:    true,
			expectedErrMsg: "creating order failed: Razorpay API error: Bad request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockrzpClient := razorpay.NewClient("sample_key", "sample_secret")

			var mockServer *httptest.Server
			if tc.mockHttpClient != nil {
				var client *http.Client
				client, mockServer = tc.mockHttpClient()
				defer mockServer.Close()

				mockrzpClient.Order.Request.BaseURL = mockServer.URL
				mockrzpClient.Order.Request.HTTPClient = client
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
		mockHttpClient func() (*http.Client, *httptest.Server)
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful order fetch",
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mocks.NewMockedHTTPClient(
					mocks.MockEndpoint{
						Path:   fmt.Sprintf(fetchOrderPathFmt, "order_123456789"),
						Method: "GET",
						Response: map[string]interface{}{
							"id":       "order_123456789",
							"amount":   float64(10000),
							"currency": "INR",
							"receipt":  "receipt-123",
							"status":   "created",
						},
					},
				)
			},
			requestArgs: map[string]interface{}{
				"order_id": "order_123456789",
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":       "order_123456789",
				"amount":   float64(10000),
				"currency": "INR",
				"receipt":  "receipt-123",
				"status":   "created",
			},
		},
		{
			name: "order not found",
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mocks.NewMockedHTTPClient(
					mocks.MockEndpoint{
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
			requestArgs: map[string]interface{}{
				"order_id": "order_invalid",
			},
			expectError:    true,
			expectedErrMsg: "fetching order failed: order not found",
		},
		{
			name:           "missing order_id parameter",
			mockHttpClient: nil, // No HTTP client needed for validation error
			requestArgs:    map[string]interface{}{},
			expectError:    true,
			expectedErrMsg: "missing required parameter: order_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockrzpClient := razorpay.NewClient("sample_key", "sample_secret")

			var mockServer *httptest.Server
			if tc.mockHttpClient != nil {
				var client *http.Client
				client, mockServer = tc.mockHttpClient()
				defer mockServer.Close()

				mockrzpClient.Order.Request.BaseURL = mockServer.URL
				mockrzpClient.Order.Request.HTTPClient = client
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
