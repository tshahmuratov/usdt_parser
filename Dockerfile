FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/app cmd/app/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/app /bin/app

ENTRYPOINT ["/bin/app"]
