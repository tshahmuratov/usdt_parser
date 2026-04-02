package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/config"
)

type Server struct {
	httpSrv *http.Server
}

func NewServer(cfg *config.Config) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &Server{
		httpSrv: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Metrics.Port),
			Handler: mux,
		},
	}
}

func (s *Server) Start() error {
	go s.httpSrv.ListenAndServe() //nolint:errcheck
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}
