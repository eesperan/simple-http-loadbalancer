#!/bin/bash

# Exit on any error
set -e

# Create test output directory if it doesn't exist
mkdir -p test/reports/integration

echo "Running integration tests..."

# Run integration tests with coverage and generate JSON output
go test -tags=integration -v -coverprofile=test/reports/integration/coverage.out \
    -json ./test/integration/... > test/reports/integration/test-output.json

# Generate HTML coverage report for integration tests
go tool cover -html=test/reports/integration/coverage.out \
    -o test/reports/integration/coverage.html

# Install go-test-report if not already installed
if ! command -v go-test-report &> /dev/null; then
    echo "Installing go-test-report..."
    go install github.com/vakenbolt/go-test-report@latest
fi

# Generate HTML test report for integration tests
go-test-report -t "Load Balancer Integration Test Results" \
    -o test/reports/integration/test-report.html < test/reports/integration/test-output.json

echo "Integration test execution completed!"
echo "Reports generated:"
echo "- Test Report: test/reports/integration/test-report.html"
echo "- Coverage Report: test/reports/integration/coverage.html"

# Print coverage statistics
go tool cover -func=test/reports/integration/coverage.out
