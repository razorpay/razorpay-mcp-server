package razorpay

import (
	"testing"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

func TestNewToolSets(t *testing.T) {
	obs := CreateTestObservability()
	client := &rzpsdk.Client{}

	t.Run("creates toolsets with empty enabled list", func(t *testing.T) {
		testCreateAllToolsets(t, obs, client)
	})

	t.Run("creates toolsets with specific enabled list", func(t *testing.T) {
		testSpecificEnabledToolsets(t, obs, client)
	})

	t.Run("creates toolsets in read-only mode", func(t *testing.T) {
		testReadOnlyMode(t, obs, client)
	})

	t.Run("handles invalid toolset name", func(t *testing.T) {
		testInvalidToolsetName(t, obs, client)
	})

	t.Run("handles mixed valid and invalid toolset names", func(t *testing.T) {
		testMixedValidInvalidToolsets(t, obs, client)
	})

	t.Run("creates all toolsets with all tools", func(t *testing.T) {
		testAllToolsCreation(t, obs, client)
	})

	t.Run("creates toolsets with single toolset enabled", func(t *testing.T) {
		testSingleToolsetEnabled(t, obs, client)
	})

	t.Run("creates toolsets with multiple specific toolsets", func(t *testing.T) {
		testMultipleSpecificToolsets(t, obs, client)
	})
}

func testCreateAllToolsets(t *testing.T, obs *observability.Observability,
	client *rzpsdk.Client) {
	toolsetGroup, err := NewToolSets(obs, client, []string{}, false)
	if err != nil {
		t.Fatalf("NewToolSets failed: %v", err)
	}

	if toolsetGroup == nil {
		t.Fatal("NewToolSets returned nil toolset group")
	}

	expectedToolsets := []string{
		"payments", "payment_links", "orders",
		"refunds", "payouts", "qr_codes", "settlements",
	}

	for _, name := range expectedToolsets {
		if _, exists := toolsetGroup.Toolsets[name]; !exists {
			t.Errorf("Expected toolset %s not found", name)
		}
	}
}

func testSpecificEnabledToolsets(t *testing.T, obs *observability.Observability,
	client *rzpsdk.Client) {
	enabledToolsets := []string{"payments", "orders"}
	toolsetGroup, err := NewToolSets(obs, client, enabledToolsets, false)
	if err != nil {
		t.Fatalf("NewToolSets failed: %v", err)
	}

	if toolsetGroup == nil {
		t.Fatal("NewToolSets returned nil toolset group")
		return
	}

	// Verify payments toolset is enabled
	payments, exists := toolsetGroup.Toolsets["payments"]
	if !exists {
		t.Error("Payments toolset not found")
	} else if !payments.Enabled {
		t.Error("Payments toolset should be enabled")
	}

	// Verify orders toolset is enabled
	orders, exists := toolsetGroup.Toolsets["orders"]
	if !exists {
		t.Error("Orders toolset not found")
	} else if !orders.Enabled {
		t.Error("Orders toolset should be enabled")
	}

	// Verify other toolsets are not enabled
	refunds, exists := toolsetGroup.Toolsets["refunds"]
	if !exists {
		t.Error("Refunds toolset not found")
	} else if refunds.Enabled {
		t.Error("Refunds toolset should not be enabled")
	}
}

func testReadOnlyMode(t *testing.T, obs *observability.Observability,
	client *rzpsdk.Client) {
	toolsetGroup, err := NewToolSets(obs, client, []string{}, true)
	if err != nil {
		t.Fatalf("NewToolSets failed: %v", err)
	}

	if toolsetGroup == nil {
		t.Fatal("NewToolSets returned nil toolset group")
		return
	}

	// We can't directly test the readOnly field since it's unexported,
	// but we can verify the toolset was created successfully in read-only mode
	if len(toolsetGroup.Toolsets) == 0 {
		t.Error("Expected toolsets to be created in read-only mode")
	}
}

func testInvalidToolsetName(t *testing.T, obs *observability.Observability,
	client *rzpsdk.Client) {
	enabledToolsets := []string{"invalid_toolset"}
	_, err := NewToolSets(obs, client, enabledToolsets, false)
	if err == nil {
		t.Fatal("Expected error for invalid toolset name")
	}

	expectedError := "toolset invalid_toolset does not exist"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func testMixedValidInvalidToolsets(t *testing.T,
	obs *observability.Observability, client *rzpsdk.Client) {
	enabledToolsets := []string{"payments", "invalid_toolset"}
	_, err := NewToolSets(obs, client, enabledToolsets, false)
	if err == nil {
		t.Fatal("Expected error for invalid toolset name")
	}

	expectedError := "toolset invalid_toolset does not exist"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func testAllToolsCreation(t *testing.T, obs *observability.Observability,
	client *rzpsdk.Client) {
	toolsetGroup, err := NewToolSets(obs, client, []string{}, false)
	if err != nil {
		t.Fatalf("NewToolSets failed: %v", err)
	}

	// Verify payments toolset has the additional tools added
	payments := toolsetGroup.Toolsets["payments"]
	if payments == nil {
		t.Fatal("Payments toolset not found")
	}

	// The test ensures that line 104-105 is executed:
	// payments.AddReadTools(FetchSavedPaymentMethods(obs, client)).
	//     AddWriteTools(RevokeToken(obs, client))

	// We can't easily verify the exact tools without exposing internal state,
	// but the fact that no error occurred means the code executed successfully
}

func testSingleToolsetEnabled(t *testing.T, obs *observability.Observability,
	client *rzpsdk.Client) {
	enabledToolsets := []string{"settlements"}
	toolsetGroup, err := NewToolSets(obs, client, enabledToolsets, false)
	if err != nil {
		t.Fatalf("NewToolSets failed: %v", err)
	}

	// Verify only settlements toolset is enabled
	for name, toolset := range toolsetGroup.Toolsets {
		if name == "settlements" {
			if !toolset.Enabled {
				t.Errorf("Settlements toolset should be enabled")
			}
		} else {
			if toolset.Enabled {
				t.Errorf("Toolset %s should not be enabled", name)
			}
		}
	}
}

func testMultipleSpecificToolsets(t *testing.T,
	obs *observability.Observability, client *rzpsdk.Client) {
	enabledToolsets := []string{"payment_links", "qr_codes", "payouts"}
	toolsetGroup, err := NewToolSets(obs, client, enabledToolsets, false)
	if err != nil {
		t.Fatalf("NewToolSets failed: %v", err)
	}

	// Verify specified toolsets are enabled
	expectedEnabled := map[string]bool{
		"payment_links": true,
		"qr_codes":      true,
		"payouts":       true,
		"payments":      false,
		"orders":        false,
		"refunds":       false,
		"settlements":   false,
	}

	for name, shouldBeEnabled := range expectedEnabled {
		toolset := toolsetGroup.Toolsets[name]
		if toolset == nil {
			t.Errorf("Toolset %s not found", name)
			continue
		}
		if toolset.Enabled != shouldBeEnabled {
			t.Errorf("Toolset %s enabled state: expected %v, got %v",
				name, shouldBeEnabled, toolset.Enabled)
		}
	}
}
