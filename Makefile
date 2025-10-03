.PHONY: all build build-backend build-frontend dev test clean install

all: build

# Build both frontend and backend
build: build-backend build-frontend

# Build Go backend
build-backend:
	@echo "Building backend..."
	cd cmd/backend && CGO_ENABLED=1 go build -o ../../dist/gpx_reporting

# Build frontend
build-frontend:
	@echo "Building frontend..."
	npm install
	npm run build

# Development mode with watch
dev:
	@echo "Starting development mode..."
	npm run watch &
	go run cmd/backend/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...
	npm test

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf dist/
	rm -rf node_modules/
	rm -rf coverage/
	rm -f *.db

# Install dependencies
install:
	@echo "Installing dependencies..."
	npm install
	go mod download

# Sign plugin (requires @grafana/sign-plugin)
sign:
	@echo "Signing plugin..."
	npx @grafana/sign-plugin

# Create distribution zip
dist: build sign
	@echo "Creating distribution package..."
	mkdir -p dist
	cp -r src/plugin.json dist/
	cp -r dist/gpx_reporting dist/
	cd dist && zip -r grafana-app-reporting.zip .
