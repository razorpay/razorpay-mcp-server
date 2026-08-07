package razorpay

import (
	"testing"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

func TestSchemaConstraints(t *testing.T) {
	obs := CreateTestObservability()
	client := &rzpsdk.Client{}

	t.Run("create_refund speed enum", func(t *testing.T) {
		schema, ok := mcpgo.ParameterSchema(CreateRefund(obs, client), "speed")
		require.True(t, ok)
		assert.Equal(t, []interface{}{"normal", "optimum"}, schema["enum"])
	})

	t.Run("create_payment_link currency pattern", func(t *testing.T) {
		schema, ok := mcpgo.ParameterSchema(
			CreatePaymentLink(obs, client), "currency")
		require.True(t, ok)
		assert.Equal(t, "^[A-Z]{3}$", schema["pattern"])
	})

	t.Run("create_payment_link reference_id max length", func(t *testing.T) {
		schema, ok := mcpgo.ParameterSchema(
			CreatePaymentLink(obs, client), "reference_id")
		require.True(t, ok)
		assert.Equal(t, 40, schema["maxLength"])
	})

	t.Run("create_upi_payment_link currency INR only", func(t *testing.T) {
		schema, ok := mcpgo.ParameterSchema(
			CreateUpiPaymentLink(obs, client), "currency")
		require.True(t, ok)
		assert.Equal(t, []interface{}{"INR"}, schema["enum"])
	})

	t.Run("create_upi_payment_link reference_id max length", func(t *testing.T) {
		schema, ok := mcpgo.ParameterSchema(
			CreateUpiPaymentLink(obs, client), "reference_id")
		require.True(t, ok)
		assert.Equal(t, 40, schema["maxLength"])
	})

	t.Run("fetch_all_payouts count bounds", func(t *testing.T) {
		schema, ok := mcpgo.ParameterSchema(
			FetchAllPayouts(obs, client), "count")
		require.True(t, ok)
		assert.Equal(t, float64(1), schema["minimum"])
		assert.Equal(t, float64(100), schema["maximum"])
	})
}
