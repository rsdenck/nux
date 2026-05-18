# Makefile for NUX

VERSION := 0.3.0
BINARY_NAME := nux

.PHONY: all clean build test install uninstall deb rpm aur snap docs

all: build

clean:
	rm -rf dist
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe

build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY_NAME) cmd/nux/main.go

test:
	go test ./...

install:
	go install ./cmd/nux

# Packaging Targets (require Linux environment)

deb:
	./scripts/build-deb.sh

rpm:
	./scripts/build-rpm.sh

aur:
	./scripts/build-aur.sh

snap:
	./scripts/build-snap.sh

# Documentation
docs:
	./scripts/generate-docs.sh
