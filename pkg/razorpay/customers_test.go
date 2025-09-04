package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func TestCreateCustomer(t *testing.T) {
	apiPath := fmt.Sprintf("/%s/customers", constants.VERSION_V1)

	// Test data based on Razorpay API documentation
	successResponse := map[string]interface{}{
		"id":            "cust_1Aa00000000003",
		"entity":        "customer",
		"name":          "John Smith",
		"email":         "john.smith@example.com",
		"contact":       "+11234567890",
		"gstin":         nil,
		"notes":         map[string]interface{}{},
		"created_at":    float64(1234567890),
		"fail_existing": "0",
	}

	testCases := []RazorpayToolTestCase{
		{
			Name: "successful customer creation with all parameters",
			Request: map[string]interface{}{
				"name":          "John Smith",
				"email":         "john.smith@example.com",
				"contact":       "+11234567890",
				"fail_existing": "0",
				"notes": map[string]interface{}{
					"customer_type": "premium",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     apiPath,
						Method:   "POST",
						Response: successResponse,
					},
				)
			},
			ExpectedResult: successResponse,
			ExpectError:    false,
		},
		{
			Name: "successful customer creation with minimal parameters",
			Request: map[string]interface{}{
				"name": "John Smith",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     apiPath,
						Method:   "POST",
						Response: successResponse,
					},
				)
			},
			ExpectedResult: successResponse,
			ExpectError:    false,
		},
		{
			Name:           "missing required name parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: name",
		},
		{
			Name: "empty name parameter",
			Request: map[string]interface{}{
				"name": "",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: name",
		},
		{
			Name: "invalid contact type",
			Request: map[string]interface{}{
				"name":    "John Smith",
				"contact": 1234567890, // Should be string
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: contact",
		},
		{
			Name: "invalid email type",
			Request: map[string]interface{}{
				"name":  "John Smith",
				"email": 12345, // Should be string
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: email",
		},
		{
			Name: "invalid fail_existing type",
			Request: map[string]interface{}{
				"name":          "John Smith",
				"fail_existing": 0, // Should be string
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: fail_existing",
		},
		{
			Name: "invalid notes type",
			Request: map[string]interface{}{
				"name":  "John Smith",
				"notes": "invalid_notes", // Should be object
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: notes",
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				"contact": 1234567890, // Invalid type
				"email":   12345,      // Invalid type
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CreateCustomer, "customer")
		})
	}
}

func TestFetchCustomer(t *testing.T) {
	customerId := "cust_1Aa00000000003"
	fetchCustomerPathFmt := fmt.Sprintf(
		"/%s/customers/%%s",
		constants.VERSION_V1,
	)

	// Test data based on Razorpay API documentation
	successResponse := map[string]interface{}{
		"id":         customerId,
		"entity":     "customer",
		"name":       "John Smith",
		"email":      "john.smith@example.com",
		"contact":    "+11234567890",
		"gstin":      nil,
		"notes":      map[string]interface{}{},
		"created_at": float64(1234567890),
	}

	testCases := []RazorpayToolTestCase{
		{
			Name: "successful customer fetch",
			Request: map[string]interface{}{
				"customer_id": customerId,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchCustomerPathFmt, customerId),
						Method:   "GET",
						Response: successResponse,
					},
				)
			},
			ExpectedResult: successResponse,
			ExpectError:    false,
		},
		{
			Name:           "missing required customer_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_id",
		},
		{
			Name: "empty customer_id parameter",
			Request: map[string]interface{}{
				"customer_id": "",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_id",
		},
		{
			Name: "invalid customer_id type",
			Request: map[string]interface{}{
				"customer_id": 1234567890, // Should be string
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: customer_id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchCustomer, "customer")
		})
	}
}

func TestEditCustomer(t *testing.T) {
	customerId := "cust_1Aa00000000003"
	editCustomerPathFmt := fmt.Sprintf(
		"/%s/customers/%%s",
		constants.VERSION_V1,
	)

	// Test data based on Razorpay API documentation
	successResponse := map[string]interface{}{
		"id":         customerId,
		"entity":     "customer",
		"name":       "John Updated",
		"email":      "john.updated@example.com",
		"contact":    "+11234567891",
		"gstin":      nil,
		"notes":      map[string]interface{}{"updated": "true"},
		"created_at": float64(1234567890),
	}

	testCases := []RazorpayToolTestCase{
		{
			Name: "successful customer edit with all parameters",
			Request: map[string]interface{}{
				"customer_id": customerId,
				"name":        "John Updated",
				"email":       "john.updated@example.com",
				"contact":     "+11234567891",
				"notes": map[string]interface{}{
					"updated": "true",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(editCustomerPathFmt, customerId),
						Method:   "PUT",
						Response: successResponse,
					},
				)
			},
			ExpectedResult: successResponse,
			ExpectError:    false,
		},
		{
			Name: "successful customer edit with minimal parameters",
			Request: map[string]interface{}{
				"customer_id": customerId,
				"name":        "John Updated",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(editCustomerPathFmt, customerId),
						Method:   "PUT",
						Response: successResponse,
					},
				)
			},
			ExpectedResult: successResponse,
			ExpectError:    false,
		},
		{
			Name:           "missing required customer_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_id",
		},
		{
			Name: "empty customer_id parameter",
			Request: map[string]interface{}{
				"customer_id": "",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_id",
		},
		{
			Name: "invalid customer_id type",
			Request: map[string]interface{}{
				"customer_id": 1234567890, // Should be string
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: customer_id",
		},
		{
			Name: "invalid name type",
			Request: map[string]interface{}{
				"customer_id": customerId,
				"name":        12345, // Should be string
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: name",
		},
		{
			Name: "invalid contact type",
			Request: map[string]interface{}{
				"customer_id": customerId,
				"contact":     1234567890, // Should be string
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: contact",
		},
		{
			Name: "invalid email type",
			Request: map[string]interface{}{
				"customer_id": customerId,
				"email":       12345, // Should be string
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: email",
		},
		{
			Name: "invalid notes type",
			Request: map[string]interface{}{
				"customer_id": customerId,
				"notes":       "invalid_notes", // Should be object
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: notes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, EditCustomer, "customer")
		})
	}
}

func TestFetchAllCustomers(t *testing.T) {
	apiPath := fmt.Sprintf("/%s/customers", constants.VERSION_V1)

	// Test data based on Razorpay API documentation
	successResponse := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":         "cust_1Aa00000000001",
				"entity":     "customer",
				"name":       "John Smith",
				"email":      "john.smith@example.com",
				"contact":    "+11234567890",
				"gstin":      nil,
				"notes":      map[string]interface{}{},
				"created_at": float64(1234567890),
			},
			map[string]interface{}{
				"id":         "cust_1Aa00000000002",
				"entity":     "customer",
				"name":       "Jane Doe",
				"email":      "jane.doe@example.com",
				"contact":    "+11234567891",
				"gstin":      nil,
				"notes":      map[string]interface{}{},
				"created_at": float64(1234567891),
			},
		},
	}

	testCases := []RazorpayToolTestCase{
		{
			Name:    "successful fetch all customers with default parameters",
			Request: map[string]interface{}{},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     apiPath,
						Method:   "GET",
						Response: successResponse,
					},
				)
			},
			ExpectedResult: successResponse,
			ExpectError:    false,
		},
		{
			Name: "successful fetch all customers with pagination",
			Request: map[string]interface{}{
				"count": 2,
				"skip":  1,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     apiPath,
						Method:   "GET",
						Response: successResponse,
					},
				)
			},
			ExpectedResult: successResponse,
			ExpectError:    false,
		},
		{
			Name: "successful fetch all customers with time range",
			Request: map[string]interface{}{
				"from": 1234567890,
				"to":   1234567999,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     apiPath,
						Method:   "GET",
						Response: successResponse,
					},
				)
			},
			ExpectedResult: successResponse,
			ExpectError:    false,
		},
		{
			Name: "invalid count type",
			Request: map[string]interface{}{
				"count": "invalid", // Should be number
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: count",
		},
		{
			Name: "invalid skip type",
			Request: map[string]interface{}{
				"skip": "invalid", // Should be number
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: skip",
		},
		{
			Name: "invalid from type",
			Request: map[string]interface{}{
				"from": "invalid", // Should be number
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: from",
		},
		{
			Name: "invalid to type",
			Request: map[string]interface{}{
				"to": "invalid", // Should be number
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: to",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchAllCustomers, "customer")
		})
	}
}

func TestFetchCustomerTokens(t *testing.T) {
	apiPath := fmt.Sprintf("/%s/customers/cust_1Aa00000000002/tokens", constants.VERSION_V1)

	successResponse := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":          "token_1Aa00000000001",
				"entity":      "token",
				"token":       "token_1Aa00000000001",
				"bank":        nil,
				"wallet":      nil,
				"method":      "card",
				"card": map[string]interface{}{
					"entity":          "card",
					"name":           "Gaurav Kumar",
					"last4":          "3335",
					"network":        "Visa",
					"type":           "debit",
					"issuer":         "HDFC",
					"international":  false,
					"emi":           false,
					"sub_type":      "business",
					"token_provision_status": "provisioned",
				},
				"vpa":           nil,
				"recurring":     true,
				"recurring_details": map[string]interface{}{
					"status":       "confirmed",
					"failure_reason": nil,
				},
				"auth_type":     nil,
				"mrn":          nil,
				"used_at":      float64(1629779657),
				"created_at":   float64(1629779657),
				"bank_details": nil,
				"max_amount":   float64(500000),
				"expired_at":   float64(1640975399),
				"dcc_enabled":  false,
				"notes": []interface{}{},
				"compliance_error": nil,
			},
			map[string]interface{}{
				"id":          "token_1Aa00000000002",
				"entity":      "token",
				"token":       "token_1Aa00000000002",
				"bank":        nil,
				"wallet":      nil,
				"method":      "card",
				"card": map[string]interface{}{
					"entity":          "card",
					"name":           "Gaurav Kumar",
					"last4":          "1111",
					"network":        "Visa",
					"type":           "credit",
					"issuer":         "ICICI",
					"international":  false,
					"emi":           true,
					"sub_type":      "consumer",
					"token_provision_status": "provisioned",
				},
				"vpa":           nil,
				"recurring":     true,
				"recurring_details": map[string]interface{}{
					"status":       "confirmed",
					"failure_reason": nil,
				},
				"auth_type":     nil,
				"mrn":          nil,
				"used_at":      float64(1629779658),
				"created_at":   float64(1629779658),
				"bank_details": nil,
				"max_amount":   float64(1000000),
				"expired_at":   float64(1640975400),
				"dcc_enabled":  false,
				"notes": []interface{}{},
				"compliance_error": nil,
			},
		},
	}

	testCases := []RazorpayToolTestCase{
		{
			Name: "successful fetch customer tokens",
			Request: map[string]interface{}{
				"customer_id": "cust_1Aa00000000002",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     apiPath,
						Method:   "GET",
						Response: successResponse,
					},
				)
			},
			ExpectedResult: successResponse,
			ExpectError:    false,
		},
		{
			Name: "missing customer_id parameter",
			Request: map[string]interface{}{},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient()
			},
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_id",
		},
		{
			Name: "empty customer_id parameter",
			Request: map[string]interface{}{
				"customer_id": "",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient()
			},
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_id",
		},
		{
			Name: "invalid customer_id format",
			Request: map[string]interface{}{
				"customer_id": "invalid_id",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient()
			},
			ExpectError:    true,
			ExpectedErrMsg: "customer_id must start with 'cust_'",
		},
		{
			Name: "wrong type for customer_id",
			Request: map[string]interface{}{
				"customer_id": 12345,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient()
			},
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: customer_id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchCustomerTokens, "customer")
		})
	}
}
