.PHONY: help build test run clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Compile the binary
	go build -o posthog-proxy .

test: ## Run tests
	CGO_ENABLED=0 go test -v ./...

run: ## Run the proxy locally
	go run .

clean: ## Remove the built binary
	rm -f posthog-proxy
