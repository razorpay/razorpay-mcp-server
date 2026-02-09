// Package integrations provides Payment Gateway integration assistance tools.
package integrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// PGStandardGetSupportedStacks returns a tool that lists all supported stacks
// for Payment Gateway Standard Checkout integration.
func PGStandardGetSupportedStacks(
	obs *observability.Observability,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{} // No parameters needed

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		stackInfoMap := GetStackInfoMap()

		stacks := make([]StackInfo, 0, len(SupportedStacks))
		for _, stackID := range SupportedStacks {
			if info, ok := stackInfoMap[stackID]; ok {
				stacks = append(stacks, info)
			}
		}

		output := SupportedStacksOutput{
			Stacks:          stacks,
			ServerLanguages: SupportedServerLanguages,
		}

		return mcpgo.NewToolResultJSON(output)
	}

	return mcpgo.NewTool(
		"pg_standard_get_supported_stacks",
		"Returns all supported stack identifiers and server languages for "+
			"Razorpay Payment Gateway Standard Checkout integration. Use this "+
			"to discover available integration options before generating plans or snippets.",
		parameters,
		handler,
	)
}

// PGStandardGetIntegrationPlan returns a tool that provides a step-by-step
// integration plan for a specific stack.
func PGStandardGetIntegrationPlan(
	obs *observability.Observability,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"stack",
			mcpgo.Description("Target platform/stack for integration. "+
				"Use pg_standard_get_supported_stacks to get valid options."),
			mcpgo.Required(),
			mcpgo.Enum(
				StackWebStandard,
				StackAndroidStandard,
				StackIOSStandard,
				StackReactNativeStandard,
				StackFlutterStandard,
				StackCordovaStandard,
				StackIonicStandard,
				StackServerSDK,
			),
		),
		mcpgo.WithString(
			"server_language",
			mcpgo.Description("Server-side programming language for backend integration."),
			mcpgo.Enum(
				LangNodeJS,
				LangPython,
				LangPHP,
				LangRuby,
				LangJava,
				LangDotNet,
				LangGo,
				LangUnknown,
			),
			mcpgo.DefaultValue(LangUnknown),
		),
		mcpgo.WithString(
			"notes",
			mcpgo.Description("Optional notes or requirements for the integration plan."),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		var stack, serverLang, notes string
		var errors []string

		argsMap, ok := r.Arguments.(map[string]interface{})
		if !ok {
			return mcpgo.NewToolResultError("invalid request arguments"), nil
		}

		// Validate stack (required)
		if val, exists := argsMap["stack"]; exists {
			if s, ok := val.(string); ok && strings.TrimSpace(s) != "" {
				stack = s
			} else {
				errors = append(errors, "stack must be a non-empty string")
			}
		} else {
			errors = append(errors, "stack is required")
		}

		// Get optional server_language
		if val, exists := argsMap["server_language"]; exists && val != nil {
			if s, ok := val.(string); ok {
				serverLang = s
			}
		}

		// Get optional notes
		if val, exists := argsMap["notes"]; exists && val != nil {
			if s, ok := val.(string); ok {
				notes = s
			}
		}

		if len(errors) > 0 {
			return mcpgo.NewToolResultError(strings.Join(errors, "; ")), nil
		}

		// Validate stack
		if !IsValidStack(stack) {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("invalid stack: %s. Use pg_standard_get_supported_stacks "+
					"to get valid options.", stack),
			), nil
		}

		// Default server language
		if serverLang == "" {
			serverLang = LangUnknown
		}

		// Validate server language if provided
		if serverLang != "" && !IsValidServerLanguage(serverLang) {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("invalid server_language: %s", serverLang),
			), nil
		}

		// Silence unused variable warning
		_ = notes

		// Generate integration plan
		plan := generateIntegrationPlan(stack, serverLang)

		return mcpgo.NewToolResultJSON(plan)
	}

	return mcpgo.NewTool(
		"pg_standard_get_integration_plan",
		"Returns a step-by-step integration plan for Razorpay Payment Gateway "+
			"Standard Checkout. Includes canonical steps, stack-specific notes, "+
			"security warnings, and recommended next tool calls.",
		parameters,
		handler,
	)
}

// PGStandardGetSnippets returns a tool that provides code snippets
// for specific integration tasks.
func PGStandardGetSnippets(
	obs *observability.Observability,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"stack",
			mcpgo.Description("Target platform/stack for integration."),
			mcpgo.Required(),
			mcpgo.Enum(
				StackWebStandard,
				StackAndroidStandard,
				StackIOSStandard,
				StackReactNativeStandard,
				StackFlutterStandard,
				StackCordovaStandard,
				StackIonicStandard,
				StackServerSDK,
			),
		),
		mcpgo.WithString(
			"server_language",
			mcpgo.Description("Server-side programming language for backend code."),
			mcpgo.Required(),
			mcpgo.Enum(
				LangNodeJS,
				LangPython,
				LangPHP,
				LangRuby,
				LangJava,
				LangDotNet,
				LangGo,
			),
		),
		mcpgo.WithString(
			"snippet_kind",
			mcpgo.Description("Type of code snippet to generate."),
			mcpgo.Required(),
			mcpgo.Enum(
				SnippetOrderCreate,
				SnippetCheckoutOpen,
				SnippetSignatureVerify,
				SnippetEnvTemplate,
				SnippetWebhookGuidance,
			),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		var stack, serverLang, snippetKind string
		var errors []string

		argsMap, ok := r.Arguments.(map[string]interface{})
		if !ok {
			return mcpgo.NewToolResultError("invalid request arguments"), nil
		}

		// Validate stack (required)
		if val, exists := argsMap["stack"]; exists {
			if s, ok := val.(string); ok && strings.TrimSpace(s) != "" {
				stack = s
			} else {
				errors = append(errors, "stack must be a non-empty string")
			}
		} else {
			errors = append(errors, "stack is required")
		}

		// Validate server_language (required)
		if val, exists := argsMap["server_language"]; exists {
			if s, ok := val.(string); ok && strings.TrimSpace(s) != "" {
				serverLang = s
			} else {
				errors = append(errors, "server_language must be a non-empty string")
			}
		} else {
			errors = append(errors, "server_language is required")
		}

		// Validate snippet_kind (required)
		if val, exists := argsMap["snippet_kind"]; exists {
			if s, ok := val.(string); ok && strings.TrimSpace(s) != "" {
				snippetKind = s
			} else {
				errors = append(errors, "snippet_kind must be a non-empty string")
			}
		} else {
			errors = append(errors, "snippet_kind is required")
		}

		if len(errors) > 0 {
			return mcpgo.NewToolResultError(strings.Join(errors, "; ")), nil
		}

		// Validate inputs
		if !IsValidStack(stack) {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("invalid stack: %s", stack),
			), nil
		}
		if !IsValidServerLanguage(serverLang) {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("invalid server_language: %s", serverLang),
			), nil
		}
		if !IsValidSnippetKind(snippetKind) {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("invalid snippet_kind: %s", snippetKind),
			), nil
		}

		// Generate snippets
		output := generateSnippets(stack, serverLang, snippetKind)

		return mcpgo.NewToolResultJSON(output)
	}

	return mcpgo.NewTool(
		"pg_standard_get_snippets",
		"Returns code snippets for specific Razorpay Payment Gateway integration "+
			"tasks. Supports order creation, checkout opening, signature verification, "+
			"environment templates, and webhook guidance.",
		parameters,
		handler,
	)
}

// PGStandardGetValidationAndTestPlan returns a tool that provides
// validation commands and test plans for integration verification.
func PGStandardGetValidationAndTestPlan(
	obs *observability.Observability,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"stack",
			mcpgo.Description("Target platform/stack for integration."),
			mcpgo.Required(),
			mcpgo.Enum(
				StackWebStandard,
				StackAndroidStandard,
				StackIOSStandard,
				StackReactNativeStandard,
				StackFlutterStandard,
				StackCordovaStandard,
				StackIonicStandard,
				StackServerSDK,
			),
		),
		mcpgo.WithString(
			"server_language",
			mcpgo.Description("Server-side programming language."),
			mcpgo.Required(),
			mcpgo.Enum(
				LangNodeJS,
				LangPython,
				LangPHP,
				LangRuby,
				LangJava,
				LangDotNet,
				LangGo,
			),
		),
		mcpgo.WithBoolean(
			"include_go_live",
			mcpgo.Description("Include go-live checklist in response."),
			mcpgo.DefaultValue(false),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		var stack, serverLang string
		var includeGoLive bool
		var errors []string

		argsMap, ok := r.Arguments.(map[string]interface{})
		if !ok {
			return mcpgo.NewToolResultError("invalid request arguments"), nil
		}

		// Validate stack (required)
		if val, exists := argsMap["stack"]; exists {
			if s, ok := val.(string); ok && strings.TrimSpace(s) != "" {
				stack = s
			} else {
				errors = append(errors, "stack must be a non-empty string")
			}
		} else {
			errors = append(errors, "stack is required")
		}

		// Validate server_language (required)
		if val, exists := argsMap["server_language"]; exists {
			if s, ok := val.(string); ok && strings.TrimSpace(s) != "" {
				serverLang = s
			} else {
				errors = append(errors, "server_language must be a non-empty string")
			}
		} else {
			errors = append(errors, "server_language is required")
		}

		// Get optional include_go_live
		if val, exists := argsMap["include_go_live"]; exists && val != nil {
			if b, ok := val.(bool); ok {
				includeGoLive = b
			}
		}

		if len(errors) > 0 {
			return mcpgo.NewToolResultError(strings.Join(errors, "; ")), nil
		}

		// Validate inputs
		if !IsValidStack(stack) {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("invalid stack: %s", stack),
			), nil
		}
		if !IsValidServerLanguage(serverLang) {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("invalid server_language: %s", serverLang),
			), nil
		}

		// Generate validation and test plan
		output := generateValidationAndTestPlan(stack, serverLang, includeGoLive)

		return mcpgo.NewToolResultJSON(output)
	}

	return mcpgo.NewTool(
		"pg_standard_get_validation_and_test_plan",
		"Returns validation commands, test plan steps, and optionally a go-live "+
			"checklist for verifying Razorpay Payment Gateway integration.",
		parameters,
		handler,
	)
}

// generateIntegrationPlan creates an integration plan for the given stack
func generateIntegrationPlan(stack, serverLang string) IntegrationPlanOutput {
	stackInfo := GetStackInfoMap()[stack]

	// Base canonical steps (common to all stacks)
	canonicalSteps := []IntegrationStep{
		{
			ID:       "install_dependencies",
			Title:    "Install SDK/Dependencies",
			Required: true,
			DocsRef:  stackInfo.DocsURL + "integration-steps/",
		},
		{
			ID:    "create_order_server",
			Title: "Create Order (server-side) - Calls POST " + RazorpayOrdersAPIURL,
			Required: true,
			DocsRef:  "https://razorpay.com/docs/api/orders/create/",
		},
		{
			ID:       "open_checkout_client",
			Title:    "Open Checkout (client-side)",
			Required: true,
			DocsRef:  stackInfo.DocsURL + "integration-steps/",
		},
		{
			ID:       "handle_payment_response",
			Title:    "Handle Payment Response",
			Required: true,
			DocsRef:  stackInfo.DocsURL + "integration-steps/",
		},
		{
			ID:       "verify_signature_server",
			Title:    "Verify Payment Signature (server-side) - HMAC-SHA256(order_id|payment_id, key_secret)",
			Required: true,
			DocsRef:  "https://razorpay.com/docs/payments/server-integration/nodejs/payment-verification/",
		},
		{
			ID:       "test_payment",
			Title:    "Test Integration with Test Mode",
			Required: true,
			DocsRef:  "https://razorpay.com/docs/payments/payments/test-integration/",
		},
		{
			ID:       "setup_webhooks",
			Title:    "Configure Webhooks (recommended)",
			Required: false,
			DocsRef:  "https://razorpay.com/docs/webhooks/",
		},
		{
			ID:       "go_live",
			Title:    "Go-live Checklist",
			Required: false,
			DocsRef:  "https://razorpay.com/docs/payments/go-live-checklist/",
		},
	}

	// Stack-specific notes
	stackNotes := getStackSpecificNotes(stack)

	// Add API endpoint clarification
	stackNotes = append([]string{
		"Razorpay Orders API: POST " + RazorpayOrdersAPIURL,
		"Your backend creates orders via SDK which calls the above API",
	}, stackNotes...)

	// Server language specific notes
	if serverLang != LangUnknown && serverLang != "" {
		langNote := fmt.Sprintf("Server-side code will use %s SDK/patterns", serverLang)
		stackNotes = append(stackNotes, langNote)
	}

	// Recommended next tool calls
	nextCalls := []string{
		fmt.Sprintf("pg_standard_get_snippets(stack=%q, server_language=%q, "+
			"snippet_kind=\"order_create\")", stack, serverLang),
		fmt.Sprintf("pg_standard_get_snippets(stack=%q, server_language=%q, "+
			"snippet_kind=\"checkout_open\")", stack, serverLang),
		fmt.Sprintf("pg_standard_get_snippets(stack=%q, server_language=%q, "+
			"snippet_kind=\"signature_verify\")", stack, serverLang),
		fmt.Sprintf("pg_standard_get_validation_and_test_plan(stack=%q, "+
			"server_language=%q, include_go_live=true)", stack, serverLang),
	}

	return IntegrationPlanOutput{
		Stack:                    stack,
		ServerLanguage:           serverLang,
		CanonicalSteps:           canonicalSteps,
		StackSpecificNotes:       stackNotes,
		SecurityWarnings:         SecurityWarnings,
		RecommendedNextToolCalls: nextCalls,
	}
}

// getStackSpecificNotes returns notes specific to each stack
func getStackSpecificNotes(stack string) []string {
	notes := map[string][]string{
		StackWebStandard: {
			"Include checkout.js script before closing </body> tag",
			"Use defer attribute for script loading in production",
			"Handle modal dismiss events for better UX",
			"Consider prefilling customer details for faster checkout",
		},
		StackAndroidStandard: {
			"Call Checkout.preload() in Application class or splash screen for faster loading",
			"Add INTERNET permission in AndroidManifest.xml",
			"Implement PaymentResultListener interface for callbacks",
			"Handle configuration changes to prevent duplicate payments",
		},
		StackIOSStandard: {
			"Add razorpay-pod via CocoaPods or Swift Package Manager",
			"Implement RazorpayPaymentCompletionProtocol for callbacks",
			"Handle view controller presentation properly",
			"Support both Swift and Objective-C",
		},
		StackReactNativeStandard: {
			"Run pod install after adding react-native-razorpay on iOS",
			"Link native modules properly for older RN versions",
			"Handle deep linking if using UPI apps",
			"Use Promise-based API for cleaner async handling",
		},
		StackFlutterStandard: {
			"Run flutter pub get after adding dependency",
			"Clear Razorpay instance in dispose() to prevent memory leaks",
			"Handle external wallet selection if needed",
			"Platform-specific setup required for iOS (pod install)",
		},
		StackCordovaStandard: {
			"Plugin works with both Cordova CLI and Ionic",
			"Rebuild platforms after plugin installation",
			"Handle platform-specific issues on Android and iOS",
			"Use callbacks for payment success/failure handling",
		},
		StackIonicStandard: {
			"Uses Cordova plugin under the hood",
			"Works with both Ionic Angular and Ionic React",
			"Import from @nickyhelali/nickyhelali-razorpay for TypeScript support",
			"Handle platform detection for proper initialization",
		},
		StackServerSDK: {
			"Server-side only - no client SDK needed",
			"Used for order creation, payment verification, and webhooks",
			"Always use environment variables for credentials",
			"Implement proper error handling and logging",
		},
	}

	if stackNotes, ok := notes[stack]; ok {
		return stackNotes
	}
	return []string{}
}

// generateSnippets creates code snippets for the given parameters
func generateSnippets(stack, serverLang, snippetKind string) SnippetsOutput {
	var files []SnippetFile
	var notes []string

	placeholders := []SnippetPlaceholder{
		{
			Name:        "RAZORPAY_KEY_ID",
			Description: "Your Razorpay Key ID (starts with rzp_test_ or rzp_live_)",
			Example:     "rzp_test_xxxxxxxxxxxx",
		},
		{
			Name:        "RAZORPAY_KEY_SECRET",
			Description: "Your Razorpay Key Secret (never expose client-side)",
			Example:     "(keep secret, never commit)",
		},
	}

	switch snippetKind {
	case SnippetOrderCreate:
		files, notes = generateOrderCreateSnippet(serverLang)

	case SnippetCheckoutOpen:
		files, notes = generateCheckoutOpenSnippet(stack)

	case SnippetSignatureVerify:
		files, notes = generateSignatureVerifySnippet(serverLang)

	case SnippetEnvTemplate:
		files = []SnippetFile{
			{
				FilenameHint: ".env.example",
				Language:     "env",
				Description:  "Environment variables template for Razorpay credentials",
				Content:      EnvTemplate,
			},
		}
		notes = []string{
			"Copy this file as .env and fill in your actual credentials",
			"Add .env to .gitignore to prevent accidental commits",
			"Use test keys during development",
			"Razorpay API endpoint: " + RazorpayOrdersAPIURL,
		}

	case SnippetWebhookGuidance:
		files = []SnippetFile{
			{
				FilenameHint: "WEBHOOKS.md",
				Language:     "markdown",
				Description:  "Webhook configuration and verification guide",
				Content:      WebhookGuidance,
			},
		}
		notes = []string{
			"Webhooks provide reliable payment status updates",
			"Always verify webhook signatures before processing",
			"Implement idempotency to handle duplicate events",
		}
	}

	return SnippetsOutput{
		SnippetKind:  snippetKind,
		Files:        files,
		Placeholders: placeholders,
		Notes:        notes,
	}
}

// generateOrderCreateSnippet returns server-side order creation code
func generateOrderCreateSnippet(serverLang string) ([]SnippetFile, []string) {
	template, ok := OrderCreateTemplates[serverLang]
	if !ok {
		template = OrderCreateTemplates[LangNodeJS] // Default to Node.js
	}

	langInfo := getLanguageInfo(serverLang)

	files := []SnippetFile{
		{
			FilenameHint: langInfo.orderCreateFile,
			Language:     langInfo.codeLanguage,
			Description:  "Server-side endpoint that calls Razorpay Orders API: POST " + RazorpayOrdersAPIURL,
			Content:      template,
		},
	}

	notes := []string{
		"Your backend calls Razorpay API: POST " + RazorpayOrdersAPIURL,
		"The SDK handles the API call internally - you don't call it directly",
		"Amount must be in smallest currency unit (paise for INR)",
		"Store order_id in your database for payment verification",
		"Return key_id (not key_secret) to the client",
	}

	return files, notes
}

// generateCheckoutOpenSnippet returns client-side checkout code
func generateCheckoutOpenSnippet(stack string) ([]SnippetFile, []string) {
	template, ok := CheckoutOpenTemplates[stack]
	if !ok {
		template = CheckoutOpenTemplates[StackWebStandard]
	}

	stackLang := getStackLanguage(stack)

	files := []SnippetFile{
		{
			FilenameHint: getCheckoutFilename(stack),
			Language:     stackLang,
			Description:  "Client-side code to open Razorpay checkout",
			Content:      template,
		},
	}

	notes := []string{
		"Get order_id and key_id from your backend before opening checkout",
		"Your backend creates orders via: POST " + RazorpayOrdersAPIURL,
		"Handle both success and error callbacks",
		"Send payment response to server for verification",
		"Never hardcode API keys in client code",
	}

	return files, notes
}

// generateSignatureVerifySnippet returns signature verification code
func generateSignatureVerifySnippet(serverLang string) ([]SnippetFile, []string) {
	template, ok := SignatureVerifyTemplates[serverLang]
	if !ok {
		template = SignatureVerifyTemplates[LangNodeJS]
	}

	langInfo := getLanguageInfo(serverLang)

	files := []SnippetFile{
		{
			FilenameHint: langInfo.verifyFile,
			Language:     langInfo.codeLanguage,
			Description:  "Server-side payment signature verification",
			Content:      template,
		},
	}

	notes := []string{
		"CRITICAL: Use order_id from YOUR database, not from client response",
		"Signature formula: HMAC-SHA256(order_id|payment_id, key_secret)",
		"Always verify before updating order status or fulfilling orders",
		"Use constant-time comparison to prevent timing attacks",
	}

	return files, notes
}

// languageInfo holds metadata about a programming language
type languageInfo struct {
	codeLanguage    string
	orderCreateFile string
	verifyFile      string
}

// getLanguageInfo returns metadata for a server language
func getLanguageInfo(lang string) languageInfo {
	info := map[string]languageInfo{
		LangNodeJS: {
			codeLanguage:    "javascript",
			orderCreateFile: "server/routes/payment.js",
			verifyFile:      "server/routes/verify.js",
		},
		LangPython: {
			codeLanguage:    "python",
			orderCreateFile: "app/routes/payment.py",
			verifyFile:      "app/routes/verify.py",
		},
		LangGo: {
			codeLanguage:    "go",
			orderCreateFile: "handlers/payment.go",
			verifyFile:      "handlers/verify.go",
		},
		LangPHP: {
			codeLanguage:    "php",
			orderCreateFile: "api/create-order.php",
			verifyFile:      "api/verify-payment.php",
		},
		LangRuby: {
			codeLanguage:    "ruby",
			orderCreateFile: "app/controllers/payments_controller.rb",
			verifyFile:      "app/controllers/payments_controller.rb",
		},
		LangJava: {
			codeLanguage:    "java",
			orderCreateFile: "src/main/java/PaymentController.java",
			verifyFile:      "src/main/java/PaymentVerification.java",
		},
		LangDotNet: {
			codeLanguage:    "csharp",
			orderCreateFile: "Controllers/PaymentController.cs",
			verifyFile:      "Services/PaymentVerification.cs",
		},
	}

	if langInfo, ok := info[lang]; ok {
		return langInfo
	}
	return info[LangNodeJS]
}

// getStackLanguage returns the primary language for a stack
func getStackLanguage(stack string) string {
	langs := map[string]string{
		StackWebStandard:         "javascript",
		StackAndroidStandard:     "kotlin",
		StackIOSStandard:         "swift",
		StackReactNativeStandard: "javascript",
		StackFlutterStandard:     "dart",
		StackCordovaStandard:     "javascript",
		StackIonicStandard:       "typescript",
	}

	if lang, ok := langs[stack]; ok {
		return lang
	}
	return "javascript"
}

// getCheckoutFilename returns an appropriate filename for checkout code
func getCheckoutFilename(stack string) string {
	filenames := map[string]string{
		StackWebStandard:         "public/js/payment.js",
		StackAndroidStandard:     "app/src/main/java/PaymentActivity.kt",
		StackIOSStandard:         "PaymentViewController.swift",
		StackReactNativeStandard: "src/screens/PaymentScreen.js",
		StackFlutterStandard:     "lib/screens/payment_screen.dart",
		StackCordovaStandard:     "www/js/payment.js",
		StackIonicStandard:       "src/app/payment/payment.page.ts",
	}

	if filename, ok := filenames[stack]; ok {
		return filename
	}
	return "payment.js"
}

// generateValidationAndTestPlan creates validation commands and test plan
func generateValidationAndTestPlan(
	stack, serverLang string, includeGoLive bool,
) ValidationAndTestPlanOutput {
	// Validation commands based on server language
	validationCmds := getValidationCommands(serverLang)

	// Add stack-specific validation
	stackCmds := getStackValidationCommands(stack)
	validationCmds = append(validationCmds, stackCmds...)

	// Test plan
	testPlan := []TestPlanStep{
		{
			Step:     1,
			Action:   "Verify test credentials are configured",
			Expected: "RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET env vars are set with test keys",
			DocsRef:  "https://razorpay.com/docs/payments/dashboard/account-settings/api-keys/",
		},
		{
			Step:     2,
			Action:   "Create a test order via your backend (calls POST " + RazorpayOrdersAPIURL + ")",
			Expected: "Order created successfully with order_id starting with 'order_'",
			DocsRef:  "https://razorpay.com/docs/api/orders/create/",
		},
		{
			Step:     3,
			Action:   "Open checkout and complete payment with test card (4111 1111 1111 1111)",
			Expected: "Checkout opens, payment completes, success handler called",
			DocsRef:  "https://razorpay.com/docs/payments/payments/test-integration/",
		},
		{
			Step:     4,
			Action:   "Verify payment signature on your server",
			Expected: "Signature verification passes, payment confirmed",
			DocsRef:  "https://razorpay.com/docs/payments/server-integration/nodejs/payment-verification/",
		},
		{
			Step:     5,
			Action:   "Test payment failure scenarios",
			Expected: "Error handler called with appropriate error message",
			DocsRef:  "https://razorpay.com/docs/payments/payments/test-integration/",
		},
		{
			Step:     6,
			Action:   "Verify webhook delivery (if configured)",
			Expected: "Webhook received with correct event payload",
			DocsRef:  "https://razorpay.com/docs/webhooks/",
		},
	}

	output := ValidationAndTestPlanOutput{
		ValidationCommands: validationCmds,
		TestPlan:           testPlan,
	}

	// Go-live checklist
	if includeGoLive {
		output.GoLiveChecklist = []string{
			"Replace test API keys with live API keys",
			"Verify server-side signature verification is implemented and working",
			"Ensure HTTPS is enabled on all endpoints",
			"Configure webhooks for production URL",
			"Test with live mode using small amount (₹1)",
			"Verify error handling and user feedback",
			"Set up monitoring and alerting for payment failures",
			"Review and comply with PCI DSS requirements",
			"Document refund and cancellation workflows",
			"Train support team on payment issue resolution",
		}
	}

	return output
}

// getValidationCommands returns language-specific validation commands
func getValidationCommands(serverLang string) []ValidationCommand {
	baseCommands := []ValidationCommand{
		{
			Command:     "echo $RAZORPAY_KEY_ID",
			Description: "Verify RAZORPAY_KEY_ID environment variable is set",
			Optional:    false,
		},
	}

	langCommands := map[string][]ValidationCommand{
		LangNodeJS: {
			{
				Command:     "npm list razorpay",
				Description: "Verify Razorpay Node.js SDK is installed",
				Optional:    false,
			},
			{
				Command:     "node -e \"require('razorpay')\"",
				Description: "Verify SDK can be loaded",
				Optional:    false,
			},
		},
		LangPython: {
			{
				Command:     "pip show razorpay",
				Description: "Verify Razorpay Python SDK is installed",
				Optional:    false,
			},
			{
				Command:     "python -c \"import razorpay\"",
				Description: "Verify SDK can be imported",
				Optional:    false,
			},
		},
		LangGo: {
			{
				Command:     "go list -m github.com/razorpay/razorpay-go",
				Description: "Verify Razorpay Go SDK is in go.mod",
				Optional:    false,
			},
		},
		LangPHP: {
			{
				Command:     "composer show razorpay/razorpay",
				Description: "Verify Razorpay PHP SDK is installed",
				Optional:    false,
			},
		},
		LangRuby: {
			{
				Command:     "gem list razorpay",
				Description: "Verify Razorpay Ruby SDK is installed",
				Optional:    false,
			},
		},
		LangJava: {
			{
				Command:     "mvn dependency:tree | grep razorpay",
				Description: "Verify Razorpay Java SDK is in dependencies",
				Optional:    false,
			},
		},
		LangDotNet: {
			{
				Command:     "dotnet list package | grep Razorpay",
				Description: "Verify Razorpay .NET SDK is installed",
				Optional:    false,
			},
		},
	}

	if cmds, ok := langCommands[serverLang]; ok {
		baseCommands = append(baseCommands, cmds...)
	}

	return baseCommands
}

// getStackValidationCommands returns stack-specific validation commands
func getStackValidationCommands(stack string) []ValidationCommand {
	stackCmds := map[string][]ValidationCommand{
		StackAndroidStandard: {
			{
				Command:     "grep 'com.razorpay:checkout' app/build.gradle",
				Description: "Verify Android SDK dependency is added",
				Optional:    false,
			},
			{
				Command:     "grep 'INTERNET' app/src/main/AndroidManifest.xml",
				Description: "Verify INTERNET permission is declared",
				Optional:    false,
			},
		},
		StackIOSStandard: {
			{
				Command:     "pod list | grep razorpay",
				Description: "Verify iOS SDK pod is installed",
				Optional:    false,
			},
		},
		StackReactNativeStandard: {
			{
				Command:     "npm list react-native-razorpay",
				Description: "Verify React Native SDK is installed",
				Optional:    false,
			},
			{
				Command:     "cd ios && pod install --dry-run | grep Razorpay",
				Description: "Verify iOS pod is linked",
				Optional:    true,
			},
		},
		StackFlutterStandard: {
			{
				Command:     "flutter pub deps | grep razorpay_flutter",
				Description: "Verify Flutter SDK is installed",
				Optional:    false,
			},
		},
	}

	if cmds, ok := stackCmds[stack]; ok {
		return cmds
	}
	return []ValidationCommand{}
}

