// Package integrations provides Payment Gateway integration assistance tools.
package integrations

// Stack identifiers for Payment Gateway Standard Checkout integration
const (
	StackWebStandard         = "web_standard"
	StackAndroidStandard     = "android_standard"
	StackIOSStandard         = "ios_standard"
	StackReactNativeStandard = "react_native_standard"
	StackFlutterStandard     = "flutter_standard"
	StackCordovaStandard     = "cordova_standard"
	StackIonicStandard       = "ionic_standard"
	StackServerSDK           = "server_sdk"
)

// Server language identifiers
const (
	LangNodeJS  = "nodejs"
	LangPython  = "python"
	LangPHP     = "php"
	LangRuby    = "ruby"
	LangJava    = "java"
	LangDotNet  = "dotnet"
	LangGo      = "go"
	LangUnknown = "unknown"
)

// Snippet kinds for code generation
const (
	SnippetOrderCreate     = "order_create"
	SnippetCheckoutOpen    = "checkout_open"
	SnippetSignatureVerify = "signature_verify"
	SnippetEnvTemplate     = "env_template"
	SnippetWebhookGuidance = "webhook_guidance"
)

// Razorpay API endpoints
const (
	RazorpayAPIBaseURL    = "https://api.razorpay.com"
	RazorpayOrdersAPIPath = "/v1/orders"
	RazorpayOrdersAPIURL  = RazorpayAPIBaseURL + RazorpayOrdersAPIPath
)

// SupportedStacks returns all supported stack identifiers
var SupportedStacks = []string{
	StackWebStandard,
	StackAndroidStandard,
	StackIOSStandard,
	StackReactNativeStandard,
	StackFlutterStandard,
	StackCordovaStandard,
	StackIonicStandard,
	StackServerSDK,
}

// SupportedServerLanguages returns all supported server languages
var SupportedServerLanguages = []string{
	LangNodeJS,
	LangPython,
	LangPHP,
	LangRuby,
	LangJava,
	LangDotNet,
	LangGo,
	LangUnknown,
}

// SupportedSnippetKinds returns all supported snippet kinds
var SupportedSnippetKinds = []string{
	SnippetOrderCreate,
	SnippetCheckoutOpen,
	SnippetSignatureVerify,
	SnippetEnvTemplate,
	SnippetWebhookGuidance,
}

// StackInfo provides information about a supported stack
type StackInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DocsURL     string `json:"docs_url"`
	Category    string `json:"category"`
}

// IntegrationStep represents a single step in the integration plan
type IntegrationStep struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Required bool   `json:"required"`
	DocsRef  string `json:"docs_ref"`
}

// IntegrationPlanOutput is the response for pg_standard_get_integration_plan
type IntegrationPlanOutput struct {
	Stack                    string            `json:"stack"`
	ServerLanguage           string            `json:"server_language,omitempty"`
	CanonicalSteps           []IntegrationStep `json:"canonical_steps"`
	StackSpecificNotes       []string          `json:"stack_specific_notes"`
	SecurityWarnings         []string          `json:"security_warnings"`
	RecommendedNextToolCalls []string          `json:"recommended_next_tool_calls"`
}

// SnippetFile represents a single file in a snippet response
type SnippetFile struct {
	FilenameHint string `json:"filename_hint"`
	Language     string `json:"language"`
	Description  string `json:"description"`
	Content      string `json:"content"`
}

// SnippetPlaceholder represents a placeholder in a snippet
type SnippetPlaceholder struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// SnippetsOutput is the response for pg_standard_get_snippets
type SnippetsOutput struct {
	SnippetKind  string               `json:"snippet_kind"`
	Files        []SnippetFile        `json:"files"`
	Placeholders []SnippetPlaceholder `json:"placeholders"`
	Notes        []string             `json:"notes"`
}

// ValidationCommand represents a validation command
type ValidationCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Optional    bool   `json:"optional"`
}

// TestPlanStep represents a step in the test plan
type TestPlanStep struct {
	Step     int    `json:"step"`
	Action   string `json:"action"`
	Expected string `json:"expected"`
	DocsRef  string `json:"docs_ref"`
}

// ValidationAndTestPlanOutput is the response for pg_standard_get_validation_and_test_plan
type ValidationAndTestPlanOutput struct {
	ValidationCommands []ValidationCommand `json:"validation_commands"`
	TestPlan           []TestPlanStep      `json:"test_plan"`
	GoLiveChecklist    []string            `json:"go_live_checklist,omitempty"`
}

// SupportedStacksOutput is the response for pg_standard_get_supported_stacks
type SupportedStacksOutput struct {
	Stacks          []StackInfo `json:"stacks"`
	ServerLanguages []string    `json:"server_languages"`
}

// GetStackInfoMap returns detailed information for each stack
func GetStackInfoMap() map[string]StackInfo {
	return map[string]StackInfo{
		StackWebStandard: {
			ID:          StackWebStandard,
			Name:        "Web Standard Checkout",
			Description: "Integrate Razorpay Standard Checkout on web applications using JavaScript",
			DocsURL:     "https://razorpay.com/docs/payments/payment-gateway/web-integration/standard/",
			Category:    "web",
		},
		StackAndroidStandard: {
			ID:          StackAndroidStandard,
			Name:        "Android Standard SDK",
			Description: "Integrate Razorpay Standard Checkout in native Android applications",
			DocsURL:     "https://razorpay.com/docs/payments/payment-gateway/android-integration/standard/",
			Category:    "mobile",
		},
		StackIOSStandard: {
			ID:          StackIOSStandard,
			Name:        "iOS Standard SDK",
			Description: "Integrate Razorpay Standard Checkout in native iOS applications (Swift/Objective-C)",
			DocsURL:     "https://razorpay.com/docs/payments/payment-gateway/ios-integration/standard/",
			Category:    "mobile",
		},
		StackReactNativeStandard: {
			ID:          StackReactNativeStandard,
			Name:        "React Native Standard SDK",
			Description: "Integrate Razorpay Standard Checkout in React Native applications",
			DocsURL:     "https://razorpay.com/docs/payments/payment-gateway/react-native-integration/standard/",
			Category:    "hybrid",
		},
		StackFlutterStandard: {
			ID:          StackFlutterStandard,
			Name:        "Flutter Standard SDK",
			Description: "Integrate Razorpay Standard Checkout in Flutter applications",
			DocsURL:     "https://razorpay.com/docs/payments/payment-gateway/flutter-integration/standard/",
			Category:    "hybrid",
		},
		StackCordovaStandard: {
			ID:          StackCordovaStandard,
			Name:        "Cordova Standard Plugin",
			Description: "Integrate Razorpay Standard Checkout in Apache Cordova applications",
			DocsURL:     "https://razorpay.com/docs/payments/payment-gateway/cordova-integration/",
			Category:    "hybrid",
		},
		StackIonicStandard: {
			ID:          StackIonicStandard,
			Name:        "Ionic Standard Plugin",
			Description: "Integrate Razorpay Standard Checkout in Ionic applications (uses Cordova plugin)",
			DocsURL:     "https://razorpay.com/docs/payments/payment-gateway/cordova-integration/",
			Category:    "hybrid",
		},
		StackServerSDK: {
			ID:          StackServerSDK,
			Name:        "Server SDK Integration",
			Description: "Server-side integration for creating orders, verifying payments, and webhooks",
			DocsURL:     "https://razorpay.com/docs/payments/server-integration/",
			Category:    "server",
		},
	}
}

// IsValidStack checks if a stack identifier is valid
func IsValidStack(stack string) bool {
	for _, s := range SupportedStacks {
		if s == stack {
			return true
		}
	}
	return false
}

// IsValidServerLanguage checks if a server language is valid
func IsValidServerLanguage(lang string) bool {
	for _, l := range SupportedServerLanguages {
		if l == lang {
			return true
		}
	}
	return false
}

// IsValidSnippetKind checks if a snippet kind is valid
func IsValidSnippetKind(kind string) bool {
	for _, k := range SupportedSnippetKinds {
		if k == kind {
			return true
		}
	}
	return false
}

