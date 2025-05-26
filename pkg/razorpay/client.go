package razorpay

import (
	"context"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// getClientFromContextOrDefault returns either the provided default 
// client or gets one from context.
func getClientFromContextOrDefault(
	ctx context.Context, 
	defaultClient *rzpsdk.Client,
) (*rzpsdk.Client, error) {
	if defaultClient != nil {
		return defaultClient, nil
	}

	clientInterface := mcpgo.ClientFromContext(ctx)
	if clientInterface == nil {
		return nil, fmt.Errorf("no client found in context")
	}

	client, ok := clientInterface.(*rzpsdk.Client)
	if !ok {
		return nil, fmt.Errorf("invalid client type in context")
	}

	return client, nil
}
