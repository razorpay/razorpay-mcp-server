package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

// Endpoint defines a route and its response
type Endpoint struct {
	Path     string
	Method   string
	Response interface{}
}

// NewHTTPClient creates and returns a mock HTTP client with configured
// endpoints
func NewHTTPClient(
	endpoints ...Endpoint,
) (*http.Client, *httptest.Server) {
	mockServer := NewServer(endpoints...)
	client := mockServer.Client()
	return client, mockServer
}

// NewServer creates a mock HTTP server for testing
func NewServer(endpoints ...Endpoint) *httptest.Server {
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
				_, err := w.Write(resp)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			case string:
				_, err := w.Write([]byte(resp))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			default:
				err := json.NewEncoder(w).Encode(resp)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
		}).Methods(method)
	}

	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"code":        "NOT_FOUND",
					"description": fmt.Sprintf("No mock for %s %s", r.Method, r.URL.Path),
				},
			})
		})

	return httptest.NewServer(router)
}
