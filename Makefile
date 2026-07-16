.PHONY: build test vet lint install clean coverage

BINARY := cine-cli
GO := go
GOFLAGS :=

build:
	$(GO) build $(GOFLAGS) -o $(BINARY) ./cmd/cine-cli

test:
	$(GO) test ./... -count=1 -timeout 30s

vet:
	$(GO) vet ./...

lint:
	golangci-lint run

install:
	$(GO) install ./cmd/cine-cli

clean:
	rm -f $(BINARY)
	rm -f coverage.out coverage.html

coverage:
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	$(GO) tool cover -html=coverage.out -o coverage.html
