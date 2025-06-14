package observability

import (
	"github.com/razorpay/razorpay-mcp-server/pkg/log"
)

// Option is used make Observability
type Option func(*Observability)

// Observability holds all the observability related dependencies
type Observability struct {
	// Logger will be passed as dependency to other services
	// which will help in pushing logs
	Logger log.Logger
}

// New will create a new Observability object and
// apply all the options to that object and returns pointer to the object
func New(opts ...Option) *Observability {
	observability := &Observability{}
	// Loop through each option
	for _, opt := range opts {
		opt(observability)
	}
	return observability
}

// WithLoggingService will set the logging dependency in Deps
func WithLoggingService(s log.Logger) Option {
	return func(observe *Observability) {
		observe.Logger = s
	}
}
