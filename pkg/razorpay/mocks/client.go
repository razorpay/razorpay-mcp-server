//nolint:lll
package mocks

// PaymentClient implements the razorpay.PaymentClient interface for testing
type PaymentClient struct {
	FetchFunc func(id string, data map[string]interface{}, extraHeaders map[string]string) (map[string]interface{}, error)
}

// Fetch implements the razorpay.PaymentClient interface
func (m *PaymentClient) Fetch(id string, data map[string]interface{}, extraHeaders map[string]string) (map[string]interface{}, error) {
	if m.FetchFunc != nil {
		return m.FetchFunc(id, data, extraHeaders)
	}
	return map[string]interface{}{}, nil
}

// OrderClient implements the razorpay.OrderClient interface for testing
type OrderClient struct {
	CreateFunc func(data map[string]interface{}, extraHeaders map[string]string) (map[string]interface{}, error)
	FetchFunc  func(id string, data map[string]interface{}, extraHeaders map[string]string) (map[string]interface{}, error)
}

// Create implements the razorpay.OrderClient interface
func (m *OrderClient) Create(data map[string]interface{}, extraHeaders map[string]string) (map[string]interface{}, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(data, extraHeaders)
	}
	return map[string]interface{}{}, nil
}

// Fetch implements the razorpay.OrderClient interface
func (m *OrderClient) Fetch(id string, data map[string]interface{}, extraHeaders map[string]string) (map[string]interface{}, error) {
	if m.FetchFunc != nil {
		return m.FetchFunc(id, data, extraHeaders)
	}
	return map[string]interface{}{}, nil
}

// PaymentLinkClient implements the razorpay.PaymentLinkClient interface for testing
type PaymentLinkClient struct {
	CreateFunc func(data map[string]interface{}, options map[string]string) (map[string]interface{}, error)
	FetchFunc  func(id string, data map[string]interface{}, options map[string]string) (map[string]interface{}, error)
}

// Create implements the razorpay.PaymentLinkClient interface
func (m *PaymentLinkClient) Create(data map[string]interface{}, options map[string]string) (map[string]interface{}, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(data, options)
	}
	return map[string]interface{}{}, nil
}

// Fetch implements the razorpay.PaymentLinkClient interface
func (m *PaymentLinkClient) Fetch(id string, data map[string]interface{}, options map[string]string) (map[string]interface{}, error) {
	if m.FetchFunc != nil {
		return m.FetchFunc(id, data, options)
	}
	return map[string]interface{}{}, nil
}
