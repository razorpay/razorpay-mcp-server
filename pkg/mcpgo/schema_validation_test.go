package mcpgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateArguments(t *testing.T) {
	parameters := []ToolParameter{
		WithInteger("amount", Required(), Min(100), Max(1000000)),
		WithString("currency", Pattern("^[A-Z]{3}$")),
		WithString("speed", Enum("normal", "optimum")),
		WithString("customer_id", Pattern("^cust_[A-Za-z0-9]+$")),
		WithNumber("rate"),
		WithObject("notes", MaxProperties(2)),
		WithArray("expand", Max(2)),
	}

	tests := []struct {
		name      string
		arguments map[string]interface{}
		expected  []string
	}{
		{
			name:      "valid arguments produce no problem",
			arguments: map[string]interface{}{"amount": 500, "currency": "INR"},
		},
		{
			name:      "a fractional amount is rejected, not truncated",
			arguments: map[string]interface{}{"amount": 100.75},
			expected: []string{
				"amount must be a whole number, got 100.75. Amounts are in " +
					"the smallest currency unit, so they cannot have a fraction",
			},
		},
		{
			name:      "an amount below the declared minimum is rejected",
			arguments: map[string]interface{}{"amount": 50},
			expected:  []string{"amount must be at least 100, got 50"},
		},
		{
			name:      "a negative amount is rejected",
			arguments: map[string]interface{}{"amount": -500},
			expected:  []string{"amount must be at least 100, got -500"},
		},
		{
			name:      "an amount above the declared maximum is rejected",
			arguments: map[string]interface{}{"amount": 2000000},
			expected:  []string{"amount must be at most 1e+06, got 2e+06"},
		},
		{
			name:      "a missing required parameter is reported",
			arguments: map[string]interface{}{"currency": "INR"},
			expected:  []string{"missing required parameter: amount"},
		},
		{
			name: "a value failing the declared pattern is rejected",
			arguments: map[string]interface{}{
				"amount": 500, "currency": "inr",
			},
			expected: []string{
				"currency has an invalid format, expected ^[A-Z]{3}$",
			},
		},
		{
			name: "a path traversal attempt in an id is rejected",
			arguments: map[string]interface{}{
				"amount": 500, "customer_id": "cust_1/../../payments",
			},
			expected: []string{
				"customer_id has an invalid format, " +
					"expected ^cust_[A-Za-z0-9]+$",
			},
		},
		{
			name: "an empty string defers to the required check",
			arguments: map[string]interface{}{
				"amount": 500, "customer_id": "",
			},
		},
		{
			name: "a value outside the declared enum is rejected",
			arguments: map[string]interface{}{
				"amount": 500, "speed": "instant",
			},
			expected: []string{"speed must be one of: normal, optimum"},
		},
		{
			name: "a wrong type reuses the existing wording",
			arguments: map[string]interface{}{
				"amount": 500, "currency": 123,
			},
			expected: []string{"invalid parameter type: currency"},
		},
		{
			name: "problems are reported in declaration order",
			arguments: map[string]interface{}{
				"currency": "inr", "speed": "instant",
			},
			expected: []string{
				"missing required parameter: amount",
				"currency has an invalid format, expected ^[A-Z]{3}$",
				"speed must be one of: normal, optimum",
			},
		},
		{
			name: "a float is still accepted where a number is declared",
			arguments: map[string]interface{}{
				"amount": 500, "rate": 1.5,
			},
		},
		{
			name: "too many object entries are rejected",
			arguments: map[string]interface{}{
				"amount": 500,
				"notes": map[string]interface{}{
					"a": 1, "b": 2, "c": 3,
				},
			},
			expected: []string{"notes must have at most 2 entries"},
		},
		{
			name: "too many array items are rejected",
			arguments: map[string]interface{}{
				"amount": 500,
				"expand": []interface{}{"a", "b", "c"},
			},
			expected: []string{"expand must have at most 2 items"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			problems := ValidateArguments(parameters, tc.arguments)
			assert.Equal(t, tc.expected, problems)
		})
	}
}

func TestValidateArgumentsAcceptsNativeAndJSONNumbers(t *testing.T) {
	parameters := []ToolParameter{WithInteger("amount", Min(1))}

	// A JSON transport decodes into float64, while an in-process caller
	// passes a Go int. Both must be accepted.
	for _, value := range []interface{}{float64(500), int(500), int64(500)} {
		problems := ValidateArguments(
			parameters,
			map[string]interface{}{"amount": value},
		)
		assert.Empty(t, problems)
	}
}

func TestValidateArgumentsIgnoresUndeclaredAndAbsent(t *testing.T) {
	parameters := []ToolParameter{WithString("note")}

	problems := ValidateArguments(parameters, map[string]interface{}{
		"unknown_parameter": 42,
	})
	assert.Empty(t, problems)

	problems = ValidateArguments(parameters, map[string]interface{}{
		"note": nil,
	})
	assert.Empty(t, problems)

	assert.Nil(t, ValidateArguments(parameters, "not a map"))
}
