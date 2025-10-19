# DevSmith Modular Platform Makefile

.PHONY: all build test clean

all: build

build:
	go build ./...

test:
	go test ./...

clean:
	rm -rf bin/
