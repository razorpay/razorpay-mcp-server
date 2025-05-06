package razorpay

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

func TestValidator(t *testing.T) {
	tests := []struct {
		name           string
		args           map[string]interface{}
		paramName      string
		validationFunc func(*Validator, map[string]interface{}, string) *Validator
		expectError    bool
		expectValue    interface{}
		expectKey      string
	}{
		// String tests
		{
			name:           "required string - valid",
			args:           map[string]interface{}{"test_param": "test_value"},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddRequiredString,
			expectError:    false,
			expectValue:    "test_value",
			expectKey:      "test_param",
		},
		{
			name:           "required string - missing",
			args:           map[string]interface{}{},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddRequiredString,
			expectError:    true,
			expectValue:    nil,
			expectKey:      "test_param",
		},
		{
			name:           "optional string - valid",
			args:           map[string]interface{}{"test_param": "test_value"},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddOptionalString,
			expectError:    false,
			expectValue:    "test_value",
			expectKey:      "test_param",
		},
		{
			name:           "optional string - empty",
			args:           map[string]interface{}{"test_param": ""},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddOptionalString,
			expectError:    false,
			expectValue:    nil,
			expectKey:      "test_param",
		},

		// Int tests
		{
			name:           "required int - valid",
			args:           map[string]interface{}{"test_param": float64(123)},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddRequiredInt,
			expectError:    false,
			expectValue:    int64(123),
			expectKey:      "test_param",
		},
		{
			name:           "optional int - valid",
			args:           map[string]interface{}{"test_param": float64(123)},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddOptionalInt,
			expectError:    false,
			expectValue:    int64(123),
			expectKey:      "test_param",
		},
		{
			name:           "optional int - zero",
			args:           map[string]interface{}{"test_param": float64(0)},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddOptionalInt,
			expectError:    false,
			expectValue:    nil,
			expectKey:      "test_param",
		},

		// Float tests
		{
			name:           "required float - valid",
			args:           map[string]interface{}{"test_param": float64(123.45)},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddRequiredFloat,
			expectError:    false,
			expectValue:    float64(123.45),
			expectKey:      "test_param",
		},
		{
			name:           "optional float - valid",
			args:           map[string]interface{}{"test_param": float64(123.45)},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddOptionalFloat,
			expectError:    false,
			expectValue:    float64(123.45),
			expectKey:      "test_param",
		},
		{
			name:           "optional float - zero",
			args:           map[string]interface{}{"test_param": float64(0)},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddOptionalFloat,
			expectError:    false,
			expectValue:    nil,
			expectKey:      "test_param",
		},

		// Bool tests
		{
			name:           "required bool - true",
			args:           map[string]interface{}{"test_param": true},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddRequiredBool,
			expectError:    false,
			expectValue:    true,
			expectKey:      "test_param",
		},
		{
			name:           "required bool - false",
			args:           map[string]interface{}{"test_param": false},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddRequiredBool,
			expectError:    false,
			expectValue:    false,
			expectKey:      "test_param",
		},
		{
			name:           "optional bool - true",
			args:           map[string]interface{}{"test_param": true},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddOptionalBool,
			expectError:    false,
			expectValue:    true,
			expectKey:      "test_param",
		},
		{
			name:           "optional bool - false",
			args:           map[string]interface{}{"test_param": false},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddOptionalBool,
			expectError:    false,
			expectValue:    false,
			expectKey:      "test_param",
		},

		// Map tests
		{
			name: "required map - valid",
			args: map[string]interface{}{
				"test_param": map[string]interface{}{"key": "value"},
			},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddRequiredMap,
			expectError:    false,
			expectValue:    map[string]interface{}{"key": "value"},
			expectKey:      "test_param",
		},
		{
			name: "optional map - valid",
			args: map[string]interface{}{
				"test_param": map[string]interface{}{"key": "value"},
			},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddOptionalMap,
			expectError:    false,
			expectValue:    map[string]interface{}{"key": "value"},
			expectKey:      "test_param",
		},
		{
			name: "optional map - empty",
			args: map[string]interface{}{
				"test_param": map[string]interface{}{},
			},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddOptionalMap,
			expectError:    false,
			expectValue:    nil,
			expectKey:      "test_param",
		},

		// Array tests
		{
			name: "required array - valid",
			args: map[string]interface{}{
				"test_param": []interface{}{"value1", "value2"},
			},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddRequiredArray,
			expectError:    false,
			expectValue:    []interface{}{"value1", "value2"},
			expectKey:      "test_param",
		},
		{
			name: "optional array - valid",
			args: map[string]interface{}{
				"test_param": []interface{}{"value1", "value2"},
			},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddOptionalArray,
			expectError:    false,
			expectValue:    []interface{}{"value1", "value2"},
			expectKey:      "test_param",
		},
		{
			name:           "optional array - empty",
			args:           map[string]interface{}{"test_param": []interface{}{}},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddOptionalArray,
			expectError:    false,
			expectValue:    nil,
			expectKey:      "test_param",
		},

		// Invalid type tests
		{
			name:           "required string - wrong type",
			args:           map[string]interface{}{"test_param": 123},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddRequiredString,
			expectError:    true,
			expectValue:    nil,
			expectKey:      "test_param",
		},
		{
			name:           "required int - wrong type",
			args:           map[string]interface{}{"test_param": "not a number"},
			paramName:      "test_param",
			validationFunc: (*Validator).ValidateAndAddRequiredInt,
			expectError:    true,
			expectValue:    nil,
			expectKey:      "test_param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make(map[string]interface{})
			request := &mcpgo.CallToolRequest{
				Arguments: tt.args,
			}
			validator := NewValidator(request)

			tt.validationFunc(validator, result, tt.paramName)

			if tt.expectError {
				assert.True(t, validator.HasErrors(), "Expected validation error")
			} else {
				assert.False(t, validator.HasErrors(), "Did not expect validation error")
				if tt.expectValue != nil {
					assert.Equal(t,
						tt.expectValue,
						result[tt.expectKey],
						"Parameter value mismatch",
					)
				} else {
					_, exists := result[tt.expectKey]
					assert.False(t, exists, "Parameter should not be added when empty")
				}
			}
		})
	}
}

func TestValidatorPagination(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		expectCount interface{}
		expectSkip  interface{}
		expectError bool
	}{
		{
			name: "valid pagination params",
			args: map[string]interface{}{
				"count": float64(10),
				"skip":  float64(5),
			},
			expectCount: int64(10),
			expectSkip:  int64(5),
			expectError: false,
		},
		{
			name:        "zero pagination params",
			args:        map[string]interface{}{"count": float64(0), "skip": float64(0)},
			expectCount: nil,
			expectSkip:  nil,
			expectError: false,
		},
		{
			name: "invalid count type",
			args: map[string]interface{}{
				"count": "not a number",
				"skip":  float64(5),
			},
			expectCount: nil,
			expectSkip:  int64(5),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make(map[string]interface{})
			request := &mcpgo.CallToolRequest{
				Arguments: tt.args,
			}
			validator := NewValidator(request)

			validator.ValidateAndAddPagination(result)

			if tt.expectError {
				assert.True(t, validator.HasErrors(), "Expected validation error")
			} else {
				assert.False(t, validator.HasErrors(), "Did not expect validation error")
			}

			if tt.expectCount != nil {
				assert.Equal(t, tt.expectCount, result["count"], "Count mismatch")
			} else {
				_, exists := result["count"]
				assert.False(t, exists, "Count should not be added")
			}

			if tt.expectSkip != nil {
				assert.Equal(t, tt.expectSkip, result["skip"], "Skip mismatch")
			} else {
				_, exists := result["skip"]
				assert.False(t, exists, "Skip should not be added")
			}
		})
	}
}

func TestValidatorExpand(t *testing.T) {
	tests := []struct {
		name         string
		args         map[string]interface{}
		expectExpand string
		expectError  bool
	}{
		{
			name:         "valid expand param",
			args:         map[string]interface{}{"expand": []interface{}{"payments"}},
			expectExpand: "payments",
			expectError:  false,
		},
		{
			name:         "empty expand array",
			args:         map[string]interface{}{"expand": []interface{}{}},
			expectExpand: "",
			expectError:  false,
		},
		{
			name:         "invalid expand type",
			args:         map[string]interface{}{"expand": "not an array"},
			expectExpand: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make(map[string]interface{})
			request := &mcpgo.CallToolRequest{
				Arguments: tt.args,
			}
			validator := NewValidator(request)

			validator.ValidateAndAddExpand(result)

			if tt.expectError {
				assert.True(t, validator.HasErrors(), "Expected validation error")
			} else {
				assert.False(t, validator.HasErrors(), "Did not expect validation error")
				if tt.expectExpand != "" {
					assert.Equal(t,
						tt.expectExpand,
						result["expand[]"],
						"Expand value mismatch",
					)
				} else {
					_, exists := result["expand[]"]
					assert.False(t, exists, "Expand should not be added")
				}
			}
		})
	}
}

// Test validator "To" functions which write to target maps
func TestValidatorToFunctions(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		paramName   string
		targetKey   string
		testFunc    func(*Validator, map[string]interface{}, string, string) *Validator
		expectValue interface{}
		expectError bool
	}{
		// ValidateAndAddOptionalStringTo tests
		{
			name:        "optional string to target - valid",
			args:        map[string]interface{}{"customer_name": "Test User"},
			paramName:   "customer_name",
			targetKey:   "name",
			testFunc:    (*Validator).ValidateAndAddOptionalStringTo,
			expectValue: "Test User",
			expectError: false,
		},
		{
			name:        "optional string to target - empty",
			args:        map[string]interface{}{"customer_name": ""},
			paramName:   "customer_name",
			targetKey:   "name",
			testFunc:    (*Validator).ValidateAndAddOptionalStringTo,
			expectValue: nil,
			expectError: false,
		},
		{
			name:        "optional string to target - missing",
			args:        map[string]interface{}{},
			paramName:   "customer_name",
			targetKey:   "name",
			testFunc:    (*Validator).ValidateAndAddOptionalStringTo,
			expectValue: nil,
			expectError: false,
		},
		{
			name:        "optional string to target - wrong type",
			args:        map[string]interface{}{"customer_name": 123},
			paramName:   "customer_name",
			targetKey:   "name",
			testFunc:    (*Validator).ValidateAndAddOptionalStringTo,
			expectValue: nil,
			expectError: true,
		},

		// ValidateAndAddOptionalBoolTo tests
		{
			name:        "optional bool to target - true",
			args:        map[string]interface{}{"notify_sms": true},
			paramName:   "notify_sms",
			targetKey:   "sms",
			testFunc:    (*Validator).ValidateAndAddOptionalBoolTo,
			expectValue: true,
			expectError: false,
		},
		{
			name:        "optional bool to target - false",
			args:        map[string]interface{}{"notify_sms": false},
			paramName:   "notify_sms",
			targetKey:   "sms",
			testFunc:    (*Validator).ValidateAndAddOptionalBoolTo,
			expectValue: false,
			expectError: false,
		},
		{
			name:        "optional bool to target - wrong type",
			args:        map[string]interface{}{"notify_sms": "not a bool"},
			paramName:   "notify_sms",
			targetKey:   "sms",
			testFunc:    (*Validator).ValidateAndAddOptionalBoolTo,
			expectValue: nil,
			expectError: true,
		},

		// ValidateAndAddOptionalIntTo tests
		{
			name:        "optional int to target - valid",
			args:        map[string]interface{}{"age": float64(25)},
			paramName:   "age",
			targetKey:   "customer_age",
			testFunc:    (*Validator).ValidateAndAddOptionalIntTo,
			expectValue: int64(25),
			expectError: false,
		},
		{
			name:        "optional int to target - zero",
			args:        map[string]interface{}{"age": float64(0)},
			paramName:   "age",
			targetKey:   "customer_age",
			testFunc:    (*Validator).ValidateAndAddOptionalIntTo,
			expectValue: nil,
			expectError: false,
		},
		{
			name:        "optional int to target - missing",
			args:        map[string]interface{}{},
			paramName:   "age",
			targetKey:   "customer_age",
			testFunc:    (*Validator).ValidateAndAddOptionalIntTo,
			expectValue: nil,
			expectError: false,
		},
		{
			name:        "optional int to target - wrong type",
			args:        map[string]interface{}{"age": "not a number"},
			paramName:   "age",
			targetKey:   "customer_age",
			testFunc:    (*Validator).ValidateAndAddOptionalIntTo,
			expectValue: nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a target map for this specific test
			target := make(map[string]interface{})

			// Create the request and validator
			request := &mcpgo.CallToolRequest{
				Arguments: tt.args,
			}
			validator := NewValidator(request)

			// Call the test function with target and verify its return value
			tt.testFunc(validator, target, tt.paramName, tt.targetKey)

			// Check if we got the expected errors
			if tt.expectError {
				assert.True(t, validator.HasErrors(), "Expected validation error")
			} else {
				assert.False(t, validator.HasErrors(), "Did not expect validation error")

				// For non-error cases, check target map value
				if tt.expectValue != nil {
					// Should have the value with the target key
					assert.Equal(t,
						tt.expectValue,
						target[tt.targetKey],
						"Target map value mismatch")
				} else {
					// Target key should not exist
					_, exists := target[tt.targetKey]
					assert.False(t, exists, "Key should not be in target map when value is empty")
				}
			}
		})
	}
}

// Test for nested validation with multiple fields into target maps
func TestValidatorNestedObjects(t *testing.T) {
	t.Run("customer object validation", func(t *testing.T) {
		// Create request with customer details
		args := map[string]interface{}{
			"customer_name":    "John Doe",
			"customer_email":   "john@example.com",
			"customer_contact": "+1234567890",
		}
		request := &mcpgo.CallToolRequest{
			Arguments: args,
		}

		// Customer target map
		customer := make(map[string]interface{})

		// Create validator and validate customer fields
		validator := NewValidator(request).
			ValidateAndAddOptionalStringTo(customer, "customer_name", "name").
			ValidateAndAddOptionalStringTo(customer, "customer_email", "email").
			ValidateAndAddOptionalStringTo(customer, "customer_contact", "contact")

		// Should not have errors
		assert.False(t, validator.HasErrors())

		// Customer map should have all three fields
		assert.Equal(t, "John Doe", customer["name"])
		assert.Equal(t, "john@example.com", customer["email"])
		assert.Equal(t, "+1234567890", customer["contact"])
	})

	t.Run("notification object validation", func(t *testing.T) {
		// Create request with notification settings
		args := map[string]interface{}{
			"notify_sms":   true,
			"notify_email": false,
		}
		request := &mcpgo.CallToolRequest{
			Arguments: args,
		}

		// Notify target map
		notify := make(map[string]interface{})

		// Create validator and validate notification fields
		validator := NewValidator(request).
			ValidateAndAddOptionalBoolTo(notify, "notify_sms", "sms").
			ValidateAndAddOptionalBoolTo(notify, "notify_email", "email")

		// Should not have errors
		assert.False(t, validator.HasErrors())

		// Notify map should have both fields
		assert.Equal(t, true, notify["sms"])
		assert.Equal(t, false, notify["email"])
	})

	t.Run("mixed object with error", func(t *testing.T) {
		// Create request with mixed valid and invalid data
		args := map[string]interface{}{
			"customer_name":  "Jane Doe",
			"customer_email": 12345, // Wrong type
		}
		request := &mcpgo.CallToolRequest{
			Arguments: args,
		}

		// Target map
		customer := make(map[string]interface{})

		// Create validator and validate fields
		validator := NewValidator(request).
			ValidateAndAddOptionalStringTo(customer, "customer_name", "name").
			ValidateAndAddOptionalStringTo(customer, "customer_email", "email")

		// Should have errors
		assert.True(t, validator.HasErrors())

		// Customer map should have only the valid field
		assert.Equal(t, "Jane Doe", customer["name"])
		_, hasEmail := customer["email"]
		assert.False(t, hasEmail, "Invalid field should not be added to target map")
	})
}
