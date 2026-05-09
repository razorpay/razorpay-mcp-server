package onboardingapis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

const (
	prodAccountsURL    = "https://api.razorpay.com/v2/accounts"
	nonProdAccountsURL = "https://api-web.ext.dev.razorpay.in/v2/accounts"
)

// accountsBaseURL returns the right base URL based on APP_ENV.
func accountsBaseURL() string {
	switch strings.ToLower(os.Getenv("APP_ENV")) {
	case "devstack", "dev":
		return nonProdAccountsURL
	default:
		return prodAccountsURL
	}
}

// doAccountsRequest makes an authenticated HTTP request to the Partnerships
// Accounts API using a Bearer token.
func doAccountsRequest(
	ctx context.Context,
	method string,
	url string,
	bearerToken string,
	body map[string]interface{},
) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshalling request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Content-Type", "application/json")

	// Devstack-specific routing headers
	if env := strings.ToLower(os.Getenv("APP_ENV")); env == "devstack" || env == "dev" {
		req.Header.Set("X-Org-Id", "org_100000razorpay")
		req.Header.Set("kong-debug", "1")
		if label := os.Getenv("DEVSTACK_LABEL"); label != "" {
			req.Header.Set("rzpctx-dev-serve-user", label)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response (status %d): %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return result, nil
}

func formatErrorMessage(prefix string, err error) string {
	if err == nil {
		return prefix + ": resource does not exist"
	}
	if msg := err.Error(); msg != "" {
		return prefix + ": " + msg
	}
	return prefix + ": resource does not exist"
}

// validator provides a fluent interface for validating and collecting
// parameters from a tool request.
type validator struct {
	request *mcpgo.CallToolRequest
	errors  []error
}

func newValidator(r *mcpgo.CallToolRequest) *validator {
	return &validator{request: r}
}

func (v *validator) addError(err error) *validator {
	if err != nil {
		v.errors = append(v.errors, err)
	}
	return v
}

func (v *validator) handleErrorsIfAny() (*mcpgo.ToolResult, error) {
	if len(v.errors) == 0 {
		return nil, nil
	}
	msgs := make([]string, 0, len(v.errors))
	for _, e := range v.errors {
		msgs = append(msgs, e.Error())
	}
	return mcpgo.NewToolResultError("Validation errors:\n- " + strings.Join(msgs, "\n- ")), nil
}

func extractParam[T any](r *mcpgo.CallToolRequest, name string, required bool) (*T, error) {
	args, ok := r.Arguments.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid arguments type")
	}

	val, ok := args[name]
	if !ok || val == nil {
		if required {
			return nil, errors.New("missing required parameter: " + name)
		}
		return nil, nil
	}

	var result T
	data, err := json.Marshal(val)
	if err != nil {
		return nil, errors.New("invalid parameter type: " + name)
	}
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, errors.New("invalid parameter type: " + name)
	}
	return &result, nil
}

func (v *validator) requireString(params map[string]interface{}, name string) *validator {
	val, err := extractParam[string](v.request, name, true)
	if err != nil {
		return v.addError(err)
	}
	if val != nil {
		params[name] = *val
	}
	return v
}

func (v *validator) optionalString(params map[string]interface{}, name string) *validator {
	val, err := extractParam[string](v.request, name, false)
	if err != nil {
		return v.addError(err)
	}
	if val != nil {
		params[name] = *val
	}
	return v
}

func (v *validator) optionalMap(params map[string]interface{}, name string) *validator {
	val, err := extractParam[map[string]interface{}](v.request, name, false)
	if err != nil {
		return v.addError(err)
	}
	if val != nil {
		params[name] = *val
	}
	return v
}

// Standalone extractors used by buildAccountPayload.

func extractString(r *mcpgo.CallToolRequest, name string, required bool) (*string, error) {
	return extractParam[string](r, name, required)
}

func extractNumber(r *mcpgo.CallToolRequest, name string, required bool) (*float64, error) {
	return extractParam[float64](r, name, required)
}

func extractStringArray(r *mcpgo.CallToolRequest, name string) ([]string, error) {
	val, err := extractParam[[]interface{}](r, name, false)
	if err != nil || val == nil {
		return nil, err
	}
	result := make([]string, 0, len(*val))
	for _, item := range *val {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result, nil
}
