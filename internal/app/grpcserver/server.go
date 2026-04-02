package grpcserver

import (
	"context"
	"fmt"
	"net"

	ratesv1 "github.com/tshahmuratov/usdt_parser/gen/rates/v1"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_handler"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/config"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type ServerParams struct {
	fx.In
	Config  *config.Config
	Logger  *zap.Logger
	Handler *rates_handler.RatesHandler
}

func NewServer(p ServerParams) *grpc.Server {
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	ratesv1.RegisterRateServiceServer(srv, p.Handler)

	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(srv)

	return srv
}

func initTracer(ctx context.Context, cfg *config.Config) (*sdktrace.TracerProvider, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTel.Endpoint),
	}
	if cfg.OTel.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("usdt-parser"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func RegisterLifecycle(lc fx.Lifecycle, srv *grpc.Server, cfg *config.Config, logger *zap.Logger) {
	var tp *sdktrace.TracerProvider

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var err error
			tp, err = initTracer(ctx, cfg)
			if err != nil {
				logger.Warn("failed to init tracer, continuing without tracing", zap.Error(err))
			}

			addr := fmt.Sprintf(":%d", cfg.GRPC.Port)
			lis, err := net.Listen("tcp", addr)
			if err != nil {
				return fmt.Errorf("listen %s: %w", addr, err)
			}

			logger.Info("gRPC server starting", zap.String("addr", addr))
			go func() {
				if err := srv.Serve(lis); err != nil {
					logger.Error("gRPC server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("gRPC server stopping")
			srv.GracefulStop()
			if tp != nil {
				if err := tp.Shutdown(ctx); err != nil {
					logger.Error("tracer shutdown error", zap.Error(err))
				}
			}
			return nil
		},
	})
}
