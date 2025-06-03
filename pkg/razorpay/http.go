package razorpay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	rzpsdk "github.com/razorpay/razorpay-go/v2"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// HTTPConfig holds the configuration for the HTTP server
type HTTPConfig struct {
	address string
	port    int
}

// HTTPConfigOpts defines a function type for applying configuration options
type HTTPConfigOpts func(*HTTPConfig)

// WithHTTPAddress returns an option to set the server address
func WithHTTPAddress(address string) HTTPConfigOpts {
	return func(config *HTTPConfig) {
		config.address = address
	}
}

// WithHTTPPort returns an option to set the server port
func WithHTTPPort(port int) HTTPConfigOpts {
	return func(config *HTTPConfig) {
		config.port = port
	}
}

// NewHTTPConfig creates a new HTTP server configuration with the provided options
func NewHTTPConfig(opts ...HTTPConfigOpts) *HTTPConfig {
	config := &HTTPConfig{
		address: "localhost",
		port:    8080,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// HTTPServer implements a JSON-RPC HTTP server
type HTTPServer struct {
	server     *Server
	config     *HTTPConfig
	httpServer *http.Server
	mux        *http.ServeMux
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Tool represents a tool that can be called
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolsListResponse represents the response to tools/list
type ToolsListResponse struct {
	Tools []Tool `json:"tools"`
}

// ToolCallParams represents the parameters for tools/call
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(server *Server, config *HTTPConfig) (*HTTPServer, error) {
	return &HTTPServer{
		server: server,
		config: config,
	}, nil
}

// Start starts the HTTP server
func (h *HTTPServer) Start() error {
	h.mux = http.NewServeMux()

	// Register health check endpoints
	h.mux.HandleFunc("/live", h.handleLiveness)
	h.mux.HandleFunc("/ready", h.handleReadiness)

	// Register JSON-RPC endpoint as default handler
	h.mux.HandleFunc("/", h.handleJSONRPC)

	// Create HTTP server
	h.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", h.config.address, h.config.port),
		Handler:      h.mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the HTTP server
	return h.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (h *HTTPServer) Shutdown(ctx context.Context) error {
	if h.httpServer != nil {
		return h.httpServer.Shutdown(ctx)
	}
	return nil
}

// handleLiveness returns 200 OK for liveness probe
func (h *HTTPServer) handleLiveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleReadiness returns 200 OK for readiness probe
func (h *HTTPServer) handleReadiness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleJSONRPC handles JSON-RPC requests
func (h *HTTPServer) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Accel-Buffering", "no") // Prevent buffering for SSE compatibility

	// Parse JSON-RPC request
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, nil, -32700, "Parse error", err.Error())
		return
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		h.writeErrorResponse(w, req.ID, -32600, "Invalid Request", "JSON-RPC version must be 2.0")
		return
	}

	// Extract and validate authentication
	ctx, err := h.authenticateRequest(r.Context(), r)
	if err != nil {
		h.writeErrorResponse(w, req.ID, -32603, "Authentication failed", err.Error())
		return
	}

	// Route based on method
	switch req.Method {
	case "tools/list":
		h.handleToolsList(ctx, w, req)
	case "tools/call":
		h.handleToolsCall(ctx, w, req)
	default:
		h.writeErrorResponse(w, req.ID, -32601, "Method not found", fmt.Sprintf("Method %s not found", req.Method))
	}
}

// authenticateRequest extracts and validates the Bearer token
func (h *HTTPServer) authenticateRequest(ctx context.Context, r *http.Request) (context.Context, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header required")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	token := parts[1]

	// Decode base64 token to get key:secret
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token encoding")
	}

	// Split into key and secret
	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return nil, fmt.Errorf("invalid credentials format")
	}

	keyID := credentials[0]
	keySecret := credentials[1]

	// Create Razorpay client
	client := rzpsdk.NewClient(keyID, keySecret)

	// Add client to context
	return mcpgo.WithClient(ctx, client), nil
}

// handleToolsList handles the tools/list method
func (h *HTTPServer) handleToolsList(ctx context.Context, w http.ResponseWriter, req JSONRPCRequest) {
	// Get all tools from the server
	tools := h.server.GetAllTools()

	// Convert to the expected format
	var toolsList []Tool
	for _, tool := range tools {
		toolsList = append(toolsList, Tool{
			Name:        tool.GetName(),
			Description: tool.GetDescription(),
			InputSchema: tool.GetInputSchema(),
		})
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  ToolsListResponse{Tools: toolsList},
	}

	h.writeJSONResponse(w, response)
}

// handleToolsCall handles the tools/call method
func (h *HTTPServer) handleToolsCall(ctx context.Context, w http.ResponseWriter, req JSONRPCRequest) {
	// Parse parameters
	var params ToolCallParams
	if req.Params != nil {
		paramsBytes, err := json.Marshal(req.Params)
		if err != nil {
			h.writeErrorResponse(w, req.ID, -32602, "Invalid params", err.Error())
			return
		}

		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			h.writeErrorResponse(w, req.ID, -32602, "Invalid params", err.Error())
			return
		}
	}

	// Call the tool
	result, err := h.server.CallTool(ctx, params.Name, params.Arguments)
	if err != nil {
		h.writeErrorResponse(w, req.ID, -32603, "Tool execution failed", err.Error())
		return
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}

	h.writeJSONResponse(w, response)
}

// writeJSONResponse writes a JSON response
func (h *HTTPServer) writeJSONResponse(w http.ResponseWriter, response JSONRPCResponse) {
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes a JSON-RPC error response
func (h *HTTPServer) writeErrorResponse(w http.ResponseWriter, id interface{}, code int, message, data string) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	h.writeJSONResponse(w, response)
}
