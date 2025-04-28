package razorpay

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mocks"
)

func Test_CreateOrder(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name           string
		setupMock      func(client *mocks.OrderClient)
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful order creation",
			setupMock: func(mock *mocks.OrderClient) {
				mock.CreateFunc = func(
					data map[string]interface{},
					_ map[string]string,
				) (map[string]interface{}, error) {
					// Validate that required fields are present
					assert.Equal(t, 10000, data["amount"])
					assert.Equal(t, "INR", data["currency"])
					assert.Equal(t, "receipt-123", data["receipt"])

					return map[string]interface{}{
						"id":       "order_123456789",
						"amount":   float64(10000),
						"currency": "INR",
						"receipt":  "receipt-123",
						"status":   "created",
					}, nil
				}
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
			setupMock: func(mock *mocks.OrderClient) {
				mock.CreateFunc = func(
					data map[string]interface{},
					_ map[string]string,
				) (map[string]interface{}, error) {
					// Validate that notes are passed correctly
					assert.Equal(t, 10000, data["amount"])
					assert.Equal(t, "INR", data["currency"])

					notesMap := data["notes"].(map[string]interface{})
					assert.Equal(t, "test-customer", notesMap["customer_name"])
					assert.Equal(t, "test-product", notesMap["product_name"])

					return map[string]interface{}{
						"id":       "order_123456789",
						"amount":   float64(10000),
						"currency": "INR",
						"notes": map[string]interface{}{
							"customer_name": "test-customer",
							"product_name":  "test-product",
						},
						"status": "created",
					}, nil
				}
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
				"notes": map[string]interface{}{
					"customer_name": "test-customer",
					"product_name":  "test-product",
				},
				"status": "created",
			},
		},
		{
			name: "order with partial payment",
			setupMock: func(mock *mocks.OrderClient) {
				mock.CreateFunc = func(
					data map[string]interface{},
					_ map[string]string,
				) (map[string]interface{}, error) {
					// Validate partial payment fields
					assert.Equal(t, 10000, data["amount"])
					assert.Equal(t, "INR", data["currency"])
					assert.Equal(t, true, data["partial_payment"])
					assert.Equal(t, 5000, data["first_payment_min_amount"])

					return map[string]interface{}{
						"id":                       "order_123456789",
						"amount":                   float64(10000),
						"currency":                 "INR",
						"partial_payment":          true,
						"first_payment_min_amount": float64(5000),
						"status":                   "created",
					}, nil
				}
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
			name: "missing required parameters",
			setupMock: func(mock *mocks.OrderClient) {
				// No need to setup mock for validation error
			},
			requestArgs: map[string]interface{}{
				"amount": float64(10000),
				// Missing currency
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: currency",
		},
		{
			name: "order creation fails",
			setupMock: func(mock *mocks.OrderClient) {
				mock.CreateFunc = func(
					data map[string]interface{},
					_ map[string]string,
				) (map[string]interface{}, error) {
					return nil, fmt.Errorf("Razorpay API error: Bad request")
				}
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
			// Set up mock client and tool
			mockClient, tool := SetupCreateOrderTest()

			if tc.setupMock != nil {
				tc.setupMock(mockClient)
			}

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := tool.GetHandler()(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.NotNil(t, result)
				assert.Contains(t, result.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tc.expectedResult != nil {
				// Parse the result to verify it contains expected data
				var returnedOrder map[string]interface{}
				err = json.Unmarshal([]byte(result.Text), &returnedOrder)
				require.NoError(t, err)

				// Verify key fields in the order
				assert.Equal(t, tc.expectedResult["id"], returnedOrder["id"])
				assert.Equal(t, tc.expectedResult["amount"], returnedOrder["amount"])
				assert.Equal(t, tc.expectedResult["currency"], returnedOrder["currency"])
				assert.Equal(t, tc.expectedResult["status"], returnedOrder["status"])

				// Check receipt if present
				if receipt, ok := tc.expectedResult["receipt"]; ok {
					assert.Equal(t, receipt, returnedOrder["receipt"])
				}

				// Check notes if present
				if notes, ok := tc.expectedResult["notes"]; ok {
					assert.Equal(t, notes, returnedOrder["notes"])
				}

				// Check partial payment fields if present
				if pp, ok := tc.expectedResult["partial_payment"]; ok {
					assert.Equal(t, pp, returnedOrder["partial_payment"])
				}

				if minAmount, ok := tc.expectedResult["first_payment_min_amount"]; ok {
					assert.Equal(t, minAmount, returnedOrder["first_payment_min_amount"])
				}
			}
		})
	}
}

func Test_FetchOrder(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name           string
		setupMock      func(client *mocks.OrderClient)
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful order fetch",
			setupMock: func(mock *mocks.OrderClient) {
				mock.FetchFunc = func(
					orderID string,
					_ map[string]interface{},
					_ map[string]string,
				) (map[string]interface{}, error) {
					assert.Equal(t, "order_123456789", orderID)
					return map[string]interface{}{
						"id":       "order_123456789",
						"amount":   float64(10000),
						"currency": "INR",
						"receipt":  "receipt-123",
						"status":   "created",
					}, nil
				}
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
			setupMock: func(mock *mocks.OrderClient) {
				mock.FetchFunc = func(
					orderID string,
					_ map[string]interface{},
					_ map[string]string,
				) (map[string]interface{}, error) {
					return nil, fmt.Errorf("order not found")
				}
			},
			requestArgs: map[string]interface{}{
				"order_id": "order_invalid",
			},
			expectError:    true,
			expectedErrMsg: "fetching order failed: order not found",
		},
		{
			name:           "missing order_id parameter",
			setupMock:      func(mock *mocks.OrderClient) {},
			requestArgs:    map[string]interface{}{},
			expectError:    true,
			expectedErrMsg: "missing required parameter: order_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mock client and tool
			mockClient, tool := SetupFetchOrderTest()

			if tc.setupMock != nil {
				tc.setupMock(mockClient)
			}

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := tool.GetHandler()(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.NotNil(t, result)
				assert.Contains(t, result.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tc.expectedResult != nil {
				// Parse the result to verify it contains expected data
				var returnedOrder map[string]interface{}
				err = json.Unmarshal([]byte(result.Text), &returnedOrder)
				require.NoError(t, err)

				// Verify key fields in the order
				assert.Equal(t, tc.expectedResult["id"], returnedOrder["id"])
				assert.Equal(t, tc.expectedResult["amount"], returnedOrder["amount"])
				assert.Equal(t, tc.expectedResult["currency"], returnedOrder["currency"])
				assert.Equal(t, tc.expectedResult["status"], returnedOrder["status"])

				// Check receipt if present
				if receipt, ok := tc.expectedResult["receipt"]; ok {
					assert.Equal(t, receipt, returnedOrder["receipt"])
				}
			}
		})
	}
}
