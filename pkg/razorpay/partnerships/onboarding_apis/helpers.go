package onboardingapis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/contextkey"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

func getClientFromContextOrDefault(
	ctx context.Context,
	defaultClient *rzpsdk.Client,
) (*rzpsdk.Client, error) {
	if defaultClient != nil {
		return defaultClient, nil
	}

	clientInterface := contextkey.ClientFromContext(ctx)
	if clientInterface == nil {
		return nil, fmt.Errorf("no client found in context")
	}

	client, ok := clientInterface.(*rzpsdk.Client)
	if !ok {
		return nil, fmt.Errorf("invalid client type in context")
	}

	return client, nil
}

func formatErrorMessage(prefix string, err error) string {
	if err == nil {
		return prefix + ": resource does not exist"
	}

	errMsg := err.Error()
	if errMsg == "" {
		return prefix + ": resource does not exist"
	}

	return prefix + ": " + errMsg
}

// validator provides a fluent interface for validating and collecting
// parameters from a tool request.
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
	return mcpgo.NewToolResultError("Validation errors:\n- " + strings.Join(msgs, "\n- ")), nil
}

func extractParam[T any](r *mcpgo.CallToolRequest, name string, required bool) (*T, error) {
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

	var result T
	data, err := json.Marshal(val)
	if err != nil {
		return nil, errors.New("invalid parameter type: " + name)
	}
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, errors.New("invalid parameter type: " + name)
	}
	return &result, nil
}

func (v *validator) requireString(params map[string]interface{}, name string) *validator {
	val, err := extractParam[string](v.request, name, true)
	if err != nil {
		return v.addError(err)
	}
	if val != nil {
		params[name] = *val
	}
	return v
}

func (v *validator) optionalString(params map[string]interface{}, name string) *validator {
	val, err := extractParam[string](v.request, name, false)
	if err != nil {
		return v.addError(err)
	}
	if val != nil {
		params[name] = *val
	}
	return v
}

func (v *validator) optionalMap(params map[string]interface{}, name string) *validator {
	val, err := extractParam[map[string]interface{}](v.request, name, false)
	if err != nil {
		return v.addError(err)
	}
	if val != nil {
		params[name] = *val
	}
	return v
}
