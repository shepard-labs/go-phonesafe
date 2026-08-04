LIBPHONENUMBER_PATH ?= ../libphonenumber
UPSTREAM_VERSION := $(shell git -C $(LIBPHONENUMBER_PATH) describe --tags --abbrev=0 2>/dev/null || echo "unknown")

.PHONY: generate test bench fuzz lint vet coverage validate-metadata update-metadata

generate:
	go run ./cmd/phonesafe-gen \
		-input $(LIBPHONENUMBER_PATH)/resources \
		-output . \
		-upstream-version $(UPSTREAM_VERSION)
	@echo "Generated metadata from libphonenumber $(UPSTREAM_VERSION)"

validate-metadata:
	go run ./cmd/phonesafe-gen \
		-input $(LIBPHONENUMBER_PATH)/resources \
		-validate-only

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

update-metadata:
	cd $(LIBPHONENUMBER_PATH) && git fetch --tags && git checkout $$(git describe --tags --abbrev=0 origin/master)
	$(MAKE) generate
	$(MAKE) test
	@echo "Metadata updated and tests pass"
