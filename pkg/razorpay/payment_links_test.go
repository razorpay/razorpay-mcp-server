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

func Test_CreatePaymentLink(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name           string
		setupMock      func(client *mocks.PaymentLinkClient)
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful payment link creation",
			setupMock: func(mock *mocks.PaymentLinkClient) {
				mock.CreateFunc = func(
					data map[string]interface{},
					_ map[string]string,
				) (map[string]interface{}, error) {
					// Validate that required fields are present
					assert.Equal(t, 50000, data["amount"])
					assert.Equal(t, "INR", data["currency"])
					assert.Equal(t, "Test payment", data["description"])

					return map[string]interface{}{
						"id":          "plink_123456789",
						"amount":      float64(50000),
						"currency":    "INR",
						"description": "Test payment",
						"status":      "created",
						"short_url":   "https://rzp.io/i/abcdef",
					}, nil
				}
			},
			requestArgs: map[string]interface{}{
				"amount":      float64(50000),
				"currency":    "INR",
				"description": "Test payment",
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":          "plink_123456789",
				"amount":      float64(50000),
				"currency":    "INR",
				"description": "Test payment",
				"status":      "created",
				"short_url":   "https://rzp.io/i/abcdef",
			},
		},
		{
			name: "payment link without description",
			setupMock: func(mock *mocks.PaymentLinkClient) {
				mock.CreateFunc = func(
					data map[string]interface{},
					_ map[string]string,
				) (map[string]interface{}, error) {
					// Validate that only required fields are present
					assert.Equal(t, 50000, data["amount"])
					assert.Equal(t, "INR", data["currency"])
					_, descExists := data["description"]
					assert.False(t, descExists)

					return map[string]interface{}{
						"id":        "plink_123456789",
						"amount":    float64(50000),
						"currency":  "INR",
						"status":    "created",
						"short_url": "https://rzp.io/i/abcdef",
					}, nil
				}
			},
			requestArgs: map[string]interface{}{
				"amount":   float64(50000),
				"currency": "INR",
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":        "plink_123456789",
				"amount":    float64(50000),
				"currency":  "INR",
				"status":    "created",
				"short_url": "https://rzp.io/i/abcdef",
			},
		},
		{
			name: "missing amount parameter",
			setupMock: func(mock *mocks.PaymentLinkClient) {
				// No need to setup mock for validation error
			},
			requestArgs: map[string]interface{}{
				"currency": "INR",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: amount",
		},
		{
			name: "missing currency parameter",
			setupMock: func(mock *mocks.PaymentLinkClient) {
				// No need to setup mock for validation error
			},
			requestArgs: map[string]interface{}{
				"amount": float64(50000),
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: currency",
		},
		{
			name: "payment link creation fails",
			setupMock: func(mock *mocks.PaymentLinkClient) {
				mock.CreateFunc = func(
					data map[string]interface{},
					_ map[string]string,
				) (map[string]interface{}, error) {
					return nil, fmt.Errorf("API error: Invalid currency")
				}
			},
			requestArgs: map[string]interface{}{
				"amount":   float64(50000),
				"currency": "XYZ", // Invalid currency
			},
			expectError:    true,
			expectedErrMsg: "creating payment link failed: API error: Invalid currency",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mock client and tool
			mockClient, tool := SetupCreatePaymentLinkTest()

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
				var returnedPaymentLink map[string]interface{}
				err = json.Unmarshal([]byte(result.Text), &returnedPaymentLink)
				require.NoError(t, err)

				// Verify key fields in the payment link
				assert.Equal(t, tc.expectedResult["id"], returnedPaymentLink["id"])
				assert.Equal(t, tc.expectedResult["amount"], returnedPaymentLink["amount"])
				assert.Equal(
					t,
					tc.expectedResult["currency"],
					returnedPaymentLink["currency"],
				)
				assert.Equal(
					t,
					tc.expectedResult["status"],
					returnedPaymentLink["status"],
				)
				assert.Equal(
					t,
					tc.expectedResult["short_url"],
					returnedPaymentLink["short_url"],
				)

				// Check description if it exists
				if desc, ok := tc.expectedResult["description"]; ok {
					assert.Equal(t, desc, returnedPaymentLink["description"])
				}
			}
		})
	}
}

func Test_FetchPaymentLink(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name           string
		setupMock      func(client *mocks.PaymentLinkClient)
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful payment link fetch",
			setupMock: func(mock *mocks.PaymentLinkClient) {
				mock.FetchFunc = func(
					id string,
					data map[string]interface{},
					options map[string]string,
				) (map[string]interface{}, error) {
					assert.Equal(t, "plink_123456789", id)
					return map[string]interface{}{
						"id":          "plink_123456789",
						"amount":      float64(50000),
						"currency":    "INR",
						"description": "Test payment",
						"status":      "paid",
						"short_url":   "https://rzp.io/i/abcdef",
					}, nil
				}
			},
			requestArgs: map[string]interface{}{
				"payment_link_id": "plink_123456789",
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":          "plink_123456789",
				"amount":      float64(50000),
				"currency":    "INR",
				"description": "Test payment",
				"status":      "paid",
				"short_url":   "https://rzp.io/i/abcdef",
			},
		},
		{
			name: "payment link not found",
			setupMock: func(mock *mocks.PaymentLinkClient) {
				mock.FetchFunc = func(
					id string,
					data map[string]interface{},
					options map[string]string,
				) (map[string]interface{}, error) {
					return nil, fmt.Errorf("payment link not found")
				}
			},
			requestArgs: map[string]interface{}{
				"payment_link_id": "plink_invalid",
			},
			expectError:    true,
			expectedErrMsg: "fetching payment link failed: payment link not found",
		},
		{
			name:           "missing payment_link_id parameter",
			setupMock:      func(mock *mocks.PaymentLinkClient) {},
			requestArgs:    map[string]interface{}{},
			expectError:    true,
			expectedErrMsg: "missing required parameter: payment_link_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mock client and tool
			mockClient, tool := SetupFetchPaymentLinkTest()

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
				var returnedPaymentLink map[string]interface{}
				err = json.Unmarshal([]byte(result.Text), &returnedPaymentLink)
				require.NoError(t, err)

				// Verify key fields in the payment link
				assert.Equal(t, tc.expectedResult["id"], returnedPaymentLink["id"])
				assert.Equal(t, tc.expectedResult["amount"], returnedPaymentLink["amount"])
				assert.Equal(
					t,
					tc.expectedResult["currency"],
					returnedPaymentLink["currency"],
				)
				assert.Equal(
					t,
					tc.expectedResult["status"],
					returnedPaymentLink["status"],
				)
				assert.Equal(
					t,
					tc.expectedResult["short_url"],
					returnedPaymentLink["short_url"],
				)

				// Check description if it exists
				if desc, ok := tc.expectedResult["description"]; ok {
					assert.Equal(t, desc, returnedPaymentLink["description"])
				}
			}
		})
	}
}
