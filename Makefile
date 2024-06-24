all: test build

.PHONY: test
test: # Run tests
	go run github.com/onsi/ginkgo/v2/ginkgo -r

build: # Build the binary
	go build -o bin/kratix main.go

.PHONY: help
help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

