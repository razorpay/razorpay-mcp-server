package contextkey

import (
	"context"

	"github.com/razorpay/goutils/passport/v4"
)

// contextKey is a type used for context value keys to avoid key collisions.
type contextKey string

// Context keys for storing various values.
const (
	authTokenKey  contextKey = "auth_token"
	clientKey     contextKey = "client"
	requestIDKey  contextKey = "request_id"
	taskIDKey     contextKey = "task_id"
	merchantIDKey contextKey = "merchant_id"
	rzpKeyKey     contextKey = "rzp_key"
	passportKey   contextKey = "passport"
	authModeKey   contextKey = "auth_mode"
)

// WithAuthToken returns a new context with the authentication token attached.
func WithAuthToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, authTokenKey, token)
}

// AuthTokenFromContext extracts the authentication token from the context.
// Returns an empty string if no token is found or if the value
// is not a string.
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

// WithClient returns a new context with the client instance attached.
func WithClient(ctx context.Context, client interface{}) context.Context {
	return context.WithValue(ctx, clientKey, client)
}

// ClientFromContext extracts the client instance from the context.
// Returns nil if no client is found.
func ClientFromContext(ctx context.Context) interface{} {
	return ctx.Value(clientKey)
}

// WithRequestID returns a new context with the request ID attached.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext extracts the request ID from the context.
// Returns an empty string if no request ID is found or if the value
// is not a string.
func RequestIDFromContext(ctx context.Context) string {
	value := ctx.Value(requestIDKey)
	if value == nil {
		return ""
	}

	requestID, ok := value.(string)
	if !ok {
		return ""
	}

	return requestID
}

// WithTaskID returns a new context with the task ID attached.
func WithTaskID(ctx context.Context, taskID string) context.Context {
	return context.WithValue(ctx, taskIDKey, taskID)
}

// TaskIDFromContext extracts the task ID from the context.
// Returns an empty string if no task ID is found or if the value
// is not a string.
func TaskIDFromContext(ctx context.Context) string {
	value := ctx.Value(taskIDKey)
	if value == nil {
		return ""
	}

	taskID, ok := value.(string)
	if !ok {
		return ""
	}

	return taskID
}

// WithMerchantID returns a new context with the merchant ID attached.
func WithMerchantID(ctx context.Context, merchantID string) context.Context {
	return context.WithValue(ctx, merchantIDKey, merchantID)
}

// MerchantIDFromContext extracts the merchant ID from the context.
// Returns an empty string if no merchant ID is found or if the value
// is not a string.
func MerchantIDFromContext(ctx context.Context) string {
	value := ctx.Value(merchantIDKey)
	if value == nil {
		return ""
	}

	merchantID, ok := value.(string)
	if !ok {
		return ""
	}

	return merchantID
}

// WithRzpKey returns a new context with the Razorpay key attached.
func WithRzpKey(ctx context.Context, rzpKey string) context.Context {
	return context.WithValue(ctx, rzpKeyKey, rzpKey)
}

// RzpKeyFromContext extracts the Razorpay key from the context.
// Returns an empty string if no key is found or if the value
// is not a string.
func RzpKeyFromContext(ctx context.Context) string {
	value := ctx.Value(rzpKeyKey)
	if value == nil {
		return ""
	}

	rzpKey, ok := value.(string)
	if !ok {
		return ""
	}

	return rzpKey
}

// WithPassport returns a new context with the passport instance attached.
func WithPassport(
	ctx context.Context,
	passport passport.IPassport,
) context.Context {
	return context.WithValue(ctx, passportKey, passport)
}

// PassportFromContext extracts the passport instance from the context.
// Returns nil if no passport is found.
func PassportFromContext(ctx context.Context) passport.IPassport {
	value := ctx.Value(passportKey)
	if value == nil {
		return nil
	}

	passport, ok := value.(passport.IPassport)
	if !ok {
		return nil
	}

	return passport
}

// WithAuthMode returns a new context with the auth mode attached.
func WithAuthMode(ctx context.Context, mode string) context.Context {
	return context.WithValue(ctx, authModeKey, mode)
}

// AuthModeFromContext extracts the auth mode from the context.
// Returns an empty string if no mode is found.
func AuthModeFromContext(ctx context.Context) string {
	value := ctx.Value(authModeKey)
	if value == nil {
		return ""
	}

	mode, ok := value.(string)
	if !ok {
		return ""
	}

	return mode
}
