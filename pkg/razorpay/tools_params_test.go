package razorpay

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

func TestValidationErrors(t *testing.T) {
	tests := []struct {
		name            string
		setupErrors     []error
		expectHasError  bool
		expectedText    string
		notExpectedText string
	}{
		{
			name:           "empty validation errors",
			setupErrors:    []error{},
			expectHasError: false,
			expectedText:   "",
		},
		{
			name:           "single error",
			setupErrors:    []error{errors.New("test error")},
			expectHasError: true,
			expectedText:   "test error",
		},
		{
			name:           "multiple errors",
			setupErrors:    []error{errors.New("error 1"), errors.New("error 2")},
			expectHasError: true,
			expectedText:   "error 1",
		},
		{
			name:           "ignore nil errors",
			setupErrors:    []error{nil},
			expectHasError: false,
			expectedText:   "",
		},
		{
			name:            "mix of nil and valid errors",
			setupErrors:     []error{nil, errors.New("valid error"), nil},
			expectHasError:  true,
			expectedText:    "valid error",
			notExpectedText: "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := NewValidationErrors()
			ve.AddErrors(tt.setupErrors...)

			// Check if errors are properly tracked
			assert.Equal(t, tt.expectHasError, ve.HasErrors(),
				"HasErrors() should return %v", tt.expectHasError)

			// Check formatted output
			formatted := ve.FormatErrors()

			if tt.expectedText != "" {
				if tt.expectHasError {
					assert.Contains(t, formatted, tt.expectedText,
						"Formatted errors should contain expected text")
				} else {
					assert.Equal(t, tt.expectedText, formatted,
						"Formatted empty errors should be empty string")
				}
			}

			if tt.notExpectedText != "" {
				assert.NotContains(t, formatted, tt.notExpectedText,
					"Formatted errors should not contain unexpected text")
			}
		})
	}
}

func TestHandleValidationErrors(t *testing.T) {
	tests := []struct {
		name            string
		setupErrors     []error
		expectNilResult bool
		expectErrorText string
	}{
		{
			name:            "no errors",
			setupErrors:     []error{},
			expectNilResult: true,
		},
		{
			name:            "with errors",
			setupErrors:     []error{errors.New("test error")},
			expectNilResult: false,
			expectErrorText: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := NewValidationErrors()
			ve.AddErrors(tt.setupErrors...)

			result, err := HandleValidationErrors(ve)

			// Always expect nil error since errors are returned via result
			assert.Nil(t, err, "Error should always be nil")

			if tt.expectNilResult {
				assert.Nil(t,
					result, "Result should be nil when there are no errors")
			} else {
				assert.NotNil(t,
					result, "Result should not be nil when there are errors")
				assert.True(t, result.IsError, "Result should be an error")
				assert.Contains(t, result.Text, tt.expectErrorText,
					"Result should contain the error message")
			}
		})
	}
}

func TestAddExpandToQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		request  map[string]interface{}
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name: "with single expand value",
			request: map[string]interface{}{
				"expand": []interface{}{"payments"},
			},
			expected: map[string]interface{}{
				"expand[]": "payments",
			},
			wantErr: false,
		},
		{
			name: "with empty expand array",
			request: map[string]interface{}{
				"expand": []interface{}{},
			},
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name: "with invalid type",
			request: map[string]interface{}{
				"expand": "payments",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "with no parameters",
			request:  map[string]interface{}{},
			expected: map[string]interface{}{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a tool request with the test parameters
			request := mcpgo.CallToolRequest{
				Arguments: tt.request,
			}

			// Initialize options map
			options := make(map[string]interface{})

			// Call the function being tested
			errors := AddExpandToQueryParams(request, options)

			// Check if an error was expected
			if tt.wantErr {
				assert.NotEmpty(t, errors, "Expected errors but got none")
			} else {
				assert.Empty(t, errors, "Expected no errors but got some")
				assert.Equal(t, tt.expected, options, "Options do not match expected")
			}
		})
	}
}

func TestAddPaginationToQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		request  map[string]interface{}
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name: "with valid count and skip",
			request: map[string]interface{}{
				"count": float64(10),
				"skip":  float64(5),
			},
			expected: map[string]interface{}{
				"count": 10,
				"skip":  5,
			},
			wantErr: false,
		},
		{
			name: "with zero count and skip",
			request: map[string]interface{}{
				"count": float64(0),
				"skip":  float64(0),
			},
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name: "with invalid count type",
			request: map[string]interface{}{
				"count": "not-a-number",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "with no parameters",
			request:  map[string]interface{}{},
			expected: map[string]interface{}{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a tool request with the test parameters
			request := mcpgo.CallToolRequest{
				Arguments: tt.request,
			}

			// Initialize options map
			options := make(map[string]interface{})

			// Call the function being tested
			errors := AddPaginationToQueryParams(request, options)

			// Check if an error was expected
			if tt.wantErr {
				assert.NotEmpty(t, errors, "Expected errors but got none")
			} else {
				assert.Empty(t, errors, "Expected no errors but got some")
				assert.Equal(t, tt.expected, options, "Options do not match expected")
			}
		})
	}
}

func Test_RequiredParam(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		paramName   string
		expected    string
		expectError bool
	}{
		{
			name:        "valid string parameter",
			params:      map[string]interface{}{"name": "test-value"},
			paramName:   "name",
			expected:    "test-value",
			expectError: false,
		},
		{
			name:        "missing parameter",
			params:      map[string]interface{}{},
			paramName:   "name",
			expected:    "",
			expectError: true,
		},
		{
			name:        "wrong type parameter",
			params:      map[string]interface{}{"name": 123},
			paramName:   "name",
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := createMCPRequest(tc.params)
			result, err := RequiredParam[string](request, tc.paramName)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func Test_OptionalParam(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		paramName   string
		expected    string
		expectError bool
	}{
		{
			name:        "valid string parameter",
			params:      map[string]interface{}{"name": "test-value"},
			paramName:   "name",
			expected:    "test-value",
			expectError: false,
		},
		{
			name:        "missing parameter",
			params:      map[string]interface{}{},
			paramName:   "name",
			expected:    "",
			expectError: false,
		},
		{
			name:        "empty string parameter",
			params:      map[string]interface{}{"name": ""},
			paramName:   "name",
			expected:    "",
			expectError: false,
		},
		{
			name:        "wrong type parameter",
			params:      map[string]interface{}{"name": 123},
			paramName:   "name",
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := createMCPRequest(tc.params)
			result, err := OptionalParam[string](request, tc.paramName)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func Test_RequiredInt(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		paramName   string
		expected    int
		expectError bool
	}{
		{
			name:        "valid number parameter",
			params:      map[string]interface{}{"count": float64(42)},
			paramName:   "count",
			expected:    42,
			expectError: false,
		},
		{
			name:        "missing parameter",
			params:      map[string]interface{}{},
			paramName:   "count",
			expected:    0,
			expectError: true,
		},
		{
			name:        "wrong type parameter",
			params:      map[string]interface{}{"count": "not-a-number"},
			paramName:   "count",
			expected:    0,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := createMCPRequest(tc.params)
			result, err := RequiredInt(request, tc.paramName)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func Test_OptionalInt(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		paramName   string
		expected    int
		expectError bool
	}{
		{
			name:        "valid number parameter",
			params:      map[string]interface{}{"count": float64(42)},
			paramName:   "count",
			expected:    42,
			expectError: false,
		},
		{
			name:        "missing parameter",
			params:      map[string]interface{}{},
			paramName:   "count",
			expected:    0,
			expectError: false,
		},
		{
			name:        "zero value",
			params:      map[string]interface{}{"count": float64(0)},
			paramName:   "count",
			expected:    0,
			expectError: false,
		},
		{
			name:        "wrong type parameter",
			params:      map[string]interface{}{"count": "not-a-number"},
			paramName:   "count",
			expected:    0,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := createMCPRequest(tc.params)
			result, err := OptionalInt(request, tc.paramName)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
