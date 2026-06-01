.PHONY: dev test lint build clean

dev:
	docker compose up --build

test: test-go test-rust

test-go:
	cd ledger-api && \
	  go env -w GONOSUMDB="*" && \
	  go mod tidy && \
	  go test ./tests/integration/... -v -count=1

test-rust:
	cd ledger-processor && cargo test

lint: lint-go lint-rust

lint-go:
	cd ledger-api && \
	  go env -w GONOSUMDB="*" && \
	  go mod tidy && \
	  go vet ./... && \
	  golangci-lint run ./...

lint-rust:
	cd ledger-processor && cargo clippy -- -D warnings

build: build-go build-rust

build-go:
	cd ledger-api && \
	  go env -w GONOSUMDB="*" && \
	  go mod tidy && \
	  CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/ledger-api ./cmd/api

build-rust:
	cd ledger-processor && cargo build --release

clean:
	docker compose down -v
	rm -rf ledger-api/bin
	cd ledger-processor && cargo clean
