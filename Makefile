LIBPHONENUMBER_PATH ?= ../libphonenumber

.PHONY: generate test bench fuzz lint vet coverage

generate:
	go run ./cmd/phonesafe-gen -input $(LIBPHONENUMBER_PATH)/resources -output .

test:
	go test ./...

bench:
	go test -bench=. -benchmem ./...

fuzz:
	go test -fuzz=. -fuzztime=60s ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
