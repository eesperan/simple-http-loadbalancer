#!/bin/bash

# Exit on any error
set -e

# Install go-test-report if not already installed
if ! command -v go-test-report &> /dev/null; then
    echo "Installing go-test-report..."
    go install github.com/vakenbolt/go-test-report@latest
fi

# Create test output directory if it doesn't exist
mkdir -p test/reports

# Run tests with coverage and generate JSON output
echo "Running tests..."
go test -v -coverprofile=test/reports/coverage.out -json ./... > test/reports/test-output.json

# Generate HTML coverage report
echo "Generating coverage report..."
go tool cover -html=test/reports/coverage.out -o test/reports/coverage.html

# Generate HTML test report
echo "Generating test report..."
go-test-report -t "Load Balancer Test Results" -o test/reports/test-report.html < test/reports/test-output.json

echo "Test execution completed!"
echo "Reports generated:"
echo "- Test Report: test/reports/test-report.html"
echo "- Coverage Report: test/reports/coverage.html"

# Print coverage statistics
go tool cover -func=test/reports/coverage.out
