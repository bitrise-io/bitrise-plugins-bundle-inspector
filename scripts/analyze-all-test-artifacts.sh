#!/bin/bash
set -e

BINARY="./bundle-inspector"

# Build if needed
if [ ! -f "$BINARY" ]; then
    echo "Building bundle-inspector..."
    go build -o bundle-inspector ./cmd/bundle-inspector
fi

echo "================================================"
echo "Analyzing Real Test Artifacts"
echo "================================================"
echo ""

# iOS IPA
if [ -f "test-artifacts/ios/lightyear.ipa" ]; then
    echo "1. Analyzing lightyear.ipa..."
    echo "-----------------------------------"
    $BINARY analyze test-artifacts/ios/lightyear.ipa
    echo ""

    echo "Generating JSON report..."
    $BINARY analyze test-artifacts/ios/lightyear.ipa -o json -f test-artifacts/ios/lightyear-report.json
    echo "✓ JSON report saved to test-artifacts/ios/lightyear-report.json"
    echo ""
fi

# iOS App Bundle
if [ -d "test-artifacts/ios/Wikipedia.app" ]; then
    echo "2. Analyzing Wikipedia.app..."
    echo "-----------------------------------"
    $BINARY analyze test-artifacts/ios/Wikipedia.app
    echo ""

    echo "Generating JSON report..."
    $BINARY analyze test-artifacts/ios/Wikipedia.app -o json -f test-artifacts/ios/wikipedia-report.json
    echo "✓ JSON report saved to test-artifacts/ios/wikipedia-report.json"
    echo ""
fi

# Android APK
if [ -f "test-artifacts/android/2048-game-2048.apk" ]; then
    echo "3. Analyzing 2048-game-2048.apk..."
    echo "-----------------------------------"
    $BINARY analyze test-artifacts/android/2048-game-2048.apk
    echo ""

    echo "Generating JSON report..."
    $BINARY analyze test-artifacts/android/2048-game-2048.apk -o json -f test-artifacts/android/2048-report.json
    echo "✓ JSON report saved to test-artifacts/android/2048-report.json"
    echo ""
fi

echo "================================================"
echo "✅ All test artifacts analyzed"
echo "================================================"
