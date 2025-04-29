package mcpgo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewToolResultJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectText  string
		shouldError bool
	}{
		{
			name:        "Simple string",
			input:       "test string",
			expectText:  `"test string"`,
			shouldError: false,
		},
		{
			name:        "Simple number",
			input:       42,
			expectText:  "42",
			shouldError: false,
		},
		{
			name:        "Simple boolean",
			input:       true,
			expectText:  "true",
			shouldError: false,
		},
		{
			name: "Simple object",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
			expectText:  `{"key1":"value1","key2":42,"key3":true}`,
			shouldError: false,
		},
		{
			name:        "Simple array",
			input:       []interface{}{"one", 2, true},
			expectText:  `["one",2,true]`,
			shouldError: false,
		},
		{
			name: "Nested object",
			input: map[string]interface{}{
				"name": "Test",
				"details": map[string]interface{}{
					"age":    30,
					"active": true,
				},
				"tags": []string{"tag1", "tag2"},
			},
			expectText: `{"details":{"active":true,"age":30},` +
				`"name":"Test","tags":["tag1","tag2"]}`,
			shouldError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := NewToolResultJSON(tc.input)
			if tc.shouldError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Nil(t, result.Content)

			var expected, actual interface{}

			err = json.Unmarshal([]byte(tc.expectText), &expected)
			assert.NoError(t, err, "Failed to parse expected JSON")

			err = json.Unmarshal([]byte(result.Text), &actual)
			assert.NoError(t, err, "Result text is not valid JSON")

			expectedBytes, _ := json.Marshal(expected)
			actualBytes, _ := json.Marshal(actual)

			assert.Equal(t,
				string(expectedBytes),
				string(actualBytes),
				"JSON doesn't match")
		})
	}
}

func TestNewToolResultText(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Empty string",
			input: "",
		},
		{
			name:  "Simple string",
			input: "Hello, world!",
		},
		{
			name:  "String with special characters",
			input: "Line 1\nLine 2\tTabbed\u2022 Unicode",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := NewToolResultText(tc.input)

			assert.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Nil(t, result.Content)
			assert.Equal(t, tc.input, result.Text)
		})
	}
}

func TestNewToolResultError(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Empty error message",
			input: "",
		},
		{
			name:  "Simple error message",
			input: "Something went wrong",
		},
		{
			name:  "Error message with special characters",
			input: "Error: resource not found\nDetails: missing ID",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := NewToolResultError(tc.input)

			assert.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Nil(t, result.Content)
			assert.Equal(t, tc.input, result.Text)
		})
	}
}
