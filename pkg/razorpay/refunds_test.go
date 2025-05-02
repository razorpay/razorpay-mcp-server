package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_CreateRefund(t *testing.T) {
	createRefundPathFmt := fmt.Sprintf(
		"/%s%s/%%s/refund",
		constants.VERSION_V1,
		constants.PAYMENT_URL,
	)

	// Define test responses
	successfulRefundResp := map[string]interface{}{
		"id":              "rfnd_FP8QHiV938haTz",
		"entity":          "refund",
		"amount":          float64(500100),
		"currency":        "INR",
		"payment_id":      "pay_29QQoUBi66xm2f",
		"notes":           map[string]interface{}{},
		"receipt":         "Receipt No. 31",
		"acquirer_data":   map[string]interface{}{"arn": nil},
		"created_at":      float64(1597078866),
		"batch_id":        nil,
		"status":          "processed",
		"speed_processed": "normal",
		"speed_requested": "normal",
	}

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "Razorpay API error: Bad request",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful full refund",
			Request: map[string]interface{}{
				"payment_id": "pay_29QQoUBi66xm2f",
				"amount":     float64(500100),
				"receipt":    "Receipt No. 31",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(createRefundPathFmt, "pay_29QQoUBi66xm2f"),
						Method:   "POST",
						Response: successfulRefundResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulRefundResp,
		},
		{
			Name: "refund with speed parameter",
			Request: map[string]interface{}{
				"payment_id": "pay_29QQoUBi66xm2f",
				"amount":     float64(500100),
				"speed":      "optimum",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				speedRefundResp := map[string]interface{}{
					"id":              "rfnd_HzAbPEkKtRq48V",
					"entity":          "refund",
					"amount":          float64(500100),
					"payment_id":      "pay_29QQoUBi66xm2f",
					"status":          "processed",
					"speed_processed": "instant",
					"speed_requested": "optimum",
				}
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(createRefundPathFmt, "pay_29QQoUBi66xm2f"),
						Method:   "POST",
						Response: speedRefundResp,
					},
				)
			},
			ExpectError: false,
			ExpectedResult: map[string]interface{}{
				"id":              "rfnd_HzAbPEkKtRq48V",
				"entity":          "refund",
				"amount":          float64(500100),
				"payment_id":      "pay_29QQoUBi66xm2f",
				"status":          "processed",
				"speed_processed": "instant",
				"speed_requested": "optimum",
			},
		},
		{
			Name: "refund API server error",
			Request: map[string]interface{}{
				"payment_id": "pay_29QQoUBi66xm2f",
				"amount":     float64(500100),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(createRefundPathFmt, "pay_29QQoUBi66xm2f"),
						Method:   "POST",
						Response: errorResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "creating refund failed: Razorpay API error: Bad request",
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				// Missing payment_id parameter
				"amount": "not-a-number",  // Wrong type for amount
				"speed":  12345,           // Wrong type for speed
				"notes":  "not-an-object", // Wrong type for notes
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:\n- " +
				"missing required parameter: payment_id\n- " +
				"invalid parameter type: amount\n- " +
				"invalid parameter type: speed\n- " +
				"invalid parameter type: notes",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CreateRefund, "Refund")
		})
	}
}

func Test_FetchRefund(t *testing.T) {
	fetchRefundPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.REFUND_URL,
	)

	// Define test response for successful refund fetch
	successfulRefundResp := map[string]interface{}{
		"id":         "rfnd_DfjjhJC6eDvUAi",
		"entity":     "refund",
		"amount":     float64(6000),
		"currency":   "INR",
		"payment_id": "pay_EpkFDYRirena0f",
		"notes": map[string]interface{}{
			"comment": "Issuing an instant refund",
		},
		"receipt": nil,
		"acquirer_data": map[string]interface{}{
			"arn": "10000000000000",
		},
		"created_at":      float64(1589521675),
		"batch_id":        nil,
		"status":          "processed",
		"speed_processed": "optimum",
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
			Name: "successful refund fetch",
			Request: map[string]interface{}{
				"refund_id": "rfnd_DfjjhJC6eDvUAi",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchRefundPathFmt, "rfnd_DfjjhJC6eDvUAi"),
						Method:   "GET",
						Response: successfulRefundResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulRefundResp,
		},
		{
			Name: "refund id not found",
			Request: map[string]interface{}{
				"refund_id": "rfnd_nonexistent",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(fetchRefundPathFmt, "rfnd_nonexistent"),
						Method:   "GET",
						Response: notFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching refund failed: The id provided does not exist",
		},
		{
			Name:           "missing refund_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: refund_id",
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				// Missing refund_id parameter
				"non_existent_param": 12345, // Additional parameter that doesn't exist
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: refund_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchRefund, "Refund")
		})
	}
}

func Test_UpdateRefund(t *testing.T) {
	updateRefundPathFmt := fmt.Sprintf(
		"/%s%s/%%s",
		constants.VERSION_V1,
		constants.REFUND_URL,
	)

	// Define test response for successful refund update
	successfulUpdateResp := map[string]interface{}{
		"id":         "rfnd_DfjjhJC6eDvUAi",
		"entity":     "refund",
		"amount":     float64(300100),
		"currency":   "INR",
		"payment_id": "pay_FIKOnlyii5QGNx",
		"notes": map[string]interface{}{
			"notes_key_1": "Beam me up Scotty.",
			"notes_key_2": "Engage",
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
			Name: "successful refund update",
			Request: map[string]interface{}{
				"refund_id": "rfnd_DfjjhJC6eDvUAi",
				"notes": map[string]interface{}{
					"notes_key_1": "Beam me up Scotty.",
					"notes_key_2": "Engage",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(updateRefundPathFmt, "rfnd_DfjjhJC6eDvUAi"),
						Method:   "PATCH",
						Response: successfulUpdateResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successfulUpdateResp,
		},
		{
			Name: "refund id not found",
			Request: map[string]interface{}{
				"refund_id": "rfnd_nonexistent",
				"notes": map[string]interface{}{
					"note_key": "Test note",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     fmt.Sprintf(updateRefundPathFmt, "rfnd_nonexistent"),
						Method:   "PATCH",
						Response: notFoundResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "updating refund failed: The id provided does not exist",
		},
		{
			Name:           "missing refund_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: refund_id",
		},
		{
			Name: "missing notes parameter",
			Request: map[string]interface{}{
				"refund_id": "rfnd_DfjjhJC6eDvUAi",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: notes",
		},
		{
			Name: "multiple validation errors",
			Request: map[string]interface{}{
				// Missing both refund_id and notes parameters
				"non_existent_param": 12345, // Additional parameter that doesn't exist
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors:\n- " +
				"missing required parameter: refund_id\n- " +
				"missing required parameter: notes",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, UpdateRefund, "Refund")
		})
	}
}
