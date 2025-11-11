package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/razorpay/razorpay-mcp-server/pkg/log"
)

func TestNew(t *testing.T) {
	t.Run("creates observability without options", func(t *testing.T) {
		obs := New()
		assert.NotNil(t, obs)
		assert.Nil(t, obs.Logger)
	})

	t.Run("creates observability with logging service option", func(t *testing.T) {
		ctx := context.Background()
		_, logger := log.New(ctx, log.NewConfig(log.WithMode(log.ModeStdio)))

		obs := New(WithLoggingService(logger))
		assert.NotNil(t, obs)
		assert.NotNil(t, obs.Logger)
		assert.Equal(t, logger, obs.Logger)
	})

	t.Run("creates observability with multiple options", func(t *testing.T) {
		ctx := context.Background()
		_, logger1 := log.New(ctx, log.NewConfig(log.WithMode(log.ModeStdio)))
		_, logger2 := log.New(ctx, log.NewConfig(log.WithMode(log.ModeStdio)))

		// Last option should override previous ones
		obs := New(
			WithLoggingService(logger1),
			WithLoggingService(logger2),
		)
		assert.NotNil(t, obs)
		assert.NotNil(t, obs.Logger)
		assert.Equal(t, logger2, obs.Logger)
	})

	t.Run("creates observability with empty options", func(t *testing.T) {
		obs := New()
		assert.NotNil(t, obs)
		assert.Nil(t, obs.Logger)
	})
}

func TestWithLoggingService(t *testing.T) {
	t.Run("returns option function", func(t *testing.T) {
		ctx := context.Background()
		_, logger := log.New(ctx, log.NewConfig(log.WithMode(log.ModeStdio)))

		opt := WithLoggingService(logger)
		assert.NotNil(t, opt)

		obs := &Observability{}
		opt(obs)

		assert.Equal(t, logger, obs.Logger)
	})

	t.Run("sets logger to nil", func(t *testing.T) {
		opt := WithLoggingService(nil)
		assert.NotNil(t, opt)

		obs := &Observability{}
		opt(obs)

		assert.Nil(t, obs.Logger)
	})

	t.Run("applies option to existing observability", func(t *testing.T) {
		ctx := context.Background()
		_, logger1 := log.New(ctx, log.NewConfig(log.WithMode(log.ModeStdio)))
		_, logger2 := log.New(ctx, log.NewConfig(log.WithMode(log.ModeStdio)))

		obs := New(WithLoggingService(logger1))
		assert.Equal(t, logger1, obs.Logger)

		// Apply new option
		opt := WithLoggingService(logger2)
		opt(obs)

		assert.Equal(t, logger2, obs.Logger)
	})
}
