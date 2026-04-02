.PHONY: build test lint run generate docker-build migrate escape load-test profile

build:
	go build -o bin/app cmd/app/main.go

test:
	go test -race ./...

lint:
	golangci-lint run

run:
	go run cmd/app/main.go

generate:
	buf generate
	mockery

docker-build:
	docker build -t usdt-parser .

migrate:
	atlas migrate apply --url "$$DATABASE_URL"

escape:
	@go build -gcflags='-m' ./... 2>&1 | grep "escapes to heap" | grep "^internal/" | sort

load-test:
	@./scripts/load-test.sh localhost:50051 50 60s

profile:
	@echo "Capturing 30s CPU profile..."
	@curl -sS -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30
	@echo "Opening profile in browser..."
	@go tool pprof -http=:8080 cpu.prof
