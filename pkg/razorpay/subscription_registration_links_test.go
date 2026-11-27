package razorpay

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_CreateSubscriptionRegistrationLink(t *testing.T) {
	path := "/v1/subscription_registration/auth_links"

	successResp := map[string]interface{}{
		"id":               "inv_ST0YqhlqxyGKVp",
		"entity":           "invoice",
		"type":             "link",
		"status":           "issued",
		"amount":           float64(100),
		"currency":         "MYR",
		"short_url":        "https://rzp.io/rzp/ISJmviU",
		"auth_link_status": "issued",
	}

	successRespMinimal := map[string]interface{}{
		"id":               "inv_MinimalResp",
		"entity":           "invoice",
		"type":             "link",
		"status":           "issued",
		"amount":           float64(100),
		"currency":         "INR",
		"short_url":        "https://rzp.io/rzp/minimal",
		"auth_link_status": "issued",
	}

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "customer details are required",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful creation with all parameters",
			Request: map[string]interface{}{
				"customer": map[string]interface{}{
					"name":    "Dhruv Mittal",
					"email":   "dhruv.mittal@razorpay.com",
					"contact": "+60102102460",
				},
				"type":     "link",
				"amount":   float64(100),
				"currency": "MYR",
				"subscription_registration": map[string]interface{}{
					"method":     "card",
					"max_amount": float64(500),
					"expire_at":  float64(1798703336),
					"frequency":  "as_presented",
				},
				"description":  "Registration Link for Nur Aisyah",
				"receipt":      "Receipt No. 1323",
				"email_notify": true,
				"sms_notify":   true,
				"expire_by":    float64(1798703336),
				"notes": map[string]interface{}{
					"note_key_1": "Beam me up Scotty",
					"note_key_2": "Tea. Earl Gray. Hot.",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     path,
					Method:   "POST",
					Response: successResp,
				})
			},
			ExpectError:    false,
			ExpectedResult: successResp,
		},
		{
			Name: "successful creation with required params only",
			Request: map[string]interface{}{
				"customer": map[string]interface{}{
					"name":    "Jane Doe",
					"email":   "jane@example.com",
					"contact": "+919876543210",
				},
				"type":     "link",
				"amount":   float64(100),
				"currency": "INR",
				"subscription_registration": map[string]interface{}{
					"method":     "emandate",
					"max_amount": float64(1000),
					"expire_at":  float64(1798703336),
					"frequency":  "monthly",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     path,
					Method:   "POST",
					Response: successRespMinimal,
				})
			},
			ExpectError:    false,
			ExpectedResult: successRespMinimal,
		},
		{
			Name:           "missing customer parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer",
		},
		{
			Name: "missing type parameter",
			Request: map[string]interface{}{
				"customer": map[string]interface{}{
					"name": "Jane Doe",
				},
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: type",
		},
		{
			Name: "missing amount parameter",
			Request: map[string]interface{}{
				"customer": map[string]interface{}{
					"name": "Jane Doe",
				},
				"type":     "link",
				"currency": "INR",
				"subscription_registration": map[string]interface{}{
					"method": "card",
				},
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: amount",
		},
		{
			Name: "wrong type for amount",
			Request: map[string]interface{}{
				"customer": map[string]interface{}{
					"name": "Jane Doe",
				},
				"type":     "link",
				"amount":   "not-a-number",
				"currency": "INR",
				"subscription_registration": map[string]interface{}{
					"method": "card",
				},
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: amount",
		},
		{
			Name: "API error response",
			Request: map[string]interface{}{
				"customer": map[string]interface{}{
					"name": "Jane Doe",
				},
				"type":     "link",
				"amount":   float64(100),
				"currency": "INR",
				"subscription_registration": map[string]interface{}{
					"method": "card",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     path,
					Method:   "POST",
					Response: errorResp,
				})
			},
			ExpectError:    true,
			ExpectedErrMsg: "creating subscription registration link failed: customer details are required", //nolint:lll
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(
				t, tc, CreateSubscriptionRegistrationLink,
				"Subscription Registration Link",
			)
		})
	}
}
