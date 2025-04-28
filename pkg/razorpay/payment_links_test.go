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

func Test_CreatePaymentLink(t *testing.T) {
	createPaymentLinkPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.PaymentLink_URL,
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
			name: "successful payment link creation",
			requestArgs: map[string]interface{}{
				"amount":      float64(50000),
				"currency":    "INR",
				"description": "Test payment",
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   createPaymentLinkPath,
						Method: "POST",
						Response: map[string]interface{}{
							"id":          "plink_ExjpAUN3gVHrPJ",
							"amount":      float64(50000),
							"currency":    "INR",
							"description": "Test payment",
							"status":      "created",
							"short_url":   "https://rzp.io/i/nxrHnLJ",
						},
					},
				)
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":          "plink_ExjpAUN3gVHrPJ",
				"amount":      float64(50000),
				"currency":    "INR",
				"description": "Test payment",
				"status":      "created",
				"short_url":   "https://rzp.io/i/nxrHnLJ",
			},
		},
		{
			name: "payment link without description",
			requestArgs: map[string]interface{}{
				"amount":   float64(50000),
				"currency": "INR",
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   createPaymentLinkPath,
						Method: "POST",
						Response: map[string]interface{}{
							"id":        "plink_ExjpAUN3gVHrPJ",
							"amount":    float64(50000),
							"currency":  "INR",
							"status":    "created",
							"short_url": "https://rzp.io/i/nxrHnLJ",
						},
					},
				)
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":        "plink_ExjpAUN3gVHrPJ",
				"amount":    float64(50000),
				"currency":  "INR",
				"status":    "created",
				"short_url": "https://rzp.io/i/nxrHnLJ",
			},
		},
		{
			name: "missing amount parameter",
			requestArgs: map[string]interface{}{
				"currency": "INR",
			},
			mockHttpClient: nil, // No HTTP client needed for validation error
			expectError:    true,
			expectedErrMsg: "missing required parameter: amount",
		},
		{
			name: "missing currency parameter",
			requestArgs: map[string]interface{}{
				"amount": float64(50000),
			},
			mockHttpClient: nil, // No HTTP client needed for validation error
			expectError:    true,
			expectedErrMsg: "missing required parameter: currency",
		},
		{
			name: "payment link creation fails",
			requestArgs: map[string]interface{}{
				"amount":   float64(50000),
				"currency": "XYZ", // Invalid currency
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   createPaymentLinkPath,
						Method: "POST",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "API error: Invalid currency",
							},
						},
					},
				)
			},
			expectError:    true,
			expectedErrMsg: "creating payment link failed: API error: Invalid currency",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockrzpClient, mockServer := newRzpMockClient(tc.mockHttpClient)
			if mockServer != nil {
				defer mockServer.Close()
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

	tests := []struct {
		name           string
		requestArgs    map[string]interface{}
		mockHttpClient func() (*http.Client, *httptest.Server)
		expectError    bool
		expectedResult map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful payment link fetch",
			requestArgs: map[string]interface{}{
				"payment_link_id": "plink_ExjpAUN3gVHrPJ",
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   fmt.Sprintf(fetchPaymentLinkPathFmt, "plink_ExjpAUN3gVHrPJ"),
						Method: "GET",
						Response: map[string]interface{}{
							"id":          "plink_ExjpAUN3gVHrPJ",
							"amount":      float64(50000),
							"currency":    "INR",
							"description": "Test payment",
							"status":      "paid",
							"short_url":   "https://rzp.io/i/nxrHnLJ",
						},
					},
				)
			},
			expectError: false,
			expectedResult: map[string]interface{}{
				"id":          "plink_ExjpAUN3gVHrPJ",
				"amount":      float64(50000),
				"currency":    "INR",
				"description": "Test payment",
				"status":      "paid",
				"short_url":   "https://rzp.io/i/nxrHnLJ",
			},
		},
		{
			name: "payment link not found",
			requestArgs: map[string]interface{}{
				"payment_link_id": "plink_invalid",
			},
			mockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   fmt.Sprintf(fetchPaymentLinkPathFmt, "plink_invalid"),
						Method: "GET",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "payment link not found",
							},
						},
					},
				)
			},
			expectError:    true,
			expectedErrMsg: "fetching payment link failed: payment link not found",
		},
		{
			name:           "missing payment_link_id parameter",
			requestArgs:    map[string]interface{}{},
			mockHttpClient: nil, // No HTTP client needed for validation error
			expectError:    true,
			expectedErrMsg: "missing required parameter: payment_link_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockrzpClient, mockServer := newRzpMockClient(tc.mockHttpClient)
			if mockServer != nil {
				defer mockServer.Close()
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
