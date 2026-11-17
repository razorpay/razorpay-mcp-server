package mcpgo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// recordingLogger keeps every formatted log line for inspection.
type recordingLogger struct {
	mu    sync.Mutex
	lines []string
}

func (l *recordingLogger) record(msg string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, fmt.Sprintf("%s %+v", msg, args))
}

func (l *recordingLogger) Infof(_ context.Context, m string, a ...interface{}) {
	l.record(m, a...)
}

func (l *recordingLogger) Errorf(_ context.Context, m string, a ...interface{}) {
	l.record(m, a...)
}

func (l *recordingLogger) Fatalf(_ context.Context, m string, a ...interface{}) {
	l.record(m, a...)
}

func (l *recordingLogger) Debugf(_ context.Context, m string, a ...interface{}) {
	l.record(m, a...)
}

func (l *recordingLogger) Warningf(
	_ context.Context, m string, a ...interface{}) {
	l.record(m, a...)
}

func (l *recordingLogger) Close() error { return nil }

func (l *recordingLogger) all() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return strings.Join(l.lines, "\n")
}

// sensitiveToolCall returns a tools/call request and result whose sensitive
// values must never reach the log.
func sensitiveToolCall() (*mcp.CallToolRequest, *mcp.CallToolResult) {
	req := &mcp.CallToolRequest{}
	req.Method = string(mcp.MethodToolsCall)
	req.Header = http.Header{"Authorization": []string{"Basic c2VjcmV0LWhkcg=="}}
	req.Params.Name = "submit_otp"
	req.Params.Arguments = map[string]any{
		"payment_id": "pay_LOGSAFE000001",
		"otp_string": "482913",
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
	"c2VjcmV0LWhkcg==", `cvv:123`,
}

func TestSetupHooks_DoNotLogSensitiveValues(t *testing.T) {
	logger := &recordingLogger{}
	hooks := SetupHooks(&observability.Observability{Logger: logger})
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

	logged := logger.all()
	assert.NotEmpty(t, logged)
	for _, secret := range sensitiveValues {
		assert.NotContains(t, logged, secret)
	}
	// Logs stay useful: tool name, safe argument values and outcome remain.
	assert.Contains(t, logged, "submit_otp")
	assert.Contains(t, logged, "pay_LOGSAFE000001")
	assert.Contains(t, logged, "otp_string")
	assert.Contains(t, logged, "boom")
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
