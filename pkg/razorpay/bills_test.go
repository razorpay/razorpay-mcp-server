package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_CreateBill(t *testing.T) {
	apiPath := fmt.Sprintf("/%s%s",
		constants.VERSION_V1, constants.INVOICE_URL)

	successResponse := map[string]interface{}{
		"id":          "inv_JZpAQqPjUHaT69",
		"entity":      "invoice",
		"type":        "invoice",
		"status":      "draft",
		"receipt":     nil,
		"description": "Test bill",
		"currency":    "INR",
		"amount":      float64(1000),
		"amount_paid": float64(0),
		"amount_due":  float64(1000),
		"customer_id": "cust_JZp80XXXXXZZ1",
		"line_items":  []interface{}{},
		"notes":       map[string]interface{}{},
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The type field is required.",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful bill creation with customer_id",
			Request: map[string]interface{}{
				"type":        "invoice",
				"customer_id": "cust_JZp80XXXXXZZ1",
				"description": "Test bill",
				"currency":    "INR",
				"line_items": []interface{}{
					map[string]interface{}{
						"name":   "Product A",
						"amount": float64(1000),
					},
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "POST",
					Response: successResponse,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name: "successful bill creation with customer details",
			Request: map[string]interface{}{
				"type":             "invoice",
				"customer_name":    "Gaurav Kumar",
				"customer_email":   "gaurav@example.com",
				"customer_contact": "9123456789",
				"description":      "Test bill",
				"currency":         "INR",
				"line_items": []interface{}{
					map[string]interface{}{
						"name":   "Product A",
						"amount": float64(1000),
					},
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "POST",
					Response: successResponse,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name: "missing type parameter",
			Request: map[string]interface{}{
				"customer_id": "cust_JZp80XXXXXZZ1",
				"line_items": []interface{}{
					map[string]interface{}{
						"name":   "Product A",
						"amount": float64(1000),
					},
				},
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: type",
		},
		{
			Name: "API error response",
			Request: map[string]interface{}{
				"type":        "invalid_type",
				"customer_id": "cust_JZp80XXXXXZZ1",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "POST",
					Response: errorResponse,
				})
			},
			ExpectError:    true,
			ExpectedErrMsg: "creating bill failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CreateBill, "Bill")
		})
	}
}

func Test_FetchBill(t *testing.T) {
	billID := "inv_JZpAQqPjUHaT69"
	apiPath := fmt.Sprintf("/%s%s/%s",
		constants.VERSION_V1, constants.INVOICE_URL, billID)

	successResponse := map[string]interface{}{
		"id":          billID,
		"entity":      "invoice",
		"type":        "invoice",
		"status":      "issued",
		"currency":    "INR",
		"amount":      float64(1000),
		"amount_paid": float64(0),
		"amount_due":  float64(1000),
		"customer_id": "cust_JZp80XXXXXZZ1",
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The id provided does not exist",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful bill fetch",
			Request: map[string]interface{}{
				"bill_id": billID,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "GET",
					Response: successResponse,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name:           "missing bill_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: bill_id",
		},
		{
			Name: "API error - bill not found",
			Request: map[string]interface{}{
				"bill_id": "inv_nonexistent",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     fmt.Sprintf("/%s%s/inv_nonexistent", constants.VERSION_V1, constants.INVOICE_URL),
					Method:   "GET",
					Response: errorResponse,
				})
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching bill failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchBill, "Bill")
		})
	}
}

func Test_FetchAllBills(t *testing.T) {
	apiPath := fmt.Sprintf("/%s%s",
		constants.VERSION_V1, constants.INVOICE_URL)

	successResponse := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":     "inv_JZpAQqPjUHaT69",
				"type":   "invoice",
				"status": "issued",
			},
			map[string]interface{}{
				"id":     "inv_KZpBQqPkUIbU79",
				"type":   "invoice",
				"status": "paid",
			},
		},
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Invalid parameters",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name:    "successful fetch all bills",
			Request: map[string]interface{}{},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "GET",
					Response: successResponse,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name: "successful fetch with filters",
			Request: map[string]interface{}{
				"type":        "invoice",
				"count":       float64(10),
				"customer_id": "cust_JZp80XXXXXZZ1",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "GET",
					Response: successResponse,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name: "API error response",
			Request: map[string]interface{}{
				"count": float64(1000),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "GET",
					Response: errorResponse,
				})
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching bills failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchAllBills, "Bill")
		})
	}
}

func Test_UpdateBill(t *testing.T) {
	billID := "inv_JZpAQqPjUHaT69"
	apiPath := fmt.Sprintf("/%s%s/%s",
		constants.VERSION_V1, constants.INVOICE_URL, billID)

	successResponse := map[string]interface{}{
		"id":     billID,
		"entity": "invoice",
		"type":   "invoice",
		"status": "issued",
		"notes": map[string]interface{}{
			"note_key_1": "Beam me up Scotty",
		},
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The id provided does not exist",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful bill update",
			Request: map[string]interface{}{
				"bill_id": billID,
				"notes": map[string]interface{}{
					"note_key_1": "Beam me up Scotty",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "PATCH",
					Response: successResponse,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name: "missing bill_id parameter",
			Request: map[string]interface{}{
				"notes": map[string]interface{}{
					"key": "value",
				},
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: bill_id",
		},
		{
			Name: "missing notes parameter",
			Request: map[string]interface{}{
				"bill_id": billID,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: notes",
		},
		{
			Name: "API error response",
			Request: map[string]interface{}{
				"bill_id": "inv_nonexistent",
				"notes": map[string]interface{}{
					"key": "value",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path: fmt.Sprintf("/%s%s/inv_nonexistent",
						constants.VERSION_V1, constants.INVOICE_URL),
					Method:   "PATCH",
					Response: errorResponse,
				})
			},
			ExpectError:    true,
			ExpectedErrMsg: "updating bill failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, UpdateBill, "Bill")
		})
	}
}

func Test_IssueBill(t *testing.T) {
	billID := "inv_JZpAQqPjUHaT69"
	apiPath := fmt.Sprintf("/%s%s/%s/issue",
		constants.VERSION_V1, constants.INVOICE_URL, billID)

	successResponse := map[string]interface{}{
		"id":     billID,
		"entity": "invoice",
		"type":   "invoice",
		"status": "issued",
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The invoice is not in draft state",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful bill issue",
			Request: map[string]interface{}{
				"bill_id": billID,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "POST",
					Response: successResponse,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name:           "missing bill_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: bill_id",
		},
		{
			Name: "API error - bill not in draft state",
			Request: map[string]interface{}{
				"bill_id": billID,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "POST",
					Response: errorResponse,
				})
			},
			ExpectError:    true,
			ExpectedErrMsg: "issuing bill failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, IssueBill, "Bill")
		})
	}
}

func Test_CancelBill(t *testing.T) {
	billID := "inv_JZpAQqPjUHaT69"
	apiPath := fmt.Sprintf("/%s%s/%s/cancel",
		constants.VERSION_V1, constants.INVOICE_URL, billID)

	successResponse := map[string]interface{}{
		"id":     billID,
		"entity": "invoice",
		"type":   "invoice",
		"status": "cancelled",
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The invoice cannot be cancelled",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful bill cancellation",
			Request: map[string]interface{}{
				"bill_id": billID,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "POST",
					Response: successResponse,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name:           "missing bill_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: bill_id",
		},
		{
			Name: "API error - bill cannot be cancelled",
			Request: map[string]interface{}{
				"bill_id": billID,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "POST",
					Response: errorResponse,
				})
			},
			ExpectError:    true,
			ExpectedErrMsg: "cancelling bill failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CancelBill, "Bill")
		})
	}
}

func Test_DeleteBill(t *testing.T) {
	billID := "inv_JZpAQqPjUHaT69"
	apiPath := fmt.Sprintf("/%s%s/%s",
		constants.VERSION_V1, constants.INVOICE_URL, billID)

	successResponse := map[string]interface{}{}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The invoice is not in draft state",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful draft bill deletion",
			Request: map[string]interface{}{
				"bill_id": billID,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "DELETE",
					Response: successResponse,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name:           "missing bill_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: bill_id",
		},
		{
			Name: "API error - bill not in draft state",
			Request: map[string]interface{}{
				"bill_id": billID,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "DELETE",
					Response: errorResponse,
				})
			},
			ExpectError:    true,
			ExpectedErrMsg: "deleting bill failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, DeleteBill, "Bill")
		})
	}
}

func Test_SendBillNotification(t *testing.T) {
	billID := "inv_JZpAQqPjUHaT69"
	emailPath := fmt.Sprintf("/%s%s/%s/notify_by/email",
		constants.VERSION_V1, constants.INVOICE_URL, billID)
	smsPath := fmt.Sprintf("/%s%s/%s/notify_by/sms",
		constants.VERSION_V1, constants.INVOICE_URL, billID)

	successResponse := map[string]interface{}{
		"success": true,
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The invoice is not in issued state",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful email notification",
			Request: map[string]interface{}{
				"bill_id": billID,
				"medium":  "email",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     emailPath,
					Method:   "POST",
					Response: successResponse,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name: "successful SMS notification",
			Request: map[string]interface{}{
				"bill_id": billID,
				"medium":  "sms",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     smsPath,
					Method:   "POST",
					Response: successResponse,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name: "missing bill_id parameter",
			Request: map[string]interface{}{
				"medium": "email",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: bill_id",
		},
		{
			Name: "missing medium parameter",
			Request: map[string]interface{}{
				"bill_id": billID,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: medium",
		},
		{
			Name: "API error - bill not in issued state",
			Request: map[string]interface{}{
				"bill_id": billID,
				"medium":  "email",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     emailPath,
					Method:   "POST",
					Response: errorResponse,
				})
			},
			ExpectError:    true,
			ExpectedErrMsg: "sending bill notification failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, SendBillNotification, "Bill")
		})
	}
}
