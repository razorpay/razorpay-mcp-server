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
			mcpgo.Description("Frontend framework: vanilla, react, nextjs, vue, angular, svelte, solid, or native (for mobile apps). "+
				"Use the frontend value from detect_stack (same vocabulary)."),
			mcpgo.Required(),
			mcpgo.Enum("vanilla", "react", "nextjs", "vue", "angular", "svelte", "solid", "native"),
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

		var output IntegrateCheckoutOutput

		frontendCode := getFrontendIntegration(frontendFramework)

		switch backendFramework {
		case "django":
			output = getDjangoIntegration(frontendCode)
		case "flask":
			output = getFlaskIntegration(frontendCode)
		case "fastapi":
			output = getFastAPIIntegration(frontendCode)
		case "gin":
			output = getGinIntegration(frontendCode)
		case "echo":
			output = getEchoIntegration(frontendCode)
		case "fiber":
			output = getFiberIntegration(frontendCode)
		case "nextjs":
			output = getNextjsReactIntegration(language)
		case "nuxt":
			output = getNuxtIntegration(language)
		case "fastify":
			output = getFastifyIntegration(language, frontendCode)
		case "koa":
			output = getKoaIntegration(language, frontendCode)
		case "spring":
			output = getSpringIntegration(frontendCode)
		case "spring-boot":
			output = getSpringBootIntegration(frontendCode)
		case "laravel":
			output = getLaravelIntegration(frontendCode)
		case "rails":
			output = getRailsIntegration(frontendCode)
		case "actix":
			output = getActixIntegration(frontendCode)
		case "aspnet":
			output = getAspNetIntegration(frontendCode)
		case "react-native":
			output = getReactNativeIntegration()
		case "flutter":
			output = getFlutterIntegration()
		case "android":
			output = getAndroidIntegration()
		case "ios":
			output = getIOSIntegration()
		case "cordova":
			output = getCordovaIntegration()
		case "ionic":
			output = getIonicIntegration()
		case "capacitor":
			output = getCapacitorIntegration()
		default:
			output = getExpressVanillaIntegration(language, frontendCode)
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
			"Map detect_stack output: language→language, framework→backendFramework, frontend→frontendFramework. "+
			"The AI should apply ALL returned files and modifications without asking the user for additional steps.",
		parameters,
		handler,
	)
}
