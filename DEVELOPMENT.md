# Development Guide

This document provides guidelines and instructions for developers working on the HTTP Load Balancer project.

## Development Environment Setup

1. Install Prerequisites:

   ```bash
   # Install Go 1.20 or higher
   # Install Docker and Docker Compose
   # Install Make
   ```

2. Initialize Development Environment:
   ```bash
   make init
   ```

This will install required development tools:

- golangci-lint for code linting
- go-test-report for test reporting

## Project Structure

```
lb-project/
├── cmd/                    # Application entry points
├── internal/              # Private application code
├── test/                  # Test suites
├── scripts/              # Development and CI scripts
├── examples/             # Example implementations
└── deployments/          # Deployment configurations
```

## Development Workflow

1. **Fork and Clone**

   ```bash
   git clone https://github.com/yourusername/lb-project.git
   cd lb-project
   ```

2. **Install Dependencies**

   ```bash
   make deps
   ```

3. **Run Tests**

   ```bash
   # Run unit tests
   make test

   # Run integration tests
   make integration-test

   # Generate coverage report
   make coverage
   ```

4. **Local Development**

   ```bash
   # Run the load balancer locally
   make run

   # Or with Docker
   make docker-run
   ```

## Testing

### Unit Tests

- Located in `*_test.go` files alongside the code they test
- Run with `make test`
- Coverage reports in `test/reports/coverage.html`

### Integration Tests

- Located in `test/integration/`
- Run with `make integration-test`
- Coverage reports in `test/reports/integration/coverage.html`

### Test Reports

- HTML test reports are generated in `test/reports/`
- Coverage information is included
- Reports are git-ignored

## Code Style and Linting

1. **Code Style**

   - Follow standard Go code style
   - Use `gofmt` for formatting
   - Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines

2. **Linting**
   ```bash
   make lint
   ```

## Making Changes

1. **Create a Branch**

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes**

   - Write tests first
   - Implement changes
   - Ensure all tests pass
   - Update documentation

3. **Commit Changes**

   ```bash
   git add .
   git commit -m "Description of changes"
   ```

4. **Submit Pull Request**
   - Push changes to your fork
   - Create pull request
   - Wait for review

## Performance Testing

1. **Load Testing**

   ```bash
   # Run example backends
   go run examples/backend/main.go -port=9001 -id=1
   go run examples/backend/main.go -port=9002 -id=2
   go run examples/backend/main.go -port=9003 -id=3

   # Run load balancer
   make run

   # In another terminal, run load test
   hey -n 1000 -c 100 http://localhost:8080/
   ```

## Debugging

1. **Logs**

   - Check load balancer logs
   - Check backend logs
   - Use metrics endpoint for monitoring

2. **Metrics**
   - Access Prometheus metrics at `:9090/metrics`
   - Monitor:
     - Request counts
     - Response times
     - Error rates
     - Backend health

## Release Process

1. **Version Bump**

   - Update version in code
   - Update CHANGELOG.md

2. **Testing**

   ```bash
   make test
   make integration-test
   ```

3. **Build Release**

   ```bash
   make build
   make docker-build
   ```

4. **Tag Release**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

## Documentation

- Update README.md for user-facing changes
- Update DEVELOPMENT.md for developer-facing changes
- Update API documentation when changing interfaces
- Keep diagrams up to date

## Best Practices

1. **Code Quality**

   - Write clear, documented code
   - Follow SOLID principles
   - Keep functions small and focused
   - Use meaningful variable names

2. **Testing**

   - Write tests for new features
   - Maintain high test coverage
   - Include integration tests
   - Test error conditions

3. **Performance**

   - Profile code when needed
   - Monitor memory usage
   - Test with realistic loads
   - Consider resource constraints

4. **Security**
   - Review dependencies
   - Follow security best practices
   - Handle errors appropriately
   - Validate input

## Getting Help

- Check existing documentation
- Review test cases for examples
- Open an issue for questions
- Contact maintainers

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
