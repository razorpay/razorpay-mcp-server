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

func Test_FetchPayment(t *testing.T) {
	// Create test cases
	tests := []struct {
		name           string
		paymentID      string
		setupMock      func(client *mocks.PaymentClient)
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name:      "successful payment fetch",
			paymentID: "pay_123456789",
			setupMock: func(mock *mocks.PaymentClient) {
				mock.FetchFunc = func(
					paymentID string,
					_ map[string]interface{},
					_ map[string]string,
				) (map[string]interface{}, error) {
					assert.Equal(t, "pay_123456789", paymentID)
					return map[string]interface{}{
						"id":     "pay_123456789",
						"amount": float64(1000),
						"status": "captured",
					}, nil
				}
			},
			requestArgs: map[string]interface{}{
				"payment_id": "pay_123456789",
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":     "pay_123456789",
				"amount": float64(1000),
				"status": "captured",
			},
		},
		{
			name:      "payment not found",
			paymentID: "pay_invalid",
			setupMock: func(mock *mocks.PaymentClient) {
				mock.FetchFunc = func(
					paymentID string,
					_ map[string]interface{},
					_ map[string]string,
				) (map[string]interface{}, error) {
					return nil, fmt.Errorf("payment not found")
				}
			},
			requestArgs: map[string]interface{}{
				"payment_id": "pay_invalid",
			},
			expectError:    true,
			expectedErrMsg: "fetching payment failed: payment not found",
		},
		{
			name:      "missing payment_id parameter",
			paymentID: "",
			setupMock: func(mock *mocks.PaymentClient) {
				// No need to set up mock for this case
			},
			requestArgs:    map[string]interface{}{},
			expectError:    true,
			expectedErrMsg: "missing required parameter: payment_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mock client
			mockPayment, tool := SetupFetchPaymentTest()

			if tc.setupMock != nil {
				tc.setupMock(mockPayment)
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
				var returnedPayment map[string]interface{}
				err = json.Unmarshal([]byte(result.Text), &returnedPayment)
				require.NoError(t, err)

				// Verify key fields in the payment
				assert.Equal(t, tc.expectedResult["id"], returnedPayment["id"])
				assert.Equal(t, tc.expectedResult["amount"], returnedPayment["amount"])
				assert.Equal(t, tc.expectedResult["status"], returnedPayment["status"])
			}
		})
	}
}
