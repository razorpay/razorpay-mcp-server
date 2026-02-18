//nolint:lll // File contains embedded code templates requiring longer lines
package integrations

import (
	"context"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// IntegrateRazorpayCheckout returns a tool for complete Razorpay checkout integration
//
//nolint:gocyclo // Complex integration logic with multiple code paths
func IntegrateRazorpayCheckout(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"language",
			mcpgo.Description("Programming language: javascript, typescript, python, go, java, php, ruby, rust, csharp, dart, swift, or kotlin"),
			mcpgo.Required(),
			mcpgo.Enum("javascript", "typescript", "python", "go", "java", "php", "ruby", "rust", "csharp", "dart", "swift", "kotlin"),
		),
		mcpgo.WithString(
			"backendFramework",
			mcpgo.Description("Backend framework: express, fastify, koa, nextjs, nuxt, django, flask, fastapi, gin, echo, fiber, spring, spring-boot, laravel, rails, actix, aspnet, react-native, flutter, android, ios, cordova, ionic, or capacitor"),
			mcpgo.Required(),
			mcpgo.Enum("express", "fastify", "koa", "nextjs", "nuxt", "django", "flask", "fastapi", "gin", "echo", "fiber", "spring", "spring-boot", "laravel", "rails", "actix", "aspnet", "react-native", "flutter", "android", "ios", "cordova", "ionic", "capacitor"),
		),
		mcpgo.WithString(
			"frontendFramework",
			mcpgo.Description("Frontend framework: vanilla, react, nextjs, vue, angular, svelte, or native (for mobile apps)"),
			mcpgo.Required(),
			mcpgo.Enum("vanilla", "react", "nextjs", "vue", "angular", "svelte", "native"),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		args, ok := r.Arguments.(map[string]interface{})
		if !ok {
			return mcpgo.NewToolResultError("Invalid arguments"), nil
		}

		language, _ := args["language"].(string)
		backendFramework, _ := args["backendFramework"].(string)
		frontendFramework, _ := args["frontendFramework"].(string)

		// Validate that a client is available
		if _, err := getClientFromContextOrDefault(ctx, client); err != nil {
			return mcpgo.NewToolResultError("Failed to get client: " + err.Error()), nil
		}

		// Security: never expose real credentials to the AI agent.
		// Placeholder values are always returned; the user fills in
		// real keys via NEXT_STEPS.md instructions.
		creds := Credentials{}

		var output IntegrateCheckoutOutput

		// Get frontend code based on frontend framework
		frontendCode := getFrontendIntegration(frontendFramework, creds)

		// Route to appropriate backend integration
		switch backendFramework {
		// Python frameworks
		case "django":
			output = getDjangoIntegration(creds, frontendCode)
		case "flask":
			output = getFlaskIntegration(creds, frontendCode)
		case "fastapi":
			output = getFastAPIIntegration(creds, frontendCode)
		// Go frameworks
		case "gin":
			output = getGinIntegration(creds, frontendCode)
		case "echo":
			output = getEchoIntegration(creds, frontendCode)
		case "fiber":
			output = getFiberIntegration(creds, frontendCode)
		// Node.js frameworks
		case "nextjs":
			output = getNextjsReactIntegration(language, creds)
		case "nuxt":
			output = getNuxtIntegration(language, creds)
		case "fastify":
			output = getFastifyIntegration(language, creds, frontendCode)
		case "koa":
			output = getKoaIntegration(language, creds, frontendCode)
		// Java
		case "spring":
			output = getSpringIntegration(creds, frontendCode)
		case "spring-boot":
			output = getSpringBootIntegration(creds, frontendCode)
		// PHP
		case "laravel":
			output = getLaravelIntegration(creds, frontendCode)
		// Ruby
		case "rails":
			output = getRailsIntegration(creds, frontendCode)
		// Rust
		case "actix":
			output = getActixIntegration(creds, frontendCode)
		// .NET
		case "aspnet":
			output = getAspNetIntegration(creds, frontendCode)
		// Mobile
		case "react-native":
			output = getReactNativeIntegration(creds)
		case "flutter":
			output = getFlutterIntegration(creds)
		case "android":
			output = getAndroidIntegration(creds)
		case "ios":
			output = getIOSIntegration(creds)
		case "cordova":
			output = getCordovaIntegration(creds)
		case "ionic":
			output = getIonicIntegration(creds)
		case "capacitor":
			output = getCapacitorIntegration(creds)
		default: // express
			output = getExpressVanillaIntegration(language, creds, frontendCode)
		}

		output.Files = append(output.Files, getNextStepsFile())
		output.AIInstructions += getNextStepsAIInstructions()

		return mcpgo.NewToolResultJSON(output)
	}

	return mcpgo.NewTool(
		"integrate_razorpay_checkout",
		"Complete Razorpay Standard Checkout integration. Returns ALL code needed - "+
			"backend routes, frontend integration, and payment verification. "+
			"IMPORTANT: Before calling this tool, ALWAYS call detect_stack first to determine the "+
			"project's language, backendFramework, and frontendFramework. Do NOT ask the user for these values. "+
			"The AI should apply ALL returned files and modifications without asking the user for additional steps.",
		parameters,
		handler,
	)
}
