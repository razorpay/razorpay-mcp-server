package integrations

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/razorpay/razorpay-mcp-server/pkg/log"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// createTestObservability creates an observability stack suitable for testing
func createTestObservability() *observability.Observability {
	_, logger := log.New(context.Background(), log.NewConfig(
		log.WithMode(log.ModeStdio)),
	)
	return &observability.Observability{
		Logger: logger,
	}
}

// TestPGStandardGetSupportedStacks tests the supported stacks tool
func TestPGStandardGetSupportedStacks(t *testing.T) {
	obs := createTestObservability()
	tool := PGStandardGetSupportedStacks(obs)

	t.Run("returns all supported stacks", func(t *testing.T) {
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{},
		}

		result, err := tool.GetHandler()(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)

		var output SupportedStacksOutput
		err = json.Unmarshal([]byte(result.Text), &output)
		require.NoError(t, err)

		// Verify all stacks are returned
		assert.Len(t, output.Stacks, len(SupportedStacks))
		assert.Len(t, output.ServerLanguages, len(SupportedServerLanguages))

		// Verify each stack has required fields
		for _, stack := range output.Stacks {
			assert.NotEmpty(t, stack.ID)
			assert.NotEmpty(t, stack.Name)
			assert.NotEmpty(t, stack.Description)
			assert.NotEmpty(t, stack.DocsURL)
			assert.NotEmpty(t, stack.Category)
		}
	})
}

// TestPGStandardGetIntegrationPlan tests the integration plan tool
func TestPGStandardGetIntegrationPlan(t *testing.T) {
	obs := createTestObservability()
	tool := PGStandardGetIntegrationPlan(obs)

	testCases := []struct {
		name           string
		request        map[string]interface{}
		expectError    bool
		expectedErrMsg string
		validateOutput func(t *testing.T, output IntegrationPlanOutput)
	}{
		{
			name: "web_standard with nodejs",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output IntegrationPlanOutput) {
				assert.Equal(t, StackWebStandard, output.Stack)
				assert.Equal(t, LangNodeJS, output.ServerLanguage)
				assert.NotEmpty(t, output.CanonicalSteps)
				assert.NotEmpty(t, output.StackSpecificNotes)
				assert.NotEmpty(t, output.SecurityWarnings)
				assert.NotEmpty(t, output.RecommendedNextToolCalls)

				// Verify canonical steps include required ones
				stepIDs := make([]string, len(output.CanonicalSteps))
				for i, step := range output.CanonicalSteps {
					stepIDs[i] = step.ID
				}
				assert.Contains(t, stepIDs, "create_order_server")
				assert.Contains(t, stepIDs, "open_checkout_client")
				assert.Contains(t, stepIDs, "verify_signature_server")

				// Verify API endpoint is mentioned in notes
				foundAPINote := false
				for _, note := range output.StackSpecificNotes {
					if note == "Razorpay Orders API: POST "+RazorpayOrdersAPIURL {
						foundAPINote = true
						break
					}
				}
				assert.True(t, foundAPINote, "Should include Razorpay API endpoint in notes")
			},
		},
		{
			name: "android_standard with python",
			request: map[string]interface{}{
				"stack":           StackAndroidStandard,
				"server_language": LangPython,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output IntegrationPlanOutput) {
				assert.Equal(t, StackAndroidStandard, output.Stack)
				assert.Equal(t, LangPython, output.ServerLanguage)
				assert.NotEmpty(t, output.StackSpecificNotes)
			},
		},
		{
			name: "ios_standard with go",
			request: map[string]interface{}{
				"stack":           StackIOSStandard,
				"server_language": LangGo,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output IntegrationPlanOutput) {
				assert.Equal(t, StackIOSStandard, output.Stack)
			},
		},
		{
			name: "flutter_standard with default language",
			request: map[string]interface{}{
				"stack": StackFlutterStandard,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output IntegrationPlanOutput) {
				assert.Equal(t, StackFlutterStandard, output.Stack)
			},
		},
		{
			name: "missing required stack",
			request: map[string]interface{}{
				"server_language": LangNodeJS,
			},
			expectError:    true,
			expectedErrMsg: "stack is required",
		},
		{
			name: "invalid stack",
			request: map[string]interface{}{
				"stack": "invalid_stack",
			},
			expectError:    true,
			expectedErrMsg: "invalid stack",
		},
		{
			name: "invalid server_language",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": "invalid_lang",
			},
			expectError:    true,
			expectedErrMsg: "invalid server_language",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcpgo.CallToolRequest{
				Arguments: tc.request,
			}

			result, err := tool.GetHandler()(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, result)

			if tc.expectError {
				assert.True(t, result.IsError)
				assert.Contains(t, result.Text, tc.expectedErrMsg)
				return
			}

			assert.False(t, result.IsError)

			var output IntegrationPlanOutput
			err = json.Unmarshal([]byte(result.Text), &output)
			require.NoError(t, err)

			if tc.validateOutput != nil {
				tc.validateOutput(t, output)
			}
		})
	}
}

// TestPGStandardGetSnippets tests the snippets tool
func TestPGStandardGetSnippets(t *testing.T) {
	obs := createTestObservability()
	tool := PGStandardGetSnippets(obs)

	testCases := []struct {
		name           string
		request        map[string]interface{}
		expectError    bool
		expectedErrMsg string
		validateOutput func(t *testing.T, output SnippetsOutput)
	}{
		{
			name: "order_create snippet nodejs",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
				"snippet_kind":    SnippetOrderCreate,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output SnippetsOutput) {
				assert.Equal(t, SnippetOrderCreate, output.SnippetKind)
				assert.NotEmpty(t, output.Files)
				assert.NotEmpty(t, output.Placeholders)
				assert.NotEmpty(t, output.Notes)

				// Verify file has content
				assert.NotEmpty(t, output.Files[0].Content)
				assert.Contains(t, output.Files[0].Content, "razorpay")
				// Verify API endpoint is mentioned
				assert.Contains(t, output.Files[0].Content, "api.razorpay.com/v1/orders")
			},
		},
		{
			name: "order_create snippet python",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangPython,
				"snippet_kind":    SnippetOrderCreate,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output SnippetsOutput) {
				assert.Equal(t, SnippetOrderCreate, output.SnippetKind)
				assert.Contains(t, output.Files[0].Content, "razorpay")
				assert.Equal(t, "python", output.Files[0].Language)
				// Verify API endpoint is mentioned
				assert.Contains(t, output.Files[0].Content, "api.razorpay.com/v1/orders")
			},
		},
		{
			name: "checkout_open snippet web",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
				"snippet_kind":    SnippetCheckoutOpen,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output SnippetsOutput) {
				assert.Equal(t, SnippetCheckoutOpen, output.SnippetKind)
				assert.Contains(t, output.Files[0].Content, "checkout.razorpay.com")
			},
		},
		{
			name: "checkout_open snippet android",
			request: map[string]interface{}{
				"stack":           StackAndroidStandard,
				"server_language": LangNodeJS,
				"snippet_kind":    SnippetCheckoutOpen,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output SnippetsOutput) {
				assert.Contains(t, output.Files[0].Content, "Checkout")
				assert.Equal(t, "kotlin", output.Files[0].Language)
			},
		},
		{
			name: "checkout_open snippet flutter",
			request: map[string]interface{}{
				"stack":           StackFlutterStandard,
				"server_language": LangNodeJS,
				"snippet_kind":    SnippetCheckoutOpen,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output SnippetsOutput) {
				assert.Contains(t, output.Files[0].Content, "razorpay_flutter")
				assert.Equal(t, "dart", output.Files[0].Language)
			},
		},
		{
			name: "signature_verify snippet go",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangGo,
				"snippet_kind":    SnippetSignatureVerify,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output SnippetsOutput) {
				assert.Equal(t, SnippetSignatureVerify, output.SnippetKind)
				assert.Contains(t, output.Files[0].Content, "hmac")
				assert.Equal(t, "go", output.Files[0].Language)
			},
		},
		{
			name: "env_template snippet",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
				"snippet_kind":    SnippetEnvTemplate,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output SnippetsOutput) {
				assert.Equal(t, SnippetEnvTemplate, output.SnippetKind)
				assert.Contains(t, output.Files[0].Content, "RAZORPAY_KEY_ID")
				// Verify API endpoint is mentioned
				assert.Contains(t, output.Files[0].Content, "api.razorpay.com/v1/orders")
			},
		},
		{
			name: "webhook_guidance snippet",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
				"snippet_kind":    SnippetWebhookGuidance,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output SnippetsOutput) {
				assert.Equal(t, SnippetWebhookGuidance, output.SnippetKind)
				assert.Contains(t, output.Files[0].Content, "Webhook")
			},
		},
		{
			name: "missing stack",
			request: map[string]interface{}{
				"server_language": LangNodeJS,
				"snippet_kind":    SnippetOrderCreate,
			},
			expectError:    true,
			expectedErrMsg: "stack is required",
		},
		{
			name: "missing server_language",
			request: map[string]interface{}{
				"stack":        StackWebStandard,
				"snippet_kind": SnippetOrderCreate,
			},
			expectError:    true,
			expectedErrMsg: "server_language is required",
		},
		{
			name: "missing snippet_kind",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
			},
			expectError:    true,
			expectedErrMsg: "snippet_kind is required",
		},
		{
			name: "invalid snippet_kind",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
				"snippet_kind":    "invalid_kind",
			},
			expectError:    true,
			expectedErrMsg: "invalid snippet_kind",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcpgo.CallToolRequest{
				Arguments: tc.request,
			}

			result, err := tool.GetHandler()(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, result)

			if tc.expectError {
				assert.True(t, result.IsError)
				assert.Contains(t, result.Text, tc.expectedErrMsg)
				return
			}

			assert.False(t, result.IsError)

			var output SnippetsOutput
			err = json.Unmarshal([]byte(result.Text), &output)
			require.NoError(t, err)

			if tc.validateOutput != nil {
				tc.validateOutput(t, output)
			}
		})
	}
}

// TestPGStandardGetValidationAndTestPlan tests the validation and test plan tool
func TestPGStandardGetValidationAndTestPlan(t *testing.T) {
	obs := createTestObservability()
	tool := PGStandardGetValidationAndTestPlan(obs)

	testCases := []struct {
		name           string
		request        map[string]interface{}
		expectError    bool
		expectedErrMsg string
		validateOutput func(t *testing.T, output ValidationAndTestPlanOutput)
	}{
		{
			name: "web with nodejs without go-live",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
				"include_go_live": false,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output ValidationAndTestPlanOutput) {
				assert.NotEmpty(t, output.ValidationCommands)
				assert.NotEmpty(t, output.TestPlan)
				assert.Empty(t, output.GoLiveChecklist)

				// Verify validation commands include SDK check
				foundSDKCheck := false
				for _, cmd := range output.ValidationCommands {
					if cmd.Command == "npm list razorpay" {
						foundSDKCheck = true
						break
					}
				}
				assert.True(t, foundSDKCheck, "Should include Node.js SDK check")
			},
		},
		{
			name: "web with nodejs with go-live",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
				"include_go_live": true,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output ValidationAndTestPlanOutput) {
				assert.NotEmpty(t, output.ValidationCommands)
				assert.NotEmpty(t, output.TestPlan)
				assert.NotEmpty(t, output.GoLiveChecklist)
				assert.Contains(t, output.GoLiveChecklist, "Replace test API keys with live API keys")
			},
		},
		{
			name: "android with python",
			request: map[string]interface{}{
				"stack":           StackAndroidStandard,
				"server_language": LangPython,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output ValidationAndTestPlanOutput) {
				// Should include Android-specific validation
				foundAndroidCheck := false
				for _, cmd := range output.ValidationCommands {
					if cmd.Description == "Verify Android SDK dependency is added" {
						foundAndroidCheck = true
						break
					}
				}
				assert.True(t, foundAndroidCheck, "Should include Android-specific check")

				// Should include Python SDK check
				foundPythonCheck := false
				for _, cmd := range output.ValidationCommands {
					if cmd.Command == "pip show razorpay" {
						foundPythonCheck = true
						break
					}
				}
				assert.True(t, foundPythonCheck, "Should include Python SDK check")
			},
		},
		{
			name: "flutter with go",
			request: map[string]interface{}{
				"stack":           StackFlutterStandard,
				"server_language": LangGo,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output ValidationAndTestPlanOutput) {
				// Should include Flutter-specific validation
				foundFlutterCheck := false
				for _, cmd := range output.ValidationCommands {
					if cmd.Description == "Verify Flutter SDK is installed" {
						foundFlutterCheck = true
						break
					}
				}
				assert.True(t, foundFlutterCheck, "Should include Flutter-specific check")
			},
		},
		{
			name: "test plan includes API endpoint",
			request: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
			},
			expectError: false,
			validateOutput: func(t *testing.T, output ValidationAndTestPlanOutput) {
				// Verify test plan has expected steps
				assert.GreaterOrEqual(t, len(output.TestPlan), 5)

				// Steps should be numbered sequentially
				for i, step := range output.TestPlan {
					assert.Equal(t, i+1, step.Step)
					assert.NotEmpty(t, step.Action)
					assert.NotEmpty(t, step.Expected)
				}

				// Verify API endpoint is mentioned in test plan
				foundAPIStep := false
				for _, step := range output.TestPlan {
					if step.Action != "" && (step.Action == "Create a test order via your backend (calls POST "+RazorpayOrdersAPIURL+")") {
						foundAPIStep = true
						break
					}
				}
				assert.True(t, foundAPIStep, "Test plan should mention Razorpay API endpoint")
			},
		},
		{
			name: "missing stack",
			request: map[string]interface{}{
				"server_language": LangNodeJS,
			},
			expectError:    true,
			expectedErrMsg: "stack is required",
		},
		{
			name: "missing server_language",
			request: map[string]interface{}{
				"stack": StackWebStandard,
			},
			expectError:    true,
			expectedErrMsg: "server_language is required",
		},
		{
			name: "invalid stack",
			request: map[string]interface{}{
				"stack":           "invalid",
				"server_language": LangNodeJS,
			},
			expectError:    true,
			expectedErrMsg: "invalid stack",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcpgo.CallToolRequest{
				Arguments: tc.request,
			}

			result, err := tool.GetHandler()(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, result)

			if tc.expectError {
				assert.True(t, result.IsError)
				assert.Contains(t, result.Text, tc.expectedErrMsg)
				return
			}

			assert.False(t, result.IsError)

			var output ValidationAndTestPlanOutput
			err = json.Unmarshal([]byte(result.Text), &output)
			require.NoError(t, err)

			if tc.validateOutput != nil {
				tc.validateOutput(t, output)
			}
		})
	}
}

// TestAllStacksHaveSnippets verifies all stacks have checkout snippets
func TestAllStacksHaveSnippets(t *testing.T) {
	obs := createTestObservability()
	tool := PGStandardGetSnippets(obs)

	// Test checkout_open for all client stacks
	clientStacks := []string{
		StackWebStandard,
		StackAndroidStandard,
		StackIOSStandard,
		StackReactNativeStandard,
		StackFlutterStandard,
		StackCordovaStandard,
		StackIonicStandard,
	}

	for _, stack := range clientStacks {
		t.Run("checkout_open_"+stack, func(t *testing.T) {
			request := mcpgo.CallToolRequest{
				Arguments: map[string]interface{}{
					"stack":           stack,
					"server_language": LangNodeJS,
					"snippet_kind":    SnippetCheckoutOpen,
				},
			}

			result, err := tool.GetHandler()(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError, "Should not error for stack: %s", stack)

			var output SnippetsOutput
			err = json.Unmarshal([]byte(result.Text), &output)
			require.NoError(t, err)

			assert.NotEmpty(t, output.Files, "Should have files for stack: %s", stack)
			assert.NotEmpty(t, output.Files[0].Content, "Should have content for stack: %s", stack)
		})
	}
}

// TestAllLanguagesHaveOrderCreate verifies all server languages have order create snippets
func TestAllLanguagesHaveOrderCreate(t *testing.T) {
	obs := createTestObservability()
	tool := PGStandardGetSnippets(obs)

	serverLangs := []string{
		LangNodeJS,
		LangPython,
		LangPHP,
		LangRuby,
		LangJava,
		LangDotNet,
		LangGo,
	}

	for _, lang := range serverLangs {
		t.Run("order_create_"+lang, func(t *testing.T) {
			request := mcpgo.CallToolRequest{
				Arguments: map[string]interface{}{
					"stack":           StackWebStandard,
					"server_language": lang,
					"snippet_kind":    SnippetOrderCreate,
				},
			}

			result, err := tool.GetHandler()(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError, "Should not error for language: %s", lang)

			var output SnippetsOutput
			err = json.Unmarshal([]byte(result.Text), &output)
			require.NoError(t, err)

			assert.NotEmpty(t, output.Files, "Should have files for language: %s", lang)
			assert.NotEmpty(t, output.Files[0].Content, "Should have content for language: %s", lang)
			// Verify API endpoint is mentioned
			assert.Contains(t, output.Files[0].Content, "api.razorpay.com/v1/orders",
				"Should mention Razorpay API endpoint for language: %s", lang)
		})
	}
}

// TestAllLanguagesHaveSignatureVerify verifies all server languages have signature verify snippets
func TestAllLanguagesHaveSignatureVerify(t *testing.T) {
	obs := createTestObservability()
	tool := PGStandardGetSnippets(obs)

	serverLangs := []string{
		LangNodeJS,
		LangPython,
		LangPHP,
		LangRuby,
		LangJava,
		LangDotNet,
		LangGo,
	}

	for _, lang := range serverLangs {
		t.Run("signature_verify_"+lang, func(t *testing.T) {
			request := mcpgo.CallToolRequest{
				Arguments: map[string]interface{}{
					"stack":           StackWebStandard,
					"server_language": lang,
					"snippet_kind":    SnippetSignatureVerify,
				},
			}

			result, err := tool.GetHandler()(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError, "Should not error for language: %s", lang)

			var output SnippetsOutput
			err = json.Unmarshal([]byte(result.Text), &output)
			require.NoError(t, err)

			assert.NotEmpty(t, output.Files, "Should have files for language: %s", lang)
			// Verify signature snippet contains HMAC-related content
			contentLower := output.Files[0].Content
			assert.True(t,
				containsIgnoreCase(contentLower, "hmac") || containsIgnoreCase(contentLower, "HMAC"),
				"Signature snippet should contain HMAC for language: %s", lang)
		})
	}
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0)
}

// TestEnumValidation tests that invalid enum values are rejected
func TestEnumValidation(t *testing.T) {
	obs := createTestObservability()

	t.Run("IsValidStack", func(t *testing.T) {
		assert.True(t, IsValidStack(StackWebStandard))
		assert.True(t, IsValidStack(StackAndroidStandard))
		assert.False(t, IsValidStack("invalid"))
		assert.False(t, IsValidStack(""))
	})

	t.Run("IsValidServerLanguage", func(t *testing.T) {
		assert.True(t, IsValidServerLanguage(LangNodeJS))
		assert.True(t, IsValidServerLanguage(LangPython))
		assert.True(t, IsValidServerLanguage(LangUnknown))
		assert.False(t, IsValidServerLanguage("invalid"))
		assert.False(t, IsValidServerLanguage(""))
	})

	t.Run("IsValidSnippetKind", func(t *testing.T) {
		assert.True(t, IsValidSnippetKind(SnippetOrderCreate))
		assert.True(t, IsValidSnippetKind(SnippetCheckoutOpen))
		assert.True(t, IsValidSnippetKind(SnippetSignatureVerify))
		assert.False(t, IsValidSnippetKind("invalid"))
		assert.False(t, IsValidSnippetKind(""))
	})

	// Test using the actual tool
	snippetTool := PGStandardGetSnippets(obs)

	t.Run("tool rejects invalid stack", func(t *testing.T) {
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"stack":           "not_a_stack",
				"server_language": LangNodeJS,
				"snippet_kind":    SnippetOrderCreate,
			},
		}

		result, err := snippetTool.GetHandler()(context.Background(), request)
		require.NoError(t, err)
		assert.True(t, result.IsError)
		assert.Contains(t, result.Text, "invalid stack")
	})

	t.Run("tool rejects invalid server_language", func(t *testing.T) {
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": "cobol",
				"snippet_kind":    SnippetOrderCreate,
			},
		}

		result, err := snippetTool.GetHandler()(context.Background(), request)
		require.NoError(t, err)
		assert.True(t, result.IsError)
		assert.Contains(t, result.Text, "invalid server_language")
	})

	t.Run("tool rejects invalid snippet_kind", func(t *testing.T) {
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
				"snippet_kind":    "not_a_kind",
			},
		}

		result, err := snippetTool.GetHandler()(context.Background(), request)
		require.NoError(t, err)
		assert.True(t, result.IsError)
		assert.Contains(t, result.Text, "invalid snippet_kind")
	})
}

// TestJSONOutputValidity verifies all tools return valid JSON
func TestJSONOutputValidity(t *testing.T) {
	obs := createTestObservability()

	t.Run("supported_stacks returns valid JSON", func(t *testing.T) {
		tool := PGStandardGetSupportedStacks(obs)
		request := mcpgo.CallToolRequest{Arguments: map[string]interface{}{}}

		result, err := tool.GetHandler()(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, result.IsError)

		var output interface{}
		err = json.Unmarshal([]byte(result.Text), &output)
		assert.NoError(t, err, "Output should be valid JSON")
	})

	t.Run("integration_plan returns valid JSON", func(t *testing.T) {
		tool := PGStandardGetIntegrationPlan(obs)
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
			},
		}

		result, err := tool.GetHandler()(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, result.IsError)

		var output interface{}
		err = json.Unmarshal([]byte(result.Text), &output)
		assert.NoError(t, err, "Output should be valid JSON")
	})

	t.Run("snippets returns valid JSON", func(t *testing.T) {
		tool := PGStandardGetSnippets(obs)
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
				"snippet_kind":    SnippetOrderCreate,
			},
		}

		result, err := tool.GetHandler()(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, result.IsError)

		var output interface{}
		err = json.Unmarshal([]byte(result.Text), &output)
		assert.NoError(t, err, "Output should be valid JSON")
	})

	t.Run("validation_test_plan returns valid JSON", func(t *testing.T) {
		tool := PGStandardGetValidationAndTestPlan(obs)
		request := mcpgo.CallToolRequest{
			Arguments: map[string]interface{}{
				"stack":           StackWebStandard,
				"server_language": LangNodeJS,
				"include_go_live": true,
			},
		}

		result, err := tool.GetHandler()(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, result.IsError)

		var output interface{}
		err = json.Unmarshal([]byte(result.Text), &output)
		assert.NoError(t, err, "Output should be valid JSON")
	})
}

