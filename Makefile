.SILENT:
.PHONY:

all:
	echo "Commands:"
	echo "[make build] build gotestiful locally"

build:
	go build -o ~/go/bin/ ./gotestiful.go

install:
	go install github.com/alex-parra/gotestiful@latest

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1 run ./...
