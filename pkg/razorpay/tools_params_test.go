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
