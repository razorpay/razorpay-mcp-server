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
	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_FetchPayment(t *testing.T) {
	fetchPaymentPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
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
			name: "successful payment fetch",
			requestArgs: map[string]interface{}{
				"payment_id": "pay_MT48CvBhIC98MQ",
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   fmt.Sprintf(fetchPaymentPathFmt, "pay_MT48CvBhIC98MQ"),
						Method: "GET",
						Response: map[string]interface{}{
							"id":     "pay_MT48CvBhIC98MQ",
							"amount": float64(1000),
							"status": "captured",
						},
					},
				)
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":     "pay_MT48CvBhIC98MQ",
				"amount": float64(1000),
				"status": "captured",
			},
		},
		{
			name: "payment not found",
			requestArgs: map[string]interface{}{
				"payment_id": "pay_invalid",
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   fmt.Sprintf(fetchPaymentPathFmt, "pay_invalid"),
						Method: "GET",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "payment not found",
							},
						},
					},
				)
			},
			expectError:    true,
			expectedErrMsg: "fetching payment failed: payment not found",
		},
		{
			name:           "missing payment_id parameter",
			requestArgs:    map[string]interface{}{},
			mockHttpClient: nil, // No HTTP client needed for validation error
			expectError:    true,
			expectedErrMsg: "missing required parameter: payment_id",
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

				mockrzpClient.Payment.Request.BaseURL = mockServer.URL
				mockrzpClient.Payment.Request.HTTPClient = client
			}

			log := CreateTestLogger()
			tool := FetchPayment(log, mockrzpClient)

			request := createMCPRequest(tc.requestArgs)

			result, err := tool.GetHandler()(context.Background(), request)

			if tc.expectError {
				require.NotNil(t, result)
				assert.Contains(t, result.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			var returnedPayment map[string]interface{}
			err = json.Unmarshal([]byte(result.Text), &returnedPayment)
			require.NoError(t, err)

			for key, expected := range tc.expectedResult {
				assert.Equal(
					t,
					expected,
					returnedPayment[key],
					"Field %s doesn't match",
					key,
				)
			}
		})
	}
}
