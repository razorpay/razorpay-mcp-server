package contextkey

import (
	"context"
)

// contextKey is a type used for context value keys to avoid key collisions.
type contextKey string

// Context keys for storing various values.
const (
	clientKey contextKey = "client"
)

// WithClient returns a new context with the client instance attached.
func WithClient(ctx context.Context, client interface{}) context.Context {
	return context.WithValue(ctx, clientKey, client)
}

// ClientFromContext extracts the client instance from the context.
// Returns nil if no client is found.
func ClientFromContext(ctx context.Context) interface{} {
	return ctx.Value(clientKey)
}
