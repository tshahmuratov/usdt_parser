package grinex_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tshahmuratov/usdt_parser/internal/pkg/grinex"
)

func generateDepthJSON(n int) string {
	var sb strings.Builder
	sb.WriteString(`{"timestamp":1700000000,"asks":[`)
	for i := range n {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf(`{"price":"%d.%02d","volume":"100.0","amount":"10000.0"}`, 81+i/100, i%100))
	}
	sb.WriteString(`],"bids":[`)
	for i := range n {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf(`{"price":"%d.%02d","volume":"200.0","amount":"20000.0"}`, 80-i/100, 99-i%100))
	}
	sb.WriteString(`]}`)
	return sb.String()
}

func BenchmarkFetchDepth_20Entries(b *testing.B) {
	body := generateDepthJSON(20)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer server.Close()

	client := grinex.NewGrinexClient(server.URL, 0, 20, nil)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, err := client.FetchDepth(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFetchDepth_200Entries(b *testing.B) {
	body := generateDepthJSON(200)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer server.Close()

	client := grinex.NewGrinexClient(server.URL, 0, 0, nil)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, err := client.FetchDepth(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
