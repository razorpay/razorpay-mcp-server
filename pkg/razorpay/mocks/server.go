package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

// MockEndpoint defines a route and its response
type MockEndpoint struct {
	Path     string
	Method   string
	Response interface{}
}

// NewMockedHTTPClient creates and returns a mock HTTP client with configured
// endpoints
func NewMockedHTTPClient(
	endpoints ...MockEndpoint,
) (*http.Client, *httptest.Server) {
	mockServer := SetupMockServer(endpoints...)
	client := mockServer.Client()
	return client, mockServer
}

// SetupMockServer creates a mock HTTP server for testing
func SetupMockServer(endpoints ...MockEndpoint) *httptest.Server {
	router := mux.NewRouter()

	for _, endpoint := range endpoints {
		path := endpoint.Path
		method := endpoint.Method
		response := endpoint.Response

		router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if respMap, ok := response.(map[string]interface{}); ok {
				if _, hasError := respMap["error"]; hasError {
					w.WriteHeader(http.StatusBadRequest)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			} else {
				w.WriteHeader(http.StatusOK)
			}

			switch resp := response.(type) {
			case []byte:
				w.Write(resp)
			case string:
				w.Write([]byte(resp))
			default:
				json.NewEncoder(w).Encode(resp)
			}
		}).Methods(method)
	}

	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)

			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"code":        "NOT_FOUND",
					"description": fmt.Sprintf("No mock for %s %s", 
					r.Method, r.URL.Path),
				},
			})
		})

	return httptest.NewServer(router)
}
