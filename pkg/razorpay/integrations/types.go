//nolint:lll // File contains embedded code templates requiring longer lines
package integrations

// Credentials holds Razorpay API credentials
type Credentials struct {
	KeyID     string
	KeySecret string
}

// FileAction represents an action to perform on a file
type FileAction struct {
	Action          string     `json:"action"`
	Path            string     `json:"path"`
	Code            string     `json:"code,omitempty"`
	Description     string     `json:"description"`
	Edits           []EditItem `json:"edits,omitempty"`
	FunctionName    string     `json:"functionName,omitempty"`
	FindCode        string     `json:"findCode,omitempty"`
	ReplaceWithCode string     `json:"replaceWithCode,omitempty"`
}

// EditItem represents a manual edit instruction
type EditItem struct {
	Line string `json:"line"`
	Add  string `json:"add"`
	Why  string `json:"why"`
}

// Dependency represents a dependency to install
type Dependency struct {
	Name           string `json:"name"`
	InstallCommand string `json:"installCommand"`
}

// EnvVar represents an environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// IntegrateCheckoutOutput is the response from integrate_razorpay_checkout
type IntegrateCheckoutOutput struct {
	Summary          string       `json:"summary"`
	Files            []FileAction `json:"files"`
	Dependencies     []Dependency `json:"dependencies"`
	EnvVars          []EnvVar     `json:"envVars"`
	TestInstructions string       `json:"testInstructions"`
	AIInstructions   string       `json:"aiInstructions"`
}

// DetectStackOutput is the response from detect_stack
type DetectStackOutput struct {
	Language       string   `json:"language"`
	Framework      string   `json:"framework"`
	Frontend       string   `json:"frontend,omitempty"`
	PackageManager string   `json:"packageManager"`
	IsFullStack    bool     `json:"isFullStack"`
	Confidence     float64  `json:"confidence"`
	Notes          []string `json:"notes"`
}

// FrontendIntegration holds frontend code for different frameworks
type FrontendIntegration struct {
	Framework   string
	Code        string
	FileName    string
	ScriptTag   string
	Description string
}
