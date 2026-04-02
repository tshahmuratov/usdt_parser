package debug

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // pprof registered on debug port only
	"runtime"

	"github.com/tshahmuratov/usdt_parser/internal/pkg/config"
)

type Server struct {
	httpSrv *http.Server
}

func NewServer(cfg *config.Config) *Server {
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(5)

	return &Server{
		httpSrv: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Debug.Port),
			Handler: http.DefaultServeMux,
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
