package grinex_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/grinex"
)

func TestGrinexClient_FetchDepth(t *testing.T) {
	t.Run("success with limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/spot/depth", r.URL.Path)
			assert.Equal(t, "usdta7a5", r.URL.Query().Get("symbol"))
			assert.Equal(t, "20", r.URL.Query().Get("limit"))
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"timestamp": 1700000000,
				"asks": [{"price": "81.24", "volume": "100.0", "amount": "8124.0"}],
				"bids": [{"price": "81.17", "volume": "200.0", "amount": "16234.0"}]
			}`))
		}))
		defer server.Close()

		client := grinex.NewGrinexClient(server.URL, 0, 20)
		depth, err := client.FetchDepth(context.Background())

		require.NoError(t, err)
		require.Len(t, depth.Asks, 1)
		require.Len(t, depth.Bids, 1)
		assert.Equal(t, 81.24, float64(depth.Asks[0].Price))
		assert.Equal(t, 81.17, float64(depth.Bids[0].Price))
	})

	t.Run("non-200 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := grinex.NewGrinexClient(server.URL, 0, 20)
		_, err := client.FetchDepth(context.Background())
		require.Error(t, err)
	})

	t.Run("malformed json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(`{bad json`))
		}))
		defer server.Close()

		client := grinex.NewGrinexClient(server.URL, 0, 20)
		_, err := client.FetchDepth(context.Background())
		require.Error(t, err)
	})
}
