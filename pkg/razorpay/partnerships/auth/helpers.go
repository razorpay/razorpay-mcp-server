//nolint:lll // long lines in function signatures are acceptable here
package auth

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

func formatErrorMessage(prefix string, err error) string {
	if err == nil {
		return prefix + ": unknown error"
	}
	if msg := err.Error(); msg != "" {
		return prefix + ": " + msg
	}
	return prefix + ": unknown error"
}

type validator struct {
	request *mcpgo.CallToolRequest
	errors  []error
}

func newValidator(r *mcpgo.CallToolRequest) *validator {
	return &validator{request: r}
}

func (v *validator) addError(err error) *validator {
	if err != nil {
		v.errors = append(v.errors, err)
	}
	return v
}

func (v *validator) handleErrorsIfAny() (*mcpgo.ToolResult, error) {
	if len(v.errors) == 0 {
		return nil, nil
	}
	msgs := make([]string, 0, len(v.errors))
	for _, e := range v.errors {
		msgs = append(msgs, e.Error())
	}
	return mcpgo.NewToolResultError(
		"Validation errors:\n- " + strings.Join(msgs, "\n- "),
	), nil
}

func extractString(
	r *mcpgo.CallToolRequest, name string, required bool,
) (*string, error) {
	args, ok := r.Arguments.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid arguments type")
	}

	val, ok := args[name]
	if !ok || val == nil {
		if required {
			return nil, errors.New("missing required parameter: " + name)
		}
		return nil, nil
	}

	data, err := json.Marshal(val)
	if err != nil {
		return nil, errors.New("invalid parameter type: " + name)
	}

	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return nil, errors.New("invalid parameter type: " + name)
	}
	return &s, nil
}

func (v *validator) requireString(
	params map[string]interface{}, name string,
) *validator {
	val, err := extractString(v.request, name, true)
	if err != nil {
		return v.addError(err)
	}
	if val != nil {
		params[name] = *val
	}
	return v
}

func (v *validator) optionalString(
	params map[string]interface{}, name string,
) *validator {
	val, err := extractString(v.request, name, false)
	if err != nil {
		return v.addError(err)
	}
	if val != nil {
		params[name] = *val
	}
	return v
}
