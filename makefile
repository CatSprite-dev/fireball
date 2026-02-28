.PHONY: setup-environment build-frontend run-server dev

setup-environment:
	@echo "Setting up development environment..."
	@command -v go >/dev/null 2>&1 || { echo >&2 "Error: Go is not installed. Install and try again."; exit 1; }
	@command -v npm >/dev/null 2>&1 || { echo >&2 "Error: Node.js is not installed. Install and try again."; exit 1; }
	@echo "Go and npm are found. Continue..."
	@go mod download
	@cd frontend && npm install
	@echo "Check if your Node.js version is compatible. If not, preferrably install 20v:\n nvm install 20\n nvm use 20"

build-frontend:
	@echo "Building frontend assets..."
	@cd frontend && npm run build

run-server:
	@echo "Starting server..."
	@go run cmd/server/main.go

dev:
	@echo "Rolling up vite..."
	@cd frontend && npm run dev