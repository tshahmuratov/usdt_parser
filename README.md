# USDT Parser

gRPC service that fetches USDT rates from the Grinex exchange, calculates ask/bid prices using configurable methods, and persists results to PostgreSQL.

## Quick Start

```bash
make build
docker-compose up -d
# Or run locally:
make run
```

## Configuration

Configuration supports both environment variables (prefix `APP_`) and CLI flags. Env vars are loaded first; CLI flags override.

| Parameter | Env Var | CLI Flag | Default |
|-----------|---------|----------|---------|
| DB Host | `APP_DB_HOST` | `--db-host` | `localhost` |
| DB Port | `APP_DB_PORT` | `--db-port` | `5432` |
| DB User | `APP_DB_USER` | `--db-user` | `postgres` |
| DB Password | `APP_DB_PASSWORD` | `--db-password` | `` |
| DB Name | `APP_DB_NAME` | `--db-name` | `usdt_parser` |
| DB SSL Mode | `APP_DB_SSLMODE` | `--db-sslmode` | `disable` |
| gRPC Port | `APP_GRPC_PORT` | `--grpc-port` | `50051` |
| Grinex Base URL | `APP_GRINEX_BASE_URL` | `--grinex-base-url` | `https://grinex.io` |
| Grinex Timeout | `APP_GRINEX_TIMEOUT` | `--grinex-timeout` | `10s` |
| OTel Endpoint | `APP_OTEL_ENDPOINT` | `--otel-endpoint` | `localhost:4317` |
| OTel Insecure | `APP_OTEL_INSECURE` | `--otel-insecure` | `true` |
| Log Level | `APP_LOG_LEVEL` | `--log-level` | `info` |
| Dev Logging | `APP_LOG_DEV` | `--log-dev` | `false` |
| DB Max Open Conns | `APP_DB_MAX_OPEN_CONNS` | `--db-max-open-conns` | `25` |
| DB Max Idle Conns | `APP_DB_MAX_IDLE_CONNS` | `--db-max-idle-conns` | `5` |
| DB Conn Max Lifetime | `APP_DB_CONN_MAX_LIFETIME` | `--db-conn-max-lifetime` | `5m` |
| Grinex Depth Limit | `APP_GRINEX_DEPTH_LIMIT` | `--grinex-depth-limit` | `20` |
| Metrics Port | `APP_METRICS_PORT` | `--metrics-port` | `9090` |
| Debug Port | `APP_DEBUG_PORT` | `--debug-port` | `6060` |

## Make Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the application binary |
| `make test` | Run tests with race detector |
| `make lint` | Run golangci-lint |
| `make run` | Run the application |
| `make generate` | Generate proto code via buf |
| `make docker-build` | Build Docker image |
| `make migrate` | Apply Atlas migrations |
| `make escape` | Run Go escape analysis on internal code |
| `make load-test` | Run gRPC load test (50 concurrency, 60s) |
| `make profile` | Capture 30s CPU profile and open in browser |

## API

### GetRates

Fetches USDT rates using a specified calculation method.

```bash
grpcurl -plaintext -d '{
  "method": {"top_n": {"n": 0}}
}' localhost:50051 rates.v1.RateService/GetRates
```

**Calculation methods:**
- `top_n` — returns the price at the Nth position (0-based) of the order book
- `avg_nm` — returns the average price over positions [N, M] (inclusive)

### Health Check

```bash
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
```

## Observability

### Tracing

Traces are exported via OTLP to Jaeger. When using docker-compose, Jaeger UI is available at http://localhost:16686.

### Metrics

Prometheus metrics are exposed at `http://localhost:9090/metrics` (configurable via `APP_METRICS_PORT`).

**RED metrics (gRPC):**

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `grpc_requests_total` | counter | `method`, `code` | Total gRPC requests by method and status code |
| `grpc_request_duration_seconds` | histogram | `method` | Request latency distribution |

**Dependency metrics:**

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `grinex_fetch_total` | counter | `status` | Exchange API calls (ok/error) |
| `grinex_fetch_duration_seconds` | histogram | — | Exchange API call latency |
| `db_persist_total` | counter | `status` | DB writes (ok/error/retry) |
| `db_persist_duration_seconds` | histogram | — | DB write latency |
| `grinex_fallback_total` | counter | — | Fallback serves from cache |

**Business metrics:**

| Metric | Type | Description |
|--------|------|-------------|
| `singleflight_requests_total` | counter | Total singleflight calls |
| `singleflight_shared_total` | counter | Calls that shared a coalesced result |

Go runtime metrics (goroutines, heap, GC) are exposed automatically.

```bash
# Quick check
curl -s http://localhost:9090/metrics | grep -E "^(grpc_|grinex_|db_|singleflight_|persist_)"
```

### Profiling

pprof endpoints are available at `http://localhost:6060/debug/pprof/` (configurable via `APP_DEBUG_PORT`, set to `0` to disable).

```bash
# CPU profile (30s capture, opens in browser)
make profile

# Or manually:
curl -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof -http=:8080 cpu.prof

# Heap profile
curl -o heap.prof http://localhost:6060/debug/pprof/heap
go tool pprof -http=:8080 heap.prof

# Goroutine dump
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```

### Load Testing

Requires [ghz](https://github.com/bojand/ghz): `go install github.com/bojand/ghz/cmd/ghz@latest`

```bash
# Default: 50 concurrency, 60s duration
make load-test

# Custom parameters
./scripts/load-test.sh localhost:50051 100 30s
```

**Profile under load** (two terminals):

```bash
# Terminal 1: start load test
make load-test

# Terminal 2: capture CPU profile during load
make profile
```

## Tech Notes

- **Atlas** is used for database migrations (exploratory choice). Atlas uses a declarative schema approach — you define the desired state in `migrations/schema.sql` and Atlas computes the diff.
- **koanf** is used for configuration management (exploratory choice). It provides a modular config loading system with support for multiple providers (env vars, CLI flags, files).
