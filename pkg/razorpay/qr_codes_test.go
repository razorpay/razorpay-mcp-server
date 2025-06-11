package razorpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/razorpay/razorpay-go/v2/constants"

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

func Test_FetchQRCodesByCustomerID(t *testing.T) {
	qrCodesPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.QRCODE_URL,
	)

	customerQRCodesResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(1),
		"items": []interface{}{
			map[string]interface{}{
				"id":                       "qr_HMsgvioW64f0vh",
				"entity":                   "qr_code",
				"created_at":               float64(1623660959),
				"name":                     "Store_1",
				"usage":                    "single_use",
				"type":                     "upi_qr",
				"image_url":                "https://rzp.io/i/DTa2eQR",
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
		},
	}

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The id provided is not a valid id",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful fetch QR codes by customer ID",
			Request: map[string]interface{}{
				"customer_id": "cust_HKsR5se84c5LTO",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     qrCodesPath,
						Method:   "GET",
						Response: customerQRCodesResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: customerQRCodesResp,
		},
		{
			Name:           "missing required customer_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: customer_id",
		},
		{
			Name: "validator error - invalid customer_id parameter type",
			Request: map[string]interface{}{
				"customer_id": 12345,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: customer_id",
		},
		{
			Name: "API error - invalid customer ID",
			Request: map[string]interface{}{
				"customer_id": "invalid_customer_id",
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
			ExpectError: true,
			ExpectedErrMsg: "fetching QR codes failed: " +
				"The id provided is not a valid id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchQRCodesByCustomerID, "QR Codes by Customer ID")
		})
	}
}

func Test_FetchQRCodesByPaymentID(t *testing.T) {
	qrCodesPath := fmt.Sprintf(
		"/%s%s",
		constants.VERSION_V1,
		constants.QRCODE_URL,
	)

	paymentQRCodesResp := map[string]interface{}{
		"entity": "collection",
		"count":  float64(1),
		"items": []interface{}{
			map[string]interface{}{
				"id":                       "qr_HMsqRoeVwKbwAF",
				"entity":                   "qr_code",
				"created_at":               float64(1623661499),
				"name":                     "Fresh Groceries",
				"usage":                    "multiple_use",
				"type":                     "upi_qr",
				"image_url":                "https://rzp.io/i/eI9XD54Q",
				"payment_amount":           nil,
				"status":                   "active",
				"description":              "Buy fresh groceries",
				"fixed_amount":             false,
				"payments_amount_received": float64(1000),
				"payments_count_received":  float64(1),
				"notes":                    []interface{}{},
				"customer_id":              "cust_HKsR5se84c5LTO",
				"close_by":                 float64(1624472999),
				"close_reason":             nil,
			},
		},
	}

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The id provided is not a valid id",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful fetch QR codes by payment ID",
			Request: map[string]interface{}{
				"payment_id": "pay_Di5iqCqA1WEHq6",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     qrCodesPath,
						Method:   "GET",
						Response: paymentQRCodesResp,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: paymentQRCodesResp,
		},
		{
			Name:           "missing required payment_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: payment_id",
		},
		{
			Name: "validator error - invalid payment_id parameter type",
			Request: map[string]interface{}{
				"payment_id": 12345,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: payment_id",
		},
		{
			Name: "API error - invalid payment ID",
			Request: map[string]interface{}{
				"payment_id": "invalid_payment_id",
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
			ExpectError: true,
			ExpectedErrMsg: "fetching QR codes failed: " +
				"The id provided is not a valid id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchQRCodesByPaymentID, "QR Codes by Payment ID")
		})
	}
}

func TestFetchQRCode(t *testing.T) {
	// Initialize necessary variables
	qrID := "qr_FuZIYx6rMbP6gs"
	apiPath := fmt.Sprintf(
		"/%s%s/%s",
		constants.VERSION_V1,
		constants.QRCODE_URL,
		qrID,
	)

	// Successful response based on Razorpay docs
	successResponse := map[string]interface{}{
		"id":                       qrID,
		"entity":                   "qr_code",
		"created_at":               float64(1623915088),
		"name":                     "Store_1",
		"usage":                    "single_use",
		"type":                     "upi_qr",
		"image_url":                "https://rzp.io/i/oCswTOcCo",
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
	}

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":        "BAD_REQUEST_ERROR",
			"description": "The QR code ID provided is invalid",
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful fetch QR code by ID",
			Request: map[string]interface{}{
				"qr_code_id": qrID,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     apiPath,
						Method:   "GET",
						Response: successResponse,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name:           "missing required qr_code_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: qr_code_id",
		},
		{
			Name: "validator error - invalid qr_code_id parameter type",
			Request: map[string]interface{}{
				"qr_code_id": 12345,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: qr_code_id",
		},
		{
			Name: "API error - invalid QR code ID",
			Request: map[string]interface{}{
				"qr_code_id": qrID,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     apiPath,
						Method:   "GET",
						Response: errorResp,
					},
				)
			},
			ExpectError: true,
			ExpectedErrMsg: "fetching QR code failed: " +
				"The QR code ID provided is invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchQRCode, "QR Code")
		})
	}
}

func TestFetchPaymentsForQRCode(t *testing.T) {
	apiPath := "/" + constants.VERSION_V1 +
		constants.QRCODE_URL + "/qr_test123/payments"

	successResponse := map[string]interface{}{
		"entity": "collection",
		"count":  float64(2),
		"items": []interface{}{
			map[string]interface{}{
				"id":              "pay_test123",
				"entity":          "payment",
				"amount":          float64(500),
				"currency":        "INR",
				"status":          "captured",
				"method":          "upi",
				"amount_refunded": float64(0),
				"refund_status":   nil,
				"captured":        true,
				"description":     "QRv2 Payment",
				"customer_id":     "cust_test123",
				"created_at":      float64(1623662800),
			},
			map[string]interface{}{
				"id":              "pay_test456",
				"entity":          "payment",
				"amount":          float64(1000),
				"currency":        "INR",
				"status":          "refunded",
				"method":          "upi",
				"amount_refunded": float64(1000),
				"refund_status":   "full",
				"captured":        true,
				"description":     "QRv2 Payment",
				"customer_id":     "cust_test123",
				"created_at":      float64(1623661533),
			},
		},
	}

	tests := []RazorpayToolTestCase{
		{
			Name: "successful fetch payments for QR code",
			Request: map[string]interface{}{
				"qr_code_id": "qr_test123",
				"count":      10,
				"from":       1623661000,
				"to":         1623663000,
				"skip":       0,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:     apiPath,
						Method:   "GET",
						Response: successResponse,
					},
				)
			},
			ExpectError:    false,
			ExpectedResult: successResponse,
		},
		{
			Name: "missing required parameter",
			Request: map[string]interface{}{
				"count": 10,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: qr_code_id",
		},
		{
			Name: "invalid parameter type",
			Request: map[string]interface{}{
				"qr_code_id": 123,
			},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "invalid parameter type: qr_code_id",
		},
		{
			Name: "API error",
			Request: map[string]interface{}{
				"qr_code_id": "qr_test123",
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
				return mock.NewHTTPClient(
					mock.Endpoint{
						Path:   apiPath,
						Method: "GET",
						Response: map[string]interface{}{
							"error": map[string]interface{}{
								"code":        "BAD_REQUEST_ERROR",
								"description": "mock error",
							},
						},
					},
				)
			},
			ExpectError:    true,
			ExpectedErrMsg: "fetching payments for QR code failed: mock error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, FetchPaymentsForQRCode, "QR Code Payments")
		})
	}
}

func TestCloseQRCode(t *testing.T) {
	successResponse := map[string]interface{}{
		"id":                       "qr_HMsVL8HOpbMcjU",
		"entity":                   "qr_code",
		"created_at":               float64(1623660301),
		"name":                     "Store_1",
		"usage":                    "single_use",
		"type":                     "upi_qr",
		"image_url":                "https://rzp.io/i/BWcUVrLp",
		"payment_amount":           float64(300),
		"status":                   "closed",
		"description":              "For Store 1",
		"fixed_amount":             true,
		"payments_amount_received": float64(0),
		"payments_count_received":  float64(0),
		"notes": map[string]interface{}{
			"purpose": "Test UPI QR Code notes",
		},
		"customer_id":  "cust_HKsR5se84c5LTO",
		"close_by":     float64(1681615838),
		"closed_at":    float64(1623660445),
		"close_reason": "on_demand",
	}

	baseAPIPath := fmt.Sprintf("/%s%s", constants.VERSION_V1, constants.QRCODE_URL)
	qrCodeID := "qr_HMsVL8HOpbMcjU"
	apiPath := fmt.Sprintf("%s/%s/close", baseAPIPath, qrCodeID)

	tests := []RazorpayToolTestCase{
		{
			Name: "successful close QR code",
			Request: map[string]interface{}{
				"qr_code_id": qrCodeID,
			},
			MockHttpClient: func() (*http.Client, *httptest.Server) {
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
			Name:           "missing required qr_code_id parameter",
			Request:        map[string]interface{}{},
			MockHttpClient: nil,
			ExpectError:    true,
			ExpectedErrMsg: "missing required parameter: qr_code_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runToolTest(t, tc, CloseQRCode, "QR Code")
		})
	}
}
