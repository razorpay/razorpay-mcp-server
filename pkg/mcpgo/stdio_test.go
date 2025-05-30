package mcpgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockInvalidServer implements Server but not mark3labsImpl
type mockInvalidServer struct{}

func (m *mockInvalidServer) AddTools(tools ...Tool) {}

// TestNewStdioServer tests the creation of a stdio server
func TestNewStdioServer(t *testing.T) {
	t.Run("Success case", func(t *testing.T) {
		server := NewServer("test-server", "1.0.0")

		stdioServer, err := NewStdioServer(server)

		assert.NoError(t, err)
		assert.NotNil(t, stdioServer)

		// Verify interface implementation
		var _ TransportServer = stdioServer
	})

	t.Run("Invalid server implementation", func(t *testing.T) {
		invalidServer := &mockInvalidServer{}

		stdioServer, err := NewStdioServer(invalidServer)

		assert.Error(t, err)
		assert.Nil(t, stdioServer)
		assert.ErrorIs(t, err, ErrInvalidServerImplementation)
	})
}
