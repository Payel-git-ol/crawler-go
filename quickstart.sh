#!/bin/bash
# Quick start script for Fyne-on

set -e

echo "=========================================="
echo "    Fyne-on - GitHub Crawler"
echo "=========================================="
echo

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

echo "✓ Go is installed: $(go version)"
echo

# Download dependencies
echo "📦 Downloading dependencies..."
go mod download
echo "✓ Dependencies downloaded"
echo

# Build
echo "🔨 Building application..."
mkdir -p bin
go build -o bin/app.exe ./cmd/app
echo "✓ Build complete: bin/app.exe"
echo

# Create data directories
echo "📁 Creating data directories..."
mkdir -p badger_data logs
echo "✓ Directories created"
echo

# Summary
echo "=========================================="
echo "    Setup Complete!"
echo "=========================================="
echo
echo "To start the application:"
echo "  ./bin/app.exe"
echo
echo "API will be available at:"
echo "  http://localhost:3000"
echo
echo "Useful endpoints:"
echo "  GET  http://localhost:3000/health       - Health check"
echo "  GET  http://localhost:3000/stats        - Database statistics"
echo "  GET  http://localhost:3000/api/routes   - List all routes"
echo "  POST http://localhost:3000/crawler/start - Start crawler"
echo
echo "Documentation:"
echo "  - README.md         - Project overview"
echo "  - EXAMPLES.md       - API usage examples"
echo "  - DEVELOPMENT.md    - Development guide"
echo
echo "GitHub token (optional but recommended):"
echo "  Get from: https://github.com/settings/tokens"
echo "  Pass with: {\"github_token\": \"your_token\"}"
echo
