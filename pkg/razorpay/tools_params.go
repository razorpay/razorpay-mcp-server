package razorpay

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

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
	// Type assert Arguments from any to map[string]interface{}
	args, ok := request.Arguments.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid arguments type")
	}

	val, ok := args[name]
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

// validateTokenMaxAmount validates the max_amount field in token.
// max_amount is required and must be a positive number representing
// the maximum amount that can be debited from the customer's account.
func (v *Validator) validateTokenMaxAmount(
	token map[string]interface{}) *Validator {
	if maxAmount, exists := token["max_amount"]; exists {
		switch amt := maxAmount.(type) {
		case float64:
			if amt <= 0 {
				return v.addError(errors.New("token.max_amount must be greater than 0"))
			}
		case int:
			if amt <= 0 {
				return v.addError(errors.New("token.max_amount must be greater than 0"))
			}
			token["max_amount"] = float64(amt) // Convert int to float64
		default:
			return v.addError(errors.New("token.max_amount must be a number"))
		}
	} else {
		return v.addError(errors.New("token.max_amount is required"))
	}
	return v
}

// validateTokenExpireAt validates the expire_at field in token.
// expire_at is optional and defaults to today + 60 days if not provided.
// If provided, it must be a positive Unix timestamp indicating when the
// mandate/token should expire.
func (v *Validator) validateTokenExpireAt(
	token map[string]interface{}) *Validator {
	if expireAt, exists := token["expire_at"]; exists {
		switch exp := expireAt.(type) {
		case float64:
			if exp <= 0 {
				return v.addError(errors.New("token.expire_at must be greater than 0"))
			}
		case int:
			if exp <= 0 {
				return v.addError(errors.New("token.expire_at must be greater than 0"))
			}
			token["expire_at"] = float64(exp) // Convert int to float64
		default:
			return v.addError(errors.New("token.expire_at must be a number"))
		}
	} else {
		// Set default value to today + 60 days
		defaultExpireAt := time.Now().AddDate(0, 0, 60).Unix()
		token["expire_at"] = float64(defaultExpireAt)
	}
	return v
}

// validateTokenFrequency validates the frequency field in token.
// frequency is required and must be one of the allowed values:
// "as_presented", "monthly", "one_time", "yearly", "weekly", "daily".
func (v *Validator) validateTokenFrequency(
	token map[string]interface{}) *Validator {
	if frequency, exists := token["frequency"]; exists {
		if freqStr, ok := frequency.(string); ok {
			validFrequencies := []string{
				"as_presented", "monthly", "one_time", "yearly", "weekly", "daily"}
			for _, validFreq := range validFrequencies {
				if freqStr == validFreq {
					return v
				}
			}
			return v.addError(errors.New(
				"token.frequency must be one of: as_presented, " +
					"monthly, one_time, yearly, weekly, daily"))
		}
		return v.addError(errors.New("token.frequency must be a string"))
	}
	return v.addError(errors.New("token.frequency is required"))
}

// validateTokenType validates the type field in token.
// type is required and must be "single_block_multiple_debit" for SBMD mandates.
func (v *Validator) validateTokenType(token map[string]interface{}) *Validator {
	if tokenType, exists := token["type"]; exists {
		if typeStr, ok := tokenType.(string); ok {
			validTypes := []string{"single_block_multiple_debit"}
			for _, validType := range validTypes {
				if typeStr == validType {
					return v
				}
			}
			return v.addError(errors.New(
				"token.type must be one of: single_block_multiple_debit"))
		}
		return v.addError(errors.New("token.type must be a string"))
	}
	return v.addError(errors.New("token.type is required"))
}

// ValidateAndAddToken validates and adds a token object with proper structure.
// The token object is used for mandate orders and must contain:
//   - max_amount: positive number (maximum debit amount)
//   - expire_at: optional Unix timestamp (mandate expiry,
//     defaults to today + 60 days)
//   - frequency: string (debit frequency: as_presented, monthly, one_time,
//     yearly, weekly, daily)
//   - type: string (mandate type: single_block_multiple_debit)
func (v *Validator) ValidateAndAddToken(
	params map[string]interface{}, name string) *Validator {
	value, err := extractValueGeneric[map[string]interface{}](
		v.request, name, false)
	if err != nil {
		return v.addError(err)
	}

	if value == nil {
		return v
	}

	token := *value

	// Validate all token fields
	v.validateTokenMaxAmount(token).
		validateTokenExpireAt(token).
		validateTokenFrequency(token).
		validateTokenType(token)

	if v.HasErrors() {
		return v
	}

	params[name] = token
	return v
}
