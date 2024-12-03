<task>
Create a robust HTTP load balancer implementation in Go with the following features and requirements:

1. Core Load Balancing Features:

   - Multiple backend support with configurable endpoints
   - Multiple frontend port configuration
   - Round-robin and weighted round-robin algorithms
   - Health check system for backend monitoring
   - Logging system for traffic patterns
   - Prometheus metrics integration
   - Dynamic configuration system
   - Graceful operations (shutdown, restart, rollout, rollback)

2. Advanced Features:

   - Circuit breaker pattern implementation
   - Rate limiting with token bucket algorithm
   - SSL/TLS termination with certificate management
   - Dynamic weight adjustment based on performance
   - Request timeout and retry mechanisms
   - Error tracking and categorization

3. Project Structure:

```
lb-project/
├── cmd/
│   └── loadbalancer/        # Main application entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── balancer/           # Core load balancing logic
│   ├── algorithm/          # Load balancing algorithms
│   ├── circuitbreaker/     # Circuit breaker implementation
│   ├── ratelimit/          # Rate limiting logic
│   ├── ssl/                # SSL/TLS handling
│   ├── metrics/            # Metrics collection
│   └── errors/             # Error definitions
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
   - Round-robin and weighted round-robin selection
   - HTTP request proxying
   - Backend health monitoring
   - Graceful shutdown support
   - Rollout/Rollback functionality

2. config Package:

   - YAML configuration parsing
   - Frontend/Backend configuration
   - Health check settings
   - SSL/TLS configuration
   - Circuit breaker settings
   - Rate limit configuration
   - Logging configuration
   - Metrics settings

3. metrics Package:

   - Prometheus metrics integration
   - Request counting
   - Response time tracking
   - Backend health status
   - Circuit breaker states
   - Rate limiter statistics
   - Connection tracking

4. circuitbreaker Package:

   - Circuit breaker implementation
   - Failure detection
   - Half-open state handling
   - Automatic recovery
   - State tracking

5. ratelimit Package:

   - Token bucket algorithm
   - Sliding window implementation
   - Burst handling
   - Rate tracking

6. ssl Package:

   - Certificate management
   - TLS configuration
   - Certificate reloading
   - Client authentication

7. Example Backend:
   - Simple HTTP server
   - Health check endpoint
   - Request logging
   - Response identification
   - Load simulation
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
11. Add circuit breaker implementation
12. Add rate limiting support
13. Implement SSL/TLS handling
14. Add weighted round-robin algorithm
15. Implement error tracking system
    </implementation_steps>

<testing_requirements>

1. Unit Tests:

   - Config package tests
   - Balancer package tests
   - Metrics package tests
   - Health check tests
   - Rollout/Rollback tests
   - Circuit breaker tests
   - Rate limiter tests
   - SSL/TLS tests
   - Algorithm tests

2. Integration Tests:

   - Full system tests
   - Load balancing verification
   - Health check verification
   - Metrics endpoint testing
   - Circuit breaker behavior
   - Rate limiting effectiveness
   - SSL/TLS functionality

3. Performance Tests:

   - Benchmark tests
   - Load tests
   - Concurrency tests
   - Resource usage monitoring
   - Latency measurements

4. Test Infrastructure:
   - HTML test reports
   - Coverage reporting
   - Integration test framework
   - Test automation scripts
   - Benchmark reporting
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
   - Performance characteristics

2. DEVELOPMENT.md:

   - Development setup
   - Testing instructions
   - Contribution guidelines
   - Code style guide
   - Release process
   - Best practices
   - Performance targets
   - Benchmark instructions

3. ARCHITECTURE.md:

   - System overview
   - Component interactions
   - Data flow descriptions
   - Performance benchmarks
   - Failure scenarios
   - Recovery procedures
   - Security considerations

4. Code Documentation:
   - Package documentation (doc.go)
   - Function documentation
   - Type documentation
   - Example usage
   - Performance notes
   - Error handling
     </documentation_requirements>

<file_specifications>

1. config.yaml:

```yaml
frontends:
  - port: 8080
  - port: 8443 # SSL port
backends:
  - url: "http://backend1:9001"
    weight: 5
  - url: "http://backend2:9002"
    weight: 3
healthcheck:
  interval: "10s"
  timeout: "2s"
  path: "/health"
ssl:
  certFile: "cert.pem"
  keyFile: "key.pem"
  caFile: "ca.pem"
ratelimit:
  enabled: true
  rate: 1000
  burst: 100
circuitbreaker:
  threshold: 5
  timeout: "30s"
  maxHalfOpen: 3
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
EXPOSE 8080 8443 9090
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
      - "8443:8443"
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
   - run-benchmarks.sh for performance testing
   - Makefile for common operations

2. Development Dependencies:

   - Go 1.20+
   - Docker and Docker Compose
   - golangci-lint
   - go-test-report
   - hey (for load testing)

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
11. Performance targets met
12. Security requirements satisfied
    </completion_criteria>

<performance_requirements>
The implementation must meet these performance targets:

1. Latency:

   - HTTP requests: < 1ms average
   - HTTPS requests: < 2ms average
   - Circuit breaker overhead: < 0.1ms
   - Rate limiter overhead: < 0.1ms

2. Resource Usage:

   - Memory: < 500MB under full load
   - CPU: < 50% per core
   - File descriptors: < 1000 per process
   - Goroutines: < 10000 at peak

3. Scalability:
   - Support 100+ backend servers
   - Handle 10,000 concurrent connections
   - Process 50,000 requests per second
     </performance_requirements>

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

3. Performance Test Verification:

   ```bash
   ./scripts/run-benchmarks.sh
   ```

4. Manual Testing:

   ```bash
   # Start the system
   docker-compose up --build

   # Test load balancing
   curl http://localhost:8080/

   # Check metrics
   curl http://localhost:9090/metrics

   # Verify health checks
   curl http://localhost:8080/health

   # Test SSL
   curl https://localhost:8443/ --cacert ca.pem
   ```

</testing_verification>
