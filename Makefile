APP_NAME = tar
GO = go
GOFLAGS = -ldflags="-s -w"
CGO_ENABLED = 0

.PHONY: build build-all test lint clean

build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -o bin/$(APP_NAME) ./cmd/tar

build-all:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe ./cmd/tar
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=arm64 $(GO) build $(GOFLAGS) -o bin/$(APP_NAME)-windows-arm64.exe ./cmd/tar
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o bin/$(APP_NAME)-linux-amd64 ./cmd/tar
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o bin/$(APP_NAME)-linux-arm64 ./cmd/tar

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

clean:
	rm -rf bin/
