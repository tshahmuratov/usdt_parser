.PHONY: build test lint run generate docker-build migrate escape

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
