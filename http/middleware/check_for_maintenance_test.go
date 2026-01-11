package middleware

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

func testHttpCheckForMaintenanceMiddleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		CheckForMaintenance()(NewTestContext(r.Context(), next, w, r))
	})
}

func TestMaintenanceMode(t *testing.T) {
	server := httptest.NewServer(testHttpCheckForMaintenanceMiddleware(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
	})))
	defer server.Close()

	client := &nethttp.Client{}

	err := file.Create(path.Storage("framework/maintenance"), "")
	require.NoError(t, err)

	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	err = file.Remove(path.Storage("framework/maintenance"))
	require.NoError(t, err)

	resp, err = client.Get(server.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
