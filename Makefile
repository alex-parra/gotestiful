.SILENT:
.PHONY:

all:
	echo "Commands:"
	echo "[make build] build gotestiful locally"

build:
	go build -o ~/go/bin/ ./gotestiful.go

install:
	go install github.com/alex-parra/gotestiful@latest
