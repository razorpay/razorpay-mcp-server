package razorpay

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

func TestAddExpandToOptions(t *testing.T) {
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
			toolResult := AddExpandToOptions(request, options)

			// Check if an error was expected
			if tt.wantErr {
				assert.NotNil(t, toolResult, "Expected an error but got nil")
			} else {
				assert.Nil(t, toolResult, "Expected no error but got one")
				assert.Equal(t, tt.expected, options, "Options do not match expected")
			}
		})
	}
}

func TestAddPaginationToOptions(t *testing.T) {
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
			toolResult := AddPaginationToOptions(request, options)

			// Check if an error was expected
			if tt.wantErr {
				assert.NotNil(t, toolResult, "Expected an error but got nil")
			} else {
				assert.Nil(t, toolResult, "Expected no error but got one")
				assert.Equal(t, tt.expected, options, "Options do not match expected")
			}
		})
	}
}
