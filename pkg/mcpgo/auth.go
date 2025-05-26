package mcpgo

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	rzpsdk "github.com/razorpay/razorpay-go"
)

// AuthenticateRequest handles authentication for a request context.
// If client is provided, it returns the context as-is (stdio mode).
// Otherwise, it validates the auth token and creates a new client (SSE mode).
func AuthenticateRequest(
	ctx context.Context,
	client *rzpsdk.Client,
) (context.Context, error) {
	// If client is provided, this is the stdio mcp server
	if client != nil {
		return ctx, nil
	}

	// Check if auth token is provided
	auth := AuthTokenFromContext(ctx)
	if auth == "" {
		return nil, fmt.Errorf("unauthorized: no auth token provided")
	}

	// Base64 decode the auth token
	token, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return nil, fmt.Errorf("unauthorized: invalid auth token")
	}

	// Split token into key:secret
	parts := strings.Split(string(token), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unauthorized: invalid auth token")
	}

	// Create a new client with the auth credentials
	newClient := rzpsdk.NewClient(parts[0], parts[1])

	// Store the client in context
	ctx = WithClient(ctx, newClient)

	return ctx, nil
}
