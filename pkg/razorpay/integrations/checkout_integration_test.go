package integrations

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/contextkey"
	"github.com/razorpay/razorpay-mcp-server/pkg/log"
	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// newMockRzpClient creates a mock Razorpay client for testing
func newMockRzpClient() *rzpsdk.Client {
	return rzpsdk.NewClient("test_key", "test_secret")
}

// createTestObservability creates an observability stack suitable for testing
func createTestObservability() *observability.Observability {
	_, logger := log.New(context.Background(), log.NewConfig(
		log.WithMode(log.ModeStdio)),
	)
	return &observability.Observability{
		Logger: logger,
	}
}

// createMCPRequest creates a CallToolRequest with the given arguments
func createMCPRequest(args any) mcpgo.CallToolRequest {
	return mcpgo.CallToolRequest{
		Arguments: args,
	}
}

func Test_DetectStack(t *testing.T) {
	tests := []struct {
		name           string
		request        map[string]interface{}
		expectError    bool
		expectedErrMsg string
		validate       func(*testing.T, DetectStackOutput)
	}{
		{
			name: "detect Flutter project",
			request: map[string]interface{}{
				"files": []interface{}{"lib/main.dart", "pubspec.yaml"},
			},
			expectError: false,
			validate: func(t *testing.T, result DetectStackOutput) {
				assert.Equal(t, "dart", result.Language)
				assert.Equal(t, "flutter", result.Framework)
				assert.Equal(t, "pub", result.PackageManager)
			},
		},
		{
			name: "detect Go project with gin",
			request: map[string]interface{}{
				"files": []interface{}{"main.go", "go.mod"},
				"goMod": "module myapp\nrequire github.com/gin-gonic/gin v1.9.0",
			},
			expectError: false,
			validate: func(t *testing.T, result DetectStackOutput) {
				assert.Equal(t, "go", result.Language)
				assert.Equal(t, "gin", result.Framework)
				assert.Equal(t, "go-mod", result.PackageManager)
				assert.Equal(t, "vanilla", result.Frontend)
			},
		},
		{
			name: "detect Go project with echo",
			request: map[string]interface{}{
				"files": []interface{}{"main.go", "go.mod"},
				"goMod": "module myapp\nrequire github.com/labstack/echo v4.0.0",
			},
			expectError: false,
			validate: func(t *testing.T, result DetectStackOutput) {
				assert.Equal(t, "go", result.Language)
				assert.Equal(t, "echo", result.Framework)
			},
		},
		{
			name: "detect Node.js Express project",
			request: map[string]interface{}{
				"files": []interface{}{"index.js", "package.json"},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"express": "^4.18.0",
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, result DetectStackOutput) {
				assert.Equal(t, "javascript", result.Language)
				assert.Equal(t, "express", result.Framework)
				assert.Equal(t, "npm", result.PackageManager)
				assert.Equal(t, "vanilla", result.Frontend)
			},
		},
		{
			name: "detect Solid.js project",
			request: map[string]interface{}{
				"files": []interface{}{"src/App.tsx", "package.json"},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"solid-js": "^1.8.0",
						"express":  "^4.18.0",
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, result DetectStackOutput) {
				assert.Equal(t, "solid", result.Frontend)
				assert.Equal(t, "express", result.Framework)
			},
		},
		{
			name: "detect TypeScript Next.js project",
			request: map[string]interface{}{
				"files": []interface{}{"pages/index.tsx", "package.json", "tsconfig.json"},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"next":  "^14.0.0",
						"react": "^18.0.0",
					},
					"devDependencies": map[string]interface{}{
						"typescript": "^5.0.0",
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, result DetectStackOutput) {
				assert.Equal(t, "typescript", result.Language)
				assert.Equal(t, "nextjs", result.Framework)
				assert.Equal(t, "react", result.Frontend)
			},
		},
		{
			name: "detect Python Django project",
			request: map[string]interface{}{
				"files":           []interface{}{"manage.py", "requirements.txt"},
				"requirementsTxt": "Django==4.2\ndjango-rest-framework==3.14",
			},
			expectError: false,
			validate: func(t *testing.T, result DetectStackOutput) {
				assert.Equal(t, "python", result.Language)
				assert.Equal(t, "django", result.Framework)
				assert.Equal(t, "pip", result.PackageManager)
			},
		},
		{
			name: "detect Python Flask project",
			request: map[string]interface{}{
				"files":           []interface{}{"app.py", "requirements.txt"},
				"requirementsTxt": "flask==3.0\ngunicorn==21.0",
			},
			expectError: false,
			validate: func(t *testing.T, result DetectStackOutput) {
				assert.Equal(t, "python", result.Language)
				assert.Equal(t, "flask", result.Framework)
			},
		},
		{
			name: "detect Rust project",
			request: map[string]interface{}{
				"files": []interface{}{"src/main.rs", "Cargo.toml"},
			},
			expectError: false,
			validate: func(t *testing.T, result DetectStackOutput) {
				assert.Equal(t, "rust", result.Language)
				assert.Equal(t, "actix", result.Framework)
				assert.Equal(t, "cargo", result.PackageManager)
			},
		},
		{
			name: "detect Java Spring project",
			request: map[string]interface{}{
				"files": []interface{}{"src/main/java/App.java", "pom.xml"},
			},
			expectError: false,
			validate: func(t *testing.T, result DetectStackOutput) {
				assert.Equal(t, "java", result.Language)
				assert.Equal(t, "spring", result.Framework)
				assert.Equal(t, "maven", result.PackageManager)
			},
		},
		{
			name: "detect yarn package manager",
			request: map[string]interface{}{
				"files": []interface{}{"index.js", "package.json", "yarn.lock"},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"express": "^4.18.0",
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, result DetectStackOutput) {
				assert.Equal(t, "yarn", result.PackageManager)
			},
		},
		{
			name:        "empty input returns unknown",
			request:     map[string]interface{}{},
			expectError: false,
			validate: func(t *testing.T, result DetectStackOutput) {
				assert.Equal(t, "unknown", result.Language)
				assert.Equal(t, "unknown", result.Framework)
			},
		},
	}

	obs := createTestObservability()
	tool := DetectStack(obs, nil)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := createMCPRequest(tc.request)
			result, err := tool.GetHandler()(context.Background(), request)

			require.NoError(t, err)
			require.NotNil(t, result)

			if tc.expectError {
				assert.True(t, result.IsError, "expected error result")
				return
			}

			var output DetectStackOutput
			err = json.Unmarshal([]byte(result.Text), &output)
			require.NoError(t, err)

			if tc.validate != nil {
				tc.validate(t, output)
			}
		})
	}
}

//nolint:gocyclo // table-driven test with multiple validate funcs
func Test_IntegrateRazorpayCheckout(t *testing.T) {
	tests := []struct {
		name           string
		request        map[string]interface{}
		expectError    bool
		expectedErrMsg string
		validate       func(*testing.T, IntegrateCheckoutOutput)
	}{
		{
			name: "integrate Express with vanilla JS",
			request: map[string]interface{}{
				"language":          "javascript",
				"backendFramework":  "express",
				"frontendFramework": "vanilla",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
				assert.NotEmpty(t, result.Dependencies)
				assert.NotEmpty(t, result.EnvVars)

				for _, env := range result.EnvVars {
					if env.Name == "RAZORPAY_KEY_ID" {
						assert.Equal(t, "rzp_test_YOUR_KEY_ID", env.Value,
							"key ID must be a placeholder, never a real credential")
					}
					if env.Name == "RAZORPAY_KEY_SECRET" {
						assert.NotContains(t, env.Value, "test_secret",
							"key secret must not contain real credentials")
					}
				}

				hasNextSteps := false
				for _, f := range result.Files {
					if f.Path == "NEXT_STEPS.md" {
						hasNextSteps = true
						break
					}
				}
				assert.True(t, hasNextSteps, "should include NEXT_STEPS.md file")
			},
		},
		{
			name: "integrate Express with Solid.js",
			request: map[string]interface{}{
				"language":          "javascript",
				"backendFramework":  "express",
				"frontendFramework": "solid",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Django with React",
			request: map[string]interface{}{
				"language":          "python",
				"backendFramework":  "django",
				"frontendFramework": "react",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
				// Should have razorpay dependency
				hasRazorpay := false
				for _, dep := range result.Dependencies {
					if dep.Name == "razorpay" {
						hasRazorpay = true
						break
					}
				}
				assert.True(t, hasRazorpay, "should have razorpay dependency")
			},
		},
		{
			name: "integrate Go Gin",
			request: map[string]interface{}{
				"language":          "go",
				"backendFramework":  "gin",
				"frontendFramework": "vanilla",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Next.js fullstack",
			request: map[string]interface{}{
				"language":          "typescript",
				"backendFramework":  "nextjs",
				"frontendFramework": "nextjs",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Flutter mobile",
			request: map[string]interface{}{
				"language":          "dart",
				"backendFramework":  "flutter",
				"frontendFramework": "native",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Spring Boot with Vue",
			request: map[string]interface{}{
				"language":          "java",
				"backendFramework":  "spring-boot",
				"frontendFramework": "vue",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Laravel with Angular",
			request: map[string]interface{}{
				"language":          "php",
				"backendFramework":  "laravel",
				"frontendFramework": "angular",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Rails with Svelte",
			request: map[string]interface{}{
				"language":          "ruby",
				"backendFramework":  "rails",
				"frontendFramework": "svelte",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		// Additional framework tests for coverage
		{
			name: "integrate Fastify with vanilla",
			request: map[string]interface{}{
				"language":          "javascript",
				"backendFramework":  "fastify",
				"frontendFramework": "vanilla",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Koa with React",
			request: map[string]interface{}{
				"language":          "javascript",
				"backendFramework":  "koa",
				"frontendFramework": "react",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate FastAPI with React",
			request: map[string]interface{}{
				"language":          "python",
				"backendFramework":  "fastapi",
				"frontendFramework": "react",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Echo with vanilla",
			request: map[string]interface{}{
				"language":          "go",
				"backendFramework":  "echo",
				"frontendFramework": "vanilla",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Fiber with vanilla",
			request: map[string]interface{}{
				"language":          "go",
				"backendFramework":  "fiber",
				"frontendFramework": "vanilla",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Spring with vanilla",
			request: map[string]interface{}{
				"language":          "java",
				"backendFramework":  "spring",
				"frontendFramework": "vanilla",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Actix with vanilla",
			request: map[string]interface{}{
				"language":          "rust",
				"backendFramework":  "actix",
				"frontendFramework": "vanilla",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate ASP.NET with vanilla",
			request: map[string]interface{}{
				"language":          "csharp",
				"backendFramework":  "aspnet",
				"frontendFramework": "vanilla",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Nuxt with Vue",
			request: map[string]interface{}{
				"language":          "typescript",
				"backendFramework":  "nuxt",
				"frontendFramework": "vue",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate React Native",
			request: map[string]interface{}{
				"language":          "javascript",
				"backendFramework":  "react-native",
				"frontendFramework": "native",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Android native",
			request: map[string]interface{}{
				"language":          "kotlin",
				"backendFramework":  "android",
				"frontendFramework": "native",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate iOS native",
			request: map[string]interface{}{
				"language":          "swift",
				"backendFramework":  "ios",
				"frontendFramework": "native",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Cordova",
			request: map[string]interface{}{
				"language":          "javascript",
				"backendFramework":  "cordova",
				"frontendFramework": "native",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Ionic",
			request: map[string]interface{}{
				"language":          "typescript",
				"backendFramework":  "ionic",
				"frontendFramework": "native",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Capacitor",
			request: map[string]interface{}{
				"language":          "typescript",
				"backendFramework":  "capacitor",
				"frontendFramework": "native",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
			},
		},
		{
			name: "integrate Flask with vanilla",
			request: map[string]interface{}{
				"language":          "python",
				"backendFramework":  "flask",
				"frontendFramework": "vanilla",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				assert.NotEmpty(t, result.Files)
				hasRazorpay := false
				for _, dep := range result.Dependencies {
					if dep.Name == "razorpay" {
						hasRazorpay = true
						break
					}
				}
				assert.True(t, hasRazorpay, "should have razorpay dependency")
			},
		},
		{
			name: "integrate Express TypeScript with vanilla",
			request: map[string]interface{}{
				"language":          "typescript",
				"backendFramework":  "express",
				"frontendFramework": "vanilla",
			},
			expectError: false,
			validate: func(t *testing.T, result IntegrateCheckoutOutput) {
				assert.NotEmpty(t, result.Summary)
				hasTS := false
				for _, f := range result.Files {
					if f.Path == "routes/razorpay.ts" {
						hasTS = true
						break
					}
				}
				assert.True(t, hasTS, "TypeScript Express should create .ts file")
			},
		},
	}

	obs := createTestObservability()
	client := newMockRzpClient()
	tool := IntegrateRazorpayCheckout(obs, client)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := createMCPRequest(tc.request)
			result, err := tool.GetHandler()(context.Background(), request)

			require.NoError(t, err)
			require.NotNil(t, result)

			if tc.expectError {
				isErr := result.IsError ||
					containsString(result.Text, tc.expectedErrMsg)
				assert.True(t, isErr,
					"expected error containing: %s", tc.expectedErrMsg)
				return
			}

			var output IntegrateCheckoutOutput
			err = json.Unmarshal([]byte(result.Text), &output)
			require.NoError(t, err, "failed to unmarshal result: %s", result.Text)

			if tc.validate != nil {
				tc.validate(t, output)
			}
		})
	}
}

func Test_normalizeDetectedFrontend(t *testing.T) {
	tests := []struct {
		backend  string
		frontend string
		want     string
	}{
		{"express", "", "vanilla"},
		{"gin", "", "vanilla"},
		{"nextjs", "", "nextjs"},
		{"nuxt", "", "vue"},
		{"flutter", "", "native"},
		{"react-native", "", "native"},
		{"express", "solid", "solid"},
		{"express", "react", "react"},
	}

	for _, tc := range tests {
		t.Run(tc.backend+"_"+tc.frontend, func(t *testing.T) {
			got := normalizeDetectedFrontend(tc.backend, tc.frontend)
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_detectStackToIntegrateCheckoutFrontendChain(t *testing.T) {
	// Express backend-only: detect_stack frontend must match checkout enum.
	detected := detectProjectStack(map[string]interface{}{
		"files": []interface{}{"index.js", "package.json"},
		"packageJson": map[string]interface{}{
			"dependencies": map[string]interface{}{
				"express": "^4.18.0",
			},
		},
	})
	assert.Equal(t, "vanilla", detected.Frontend)

	obs := createTestObservability()
	tool := IntegrateRazorpayCheckout(obs, newMockRzpClient())
	request := createMCPRequest(map[string]interface{}{
		"language":          detected.Language,
		"backendFramework":  detected.Framework,
		"frontendFramework": detected.Frontend,
	})
	result, err := tool.GetHandler()(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(
		t, result.IsError, "detect_stack frontend should pass checkout schema")
}

func Test_detectProjectStack(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		expected DetectStackOutput
	}{
		{
			name: "empty input returns unknown",
			args: map[string]interface{}{
				"files": []interface{}{},
			},
			expected: DetectStackOutput{
				Language:   "unknown",
				Framework:  "unknown",
				Confidence: 0.0,
			},
		},
		{
			name: "flutter detection",
			args: map[string]interface{}{
				"files":       []interface{}{"pubspec.yaml"},
				"pubspecYaml": "name: myapp\ndependencies:\n  flutter:",
			},
			expected: DetectStackOutput{
				Language:       "dart",
				Framework:      "flutter",
				PackageManager: "pub",
				IsFullStack:    false,
				Confidence:     0.95,
			},
		},
		{
			name: "go with fiber",
			args: map[string]interface{}{
				"files": []interface{}{"go.mod", "main.go"},
				"goMod": "module test\nrequire github.com/gofiber/fiber v2.0.0",
			},
			expected: DetectStackOutput{
				Language:       "go",
				Framework:      "fiber",
				PackageManager: "go-mod",
				IsFullStack:    true,
				Confidence:     0.9,
			},
		},
		{
			name: "go with echo",
			args: map[string]interface{}{
				"files": []interface{}{"go.mod", "main.go"},
				"goMod": "module test\nrequire github.com/labstack/echo v4.0.0",
			},
			expected: DetectStackOutput{
				Language:       "go",
				Framework:      "echo",
				PackageManager: "go-mod",
				IsFullStack:    true,
				Confidence:     0.9,
			},
		},
		{
			name: "rust detection",
			args: map[string]interface{}{
				"files": []interface{}{"Cargo.toml", "src/main.rs"},
			},
			expected: DetectStackOutput{
				Language:       "rust",
				Framework:      "actix",
				PackageManager: "cargo",
				IsFullStack:    true,
				Confidence:     0.9,
			},
		},
		{
			name: "java spring detection",
			args: map[string]interface{}{
				"files": []interface{}{"pom.xml", "src/main/java/App.java"},
			},
			expected: DetectStackOutput{
				Language:       "java",
				Framework:      "spring",
				PackageManager: "maven",
			},
		},
		{
			name: "java gradle detection",
			args: map[string]interface{}{
				"files": []interface{}{"build.gradle", "src/main/java/App.java"},
			},
			expected: DetectStackOutput{
				Language:       "java",
				Framework:      "spring",
				PackageManager: "maven",
			},
		},
		{
			name: "python fastapi detection",
			args: map[string]interface{}{
				"files":           []interface{}{"main.py", "requirements.txt"},
				"requirementsTxt": "fastapi==0.100.0\nuvicorn==0.23.0",
			},
			expected: DetectStackOutput{
				Language:       "python",
				Framework:      "fastapi",
				PackageManager: "pip",
			},
		},
		{
			name: "node express with pnpm",
			args: map[string]interface{}{
				"files": []interface{}{
					"index.js",
					"package.json",
					"pnpm-lock.yaml",
				},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"express": "^4.18.0",
					},
				},
			},
			expected: DetectStackOutput{
				Language:       "javascript",
				Framework:      "express",
				PackageManager: "pnpm",
			},
		},
		{
			name: "node with bun",
			args: map[string]interface{}{
				"files": []interface{}{
					"index.js",
					"package.json",
					"bun.lockb",
				},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"express": "^4.18.0",
					},
				},
			},
			expected: DetectStackOutput{
				Language:       "javascript",
				Framework:      "express",
				PackageManager: "bun",
			},
		},
		{
			name: "vue nuxt detection",
			args: map[string]interface{}{
				"files": []interface{}{"nuxt.config.js", "package.json"},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"nuxt": "^3.0.0",
						"vue":  "^3.0.0",
					},
				},
			},
			expected: DetectStackOutput{
				Language:  "javascript",
				Framework: "nuxt",
			},
		},
		{
			name: "react native detection",
			args: map[string]interface{}{
				"files": []interface{}{"App.js", "package.json"},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"react-native": "^0.72.0",
					},
				},
			},
			expected: DetectStackOutput{
				Language:  "javascript",
				Framework: "react-native",
			},
		},
		{
			name: "php laravel detection",
			args: map[string]interface{}{
				"files": []interface{}{"composer.json", "artisan", "app/Http/Controllers"},
			},
			expected: DetectStackOutput{
				Language:       "php",
				Framework:      "laravel",
				PackageManager: "composer",
			},
		},
		{
			name: "ruby rails detection",
			args: map[string]interface{}{
				"files": []interface{}{"Gemfile", "config/routes.rb", "app/controllers"},
			},
			expected: DetectStackOutput{
				Language:       "ruby",
				Framework:      "rails",
				PackageManager: "bundler",
			},
		},
		{
			name: "dotnet detection",
			args: map[string]interface{}{
				"files": []interface{}{"Program.cs", "MyApp.csproj"},
			},
			expected: DetectStackOutput{
				Language:       "csharp",
				Framework:      "aspnet",
				PackageManager: "nuget",
			},
		},
		{
			name: "dotnet sln detection",
			args: map[string]interface{}{
				"files": []interface{}{"MyApp.sln", "Program.cs"},
			},
			expected: DetectStackOutput{
				Language:       "csharp",
				Framework:      "aspnet",
				PackageManager: "nuget",
			},
		},
		{
			name: "python with pyproject.toml",
			args: map[string]interface{}{
				"files": []interface{}{"pyproject.toml", "src/main.py"},
			},
			expected: DetectStackOutput{
				Language:       "python",
				Framework:      "flask",
				PackageManager: "pip",
			},
		},
		{
			name: "python django via manage.py",
			args: map[string]interface{}{
				"files":           []interface{}{"requirements.txt", "manage.py"},
				"requirementsTxt": "Django==4.2\ngunicorn",
			},
			expected: DetectStackOutput{
				Language:       "python",
				Framework:      "django",
				PackageManager: "pip",
			},
		},
		{
			name: "typescript via tsconfig.json",
			args: map[string]interface{}{
				"files": []interface{}{"package.json", "tsconfig.json", "src/index.ts"},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"express": "^4.18.0",
					},
				},
			},
			expected: DetectStackOutput{
				Language:       "typescript",
				Framework:      "express",
				PackageManager: "npm",
			},
		},
		{
			name: "typescript via dependency",
			args: map[string]interface{}{
				"files": []interface{}{"package.json", "src/index.ts"},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"express": "^4.18.0",
					},
					"devDependencies": map[string]interface{}{
						"typescript": "^5.0.0",
					},
				},
			},
			expected: DetectStackOutput{
				Language:       "typescript",
				Framework:      "express",
				PackageManager: "npm",
			},
		},
		{
			name: "node with yarn",
			args: map[string]interface{}{
				"files": []interface{}{"package.json", "yarn.lock"},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"express": "^4.18.0",
					},
				},
			},
			expected: DetectStackOutput{
				Language:       "javascript",
				Framework:      "express",
				PackageManager: "yarn",
			},
		},
		{
			name: "node koa detection",
			args: map[string]interface{}{
				"files": []interface{}{"package.json"},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"koa": "^2.14.0",
					},
				},
			},
			expected: DetectStackOutput{
				Language:  "javascript",
				Framework: "koa",
			},
		},
		{
			name: "node fastify detection",
			args: map[string]interface{}{
				"files": []interface{}{"package.json"},
				"packageJson": map[string]interface{}{
					"dependencies": map[string]interface{}{
						"fastify": "^4.0.0",
					},
				},
			},
			expected: DetectStackOutput{
				Language:  "javascript",
				Framework: "fastify",
			},
		},
		{
			name: "go with gin default",
			args: map[string]interface{}{
				"files": []interface{}{"go.mod", "main.go"},
				"goMod": "module myapp\ngo 1.21\nrequire github.com/gin-gonic/gin v1.9.1",
			},
			expected: DetectStackOutput{
				Language:       "go",
				Framework:      "gin",
				PackageManager: "go-mod",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detectProjectStack(tc.args)
			assert.Equal(t, tc.expected.Language, result.Language)
			assert.Equal(t, tc.expected.Framework, result.Framework)
			if tc.expected.PackageManager != "" {
				assert.Equal(t, tc.expected.PackageManager, result.PackageManager)
			}
		})
	}
}

func Test_DetectStack_InvalidArgs(t *testing.T) {
	obs := createTestObservability()
	client := newMockRzpClient()
	tool := DetectStack(obs, client)

	request := createMCPRequest("not a map")
	result, err := tool.GetHandler()(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError, "should return error for invalid args")
}

func Test_IntegrateCheckout_InvalidArgs(t *testing.T) {
	obs := createTestObservability()
	client := newMockRzpClient()
	tool := IntegrateRazorpayCheckout(obs, client)

	request := createMCPRequest("not a map")
	result, err := tool.GetHandler()(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError, "should return error for invalid args")
}

func Test_IntegrateCheckout_NoClient(t *testing.T) {
	obs := createTestObservability()
	tool := IntegrateRazorpayCheckout(obs, nil)

	request := createMCPRequest(map[string]interface{}{
		"language":          "javascript",
		"backendFramework":  "express",
		"frontendFramework": "vanilla",
	})
	result, err := tool.GetHandler()(context.Background(), request)
	require.NoError(t, err)
	assert.True(t, result.IsError, "should return error when no client available")
}

func Test_getClientFromContextOrDefault(t *testing.T) {
	t.Run("returns default client when provided", func(t *testing.T) {
		client := newMockRzpClient()
		got, err := getClientFromContextOrDefault(context.Background(), client)
		require.NoError(t, err)
		assert.Equal(t, client, got)
	})

	t.Run("error when no client and no default", func(t *testing.T) {
		got, err := getClientFromContextOrDefault(context.Background(), nil)
		assert.Nil(t, got)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no client found")
	})

	t.Run("returns client from context", func(t *testing.T) {
		client := newMockRzpClient()
		ctx := contextkey.WithClient(context.Background(), client)
		got, err := getClientFromContextOrDefault(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, client, got)
	})

	t.Run("error on invalid client type in context", func(t *testing.T) {
		ctx := contextkey.WithClient(context.Background(), "not a client")
		got, err := getClientFromContextOrDefault(ctx, nil)
		assert.Nil(t, got)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid client type")
	})
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
