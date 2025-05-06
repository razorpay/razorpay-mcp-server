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
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: type",
		},
		{
			Name: "missing required usage parameter",
			Request: map[string]interface{}{
				"type": "upi_qr",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: usage",
		},
		{
			Name: "validator error - invalid parameter type",
			Request: map[string]interface{}{
				"type":  123,
				"usage": "single_use",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors",
		},
		{
			Name: "fixed_amount true but payment_amount missing",
			Request: map[string]interface{}{
				"type":         "upi_qr",
				"usage":        "single_use",
				"fixed_amount": true,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "payment_amount is required when fixed_amount is true",
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

func Test_FetchAllQRCodes(t *testing.T) {
	qrCodesPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.QRCODE_URL,
	)

	allQRCodesResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":                       "qr_HO2jGkWReVBMNu",
				"entity":                   "qr_code",
				"created_at":               float64(1623914648),
				"name":                     "Store_1",
				"usage":                    "single_use",
				"type":                     "upi_qr",
				"image_url":                "https://rzp.io/i/w2CEwYmkAu",
				"payment_amount":           float64(300),
				"status":                   "active",
				"description":              "For Store 1",
				"fixed_amount":             true,
				"payments_amount_received": float64(0),
				"payments_count_received":  float64(0),
				"notes": map[string]interface{}{
					"purpose": "Test UPI QR Code notes",
				},
				"customer_id":  "cust_HKsR5se84c5LTO",
				"close_by":     float64(1681615838),
				"closed_at":    nil,
				"close_reason": nil,
			},
			map[string]interface{}{
				"id":                       "qr_HO2e0813YlchUn",
				"entity":                   "qr_code",
				"created_at":               float64(1623914349),
				"name":                     "Acme Groceries",
				"usage":                    "multiple_use",
				"type":                     "upi_qr",
				"image_url":                "https://rzp.io/i/X6QM7LL",
				"payment_amount":           nil,
				"status":                   "closed",
				"description":              "Buy fresh groceries",
				"fixed_amount":             false,
				"payments_amount_received": float64(200),
				"payments_count_received":  float64(1),
				"notes": map[string]interface{}{
					"Branch": "Bangalore - Rajaji Nagar",
				},
				"customer_id":  "cust_HKsR5se84c5LTO",
				"close_by":     float64(1625077799),
				"closed_at":    float64(1623914515),
				"close_reason": "on_demand",
			},
		},
	}

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The query parameters are invalid",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name:    "successful fetch all QR codes with no parameters",
			Request: map[string]interface{}{},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     qrCodesPath,
						Method:   "GET",
						Response: allQRCodesResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: allQRCodesResp,
		},
		{
			Name: "successful fetch all QR codes with count parameter",
			Request: map[string]interface{}{
				"count": float64(2),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     qrCodesPath,
						Method:   "GET",
						Response: allQRCodesResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: allQRCodesResp,
		},
		{
			Name: "successful fetch all QR codes with pagination parameters",
			Request: map[string]interface{}{
				"from":  float64(1622000000),
				"to":    float64(1625000000),
				"count": float64(2),
				"skip":  float64(0),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     qrCodesPath,
						Method:   "GET",
						Response: allQRCodesResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: allQRCodesResp,
		},
		{
			Name: "invalid parameters - caught by SDK",
			Request: map[string]interface{}{
				"count": float64(-1),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   qrCodesPath,
						Method: "GET",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "The count value should be greater than or equal to 1",
							},
						},
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "fetching QR codes failed: " +
				"The count value should be greater than or equal to 1",
		},
		{
			Name: "validator error - invalid count parameter type",
			Request: map[string]interface{}{
				"count": "not-a-number",
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "Validation errors",
		},
		{
			Name: "API error response",
			Request: map[string]interface{}{
				"count": float64(1000),
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     qrCodesPath,
						Method:   "GET",
						Response: errorResp,
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching QR codes failed: The query parameters are invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchAllQRCodes, "QR Codes")
		})
	}
}
