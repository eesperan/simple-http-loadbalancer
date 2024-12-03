<task>
Create a robust HTTP load balancer implementation in Go with the following features and requirements:

1. Core Load Balancing Features:

   - Multiple backend support with configurable endpoints
   - Multiple frontend port configuration
   - Round-robin load balancing algorithm
   - Health check system for backend monitoring
   - Logging system for traffic patterns
   - Prometheus metrics integration
   - Dynamic configuration system
   - Graceful operations (shutdown, restart, rollout, rollback)

2. Project Structure:

```
lb-project/
├── cmd/
│   └── loadbalancer/        # Main application entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── health/             # Health check implementation
│   ├── logging/            # Logging system
│   ├── metrics/            # Metrics collection
│   └── balancer/           # Core load balancing logic
├── test/
│   ├── integration/        # Integration tests
│   └── reports/            # Test reports (gitignored)
├── scripts/                # Utility scripts
├── examples/               # Example backends
└── deployments/            # Deployment configurations
```

</task>

<package_implementations>

1. balancer Package:

   - LoadBalancer struct with backends management
   - Round-robin backend selection
   - HTTP request proxying
   - Backend health monitoring
   - Graceful shutdown support
   - Rollout/Rollback functionality

2. config Package:

   - YAML configuration parsing
   - Frontend/Backend configuration
   - Health check settings
   - Logging configuration
   - Metrics settings

3. metrics Package:

   - Prometheus metrics integration
   - Request counting
   - Response time tracking
   - Backend health status
   - Connection tracking

4. Example Backend:
   - Simple HTTP server
   - Health check endpoint
   - Request logging
   - Response identification
     </package_implementations>

<implementation_steps>

1. Create project structure and initialize Go module
2. Implement configuration system using YAML
3. Create core load balancer logic with round-robin algorithm
4. Add health check functionality
5. Implement metrics collection using Prometheus
6. Add logging system
7. Implement graceful operations
8. Create example backend servers
9. Add Docker support
10. Implement testing infrastructure
    </implementation_steps>

<testing_requirements>

1. Unit Tests:

   - Config package tests
   - Balancer package tests
   - Metrics package tests
   - Health check tests
   - Rollout/Rollback tests

2. Integration Tests:

   - Full system tests
   - Load balancing verification
   - Health check verification
   - Metrics endpoint testing

3. Test Infrastructure:
   - HTML test reports
   - Coverage reporting
   - Integration test framework
   - Test automation scripts
     </testing_requirements>

<documentation_requirements>

1. README.md:

   - Project overview
   - Features list
   - Architecture diagram (Mermaid)
   - Installation instructions
   - Usage examples
   - Configuration guide
   - API documentation

2. DEVELOPMENT.md:

   - Development setup
   - Testing instructions
   - Contribution guidelines
   - Code style guide
   - Release process
   - Best practices

3. Code Documentation:
   - Package documentation
   - Function documentation
   - Type documentation
   - Example usage
     </documentation_requirements>

<file_specifications>

1. config.yaml:

```yaml
frontends:
  - port: 8080
  - port: 8081
backends:
  - "http://backend1:9001"
  - "http://backend2:9002"
  - "http://backend3:9003"
healthcheck:
  interval: "10s"
  timeout: "2s"
  path: "/health"
logging:
  level: "info"
  format: "json"
metrics:
  enabled: true
  port: 9090
```

2. Dockerfile:

```dockerfile
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o loadbalancer ./cmd/loadbalancer

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/loadbalancer .
COPY --from=builder /app/config.yaml .
EXPOSE 8080 9090
CMD ["./loadbalancer"]
```

3. docker-compose.yml:

```yaml
version: "3"
services:
  loadbalancer:
    build: .
    ports:
      - "8080:8080"
      - "8081:8081"
      - "9090:9090"
    depends_on:
      - backend1
      - backend2
      - backend3
  backend1:
    build:
      context: .
      dockerfile: examples/backend/Dockerfile
    command: ["/app/backend", "-port=9001", "-id=1"]
  backend2:
    build:
      context: .
      dockerfile: examples/backend/Dockerfile
    command: ["/app/backend", "-port=9002", "-id=2"]
  backend3:
    build:
      context: .
      dockerfile: examples/backend/Dockerfile
    command: ["/app/backend", "-port=9003", "-id=3"]
```

4. Required Go Dependencies:

```go
require (
    github.com/prometheus/client_golang v1.16.0
    github.com/gorilla/mux v1.8.0
    gopkg.in/yaml.v2 v2.4.0
    go.uber.org/zap v1.24.0
)
```

</file_specifications>

<development_tools>

1. Required Scripts:

   - run-tests.sh for test execution
   - run-integration-tests.sh for integration testing
   - Makefile for common operations

2. Development Dependencies:

   - Go 1.20+
   - Docker and Docker Compose
   - golangci-lint
   - go-test-report

3. Build Tools:
   - Make
   - Docker
   - Go toolchain
     </development_tools>

<completion_criteria>
The implementation must satisfy:

1. All core features implemented and tested
2. Complete test coverage with reports
3. Comprehensive documentation
4. Working Docker configuration
5. Successful integration tests
6. Proper error handling
7. Metrics collection
8. Health checking
9. Graceful operations
10. Example backend servers
    </completion_criteria>

<license_requirements>
Use MIT License for the project to allow free redistribution while maintaining basic protections.
</license_requirements>

<testing_verification>

1. Unit Test Verification:

   ```bash
   ./scripts/run-tests.sh
   ```

2. Integration Test Verification:

   ```bash
   ./scripts/run-integration-tests.sh
   ```

3. Manual Testing:

   ```bash
   # Start the system
   docker-compose up --build

   # Test load balancing
   curl http://localhost:8080/

   # Check metrics
   curl http://localhost:9090/metrics

   # Verify health checks
   curl http://localhost:8080/health
   ```

   </testing_verification>
