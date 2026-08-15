//nolint:lll // File contains embedded code templates requiring longer lines
package integrations

import (
	"context"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// DetectStack returns a tool for detecting project technology stack
func DetectStack(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithArray(
			"files",
			mcpgo.Description("List of file paths in the project"),
			mcpgo.Required(),
			mcpgo.Items(map[string]interface{}{"type": "string"}),
		),
		mcpgo.WithObject(
			"packageJson",
			mcpgo.Description("Contents of package.json if it exists (Node.js)"),
		),
		mcpgo.WithString(
			"requirementsTxt",
			mcpgo.Description("Contents of requirements.txt if it exists (Python)"),
		),
		mcpgo.WithString(
			"goMod",
			mcpgo.Description("Contents of go.mod if it exists (Go)"),
		),
		mcpgo.WithString(
			"pubspecYaml",
			mcpgo.Description("Contents of pubspec.yaml if it exists (Flutter)"),
		),
		mcpgo.WithString(
			"composerJson",
			mcpgo.Description("Contents of composer.json if it exists (PHP)"),
		),
		mcpgo.WithString(
			"gemfile",
			mcpgo.Description("Contents of Gemfile if it exists (Ruby)"),
		),
		mcpgo.WithString(
			"cargoToml",
			mcpgo.Description("Contents of Cargo.toml if it exists (Rust)"),
		),
		mcpgo.WithString(
			"pomXml",
			mcpgo.Description("Contents of pom.xml if it exists (Java/Maven)"),
		),
		mcpgo.WithString(
			"csproj",
			mcpgo.Description("Contents of .csproj if it exists (.NET)"),
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

		output := detectProjectStack(args)
		return mcpgo.NewToolResultJSON(output)
	}

	return mcpgo.NewTool(
		"detect_stack",
		"Detect the technology stack of a project based on file information. "+
			"Returns language, framework, frontend, and package manager. "+
			"IMPORTANT: Always call this tool FIRST before calling integrate_razorpay_checkout. "+
			"Before calling this tool, you MUST: "+
			"1) List the project's files and pass them in the 'files' parameter, "+
			"2) Read the relevant dependency file (package.json for Node.js, requirements.txt for Python, "+
			"go.mod for Go, pubspec.yaml for Flutter, Cargo.toml for Rust, pom.xml for Java, etc.) "+
			"and pass its contents in the corresponding parameter. "+
			"Then pass language, framework as backendFramework, and frontend as frontendFramework "+
			"to integrate_razorpay_checkout. Frontend values match that tool's enum "+
			"(vanilla, react, nextjs, vue, angular, svelte, solid, native); "+
			"empty or backend-only projects default to vanilla.",
		parameters,
		handler,
	)
}

// =============================================================================
// DETECT STACK HELPER
// =============================================================================

//nolint:gocyclo // Complex detection logic requires multiple conditions
func detectProjectStack(args map[string]interface{}) DetectStackOutput {
	files := []string{}
	if f, ok := args["files"].([]interface{}); ok {
		for _, v := range f {
			if s, ok := v.(string); ok {
				files = append(files, s)
			}
		}
	}

	packageJsonRaw, hasPackageJson := args["packageJson"].(map[string]interface{})
	requirementsTxt, _ := args["requirementsTxt"].(string)
	goMod, _ := args["goMod"].(string)
	pubspecYaml, _ := args["pubspecYaml"].(string)

	notes := []string{}

	// Flutter detection
	if pubspecYaml != "" || containsSuffix(files, "pubspec.yaml") {
		return withNormalizedFrontend(DetectStackOutput{
			Language:       "dart",
			Framework:      "flutter",
			PackageManager: "pub",
			IsFullStack:    false,
			Confidence:     0.95,
			Notes:          []string{"Flutter mobile app detected"},
		})
	}

	// Go detection
	if goMod != "" || containsSuffix(files, "go.mod") {
		framework := "gin" // default to gin for Go web
		//nolint:gocritic // if-else chain is clearer here
		if contains(goMod, "github.com/labstack/echo") {
			framework = "echo"
		} else if contains(goMod, "github.com/gofiber/fiber") {
			framework = "fiber"
		}

		return withNormalizedFrontend(DetectStackOutput{
			Language:       "go",
			Framework:      framework,
			PackageManager: "go-mod",
			IsFullStack:    true,
			Confidence:     0.9,
			Notes:          []string{"Go project with " + framework},
		})
	}

	// Rust detection
	if containsSuffix(files, "Cargo.toml") {
		framework := "actix" // default for Rust web
		return withNormalizedFrontend(DetectStackOutput{
			Language:       "rust",
			Framework:      framework,
			PackageManager: "cargo",
			IsFullStack:    true,
			Confidence:     0.9,
			Notes:          []string{"Rust project detected"},
		})
	}

	// Java/Spring detection
	if containsSuffix(files, "pom.xml") || containsSuffix(files, "build.gradle") {
		return withNormalizedFrontend(DetectStackOutput{
			Language:       "java",
			Framework:      "spring",
			PackageManager: "maven",
			IsFullStack:    true,
			Confidence:     0.85,
			Notes:          []string{"Java project detected - assuming Spring Boot"},
		})
	}

	// PHP/Laravel detection
	if containsSuffix(files, "composer.json") {
		framework := "laravel" // default for PHP
		if containsPath(files, "artisan") {
			framework = "laravel"
		}
		return withNormalizedFrontend(DetectStackOutput{
			Language:       "php",
			Framework:      framework,
			PackageManager: "composer",
			IsFullStack:    true,
			Confidence:     0.85,
			Notes:          []string{"PHP project detected"},
		})
	}

	// Ruby/Rails detection
	if containsSuffix(files, "Gemfile") {
		framework := "rails" // default for Ruby web
		if containsPath(files, "config/routes.rb") {
			framework = "rails"
		}
		return withNormalizedFrontend(DetectStackOutput{
			Language:       "ruby",
			Framework:      framework,
			PackageManager: "bundler",
			IsFullStack:    true,
			Confidence:     0.85,
			Notes:          []string{"Ruby project detected"},
		})
	}

	// .NET detection
	if containsSuffix(files, ".csproj") || containsSuffix(files, ".sln") {
		return withNormalizedFrontend(DetectStackOutput{
			Language:       "csharp",
			Framework:      "aspnet",
			PackageManager: "nuget",
			IsFullStack:    true,
			Confidence:     0.85,
			Notes:          []string{".NET project detected - assuming ASP.NET Core"},
		})
	}

	// Python detection
	if requirementsTxt != "" || containsSuffix(files, "requirements.txt") || containsSuffix(files, "pyproject.toml") {
		framework := "flask" // default
		pythonFrameworks := []string{"django", "flask", "fastapi", "starlette"}
		for _, fw := range pythonFrameworks {
			if contains(requirementsTxt, fw) {
				framework = fw
				break
			}
		}

		if containsPath(files, "manage.py") {
			framework = "django"
		}
		if containsSuffix(files, "app.py") && framework == "flask" {
			framework = "flask"
		}

		return withNormalizedFrontend(DetectStackOutput{
			Language:       "python",
			Framework:      framework,
			PackageManager: "pip",
			IsFullStack:    true,
			Confidence:     0.85,
			Notes:          []string{"Python project with " + framework},
		})
	}

	// Node.js detection
	if hasPackageJson || containsSuffix(files, "package.json") {
		deps := map[string]bool{}
		if packageJsonRaw != nil {
			if d, ok := packageJsonRaw["dependencies"].(map[string]interface{}); ok {
				for k := range d {
					deps[k] = true
				}
			}
			if d, ok := packageJsonRaw["devDependencies"].(map[string]interface{}); ok {
				for k := range d {
					deps[k] = true
				}
			}
		}

		isTypeScript := containsSuffix(files, "tsconfig.json") || deps["typescript"]
		language := "javascript"
		if isTypeScript {
			language = "typescript"
		}

		// Detect package manager
		packageManager := "npm"
		//nolint:gocritic // if-else chain is clearer here
		if containsPath(files, "yarn.lock") {
			packageManager = "yarn"
		} else if containsPath(files, "pnpm-lock.yaml") {
			packageManager = "pnpm"
		} else if containsPath(files, "bun.lockb") {
			packageManager = "bun"
		}

		// Detect backend framework
		framework := "express" // default for Node
		nodeFrameworks := map[string]string{
			"next":         "nextjs",
			"express":      "express",
			"fastify":      "fastify",
			"koa":          "koa",
			"hono":         "hono",
			"nuxt":         "nuxt",
			"@nestjs/core": "nestjs",
		}
		for pkg, fw := range nodeFrameworks {
			if deps[pkg] {
				framework = fw
				notes = append(notes, "Found "+pkg+" in dependencies")
				break
			}
		}

		// Detect frontend framework
		frontend := ""
		frontendFrameworks := map[string]string{
			"react":         "react",
			"vue":           "vue",
			"@angular/core": "angular",
			"svelte":        "svelte",
			"solid-js":      "solid",
			"react-native":  "react-native",
			"expo":          "react-native",
		}
		for pkg, fw := range frontendFrameworks {
			if deps[pkg] {
				frontend = fw
				notes = append(notes, "Found "+pkg+" for frontend")
				break
			}
		}

		// React Native special case
		if frontend == "react-native" {
			return withNormalizedFrontend(DetectStackOutput{
				Language:       language,
				Framework:      "react-native",
				PackageManager: packageManager,
				IsFullStack:    false,
				Confidence:     0.95,
				Notes:          []string{"React Native mobile app detected"},
			})
		}

		// Determine if fullstack
		isFullStack := framework == "nextjs" || framework == "nuxt" || framework == "nestjs" ||
			(framework != "node" && frontend == "")

		return withNormalizedFrontend(DetectStackOutput{
			Language:       language,
			Framework:      framework,
			Frontend:       frontend,
			PackageManager: packageManager,
			IsFullStack:    isFullStack,
			Confidence:     0.9,
			Notes:          notes,
		})
	}

	// Default fallback
	return DetectStackOutput{
		Language:       "unknown",
		Framework:      "unknown",
		PackageManager: "unknown",
		IsFullStack:    false,
		Confidence:     0.1,
		Notes:          []string{"Could not detect project stack"},
	}
}
