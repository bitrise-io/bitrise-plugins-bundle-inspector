#!/bin/bash
set -e

echo "Running integration tests with real artifacts..."
echo ""

# Check if test artifacts exist
if [ ! -d "test-artifacts" ]; then
    echo "❌ test-artifacts directory not found"
    echo "Please ensure test artifacts are in place"
    exit 1
fi

# Run integration tests
go test -v -tags=integration -timeout=5m ./...

echo ""
echo "✅ Integration tests completed"
