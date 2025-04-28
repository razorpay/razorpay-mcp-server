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

func Test_CreatePaymentLink(t *testing.T) {
	createPaymentLinkPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.PaymentLink_URL,
	)

	samplePaymentLink := map[string]interface{}{
		"id":          "plink_123456789",
		"amount":      float64(50000),
		"currency":    "INR",
		"description": "Test payment",
		"status":      "created",
		"short_url":   "https://rzp.io/i/abcdef",
	}

	paymentLinkWithoutDesc := map[string]interface{}{
		"id":        "plink_123456789",
		"amount":    float64(50000),
		"currency":  "INR",
		"status":    "created",
		"short_url": "https://rzp.io/i/abcdef",
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "API error: Invalid currency",
		},
	}

	tests := []struct {
		name           string
		mockHttpClient func() (*http.Client, *httptest.Server)
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful payment link creation",
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mocks.NewMockedHTTPClient(
					mocks.MockEndpoint{
						Path:     createPaymentLinkPath,
						Method:   "POST",
						Response: samplePaymentLink,
					},
				)
			},
			requestArgs: map[string]interface{}{
				"amount":      float64(50000),
				"currency":    "INR",
				"description": "Test payment",
			},
			expectError:    false,
			expectedResult: samplePaymentLink,
		},
		{
			name: "payment link without description",
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mocks.NewMockedHTTPClient(
					mocks.MockEndpoint{
						Path:     createPaymentLinkPath,
						Method:   "POST",
						Response: paymentLinkWithoutDesc,
					},
				)
			},
			requestArgs: map[string]interface{}{
				"amount":   float64(50000),
				"currency": "INR",
			},
			expectError:    false,
			expectedResult: paymentLinkWithoutDesc,
		},
		{
			name:           "missing amount parameter",
			mockHttpClient: nil, // No HTTP client needed for validation error
			requestArgs: map[string]interface{}{
				"currency": "INR",
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: amount",
		},
		{
			name:           "missing currency parameter",
			mockHttpClient: nil, // No HTTP client needed for validation error
			requestArgs: map[string]interface{}{
				"amount": float64(50000),
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: currency",
		},
		{
			name: "payment link creation fails",
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mocks.NewMockedHTTPClient(
					mocks.MockEndpoint{
						Path:     createPaymentLinkPath,
						Method:   "POST",
						Response: errorResponse,
					},
				)
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
			mockrzpClient := razorpay.NewClient("sample_key", "sample_secret")

			var mockServer *httptest.Server
			if tc.mockHttpClient != nil {
				var client *http.Client
				client, mockServer = tc.mockHttpClient()
				defer mockServer.Close()

				mockrzpClient.PaymentLink.Request.BaseURL = mockServer.URL
				mockrzpClient.PaymentLink.Request.HTTPClient = client
			}

			log := CreateTestLogger()
			tool := CreatePaymentLink(log, mockrzpClient)

			request := createMCPRequest(tc.requestArgs)

			result, err := tool.GetHandler()(context.Background(), request)

			if tc.expectError {
				require.NotNil(t, result)
				assert.Contains(t, result.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			var returnedPaymentLink map[string]interface{}
			err = json.Unmarshal([]byte(result.Text), &returnedPaymentLink)
			require.NoError(t, err)

			for key, expected := range tc.expectedResult {
				assert.Equal(
					t,
					expected,
					returnedPaymentLink[key],
					"Field %s doesn't match",
					key,
				)
			}
		})
	}
}

func Test_FetchPaymentLink(t *testing.T) {
	fetchPaymentLinkPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.PaymentLink_URL,
	)

	samplePaymentLink := map[string]interface{}{
		"id":          "plink_123456789",
		"amount":      float64(50000),
		"currency":    "INR",
		"description": "Test payment",
		"status":      "paid",
		"short_url":   "https://rzp.io/i/abcdef",
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "payment link not found",
		},
	}

	tests := []struct {
		name           string
		mockHttpClient func() (*http.Client, *httptest.Server)
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful payment link fetch",
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mocks.NewMockedHTTPClient(
					mocks.MockEndpoint{
						Path:     fmt.Sprintf(fetchPaymentLinkPathFmt, "plink_123456789"),
						Method:   "GET",
						Response: samplePaymentLink,
					},
				)
			},
			requestArgs: map[string]interface{}{
				"payment_link_id": "plink_123456789",
			},
			expectError:    false,
			expectedResult: samplePaymentLink,
		},
		{
			name: "payment link not found",
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mocks.NewMockedHTTPClient(
					mocks.MockEndpoint{
						Path:     fmt.Sprintf(fetchPaymentLinkPathFmt, "plink_invalid"),
						Method:   "GET",
						Response: errorResponse,
					},
				)
			},
			requestArgs: map[string]interface{}{
				"payment_link_id": "plink_invalid",
			},
			expectError:    true,
			expectedErrMsg: "fetching payment link failed: payment link not found",
		},
		{
			name:           "missing payment_link_id parameter",
			mockHttpClient: nil, // No HTTP client needed for validation error
			requestArgs:    map[string]interface{}{},
			expectError:    true,
			expectedErrMsg: "missing required parameter: payment_link_id",
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

				mockrzpClient.PaymentLink.Request.BaseURL = mockServer.URL
				mockrzpClient.PaymentLink.Request.HTTPClient = client
			}

			log := CreateTestLogger()
			tool := FetchPaymentLink(log, mockrzpClient)

			request := createMCPRequest(tc.requestArgs)

			result, err := tool.GetHandler()(context.Background(), request)

			if tc.expectError {
				require.NotNil(t, result)
				assert.Contains(t, result.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			var returnedPaymentLink map[string]interface{}
			err = json.Unmarshal([]byte(result.Text), &returnedPaymentLink)
			require.NoError(t, err)

			for key, expected := range tc.expectedResult {
				assert.Equal(
					t,
					expected,
					returnedPaymentLink[key],
					"Field %s doesn't match",
					key,
				)
			}
		})
	}
}
