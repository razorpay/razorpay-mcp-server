package mcpgo

import (
	"encoding/json"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// redactedValue replaces the value of a sensitive field in log output.
const redactedValue = "[REDACTED]"

// unloggableValue stands in for a value that cannot be rendered safely.
const unloggableValue = "[UNLOGGABLE]"

// sensitiveLogKeys are field names whose values are never logged. Keys are
// compared lower-cased.
var sensitiveLogKeys = map[string]bool{
	"otp":                true,
	"otp_string":         true,
	"card":               true,
	"card_number":        true,
	"number":             true,
	"cvv":                true,
	"expiry":             true,
	"expiry_month":       true,
	"expiry_year":        true,
	"token":              true,
	"access_token":       true,
	"refresh_token":      true,
	"authorization":      true,
	"cookie":             true,
	"password":           true,
	"secret":             true,
	"key_secret":         true,
	"api_key":            true,
	"signature":          true,
	"razorpay_signature": true,
	"contact":            true,
	"email":              true,
	"vpa":                true,
}

// sensitiveLogKeyParts are substrings that mark a field name as sensitive.
var sensitiveLogKeyParts = []string{
	"secret", "password", "otp", "cvv", "authorization", "signature",
}

func isSensitiveLogKey(key string) bool {
	k := strings.ToLower(key)
	if sensitiveLogKeys[k] {
		return true
	}
	for _, part := range sensitiveLogKeyParts {
		if strings.Contains(k, part) {
			return true
		}
	}
	return false
}

// RedactForLog returns a copy of v that is safe to log: the value is
// normalised through JSON (so fields tagged json:"-", such as transport
// headers, are dropped) and the value of every sensitive field, at any depth,
// is replaced with redactedValue. The input is not modified.
func RedactForLog(v any) any {
	if v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return unloggableValue
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return unloggableValue
	}
	return redactDecoded(decoded)
}

func redactDecoded(v any) any {
	switch typed := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			if isSensitiveLogKey(key) {
				out[key] = redactedValue
				continue
			}
			out[key] = redactDecoded(value)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, value := range typed {
			out[i] = redactDecoded(value)
		}
		return out
	default:
		return typed
	}
}

// toolRequestForLog summarises a tools/call request: the tool name and its
// arguments with sensitive values redacted.
func toolRequestForLog(req *mcp.CallToolRequest) map[string]any {
	if req == nil {
		return nil
	}
	return map[string]any{
		"tool":      req.Params.Name,
		"arguments": RedactForLog(req.Params.Arguments),
	}
}

// toolResultForLog summarises a tool result without its content, which can
// hold card details, tokens and full API responses.
func toolResultForLog(result *mcp.CallToolResult) map[string]any {
	if result == nil {
		return nil
	}
	return map[string]any{
		"is_error":      result.IsError,
		"content_items": len(result.Content),
	}
}

// messageForLog returns a log-safe form of an MCP request message.
func messageForLog(message any) any {
	if req, ok := message.(*mcp.CallToolRequest); ok {
		return toolRequestForLog(req)
	}
	return RedactForLog(message)
}

// resultForLog returns a log-safe form of an MCP method result.
func resultForLog(method mcp.MCPMethod, result any) any {
	switch typed := result.(type) {
	case *mcp.CallToolResult:
		return toolResultForLog(typed)
	case *mcp.ListToolsResult:
		if method == mcp.MethodToolsList {
			names := make([]string, 0, len(typed.Tools))
			for _, tool := range typed.Tools {
				names = append(names, tool.Name)
			}
			return map[string]any{"tools": names}
		}
	}
	return RedactForLog(result)
}
