.PHONY: setup-environment build-frontend run-server dev

setup-environment:
	@"Setting up development environment..."
	@command -v go >/dev/null 2>&1 || { echo >&2 "Error: Go is not installed. Install and try again."; exit 1; }
	@command -v npm >/dev/null 2>&1 || { echo >&2 "Error: Node.js is not installed. Install and try again."; exit 1; }
	@echo "Go and npm are found. Continue..."
	@go mod download
	@cd frontend && npm install

build-frontend:
	@echo "Building frontend assets..."
	@cd frontend && npm run build

run-server:
	@echo "Starting server..."
	@go run cmd/server/main.go

dev:
	@echo "Configuring watchers.."
	@cd frontend && npm run dev