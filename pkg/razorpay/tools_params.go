package razorpay

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// ValidationErrors collects multiple validation errors
type ValidationErrors struct {
	errors []error
}

// NewValidationErrors creates a new ValidationErrors
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		errors: []error{},
	}
}

// AddErrors adds multiple errors to the collection, ignoring any nil errors
func (ve *ValidationErrors) AddErrors(errs ...error) {
	for _, err := range errs {
		if err != nil {
			ve.errors = append(ve.errors, err)
		}
	}
}

// HasErrors returns true if there are any errors
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.errors) > 0
}

// FormatErrors formats all errors as a single string
func (ve *ValidationErrors) FormatErrors() string {
	if !ve.HasErrors() {
		return ""
	}

	messages := make([]string, 0, len(ve.errors))
	for _, err := range ve.errors {
		messages = append(messages, err.Error())
	}

	return "Validation errors:\n- " + strings.Join(messages, "\n- ")
}

// HandleValidationErrors processes validation errors and returns appropriate
// result
func HandleValidationErrors(ve *ValidationErrors) (*mcpgo.ToolResult, error) {
	if ve.HasErrors() {
		return mcpgo.NewToolResultError(ve.FormatErrors()), nil
	}
	return nil, nil
}

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
func RequiredInt(r mcpgo.CallToolRequest, name string) (int64, error) {
	v, err := RequiredParam[float64](r, name)
	if err != nil {
		return 0, err
	}

	return int64(v), nil
}

// OptionalInt extracts an optional integer parameter from the request
func OptionalInt(r mcpgo.CallToolRequest, name string) (int64, error) {
	v, err := OptionalParam[float64](r, name)
	if err != nil {
		return 0, err
	}

	return int64(v), nil
}

// AddPaginationToQueryParams processes and adds pagination parameters
// (count and skip) to the options map and returns any validation errors
func AddPaginationToQueryParams(
	r mcpgo.CallToolRequest,
	options map[string]interface{},
) []error {
	var errors []error

	// Process optional count parameter
	count, err := OptionalInt(r, "count")
	if err != nil {
		errors = append(errors, err)
	} else if count > 0 {
		options["count"] = count
	}

	// Process optional skip parameter
	skip, err := OptionalInt(r, "skip")
	if err != nil {
		errors = append(errors, err)
	} else if skip > 0 {
		options["skip"] = skip
	}

	return errors
}

// AddExpandToQueryParams handles and adds expand parameters to the options map
// and returns any validation errors
func AddExpandToQueryParams(
	r mcpgo.CallToolRequest,
	options map[string]interface{},
) []error {
	var errors []error

	// Process optional expand parameter as an array
	expand, err := OptionalParam[[]interface{}](r, "expand")
	if err != nil {
		errors = append(errors, err)
	} else if len(expand) > 0 {
		// Add each expand value with the expand[] key
		for _, val := range expand {
			if strVal, ok := val.(string); ok && strVal != "" {
				options["expand[]"] = strVal
			}
		}
	}

	return errors
}
