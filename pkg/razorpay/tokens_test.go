package razorpay

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/contextkey"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_FetchSavedPaymentMethods(t *testing.T) {
	// URL patterns for mocking
	createCustomerPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.CUSTOMER_URL,
	)

	fetchTokensPathFmt := fmt.Sprintf(
		"/%s/customers/%%s/tokens",
		constants.VERSION_V1,
	)

	// Sample successful customer creation/fetch response
	customerResp := map[string]interface{}{
		"id":         "cust_1Aa00000000003",
		"entity":     "customer",
		"name":       "",
		"email":      "",
		"contact":    "9876543210",
		"gstin":      nil,
		"notes":      map[string]interface{}{},
		"created_at": float64(1234567890),
	}

	// Sample successful tokens response
	tokensResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":     "token_EhYXHrLsJdwRhM",
				"entity": "token",
				"token":  "EhYXHrLsJdwRhM",
				"bank":   nil,
				"wallet": nil,
				"method": "card",
				"card": map[string]interface{}{
					"entity":        "card",
					"name":          "Gaurav Kumar",
					"last4":         "1111",
					"network":       "Visa",
					"type":          "debit",
					"issuer":        "HDFC",
					"international": false,
					"emi":           false,
					"sub_type":      "consumer",
				},
				"vpa":       nil,
				"recurring": true,
				"recurring_details": map[string]interface{}{
					"status":         "confirmed",
					"failure_reason": nil,
				},
				"auth_type":   nil,
				"mrn":         nil,
				"used_at":     float64(1629779657),
				"created_at":  float64(1629779657),
				"expired_at":  float64(1640918400),
				"dcc_enabled": false,
			},
			map[string]interface{}{
				"id":     "token_EhYXHrLsJdwRhN",
				"entity": "token",
				"token":  "EhYXHrLsJdwRhN",
				"bank":   nil,
				"wallet": nil,
				"method": "upi",
				"card":   nil,
				"vpa": map[string]interface{}{
					"username": "gauravkumar",
					"handle":   "okhdfcbank",
					"name":     "Gaurav Kumar",
				},
				"recurring": true,
				"recurring_details": map[string]interface{}{
					"status":         "confirmed",
					"failure_reason": nil,
				},
				"auth_type":   nil,
				"mrn":         nil,
				"used_at":     float64(1629779657),
				"created_at":  float64(1629779657),
				"expired_at":  float64(1640918400),
				"dcc_enabled": false,
			},
		},
	}

	// Expected combined response
	expectedSuccessResp := map[string]interface{}{
		"customer":              customerResp,
		"saved_payment_methods": tokensResp,
	}

	// Error responses
	customerCreationFailedResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Contact number is invalid",
		},
	}

	tokensAPIFailedResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Customer not found",
		},
	}

	// Customer response without ID (invalid)
	invalidCustomerResp := map[string]interface{}{
		"entity":     "customer",
		"name":       "",
		"email":      "",
		"contact":    "9876543210",
		"gstin":      nil,
		"notes":      map[string]interface{}{},
		"created_at": float64(1234567890),
		// Missing "id" field
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful fetch of saved cards with valid contact",
			Request: map[string]interface{}{
				"contact": "9876543210",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerResp,
					},
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchTokensPathFmt, "cust_1Aa00000000003"),
						Method:   "GET",
						Response: tokensResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: expectedSuccessResp,
		},
		{
			Name: "successful fetch with international contact format",
			Request: map[string]interface{}{
				"contact": "+919876543210",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				customerRespIntl := map[string]interface{}{
					"id":         "cust_1Aa00000000004",
					"entity":     "customer",
					"name":       "",
					"email":      "",
					"contact":    "+919876543210",
					"gstin":      nil,
					"notes":      map[string]interface{}{},
					"created_at": float64(1234567890),
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerRespIntl,
					},
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchTokensPathFmt, "cust_1Aa00000000004"),
						Method:   "GET",
						Response: tokensResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"customer": map[string]interface{}{
					"id":         "cust_1Aa00000000004",
					"entity":     "customer",
					"name":       "",
					"email":      "",
					"contact":    "+919876543210",
					"gstin":      nil,
					"notes":      map[string]interface{}{},
					"created_at": float64(1234567890),
				},
				"saved_payment_methods": tokensResp,
			},
		},
		{
			Name: "customer creation/fetch failure",
			Request: map[string]interface{}{
				"contact": "invalid_contact",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerCreationFailedResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "Failed to create/fetch customer with " +
				"contact invalid_contact: Contact number is invalid",
		},
		{
			Name: "tokens API failure after successful customer creation",
			Request: map[string]interface{}{
				"contact": "9876543210",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerResp,
					},
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchTokensPathFmt, "cust_1Aa00000000003"),
						Method:   "GET",
						Response: tokensAPIFailedResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "Failed to fetch saved payment methods for " +
				"customer cust_1Aa00000000003: Customer not found",
		},
		{
			Name: "invalid customer response - missing customer ID",
			Request: map[string]interface{}{
				"contact": "9876543210",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: invalidCustomerResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "Customer ID not found in response",
		},
		{
			Name:    "missing contact parameter",
			Request: map[string]interface{}{
				// No contact parameter
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: contact",
		},
		{
			Name: "empty contact parameter",
			Request: map[string]interface{}{
				"contact": "",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: contact",
		},
		{
			Name: "null contact parameter",
			Request: map[string]interface{}{
				"contact": nil,
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: contact",
		},
		{
			Name: "successful fetch with empty tokens list",
			Request: map[string]interface{}{
				"contact": "9876543210",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				emptyTokensResp := map[string]interface{}{
					"entity": "collection",
					"count":  float64(0),
					"items":  []interface{}{},
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: customerResp,
					},
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchTokensPathFmt, "cust_1Aa00000000003"),
						Method:   "GET",
						Response: emptyTokensResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"customer": customerResp,
				"saved_payment_methods": map[string]interface{}{
					"entity": "collection",
					"count":  float64(0),
					"items":  []interface{}{},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchSavedPaymentMethods, "Saved Cards")
		})
	}
}

// Test_FetchSavedPaymentMethods_ClientContextScenarios tests scenarios
// related to client context handling for 100% code coverage
func Test_FetchSavedPaymentMethods_ClientContextScenarios(t *testing.T) {
	obs := CreateTestObservability()

	t.Run("no client in context and default is nil", func(t *testing.T) {
		// Create tool with nil client
		tool := FetchSavedPaymentMethods(obs, nil)

		// Create context without client
		ctx := context.Background()
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"contact": "9876543210",
			},
		}

		result, err := tool.GetHandler()(ctx, request)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		if result.Text == "" {
			t.Fatal("Expected error message in result")
		}

		expectedErrMsg := "no client found in context"
		if !strings.Contains(result.Text, expectedErrMsg) {
			t.Errorf(
				"Expected error message to contain '%s', got '%s'",
				expectedErrMsg,
				result.Text,
			)
		}
	})

	t.Run("invalid client type in context", func(t *testing.T) {
		// Create tool with nil client
		tool := FetchSavedPaymentMethods(obs, nil)

		// Create context with invalid client type
		ctx := contextkey.WithClient(context.Background(), "invalid_client_type")
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"contact": "9876543210",
			},
		}

		result, err := tool.GetHandler()(ctx, request)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		if result.Text == "" {
			t.Fatal("Expected error message in result")
		}

		expectedErrMsg := "invalid client type in context"
		if !strings.Contains(result.Text, expectedErrMsg) {
			t.Errorf(
				"Expected error message to contain '%s', got '%s'",
				expectedErrMsg,
				result.Text,
			)
		}
	})
}
