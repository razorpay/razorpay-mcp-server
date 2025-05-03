// Package mcpgo provides Model Context Protocol (MCP) server implementations.
package mcpgo

import (
	"context"
)

// contextKey is a type used for context value keys to avoid key collisions.
type contextKey string

// Context keys for storing various values.
const (
	authTokenKey contextKey = "auth_token"
)

// WithAuthToken returns a new context with the authentication token attached.
func WithAuthToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, authTokenKey, token)
}

// AuthTokenFromContext extracts the authentication token from the context.
// Returns an empty string if no token is found or if the value is not a string.
func AuthTokenFromContext(ctx context.Context) string {
	value := ctx.Value(authTokenKey)
	if value == nil {
		return ""
	}

	token, ok := value.(string)
	if !ok {
		return ""
	}

	return token
}
