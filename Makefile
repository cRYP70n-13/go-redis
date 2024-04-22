.PHONY: build
build:
	@go build -race -o bin/goredis .

.PHONY: test
test:
	@go test -timeout=10s -v ./...

.PHONY: build
run: build
	@./bin/goredis

.PHONY: fmt
fmt:
	@go fmt ./...
