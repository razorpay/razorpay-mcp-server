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

// Field names are compared in normalised form: lower-cased, with '_', '-',
// '.' and ' ' removed, so "customer_contact", "customerContact" and
// "Customer-Contact" all compare as "customercontact".

// sensitiveLogKeys are exact normalised field names whose values are never
// logged. They are too short or too generic to match as substrings.
var sensitiveLogKeys = map[string]bool{
	"card":   true,
	"number": true,
	"pin":    true,
	"mpin":   true,
	"pan":    true,
	"vpa":    true,
	"upi":    true,
	"dob":    true,
	"cookie": true,
}

// secretLogKeyParts mark a field as a credential or card secret. They are
// redacted wherever they appear in a name, including names ending in an id
// suffix such as session_id or token_id.
var secretLogKeyParts = []string{
	"secret", "password", "passwd", "passcode", "otp", "cvv", "cvc",
	"token", "authorization", "signature", "apikey", "credential",
	"privatekey", "session", "cookie", "cardnumber", "expiry",
}

// piiLogKeyParts mark a field as personal data. A field whose name ends in
// an id suffix (customer_id, contactId) is an identifier, not the data
// itself, and is logged unless its stem is contact data (email_id, upi_id).
// "upi" is matched exactly (sensitiveLogKeys), not as a part, because it
// occurs inside "group".
var piiLogKeyParts = []string{
	"contact", "email", "phone", "mobile", "vpa", "address",
	"customername", "fullname", "firstname", "lastname", "holdername",
	"billingname", "beneficiaryname", "accountnumber", "aadhaar",
	"birth",
}

// contactIDStems are id-suffixed stems that hold contact data, not a
// reference: "email_id" is an email address and "upi_id" is a VPA.
// "contact_id" is not here: Razorpay uses it for a cont_ reference.
var contactIDStems = map[string]bool{
	"email": true, "mail": true, "upi": true, "vpa": true,
	"phone": true, "mobile": true,
}

// normaliseLogKey lower-cases a field name and removes separators.
func normaliseLogKey(key string) string {
	var b strings.Builder
	b.Grow(len(key))
	for _, r := range strings.ToLower(key) {
		switch r {
		case '_', '-', '.', ' ':
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// idStem returns the normalised name before an id suffix ("customer" for
// "customer_id", "customerId" or "customer-ID") and whether the name has one.
func idStem(key string) (string, bool) {
	lower := strings.ToLower(key)
	for _, suffix := range []string{"_id", "-id", ".id"} {
		if strings.HasSuffix(lower, suffix) && len(lower) > len(suffix) {
			return normaliseLogKey(key[:len(key)-len(suffix)]), true
		}
	}
	// camelCase: "customerId" or "customerID".
	if len(key) > 2 && (strings.HasSuffix(key, "Id") ||
		strings.HasSuffix(key, "ID")) {
		prev := key[len(key)-3]
		if prev >= 'a' && prev <= 'z' {
			return normaliseLogKey(key[:len(key)-2]), true
		}
	}
	return "", false
}

func containsAny(s string, parts []string) bool {
	for _, part := range parts {
		if strings.Contains(s, part) {
			return true
		}
	}
	return false
}

// IsSensitiveLogKey reports whether the value of a field with this name must
// be redacted before it is logged.
func IsSensitiveLogKey(key string) bool {
	k := normaliseLogKey(key)
	if sensitiveLogKeys[k] || containsAny(k, secretLogKeyParts) {
		return true
	}
	if stem, isID := idStem(key); isID {
		return contactIDStems[stem]
	}
	return containsAny(k, piiLogKeyParts)
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
			if IsSensitiveLogKey(key) {
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

// ToolRequestForLog summarises a tools/call request: the tool name and its
// arguments with sensitive values redacted.
func ToolRequestForLog(req *mcp.CallToolRequest) map[string]any {
	if req == nil {
		return nil
	}
	return map[string]any{
		"tool":      req.Params.Name,
		"arguments": RedactForLog(req.Params.Arguments),
	}
}

// ToolResultForLog summarises a tool result without its content, which can
// hold card details, tokens and full API responses.
func ToolResultForLog(result *mcp.CallToolResult) map[string]any {
	if result == nil {
		return nil
	}
	return map[string]any{
		"is_error":      result.IsError,
		"content_items": len(result.Content),
	}
}

// MessageForLog returns a log-safe form of an MCP request message.
func MessageForLog(message any) any {
	if req, ok := message.(*mcp.CallToolRequest); ok {
		return ToolRequestForLog(req)
	}
	return RedactForLog(message)
}

// ResultForLog returns a log-safe form of an MCP method result.
func ResultForLog(method mcp.MCPMethod, result any) any {
	switch typed := result.(type) {
	case *mcp.CallToolResult:
		return ToolResultForLog(typed)
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
