#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Plugin information
PLUGIN_ID="sheduled-reports-app"
PLUGIN_VERSION=$(grep '"version":' src/plugin.json | head -1 | sed 's/.*"version": *"\([^"]*\)".*/\1/')

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Building Grafana Reporting Plugin${NC}"
echo -e "${BLUE}  Version: ${PLUGIN_VERSION}${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Step 1: Clean previous builds
echo -e "${YELLOW}[1/6]${NC} Cleaning previous builds..."
rm -rf dist/
mkdir -p dist/
echo -e "${GREEN}✓${NC} Clean complete"
echo ""

# Step 2: Install frontend dependencies
echo -e "${YELLOW}[2/6]${NC} Installing frontend dependencies..."
if ! npm install --silent; then
    echo -e "${RED}✗ Failed to install frontend dependencies${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} Frontend dependencies installed"
echo ""

# Step 3: Build frontend
echo -e "${YELLOW}[3/6]${NC} Building frontend..."
if ! npm run build; then
    echo -e "${RED}✗ Frontend build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓${NC} Frontend built successfully"
echo ""

# Step 4: Build backend
echo -e "${YELLOW}[4/6]${NC} Building backend..."

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        GOARCH="amd64"
        ;;
    aarch64|arm64)
        GOARCH="arm64"
        ;;
    armv7l)
        GOARCH="arm"
        ;;
    *)
        echo -e "${RED}✗ Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

case "$OS" in
    linux)
        GOOS="linux"
        BINARY_NAME="gpx_reporting"
        ;;
    darwin)
        GOOS="darwin"
        BINARY_NAME="gpx_reporting"
        ;;
    mingw*|msys*|cygwin*)
        GOOS="windows"
        BINARY_NAME="gpx_reporting.exe"
        ;;
    *)
        echo -e "${RED}✗ Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

echo "  Building for: ${GOOS}/${GOARCH}"

# Build the backend
if ! CGO_ENABLED=1 GOOS=$GOOS GOARCH=$GOARCH go build -o dist/$BINARY_NAME ./cmd/backend; then
    echo -e "${RED}✗ Backend build failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓${NC} Backend built successfully"
echo ""

# Step 5: Package plugin
echo -e "${YELLOW}[5/6]${NC} Packaging plugin..."

# Verify all required files exist
if [ ! -f "dist/module.js" ]; then
    echo -e "${RED}✗ Frontend build output missing (dist/module.js)${NC}"
    exit 1
fi

if [ ! -f "dist/$BINARY_NAME" ]; then
    echo -e "${RED}✗ Backend build output missing (dist/$BINARY_NAME)${NC}"
    exit 1
fi

# Copy additional files
cp -r src/img dist/ 2>/dev/null || true
cp src/plugin.json dist/
cp README.md dist/ 2>/dev/null || true

# Make binary executable
chmod +x dist/$BINARY_NAME

# Show package contents
echo "  Package contents:"
ls -lh dist/ | grep -v "^total" | awk '{printf "    %s %s\n", $9, $5}'

echo -e "${GREEN}✓${NC} Plugin packaged"
echo ""

# Step 6: Create archive
echo -e "${YELLOW}[6/6]${NC} Creating distribution archive..."

ARCHIVE_NAME="${PLUGIN_ID}-${PLUGIN_VERSION}-${GOOS}-${GOARCH}.zip"

cd dist
if ! zip -r ../$ARCHIVE_NAME . -q; then
    echo -e "${RED}✗ Failed to create archive${NC}"
    exit 1
fi
cd ..

ARCHIVE_SIZE=$(du -h "$ARCHIVE_NAME" | cut -f1)
echo -e "${GREEN}✓${NC} Archive created: ${ARCHIVE_NAME} (${ARCHIVE_SIZE})"
echo ""

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✓ Build Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Plugin files:"
echo "  • dist/               - Plugin directory"
echo "  • $ARCHIVE_NAME       - Distribution archive"
echo ""
echo "Installation:"
echo "  1. Extract to Grafana plugins directory:"
echo "     unzip $ARCHIVE_NAME -d /var/lib/grafana/plugins/$PLUGIN_ID/"
echo ""
echo "  2. Configure Grafana to allow unsigned plugins:"
echo "     [plugins]"
echo "     allow_loading_unsigned_plugins = $PLUGIN_ID"
echo ""
echo "  3. Restart Grafana:"
echo "     sudo systemctl restart grafana-server"
echo ""
echo "  4. Configure plugin in Grafana UI:"
echo "     Settings → Set Grafana URL to: https://your-host:3000/dna"
echo ""
