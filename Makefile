build:
	@go build -o bin/goredis .

test:
	@go test -race -v ./...

run: build
	@./bin/goredis

.PHONEY: run
