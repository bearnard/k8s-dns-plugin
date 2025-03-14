# Makefile for k8s-dns-plugin

.PHONY: all build clean test

all: build

build:
	go build -o k8sdns-plugin ./cmd/main.go

clean:
	go clean
	rm -f k8sdns-plugin

test:
	go test ./... -v