package razorpay

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// Validator provides a fluent interface for validating parameters
// and collecting errors
type Validator struct {
	request *mcpgo.CallToolRequest
	errors  []error
}

// NewValidator creates a new validator for the given request
func NewValidator(r *mcpgo.CallToolRequest) *Validator {
	return &Validator{
		request: r,
		errors:  []error{},
	}
}

// addError adds a non-nil error to the collection
func (v *Validator) addError(err error) *Validator {
	if err != nil {
		v.errors = append(v.errors, err)
	}
	return v
}

// HasErrors returns true if there are any validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// HandleErrorsIfAny formats all errors and returns an appropriate tool result
func (v *Validator) HandleErrorsIfAny() (*mcpgo.ToolResult, error) {
	if v.HasErrors() {
		messages := make([]string, 0, len(v.errors))
		for _, err := range v.errors {
			messages = append(messages, err.Error())
		}
		errorMsg := "Validation errors:\n- " + strings.Join(messages, "\n- ")
		return mcpgo.NewToolResultError(errorMsg), nil
	}
	return nil, nil
}

// extractValueGeneric is a standalone generic function to extract a parameter
// of type T
func extractValueGeneric[T any](
	request *mcpgo.CallToolRequest,
	name string,
	required bool,
) (*T, error) {
	val, ok := request.Arguments[name]
	if !ok || val == nil {
		if required {
			return nil, errors.New("missing required parameter: " + name)
		}
		return nil, nil // Not an error for optional params
	}

	var result T
	data, err := json.Marshal(val)
	if err != nil {
		return nil, errors.New("invalid parameter type: " + name)
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, errors.New("invalid parameter type: " + name)
	}

	return &result, nil
}

// Generic validation functions

// validateAndAddRequired validates and adds a required parameter of any type
func validateAndAddRequired[T any](
	v *Validator,
	params map[string]interface{},
	name string,
) *Validator {
	value, err := extractValueGeneric[T](v.request, name, true)
	if err != nil {
		return v.addError(err)
	}

	if value == nil {
		return v
	}

	params[name] = *value
	return v
}

// validateAndAddOptional validates and adds an optional parameter of any type
// if not empty
func validateAndAddOptional[T any](
	v *Validator,
	params map[string]interface{},
	name string,
) *Validator {
	value, err := extractValueGeneric[T](v.request, name, false)
	if err != nil {
		return v.addError(err)
	}

	if value == nil {
		return v
	}

	params[name] = *value

	return v
}

// validateAndAddToPath is a generic helper to extract a value and write it into
// `target[targetKey]` if non-empty
func validateAndAddToPath[T any](
	v *Validator,
	target map[string]interface{},
	paramName string,
	targetKey string,
) *Validator {
	value, err := extractValueGeneric[T](v.request, paramName, false)
	if err != nil {
		return v.addError(err)
	}

	if value == nil {
		return v
	}

	target[targetKey] = *value

	return v
}

// ValidateAndAddOptionalStringToPath validates an optional string
// and writes it into target[targetKey]
func (v *Validator) ValidateAndAddOptionalStringToPath(
	target map[string]interface{},
	paramName, targetKey string,
) *Validator {
	return validateAndAddToPath[string](v, target, paramName, targetKey) // nolint:lll
}

// ValidateAndAddOptionalBoolToPath validates an optional bool
// and writes it into target[targetKey]
// only if it was explicitly provided in the request
func (v *Validator) ValidateAndAddOptionalBoolToPath(
	target map[string]interface{},
	paramName, targetKey string,
) *Validator {
	// Now validate and add the parameter
	value, err := extractValueGeneric[bool](v.request, paramName, false)
	if err != nil {
		return v.addError(err)
	}

	if value == nil {
		return v
	}

	target[targetKey] = *value
	return v
}

// ValidateAndAddOptionalIntToPath validates an optional integer
// and writes it into target[targetKey]
func (v *Validator) ValidateAndAddOptionalIntToPath(
	target map[string]interface{},
	paramName, targetKey string,
) *Validator {
	return validateAndAddToPath[int64](v, target, paramName, targetKey)
}

// Type-specific validator methods

// ValidateAndAddRequiredString validates and adds a required string parameter
func (v *Validator) ValidateAndAddRequiredString(
	params map[string]interface{},
	name string,
) *Validator {
	return validateAndAddRequired[string](v, params, name)
}

// ValidateAndAddOptionalString validates and adds an optional string parameter
func (v *Validator) ValidateAndAddOptionalString(
	params map[string]interface{},
	name string,
) *Validator {
	return validateAndAddOptional[string](v, params, name)
}

// ValidateAndAddRequiredMap validates and adds a required map parameter
func (v *Validator) ValidateAndAddRequiredMap(
	params map[string]interface{},
	name string,
) *Validator {
	return validateAndAddRequired[map[string]interface{}](v, params, name)
}

// ValidateAndAddOptionalMap validates and adds an optional map parameter
func (v *Validator) ValidateAndAddOptionalMap(
	params map[string]interface{},
	name string,
) *Validator {
	return validateAndAddOptional[map[string]interface{}](v, params, name)
}

// ValidateAndAddRequiredArray validates and adds a required array parameter
func (v *Validator) ValidateAndAddRequiredArray(
	params map[string]interface{},
	name string,
) *Validator {
	return validateAndAddRequired[[]interface{}](v, params, name)
}

// ValidateAndAddOptionalArray validates and adds an optional array parameter
func (v *Validator) ValidateAndAddOptionalArray(
	params map[string]interface{},
	name string,
) *Validator {
	return validateAndAddOptional[[]interface{}](v, params, name)
}

// ValidateAndAddPagination validates and adds pagination parameters
// (count and skip)
func (v *Validator) ValidateAndAddPagination(
	params map[string]interface{},
) *Validator {
	return v.ValidateAndAddOptionalInt(params, "count").
		ValidateAndAddOptionalInt(params, "skip")
}

// ValidateAndAddExpand validates and adds expand parameters
func (v *Validator) ValidateAndAddExpand(
	params map[string]interface{},
) *Validator {
	expand, err := extractValueGeneric[[]string](v.request, "expand", false)
	if err != nil {
		return v.addError(err)
	}

	if expand == nil {
		return v
	}

	if len(*expand) > 0 {
		for _, val := range *expand {
			params["expand[]"] = val
		}
	}
	return v
}

// ValidateAndAddRequiredInt validates and adds a required integer parameter
func (v *Validator) ValidateAndAddRequiredInt(
	params map[string]interface{},
	name string,
) *Validator {
	return validateAndAddRequired[int64](v, params, name)
}

// ValidateAndAddOptionalInt validates and adds an optional integer parameter
func (v *Validator) ValidateAndAddOptionalInt(
	params map[string]interface{},
	name string,
) *Validator {
	return validateAndAddOptional[int64](v, params, name)
}

// ValidateAndAddRequiredFloat validates and adds a required float parameter
func (v *Validator) ValidateAndAddRequiredFloat(
	params map[string]interface{},
	name string,
) *Validator {
	return validateAndAddRequired[float64](v, params, name)
}

// ValidateAndAddOptionalFloat validates and adds an optional float parameter
func (v *Validator) ValidateAndAddOptionalFloat(
	params map[string]interface{},
	name string,
) *Validator {
	return validateAndAddOptional[float64](v, params, name)
}

// ValidateAndAddRequiredBool validates and adds a required boolean parameter
func (v *Validator) ValidateAndAddRequiredBool(
	params map[string]interface{},
	name string,
) *Validator {
	return validateAndAddRequired[bool](v, params, name)
}

// ValidateAndAddOptionalBool validates and adds an optional boolean parameter
// Note: This adds the boolean value only
// if it was explicitly provided in the request
func (v *Validator) ValidateAndAddOptionalBool(
	params map[string]interface{},
	name string,
) *Validator {
	// Now validate and add the parameter
	value, err := extractValueGeneric[bool](v.request, name, false)
	if err != nil {
		return v.addError(err)
	}

	if value == nil {
		return v
	}

	params[name] = *value
	return v
}
