.PHONY: build fmt lint test

build:
	go build ./...

fmt:
	go fmt ./...

test:
	go test ./...
