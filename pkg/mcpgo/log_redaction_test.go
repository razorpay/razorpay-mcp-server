package mcpgo

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/razorpay/razorpay-mcp-server/pkg/log"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// newCapturingLogger returns a production slog logger that writes to buf
// through the given handler kind, so the test sees exactly the bytes a
// deployed server would write.
func newCapturingLogger(kind string, buf *bytes.Buffer) log.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		// Drop the timestamp so digit values cannot match it by chance.
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}
	if kind == "json" {
		return log.NewSloggerWithHandler(slog.NewJSONHandler(buf, opts))
	}
	return log.NewSloggerWithHandler(slog.NewTextHandler(buf, opts))
}

// testAuthHeaderValue is a fake Basic credential built at runtime, so no
// encoded credential literal sits in the source.
var testAuthHeaderValue = base64.StdEncoding.EncodeToString(
	[]byte("redaction-test:not-a-real-secret"))

// sensitiveToolCall returns a tools/call request and result whose sensitive
// values must never reach the log.
func sensitiveToolCall() (*mcp.CallToolRequest, *mcp.CallToolResult) {
	req := &mcp.CallToolRequest{}
	req.Method = string(mcp.MethodToolsCall)
	req.Header = http.Header{
		"Authorization": []string{"Basic " + testAuthHeaderValue},
	}
	req.Params.Name = "submit_otp"
	req.Params.Arguments = map[string]any{
		"payment_id":       "pay_LOGSAFE000001",
		"customer_id":      "cust_LOGSAFE000002",
		"otp_string":       "482913",
		"customer_contact": "+919812345670",
		"customer_email":   "payer.logsafe@example.com",
		"customer_name":    "Logsafe Payer",
		"customerEmail":    "camel.logsafe@example.com",
		"accessToken":      "camel-token-logsafe",
		"x-api-key":        "dash-key-logsafe",
		"email_id":         "id.logsafe@example.com",
		"card": map[string]any{
			"number": "4111111111111111",
			"cvv":    "123",
		},
	}
	result := mcp.NewToolResultText(
		`{"card":{"number":"5555555555554444"},"token":"tok-res-secret"}`)
	return req, result
}

var sensitiveValues = []string{
	"482913", "4111111111111111", "5555555555554444", "tok-res-secret",
	testAuthHeaderValue, "+919812345670", "payer.logsafe@example.com",
	"Logsafe Payer", "camel.logsafe@example.com", "camel-token-logsafe",
	"dash-key-logsafe", "id.logsafe@example.com",
	// The card object is redacted whole, so its keys never appear either.
	"cvv", "4111",
}

func runAllHooks(obs *observability.Observability) {
	hooks := SetupHooks(obs)
	ctx := context.Background()
	req, result := sensitiveToolCall()

	for _, h := range hooks.OnBeforeAny {
		h(ctx, 1, mcp.MethodToolsCall, req)
	}
	for _, h := range hooks.OnBeforeCallTool {
		h(ctx, 1, req)
	}
	for _, h := range hooks.OnAfterCallTool {
		h(ctx, 1, req, result)
	}
	for _, h := range hooks.OnSuccess {
		h(ctx, 1, mcp.MethodToolsCall, req, result)
	}
	for _, h := range hooks.OnError {
		h(ctx, 1, mcp.MethodToolsCall, req, errors.New("boom"))
	}
}

func TestSetupHooks_DoNotLogSensitiveValues(t *testing.T) {
	for _, kind := range []string{"json", "text"} {
		t.Run(kind, func(t *testing.T) {
			var buf bytes.Buffer
			runAllHooks(&observability.Observability{
				Logger: newCapturingLogger(kind, &buf),
			})

			logged := buf.String()
			assert.NotEmpty(t, logged)
			for _, secret := range sensitiveValues {
				assert.NotContains(t, logged, secret)
			}
			// Logs stay useful: tool name, ids, key names and outcome remain.
			assert.Contains(t, logged, "submit_otp")
			assert.Contains(t, logged, "pay_LOGSAFE000001")
			assert.Contains(t, logged, "cust_LOGSAFE000002")
			assert.Contains(t, logged, "otp_string")
			assert.Contains(t, logged, redactedValue)
			assert.Contains(t, logged, "boom")
		})
	}
}

func TestRedactForLog(t *testing.T) {
	in := map[string]any{
		"payment_id":    "pay_1",
		"OTP":           "111111",
		"key_secret":    "s3",
		"Authorization": "Bearer abc",
		"contact":       "+919999999999",
		"nested": []any{
			map[string]any{"cvv": "999", "amount": 100},
		},
	}
	out := RedactForLog(in).(map[string]any)

	assert.Equal(t, "pay_1", out["payment_id"])
	assert.Equal(t, redactedValue, out["OTP"])
	assert.Equal(t, redactedValue, out["key_secret"])
	assert.Equal(t, redactedValue, out["Authorization"])
	assert.Equal(t, redactedValue, out["contact"])
	nested := out["nested"].([]any)[0].(map[string]any)
	assert.Equal(t, redactedValue, nested["cvv"])
	assert.EqualValues(t, 100, nested["amount"])
	// The input is not modified.
	assert.Equal(t, "111111", in["OTP"])
}

func TestIsSensitiveLogKey(t *testing.T) {
	sensitive := []string{
		// secrets, in any casing or separator style
		"otp", "otp_string", "OTP", "cvv", "card", "number", "expiry_month",
		"token", "access_token", "accessToken", "refresh-token", "api_key",
		"x-api-key", "apiKey", "Authorization", "cookie", "Set-Cookie",
		"password", "key_secret", "keySecret", "razorpay_signature",
		"card_number", "cardNumber",
		// secrets win over an id suffix
		"session_id", "sessionId", "token_id",
		// PII
		"contact", "customer_contact", "customerContact", "email",
		"customer_email", "customerEmail", "Customer-Email", "phone",
		"mobile_number", "customer_name", "customerName", "first_name",
		"vpa", "upi", "billing_address", "account_number", "date_of_birth",
		// id-suffixed contact data
		"email_id", "emailId", "upi_id", "vpa_id",
	}
	for _, key := range sensitive {
		assert.Truef(t, IsSensitiveLogKey(key), "%q should be redacted", key)
	}

	safe := []string{
		"payment_id", "order_id", "customer_id", "customerId", "contact_id",
		"contactID", "payment_link_id", "payout_id", "id", "amount",
		"currency", "status", "method", "description", "notes", "medium",
		"group_name", "groupId", "upi_link", "expire_by", "reference_id",
	}
	for _, key := range safe {
		assert.Falsef(t, IsSensitiveLogKey(key), "%q should be logged", key)
	}
}
