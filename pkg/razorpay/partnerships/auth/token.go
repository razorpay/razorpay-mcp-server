package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

const tokenURL = "https://auth.razorpay.com/token"

// GenerateAccessToken returns a tool that generates an OAuth access token
// for the Razorpay Partnerships Onboarding API using the client_credentials grant.
// The returned access_token is used as a Bearer token in subsequent onboarding API calls.
func GenerateAccessToken(
	obs *observability.Observability,
	_ *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"client_id",
			mcpgo.Description("OAuth application client ID provided by Razorpay "+
				"when the cobranded_onboarding feature is enabled."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"client_secret",
			mcpgo.Description("OAuth application client secret provided by Razorpay."),
			mcpgo.Required(),
		),
		mcpgo.WithString(
			"mode",
			mcpgo.Description("API mode. Accepted values: test, live. Defaults to live."),
			mcpgo.Enum("test", "live"),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		params := make(map[string]interface{})

		v := newValidator(&r).
			requireString(params, "client_id").
			requireString(params, "client_secret").
			optionalString(params, "mode")

		if result, err := v.handleErrorsIfAny(); result != nil {
			return result, err
		}

		body := map[string]string{
			"grant_type":    "client_credentials",
			"client_id":     params["client_id"].(string),
			"client_secret": params["client_secret"].(string),
		}
		if mode, ok := params["mode"].(string); ok {
			body["mode"] = mode
		}

		payload, err := json.Marshal(body)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("marshalling token request failed", err),
			), nil
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(payload))
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("creating token request failed", err),
			), nil
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("token request failed", err),
			), nil
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("reading token response failed", err),
			), nil
		}

		if resp.StatusCode != http.StatusOK {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("token request failed with status %d: %s", resp.StatusCode, string(respBody)),
			), nil
		}

		var result map[string]interface{}
		if err := json.Unmarshal(respBody, &result); err != nil {
			return mcpgo.NewToolResultError(
				formatErrorMessage("parsing token response failed", err),
			), nil
		}

		return mcpgo.NewToolResultJSON(result)
	}

	return mcpgo.NewTool(
		"generate_access_token",
		"Generate an OAuth access token for the Razorpay Partnerships Onboarding API "+
			"using the client_credentials grant (POST https://auth.razorpay.com/token). "+
			"\n\nPrerequisite: the cobranded_onboarding feature must be enabled on your Razorpay account. "+
			"\n\nReturns access_token (use as 'Authorization: Bearer <token>' in onboarding API calls), "+
			"token_type (Bearer), expires_in (TTL in seconds), and razorpay_account_id. "+
			"\n\nTokens expire and must be regenerated. Store expires_in to know when to refresh.",
		parameters,
		handler,
	)
}
