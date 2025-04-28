package mock

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClient(t *testing.T) {
	client, server := NewHTTPClient(
		Endpoint{
			Path:     "/test",
			Method:   "GET",
			Response: map[string]interface{}{"status": "ok"},
		},
	)
	defer server.Close()

	require.NotNil(t, client)
	require.NotNil(t, server)

	resp, err := client.Get(server.URL + "/test")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "ok", result["status"])
}

func TestNewServer(t *testing.T) {
	testCases := []struct {
		name           string
		endpoints      []Endpoint
		requestPath    string
		requestMethod  string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful GET with JSON response",
			endpoints: []Endpoint{
				{
					Path:     "/test",
					Method:   "GET",
					Response: map[string]interface{}{"result": "success"},
				},
			},
			requestPath:    "/test",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"result":"success"}`,
		},
		{
			name: "error response",
			endpoints: []Endpoint{
				{
					Path:   "/error",
					Method: "GET",
					Response: map[string]interface{}{
						"error": map[string]interface{}{
							"code":        "BAD_REQUEST",
							"description": "Test error",
						},
					},
				},
			},
			requestPath:    "/error",
			requestMethod:  "GET",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":{"code":"BAD_REQUEST","description":"Test error"}}`,
		},
		{
			name: "string response",
			endpoints: []Endpoint{
				{
					Path:     "/string",
					Method:   "GET",
					Response: "plain text response",
				},
			},
			requestPath:    "/string",
			requestMethod:  "GET",
			expectedStatus: http.StatusOK,
			expectedBody:   "plain text response",
		},
		{
			name: "byte array response",
			endpoints: []Endpoint{
				{
					Path:     "/bytes",
					Method:   "POST",
					Response: []byte(`{"raw":"data"}`),
				},
			},
			requestPath:    "/bytes",
			requestMethod:  "POST",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"raw":"data"}`,
		},
		{
			name:           "not found",
			endpoints:      []Endpoint{},
			requestPath:    "/nonexistent",
			requestMethod:  "GET",
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":{"code":"NOT_FOUND","description":"No mock for GET /nonexistent"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := NewServer(tc.endpoints...)
			defer server.Close()

			var req *http.Request
			var err error
			if tc.requestMethod == "GET" {
				req, err = http.NewRequest(tc.requestMethod, server.URL+tc.requestPath, nil)
			} else {
				req, err = http.NewRequest(tc.requestMethod, server.URL+tc.requestPath,
					strings.NewReader("test body"))
			}
			require.NoError(t, err)

			client := server.Client()
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			actualBody := strings.TrimSpace(string(body))
			if strings.HasPrefix(actualBody, "{") {
				var expected, actual interface{}
				err = json.Unmarshal([]byte(tc.expectedBody), &expected)
				require.NoError(t, err)

				err = json.Unmarshal(body, &actual)
				require.NoError(t, err)
				assert.Equal(t, expected, actual)
			} else {
				assert.Equal(t, tc.expectedBody, actualBody)
			}
		})
	}
}

func TestMultipleEndpoints(t *testing.T) {
	server := NewServer(
		Endpoint{
			Path:   "/path1",
			Method: "GET",
			Response: map[string]interface{}{
				"endpoint": "path1",
			},
		},
		Endpoint{
			Path:   "/path2",
			Method: "POST",
			Response: map[string]interface{}{
				"endpoint": "path2",
			},
		},
	)
	defer server.Close()

	client := server.Client()

	resp1, err := client.Get(server.URL + "/path1")
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	var result1 map[string]interface{}
	err = json.NewDecoder(resp1.Body).Decode(&result1)
	require.NoError(t, err)
	assert.Equal(t, "path1", result1["endpoint"])

	req2, err := http.NewRequest("POST", server.URL+"/path2", nil)
	require.NoError(t, err)
	resp2, err := client.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var result2 map[string]interface{}
	err = json.NewDecoder(resp2.Body).Decode(&result2)
	require.NoError(t, err)
	assert.Equal(t, "path2", result2["endpoint"])
}
