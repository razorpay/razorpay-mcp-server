package razorpay

import (
	"io"
	"log/slog"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/razorpay/mocks"
)

// CreateTestLogger creates a logger suitable for testing
func CreateTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// SetupFetchPaymentTest creates a mock payment client and the
// FetchPayment tool for testing
func SetupFetchPaymentTest() (*mocks.PaymentClient, mcpgo.Tool) {
	mockPaymentClient := &mocks.PaymentClient{}
	log := CreateTestLogger()
	fetchPaymentTool := FetchPayment(log, mockPaymentClient)

	return mockPaymentClient, fetchPaymentTool
}

// SetupCreateOrderTest creates a mock order client and the
// CreateOrder tool for testing
func SetupCreateOrderTest() (*mocks.OrderClient, mcpgo.Tool) {
	mockOrderClient := &mocks.OrderClient{}
	log := CreateTestLogger()
	createOrderTool := CreateOrder(log, mockOrderClient)

	return mockOrderClient, createOrderTool
}

// SetupFetchOrderTest creates a mock order client and
// the FetchOrder tool for testing
func SetupFetchOrderTest() (*mocks.OrderClient, mcpgo.Tool) {
	mockOrderClient := &mocks.OrderClient{}
	log := CreateTestLogger()
	fetchOrderTool := FetchOrder(log, mockOrderClient)

	return mockOrderClient, fetchOrderTool
}

// SetupCreatePaymentLinkTest creates a mock payment link
// client and the CreatePaymentLink tool for testing
func SetupCreatePaymentLinkTest() (*mocks.PaymentLinkClient, mcpgo.Tool) {
	mockPaymentLinkClient := &mocks.PaymentLinkClient{}
	log := CreateTestLogger()
	createPaymentLinkTool := CreatePaymentLink(log, mockPaymentLinkClient)

	return mockPaymentLinkClient, createPaymentLinkTool
}

// SetupFetchPaymentLinkTest creates a mock payment link client
// and the FetchPaymentLink tool for testing
func SetupFetchPaymentLinkTest() (*mocks.PaymentLinkClient, mcpgo.Tool) {
	mockPaymentLinkClient := &mocks.PaymentLinkClient{}
	log := CreateTestLogger()
	fetchPaymentLinkTool := FetchPaymentLink(log, mockPaymentLinkClient)

	return mockPaymentLinkClient, fetchPaymentLinkTool
}

// createMCPRequest creates a CallToolRequest with the given arguments
func createMCPRequest(args map[string]interface{}) mcpgo.CallToolRequest {
	return mcpgo.CallToolRequest{
		Arguments: args,
	}
}
