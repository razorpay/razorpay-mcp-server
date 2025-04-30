package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/constants"

	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mock"
)

func Test_CreateQRCode(t *testing.T) {
	createQRCodePath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.QRCODE_URL,
	)

	// Define common response maps to be reused
	qrCodeWithAllParamsResp := map[string]interface{}{
		"id":                       "qr_HMsVL8HOpbMcjU",
		"entity":                   "qr_code",
		"created_at":               float64(1623660301),
		"name":                     "Store Front Display",
		"usage":                    "single_use",
		"type":                     "upi_qr",
		"image_url":                "https://rzp.io/i/BWcUVrLp",
		"payment_amount":           float64(300),
		"status":                   "active",
		"description":              "For Store 1",
		"fixed_amount":             true,
		"payments_amount_received": float64(0),
		"payments_count_received":  float64(0),
		"notes": map[string]interface{}{
			"purpose": "Test UPI QR Code notes",
		},
		"customer_id": "cust_HKsR5se84c5LTO",
		"close_by":    float64(1681615838),
	}

	qrCodeWithRequiredParamsResp := map[string]interface{}{
		"id":                       "qr_HMsVL8HOpbMcjU",
		"entity":                   "qr_code",
		"created_at":               float64(1623660301),
		"usage":                    "multiple_use",
		"type":                     "upi_qr",
		"image_url":                "https://rzp.io/i/BWcUVrLp",
		"status":                   "active",
		"fixed_amount":             false,
		"payments_amount_received": float64(0),
		"payments_count_received":  float64(0),
	}

	qrCodeWithoutPaymentAmountResp := map[string]interface{}{
		"id":                       "qr_HMsVL8HOpbMcjU",
		"entity":                   "qr_code",
		"created_at":               float64(1623660301),
		"name":                     "Store Front Display",
		"usage":                    "single_use",
		"type":                     "upi_qr",
		"image_url":                "https://rzp.io/i/BWcUVrLp",
		"status":                   "active",
		"description":              "For Store 1",
		"fixed_amount":             false,
		"payments_amount_received": float64(0),
		"payments_count_received":  float64(0),
	}

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The type field is invalid",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful QR code creation with all parameters",
			Request: map[string]interface{}{
				"type":           "upi_qr",
				"name":           "Store Front Display",
				"usage":          "single_use",
				"fixed_amount":   true,
				"payment_amount": float64(300),
				"description":    "For Store 1",
				"customer_id":    "cust_HKsR5se84c5LTO",
				"close_by":       float64(1681615838),
				"notes": map[string]interface{}{
					"purpose": "Test UPI QR Code notes",
				},
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createQRCodePath,
						Method:   "POST",
						Response: qrCodeWithAllParamsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: qrCodeWithAllParamsResp,
		},
		{
			Name: "successful QR code creation with required params only",
			Request: map[string]interface{}{
				"type":  "upi_qr",
				"usage": "multiple_use",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createQRCodePath,
						Method:   "POST",
						Response: qrCodeWithRequiredParamsResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: qrCodeWithRequiredParamsResp,
		},
		{
			Name: "successful QR code creation without payment amount",
			Request: map[string]interface{}{
				"type":         "upi_qr",
				"name":         "Store Front Display",
				"usage":        "single_use",
				"fixed_amount": false,
				"description":  "For Store 1",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createQRCodePath,
						Method:   "POST",
						Response: qrCodeWithoutPaymentAmountResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: qrCodeWithoutPaymentAmountResp,
		},
		{
			Name: "missing required type parameter",
			Request: map[string]interface{}{
				"usage": "single_use",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: type",
		},
		{
			Name: "missing required usage parameter",
			Request: map[string]interface{}{
				"type": "upi_qr",
			},
			MockHttpClient: nil, // No HTTP client needed for validation error
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: usage",
		},
		{
			Name: "invalid type parameter",
			Request: map[string]interface{}{
				"type":  "invalid_type",
				"usage": "single_use",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     createQRCodePath,
						Method:   "POST",
						Response: errorResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "creating QR code failed: The type field is invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CreateQRCode, "QR Code")
		})
	}
}
