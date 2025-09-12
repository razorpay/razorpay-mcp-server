package razorpay

import (
	"testing"

	rzpsdk "github.com/razorpay/razorpay-go"
)

func TestNewToolSets(t *testing.T) {
	// Create test observability
	obs := CreateTestObservability()

	// Create a test client
	client := &rzpsdk.Client{}

	// Test with empty enabled toolsets
	toolsetGroup, err := NewToolSets(obs, client, []string{}, false)
	if err != nil {
		t.Fatalf("NewToolSets failed: %v", err)
	}

	if toolsetGroup == nil {
		t.Fatal("NewToolSets returned nil toolset group")
	}

	// This test ensures that the FetchSavedPaymentMethods line is executed
	// providing the missing code coverage
}
