package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_CreateSubscriptionRegistrationAuthLink(t *testing.T) {
	authLinkPath := fmt.Sprintf(
		"/%s/subscription_registration/auth_links",
		constants.VERSION_V1,
	)

	successfulAuthLinkResp := map[string]interface{}{
		"id":     "inv_ST0YqhlqxyGKVp",
		"entity": "invoice",
		"customer_details": map[string]interface{}{
			"name":    "Dhruv Mittal",
			"email":   "dhruv.mittal@razorpay.com",
			"contact": "+60102102460",
		},
		"amount":           float64(100),
		"currency":         "MYR",
		"status":           "issued",
		"short_url":        "https://rzp.io/rzp/ISJmviU",
		"auth_link_status": "issued",
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful auth link creation with all required fields",
			Request: map[string]interface{}{
				"customer_name":    "Dhruv Mittal",
				"customer_email":   "dhruv.mittal@razorpay.com",
				"customer_contact": "+60102102460",
				"type":             "link",
				"amount":           float64(100),
				"currency":         "MYR",
				"method":           "card",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     authLinkPath,
						Method:   "POST",
						Response: successfulAuthLinkResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulAuthLinkResp,
		},
		{
			Name: "successful auth link creation with optional fields",
			Request: map[string]interface{}{
				"customer_name":    "Dhruv Mittal",
				"customer_email":   "dhruv.mittal@razorpay.com",
				"customer_contact": "+60102102460",
				"type":             "link",
				"amount":           float64(100),
				"currency":         "MYR",
				"method":           "card",
				"description":      "Registration Link for Nur Aisyah",
				"receipt":          "Receipt No. 1323",
				"email_notify":     true,
				"sms_notify":       true,
				"expire_by":        float64(1798703336),
				"max_amount":       float64(500),
				"expire_at":        float64(1798703336),
				"frequency":        "as_presented",
				"notes": map[string]interface{}{
					"note_key 1": "Beam me up Scotty",
					"note_key 2": "Tea. Earl Gray. Hot.",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     authLinkPath,
						Method:   "POST",
						Response: successfulAuthLinkResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulAuthLinkResp,
		},
		{
			Name: "missing customer_name parameter",
			Request: map[string]interface{}{
				"customer_email":   "dhruv.mittal@razorpay.com",
				"customer_contact": "+60102102460",
				"type":             "link",
				"amount":           float64(100),
				"currency":         "MYR",
				"method":           "card",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_name",
		},
		{
			Name: "missing customer_email parameter",
			Request: map[string]interface{}{
				"customer_name":    "Dhruv Mittal",
				"customer_contact": "+60102102460",
				"type":             "link",
				"amount":           float64(100),
				"currency":         "MYR",
				"method":           "card",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_email",
		},
		{
			Name: "missing customer_contact parameter",
			Request: map[string]interface{}{
				"customer_name":  "Dhruv Mittal",
				"customer_email": "dhruv.mittal@razorpay.com",
				"type":           "link",
				"amount":         float64(100),
				"currency":       "MYR",
				"method":         "card",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_contact",
		},
		{
			Name: "missing amount parameter",
			Request: map[string]interface{}{
				"customer_name":    "Dhruv Mittal",
				"customer_email":   "dhruv.mittal@razorpay.com",
				"customer_contact": "+60102102460",
				"type":             "link",
				"currency":         "MYR",
				"method":           "card",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: amount",
		},
		{
			Name: "missing currency parameter",
			Request: map[string]interface{}{
				"customer_name":    "Dhruv Mittal",
				"customer_email":   "dhruv.mittal@razorpay.com",
				"customer_contact": "+60102102460",
				"type":             "link",
				"amount":           float64(100),
				"method":           "card",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: currency",
		},
		{
			Name: "missing method parameter",
			Request: map[string]interface{}{
				"customer_name":    "Dhruv Mittal",
				"customer_email":   "dhruv.mittal@razorpay.com",
				"customer_contact": "+60102102460",
				"type":             "link",
				"amount":           float64(100),
				"currency":         "MYR",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: method",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CreateSubscriptionRegistrationAuthLink, "auth_link")
		})
	}
}
