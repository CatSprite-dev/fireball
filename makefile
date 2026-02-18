.PHONY: dev build-frontend run-server clean

build-frontend:
	@echo "Building frontend assets..."
	@cd frontend && npm run build

run-server:
	@echo "Starting server..."
	@go run cmd/server/main.go

clean:
	@echo "Cleaning dist folder..."
	@cd frontend && npm run clean