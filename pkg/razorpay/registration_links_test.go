package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_CreateRegistrationLink(t *testing.T) {
	apiPath := fmt.Sprintf(
		"/%s/subscription_registration/auth_links",
		constants.VERSION_V1,
	)

	successResponse := map[string]interface{}{
		"id":          "inv_ST0YqhlqxyGKVp",
		"entity":      "invoice",
		"receipt":     "Receipt No. 1323",
		"customer_id": "cust_RFMQlE3iXLosoP",
		"customer_details": map[string]interface{}{
			"id":               "cust_RFMQlE3iXLosoP",
			"name":             "Dhruv Mittal",
			"email":            "dhruv.mittal@razorpay.com",
			"contact":          "+60102102460",
			"gstin":            nil,
			"customer_name":    "Dhruv Mittal",
			"customer_email":   "dhruv.mittal@razorpay.com",
			"customer_contact": "+60102102460",
		},
		"order_id":         "order_ST0YqsGGV4gGv3",
		"status":           "issued",
		"expire_by":        float64(1798703336),
		"amount":           float64(100),
		"amount_paid":      float64(0),
		"amount_due":       float64(100),
		"currency":         "MYR",
		"description":      "Registration Link for  Nur Aisyah",
		"short_url":        "https://rzp.io/rzp/ISJmviU",
		"type":             "link",
		"auth_link_status": "issued",
		"notes": map[string]interface{}{
			"note_key 1": "Beam me up Scotty",
			"note_key 2": "Tea. Earl Gray. Hot.",
		},
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Invalid currency",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful registration link creation",
			Request: map[string]interface{}{
				"type":        "link",
				"amount":      float64(100),
				"currency":    "MYR",
				"description": "Registration Link for  Nur Aisyah", //nolint:lll
				"subscription_registration": map[string]interface{}{
					"method":     "card",
					"max_amount": "500",
					"expire_at":  "1798703336",
					"frequency":  "as_presented",
				},
				"customer_name":    "Dhruv Mittal",
				"customer_email":   "dhruv.mittal@razorpay.com", //nolint:lll
				"customer_contact": "+60102102460",
				"receipt":          "Receipt No. 1323",
				"email_notify":     true,
				"sms_notify":       true,
				"expire_by":        float64(1798703336),
				"notes": map[string]interface{}{
					"note_key 1": "Beam me up Scotty",
					"note_key 2": "Tea. Earl Gray. Hot.",
				},
			},
			MockHttpClient: func() (
				*http.Client, *httptest.Server,
			) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     apiPath,
						Method:   "POST",
						Response: successResponse,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name: "missing type parameter",
			Request: map[string]interface{}{
				"amount":      float64(100),
				"currency":    "MYR",
				"description": "Test",
				"subscription_registration": map[string]interface{}{
					"method": "card",
				},
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: type",
		},
		{
			Name: "missing amount parameter",
			Request: map[string]interface{}{
				"type":        "link",
				"currency":    "MYR",
				"description": "Test",
				"subscription_registration": map[string]interface{}{
					"method": "card",
				},
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: amount",
		},
		{
			Name: "missing currency parameter",
			Request: map[string]interface{}{
				"type":        "link",
				"amount":      float64(100),
				"description": "Test",
				"subscription_registration": map[string]interface{}{
					"method": "card",
				},
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: currency", //nolint:lll
		},
		{
			Name: "missing description parameter",
			Request: map[string]interface{}{
				"type":     "link",
				"amount":   float64(100),
				"currency": "MYR",
				"subscription_registration": map[string]interface{}{
					"method": "card",
				},
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: description", //nolint:lll
		},
		{
			Name: "missing subscription_registration",
			Request: map[string]interface{}{
				"type":        "link",
				"amount":      float64(100),
				"currency":    "MYR",
				"description": "Test",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: " +
				"subscription_registration",
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				"customer_name": 12345,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:\n- " +
				"missing required parameter: type\n- " +
				"missing required parameter: amount\n- " +
				"missing required parameter: currency\n- " +
				"missing required parameter: " +
				"description\n- " +
				"missing required parameter: " +
				"subscription_registration\n- " +
				"invalid parameter type: customer_name",
		},
		{
			Name: "API error response",
			Request: map[string]interface{}{
				"type":        "link",
				"amount":      float64(100),
				"currency":    "INVALID",
				"description": "Test",
				"subscription_registration": map[string]interface{}{
					"method": "card",
				},
			},
			MockHttpClient: func() (
				*http.Client, *httptest.Server,
			) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     apiPath,
						Method:   "POST",
						Response: errorResponse,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "creating registration link " +
				"failed: Invalid currency",
		},
		{
			Name: "successful creation without customer fields",
			Request: map[string]interface{}{
				"type":        "link",
				"amount":      float64(100),
				"currency":    "MYR",
				"description": "Registration Link without customer",
				"subscription_registration": map[string]interface{}{
					"method": "card",
				},
			},
			MockHttpClient: func() (
				*http.Client, *httptest.Server,
			) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     apiPath,
						Method:   "POST",
						Response: successResponse,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(
				t, tc,
				CreateRegistrationLink,
				"Registration Link",
			)
		})
	}
}
