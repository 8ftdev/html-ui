.PHONY: build test check

build:
	go build -trimpath -o bin/html-ui ./cmd/html-ui

test:
	go test ./...

check:
	go test -race ./...
	go vet ./...
	bun run test
