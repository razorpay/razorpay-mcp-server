package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_CreateCustomer(t *testing.T) {
	createCustomerPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.CUSTOMER_URL,
	)

	successResponse := map[string]interface{}{
		"id":      "cust_1Aa00000000001",
		"entity":  "customer",
		"name":    "John Doe",
		"email":   "john.doe@example.com",
		"contact": "+919876543210",
		"gstin":   nil,
		"notes": map[string]interface{}{
			"purpose": "Test customer creation",
		},
		"created_at": float64(1234567890),
	}

	validationErrorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The name field is required.",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful customer creation with all parameters",
			Request: map[string]interface{}{
				"name":    "John Doe",
				"email":   "john.doe@example.com",
				"contact": "+919876543210",
				"notes": map[string]interface{}{
					"purpose": "Test customer creation",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: successResponse,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name: "successful customer creation with only name",
			Request: map[string]interface{}{
				"name": "Jane Smith",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				minimalResponse := map[string]interface{}{
					"id":         "cust_1Aa00000000002",
					"entity":     "customer",
					"name":       "Jane Smith",
					"email":      nil,
					"contact":    nil,
					"gstin":      nil,
					"notes":      map[string]interface{}{},
					"created_at": float64(1234567891),
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: minimalResponse,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"id":         "cust_1Aa00000000002",
				"entity":     "customer",
				"name":       "Jane Smith",
				"email":      nil,
				"contact":    nil,
				"gstin":      nil,
				"notes":      map[string]interface{}{},
				"created_at": float64(1234567891),
			},
		},
		{
			Name: "missing required name parameter",
			Request: map[string]interface{}{
				"email":   "test@example.com",
				"contact": "+919876543210",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: name",
		},
		{
			Name:           "missing all parameters",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: name",
		},
		{
			Name: "invalid parameter type - name as number",
			Request: map[string]interface{}{
				"name":  12345,
				"email": "test@example.com",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: name",
		},
		{
			Name: "invalid parameter type - email as number",
			Request: map[string]interface{}{
				"name":  "John Doe",
				"email": 12345,
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: email",
		},
		{
			Name: "invalid parameter type - contact as number",
			Request: map[string]interface{}{
				"name":    "John Doe",
				"contact": 12345,
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: contact",
		},
		{
			Name: "invalid parameter type - notes as string",
			Request: map[string]interface{}{
				"name":  "John Doe",
				"notes": "invalid notes format",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: notes",
		},
		{
			Name: "API error - validation failure",
			Request: map[string]interface{}{
				"name": "", // Empty name should cause API validation error
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createCustomerPath,
						Method:   "POST",
						Response: validationErrorResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "creating customer failed: The name field is required.",
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				"name":    12345,          // Wrong type
				"email":   true,           // Wrong type
				"contact": []int{1, 2, 3}, // Wrong type
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: name", // First error returned
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CreateCustomer, "Customer")
		})
	}
}
