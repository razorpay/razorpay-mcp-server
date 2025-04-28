package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_FetchPayment(t *testing.T) {
	fetchPaymentPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	paymentResp := map[string]interface{}{
		"id":     "pay_MT48CvBhIC98MQ",
		"amount": float64(1000),
		"status": "captured",
	}

	paymentNotFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "payment not found",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful payment fetch",
			Request: map[string]interface{}{
				"payment_id": "pay_MT48CvBhIC98MQ",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchPaymentPathFmt, "pay_MT48CvBhIC98MQ"),
						Method:   "GET",
						Response: paymentResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: paymentResp,
		},
		{
			Name: "payment not found",
			Request: map[string]interface{}{
				"payment_id": "pay_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchPaymentPathFmt, "pay_invalid"),
						Method:   "GET",
						Response: paymentNotFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching payment failed: payment not found",
		},
		{
			Name:           "missing payment_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchPayment, "Payment")
		})
	}
}
