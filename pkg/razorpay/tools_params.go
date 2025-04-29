package razorpay

import (
	"encoding/json"
	"errors"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// RequiredParam extracts a required parameter of type T from the request
func RequiredParam[T any](r mcpgo.CallToolRequest, name string) (T, error) {
	var zero T
	v, ok := r.Arguments[name]
	if !ok || v == nil {
		return zero, errors.New("missing required parameter: " + name)
	}

	var result T
	data, err := json.Marshal(v)
	if err != nil {
		return zero, errors.New("invalid parameter type: " + name)
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return zero, errors.New("invalid parameter type: " + name)
	}

	return result, nil
}

// OptionalParam extracts an optional parameter of type T from the request
func OptionalParam[T any](r mcpgo.CallToolRequest, name string) (T, error) {
	var zero T
	v, ok := r.Arguments[name]
	if !ok || v == nil {
		return zero, nil
	}

	var result T
	data, err := json.Marshal(v)
	if err != nil {
		return zero, errors.New("invalid parameter type: " + name)
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return zero, errors.New("invalid parameter type: " + name)
	}

	return result, nil
}

// RequiredInt extracts a required integer parameter from the request
func RequiredInt(r mcpgo.CallToolRequest, name string) (int, error) {
	v, err := RequiredParam[float64](r, name)
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

// OptionalInt extracts an optional integer parameter from the request
func OptionalInt(r mcpgo.CallToolRequest, name string) (int, error) {
	v, err := OptionalParam[float64](r, name)
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

// HandleValidationError handles parameter validation errors
func HandleValidationError(err error) (*mcpgo.ToolResult, error) {
	if err != nil {
		return mcpgo.NewToolResultError(err.Error()), nil
	}
	return nil, nil
}

// AddPaginationToOptions processes and adds pagination parameters
// (count and skip) to the options map
func AddPaginationToOptions(
	r mcpgo.CallToolRequest,
	options map[string]interface{},
) *mcpgo.ToolResult {
	// Process optional count parameter
	count, err := OptionalInt(r, "count")
	if result, _ := HandleValidationError(err); result != nil {
		return result
	}
	if count > 0 {
		options["count"] = count
	}

	// Process optional skip parameter
	skip, err := OptionalInt(r, "skip")
	if result, _ := HandleValidationError(err); result != nil {
		return result
	}
	if skip > 0 {
		options["skip"] = skip
	}

	return nil
}

// AddExpandToOptions handles and adds expand parameters to the options map
func AddExpandToOptions(
	r mcpgo.CallToolRequest,
	options map[string]interface{},
) *mcpgo.ToolResult {
	// Process optional expand parameter as an array
	expand, err := OptionalParam[[]interface{}](r, "expand")
	if result, _ := HandleValidationError(err); result != nil {
		return result
	}
	if len(expand) > 0 {
		// Add each expand value with the expand[] key
		for _, val := range expand {
			if strVal, ok := val.(string); ok && strVal != "" {
				options["expand[]"] = strVal
			}
		}
	}

	return nil
}
