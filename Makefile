.PHONEY: run

build:
	@go build -race -o bin/goredis .

test:
	@go test -timeout=10s -v ./...

run: build
	@./bin/goredis
