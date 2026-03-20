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
	apiPath := fmt.Sprintf("/%s/subscription_registration/auth_links",
		constants.VERSION_V1)

	successResponse := map[string]interface{}{
		"id":          "inv_xxxxxxxxxxxxx",
		"entity":      "invoice",
		"type":        "link",
		"amount":      float64(100),
		"currency":    "MYR",
		"description": "Registration Link for Nur Aisyah",
		"customer": map[string]interface{}{
			"name":    "Dhruv Mittal",
			"email":   "dhruv.mittal@razorpay.com",
			"contact": "+60102102460",
		},
		"subscription_registration": map[string]interface{}{
			"method":       "card",
			"max_amount":   "500",
			"expire_at":    "1798703336",
			"frequency":    "as_presented",
			"auth_type":    "netbanking",
			"bank_account": map[string]interface{}{},
		},
		"receipt":      "Receipt No. 1323",
		"email_notify": true,
		"sms_notify":   true,
		"expire_by":    float64(1798703336),
		"notes": map[string]interface{}{
			"note_key 1": "Beam me up Scotty",
			"note_key 2": "Tea. Earl Gray. Hot.",
		},
		"short_url": "https://rzp.io/i/xxxxx",
		"status":    "issued",
	}

	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Invalid request parameters",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful creation with all parameters",
			Request: map[string]interface{}{
				"customer_name":                        "Dhruv Mittal",
				"customer_email":                       "dhruv.mittal@razorpay.com",
				"customer_contact":                     "+60102102460",
				"type":                                 "link",
				"amount":                               "100",
				"currency":                             "MYR",
				"description":                          "Registration Link for Nur Aisyah",
				"subscription_registration_method":     "card",
				"subscription_registration_max_amount": "500",
				"subscription_registration_expire_at":  "1798703336",
				"subscription_registration_frequency":  "as_presented",
				"receipt":                              "Receipt No. 1323",
				"email_notify":                         true,
				"sms_notify":                           true,
				"expire_by":                            float64(1798703336),
				"notes": map[string]interface{}{
					"note_key 1": "Beam me up Scotty",
					"note_key 2": "Tea. Earl Gray. Hot.",
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
			Name: "successful creation with only required parameters",
			Request: map[string]interface{}{
				"customer_name":                        "John Doe",
				"customer_email":                       "john@example.com",
				"customer_contact":                     "+919876543210",
				"type":                                 "link",
				"amount":                               "100",
				"currency":                             "INR",
				"description":                          "Test registration",
				"subscription_registration_method":     "emandate",
				"subscription_registration_max_amount": "1000",
				"subscription_registration_expire_at":  "1798703336",
				"subscription_registration_frequency":  "monthly",
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
			Name: "missing required parameter: customer_name",
			Request: map[string]interface{}{
				"customer_email":                       "dhruv.mittal@razorpay.com",
				"customer_contact":                     "+60102102460",
				"type":                                 "link",
				"amount":                               "100",
				"currency":                             "MYR",
				"description":                          "Registration Link",
				"subscription_registration_method":     "card",
				"subscription_registration_max_amount": "500",
				"subscription_registration_expire_at":  "1798703336",
				"subscription_registration_frequency":  "as_presented",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_name",
		},
		{
			Name: "missing required parameter: customer_email",
			Request: map[string]interface{}{
				"customer_name":                        "Dhruv Mittal",
				"customer_contact":                     "+60102102460",
				"type":                                 "link",
				"amount":                               "100",
				"currency":                             "MYR",
				"description":                          "Registration Link",
				"subscription_registration_method":     "card",
				"subscription_registration_max_amount": "500",
				"subscription_registration_expire_at":  "1798703336",
				"subscription_registration_frequency":  "as_presented",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_email",
		},
		{
			Name: "missing required parameter: customer_contact",
			Request: map[string]interface{}{
				"customer_name":                        "Dhruv Mittal",
				"customer_email":                       "dhruv.mittal@razorpay.com",
				"type":                                 "link",
				"amount":                               "100",
				"currency":                             "MYR",
				"description":                          "Registration Link",
				"subscription_registration_method":     "card",
				"subscription_registration_max_amount": "500",
				"subscription_registration_expire_at":  "1798703336",
				"subscription_registration_frequency":  "as_presented",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_contact",
		},
		{
			Name: "missing required parameter: type",
			Request: map[string]interface{}{
				"customer_name":                        "Dhruv Mittal",
				"customer_email":                       "dhruv.mittal@razorpay.com",
				"customer_contact":                     "+60102102460",
				"amount":                               "100",
				"currency":                             "MYR",
				"description":                          "Registration Link",
				"subscription_registration_method":     "card",
				"subscription_registration_max_amount": "500",
				"subscription_registration_expire_at":  "1798703336",
				"subscription_registration_frequency":  "as_presented",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: type",
		},
		{
			Name: "missing required parameter: amount",
			Request: map[string]interface{}{
				"customer_name":                        "Dhruv Mittal",
				"customer_email":                       "dhruv.mittal@razorpay.com",
				"customer_contact":                     "+60102102460",
				"type":                                 "link",
				"currency":                             "MYR",
				"description":                          "Registration Link",
				"subscription_registration_method":     "card",
				"subscription_registration_max_amount": "500",
				"subscription_registration_expire_at":  "1798703336",
				"subscription_registration_frequency":  "as_presented",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: amount",
		},
		{
			Name: "missing required parameter: currency",
			Request: map[string]interface{}{
				"customer_name":                        "Dhruv Mittal",
				"customer_email":                       "dhruv.mittal@razorpay.com",
				"customer_contact":                     "+60102102460",
				"type":                                 "link",
				"amount":                               "100",
				"description":                          "Registration Link",
				"subscription_registration_method":     "card",
				"subscription_registration_max_amount": "500",
				"subscription_registration_expire_at":  "1798703336",
				"subscription_registration_frequency":  "as_presented",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: currency",
		},
		{
			Name: "missing required parameter: description",
			Request: map[string]interface{}{
				"customer_name":                        "Dhruv Mittal",
				"customer_email":                       "dhruv.mittal@razorpay.com",
				"customer_contact":                     "+60102102460",
				"type":                                 "link",
				"amount":                               "100",
				"currency":                             "MYR",
				"subscription_registration_method":     "card",
				"subscription_registration_max_amount": "500",
				"subscription_registration_expire_at":  "1798703336",
				"subscription_registration_frequency":  "as_presented",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: description",
		},
		{
			Name: "missing subscription_registration_method",
			Request: map[string]interface{}{
				"customer_name":                        "Dhruv Mittal",
				"customer_email":                       "dhruv.mittal@razorpay.com",
				"customer_contact":                     "+60102102460",
				"type":                                 "link",
				"amount":                               "100",
				"currency":                             "MYR",
				"description":                          "Registration Link",
				"subscription_registration_max_amount": "500",
				"subscription_registration_expire_at":  "1798703336",
				"subscription_registration_frequency":  "as_presented",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: " +
				"subscription_registration_method",
		},
		{
			Name: "missing subscription_registration_max_amount",
			Request: map[string]interface{}{
				"customer_name":                       "Dhruv Mittal",
				"customer_email":                      "dhruv.mittal@razorpay.com",
				"customer_contact":                    "+60102102460",
				"type":                                "link",
				"amount":                              "100",
				"currency":                            "MYR",
				"description":                         "Registration Link",
				"subscription_registration_method":    "card",
				"subscription_registration_expire_at": "1798703336",
				"subscription_registration_frequency": "as_presented",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: " +
				"subscription_registration_max_amount",
		},
		{
			Name: "missing subscription_registration_expire_at",
			Request: map[string]interface{}{
				"customer_name":                        "Dhruv Mittal",
				"customer_email":                       "dhruv.mittal@razorpay.com",
				"customer_contact":                     "+60102102460",
				"type":                                 "link",
				"amount":                               "100",
				"currency":                             "MYR",
				"description":                          "Registration Link",
				"subscription_registration_method":     "card",
				"subscription_registration_max_amount": "500",
				"subscription_registration_frequency":  "as_presented",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: " +
				"subscription_registration_expire_at",
		},
		{
			Name: "missing subscription_registration_frequency",
			Request: map[string]interface{}{
				"customer_name":                        "Dhruv Mittal",
				"customer_email":                       "dhruv.mittal@razorpay.com",
				"customer_contact":                     "+60102102460",
				"type":                                 "link",
				"amount":                               "100",
				"currency":                             "MYR",
				"description":                          "Registration Link",
				"subscription_registration_method":     "card",
				"subscription_registration_max_amount": "500",
				"subscription_registration_expire_at":  "1798703336",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: " +
				"subscription_registration_frequency",
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				"customer_name": "Dhruv Mittal",
				"type":          123,
				"amount":        true,
				"currency":      456,
			},
			MockHttpClient: nil,
			ExpectError:    true,
		},
		{
			Name: "API error response",
			Request: map[string]interface{}{
				"customer_name":                        "Dhruv Mittal",
				"customer_email":                       "dhruv.mittal@razorpay.com",
				"customer_contact":                     "+60102102460",
				"type":                                 "link",
				"amount":                               "100",
				"currency":                             "MYR",
				"description":                          "Registration Link",
				"subscription_registration_method":     "card",
				"subscription_registration_max_amount": "500",
				"subscription_registration_expire_at":  "1798703336",
				"subscription_registration_frequency":  "as_presented",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(mock.Endpoint{
					Path:     apiPath,
					Method:   "POST",
					Response: errorResponse,
				})
			},
			ExpectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(
				t,
				tc,
				CreateSubscriptionRegistrationAuthLink,
				"SubscriptionRegistration",
			)
		})
	}
}
