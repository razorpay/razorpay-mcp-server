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

func Test_FetchMultipleRefundsForPayment(t *testing.T) {
	fetchMultipleRefundsPathFmt := fmt.Sprintf(
		"/%s%s/%%s/refunds",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	// Define test response for successful multiple refunds fetch
	successfulMultipleRefundsResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":         "rfnd_FP8DDKxqJif6ca",
				"entity":     "refund",
				"amount":     float64(300100),
				"currency":   "INR",
				"payment_id": "pay_29QQoUBi66xm2f",
				"notes": map[string]interface{}{
					"comment": "Comment for refund",
				},
				"receipt": nil,
				"acquirer_data": map[string]interface{}{
					"arn": "10000000000000",
				},
				"created_at":      float64(1597078124),
				"batch_id":        nil,
				"status":          "processed",
				"speed_processed": "normal",
				"speed_requested": "optimum",
			},
			map[string]interface{}{
				"id":         "rfnd_FP8DRfu3ygfOaC",
				"entity":     "refund",
				"amount":     float64(200000),
				"currency":   "INR",
				"payment_id": "pay_29QQoUBi66xm2f",
				"notes": map[string]interface{}{
					"comment": "Comment for refund",
				},
				"receipt": nil,
				"acquirer_data": map[string]interface{}{
					"arn": "10000000000000",
				},
				"created_at":      float64(1597078137),
				"batch_id":        nil,
				"status":          "processed",
				"speed_processed": "normal",
				"speed_requested": "optimum",
			},
		},
	}

	// Define error responses
	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Bad request",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful fetch multiple refunds",
			Request: map[string]interface{}{
				"payment_id": "pay_29QQoUBi66xm2f",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							fetchMultipleRefundsPathFmt,
							"pay_29QQoUBi66xm2f",
						),
						Method:   "GET",
						Response: successfulMultipleRefundsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulMultipleRefundsResp,
		},
		{
			Name: "fetch multiple refunds with query params",
			Request: map[string]interface{}{
				"payment_id": "pay_29QQoUBi66xm2f",
				"from":       float64(1500826740),
				"to":         float64(1500826760),
				"count":      float64(10),
				"skip":       float64(0),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							fetchMultipleRefundsPathFmt,
							"pay_29QQoUBi66xm2f",
						),
						Method:   "GET",
						Response: successfulMultipleRefundsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulMultipleRefundsResp,
		},
		{
			Name: "fetch multiple refunds api error",
			Request: map[string]interface{}{
				"payment_id": "pay_invalid",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							fetchMultipleRefundsPathFmt,
							"pay_invalid",
						),
						Method:   "GET",
						Response: errorResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching multiple refunds failed: Bad request",
		},
		{
			Name:           "missing payment_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchMultipleRefundsForPayment, "Payment")
		})
	}
}

func Test_FetchSpecificRefundForPayment(t *testing.T) {
	fetchSpecificRefundPathFmt := fmt.Sprintf(
		"/%s%s/%%s/refunds/%%s",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	// Define test response for successful specific refund fetch
	successfulSpecificRefundResp := map[string]interface{}{
		"id":         "rfnd_AABBdHIieexn5c",
		"entity":     "refund",
		"amount":     float64(300100),
		"currency":   "INR",
		"payment_id": "pay_FIKOnlyii5QGNx",
		"notes": map[string]interface{}{
			"comment": "Comment for refund",
		},
		"receipt":         nil,
		"acquirer_data":   map[string]interface{}{"arn": "10000000000000"},
		"created_at":      float64(1597078124),
		"batch_id":        nil,
		"status":          "processed",
		"speed_processed": "normal",
		"speed_requested": "optimum",
	}

	// Define error responses
	notFoundResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The id provided does not exist",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful specific refund fetch",
			Request: map[string]interface{}{
				"payment_id": "pay_FIKOnlyii5QGNx",
				"refund_id":  "rfnd_AABBdHIieexn5c",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							fetchSpecificRefundPathFmt,
							"pay_FIKOnlyii5QGNx",
							"rfnd_AABBdHIieexn5c",
						),
						Method:   "GET",
						Response: successfulSpecificRefundResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulSpecificRefundResp,
		},
		{
			Name: "refund id not found",
			Request: map[string]interface{}{
				"payment_id": "pay_FIKOnlyii5QGNx",
				"refund_id":  "rfnd_nonexistent",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path: fmt.Sprintf(
							fetchSpecificRefundPathFmt,
							"pay_FIKOnlyii5QGNx",
							"rfnd_nonexistent",
						),
						Method:   "GET",
						Response: notFoundResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "fetching specific refund for payment failed: " +
				"The id provided does not exist",
		},
		{
			Name: "missing payment_id parameter",
			Request: map[string]interface{}{
				"refund_id": "rfnd_AABBdHIieexn5c",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_id",
		},
		{
			Name: "missing refund_id parameter",
			Request: map[string]interface{}{
				"payment_id": "pay_FIKOnlyii5QGNx",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: refund_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchSpecificRefundForPayment, "Payment")
		})
	}
}
