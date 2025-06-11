package razorpay

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPServer_ToolsList(t *testing.T) {
	// Create a test server
	server, err := NewServer(nil, nil, "1.0.0", []string{}, false)
	require.NoError(t, err)

	httpServer, err := NewHTTPServer(server, NewHTTPConfig())
	require.NoError(t, err)

	// Create test request
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "test-request",
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Create test credentials (test_key:test_secret)
	credentials := base64.StdEncoding.EncodeToString([]byte("test_key:test_secret"))
	req.Header.Set("Authorization", "Bearer "+credentials)

	w := httptest.NewRecorder()

	// Call the handler
	httpServer.handleJSONRPC(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "no", w.Header().Get("X-Accel-Buffering"))

	var response JSONRPCResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, "test-request", response.ID)
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)

	// Check that we have tools in the response
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok)

	tools, ok := result["tools"].([]interface{})
	require.True(t, ok)
	assert.Greater(t, len(tools), 0)
}

func TestHTTPServer_InvalidJSONRPC(t *testing.T) {
	server, err := NewServer(nil, nil, "1.0.0", []string{}, false)
	require.NoError(t, err)

	httpServer, err := NewHTTPServer(server, NewHTTPConfig())
	require.NoError(t, err)

	// Test invalid JSON-RPC version
	requestBody := map[string]interface{}{
		"jsonrpc": "1.0", // Invalid version
		"id":      "test-request",
		"method":  "tools/list",
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	credentials := base64.StdEncoding.EncodeToString([]byte("test_key:test_secret"))
	req.Header.Set("Authorization", "Bearer "+credentials)

	w := httptest.NewRecorder()
	httpServer.handleJSONRPC(w, req)

	var response JSONRPCResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.NotNil(t, response.Error)
	assert.Equal(t, -32600, response.Error.Code)
	assert.Equal(t, "Invalid Request", response.Error.Message)
}

func TestHTTPServer_MissingAuth(t *testing.T) {
	server, err := NewServer(nil, nil, "1.0.0", []string{}, false)
	require.NoError(t, err)

	httpServer, err := NewHTTPServer(server, NewHTTPConfig())
	require.NoError(t, err)

	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "test-request",
		"method":  "tools/list",
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	httpServer.handleJSONRPC(w, req)

	var response JSONRPCResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Error)
	assert.Equal(t, -32603, response.Error.Code)
	assert.Equal(t, "Authentication failed", response.Error.Message)
}

func TestHTTPServer_HealthChecks(t *testing.T) {
	server, err := NewServer(nil, nil, "1.0.0", []string{}, false)
	require.NoError(t, err)

	httpServer, err := NewHTTPServer(server, NewHTTPConfig())
	require.NoError(t, err)

	// Test liveness endpoint
	req := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()
	httpServer.handleLiveness(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Test readiness endpoint
	req = httptest.NewRequest("GET", "/ready", nil)
	w = httptest.NewRecorder()
	httpServer.handleReadiness(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestHTTPServer_AuthenticateRequest(t *testing.T) {
	server, err := NewServer(nil, nil, "1.0.0", []string{}, false)
	require.NoError(t, err)

	httpServer, err := NewHTTPServer(server, NewHTTPConfig())
	require.NoError(t, err)

	// Test valid Bearer token
	req := httptest.NewRequest("POST", "/", nil)
	credentials := base64.StdEncoding.EncodeToString([]byte("test_key:test_secret"))
	req.Header.Set("Authorization", "Bearer "+credentials)

	ctx, err := httpServer.authenticateRequest(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	// Test invalid Bearer token format
	req = httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")

	_, err = httpServer.authenticateRequest(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token encoding")

	// Test missing Authorization header
	req = httptest.NewRequest("POST", "/", nil)

	_, err = httpServer.authenticateRequest(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization header required")
}
