# Backend - KubeTools API

## Overview
Go-based REST API server providing Kubernetes tooling capabilities.

## Tech Stack
- **Language**: Go 1.21+
- **Framework**: Fiber (recommended) or Gin
- **K8s Client**: client-go
- **Config**: Viper
- **Logging**: Zap
- **Validation**: go-playground/validator
- **API Docs**: Swagger/OpenAPI

## Directory Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/                 # Private application code
│   ├── api/
│   │   ├── handlers/         # HTTP request handlers
│   │   │   ├── health.go
│   │   │   └── k8s.go
│   │   ├── middleware/       # HTTP middleware
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   └── logger.go
│   │   └── routes/           # Route definitions
│   │       └── routes.go
│   ├── config/               # Configuration
│   │   └── config.go
│   ├── k8s/                  # Kubernetes client logic
│   │   ├── client.go
│   │   └── resources.go
│   ├── models/               # Data models/structs
│   │   ├── request.go
│   │   └── response.go
│   ├── services/             # Business logic
│   │   └── k8s_service.go
│   └── utils/                # Utility functions
│       ├── logger.go
│       └── errors.go
├── pkg/                      # Public libraries (if needed)
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
├── .env.example              # Example environment variables
└── Dockerfile                # Container image definition
```

## Key Components

### cmd/server/main.go
- Application entry point
- Server initialization
- Graceful shutdown handling

### internal/api/handlers/
- HTTP request handlers
- Input validation
- Response formatting
- Keep handlers thin - delegate to services

### internal/api/middleware/
- **CORS**: Cross-origin resource sharing
- **Logger**: Request/response logging
- **Auth**: Authentication/authorization (future)
- **Recovery**: Panic recovery

### internal/k8s/
- Kubernetes client initialization
- Kubeconfig loading
- Resource operations (CRUD)
- Client caching/connection pooling

### internal/services/
- Business logic layer
- Orchestrates multiple operations
- Error handling and validation
- Independent of HTTP layer

## Configuration

### Environment Variables
```bash
# Server
PORT=8080
HOST=localhost
ENV=development

# Kubernetes
KUBECONFIG=/path/to/kubeconfig
K8S_CONTEXT=default

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# CORS
CORS_ORIGINS=http://localhost:5173
CORS_CREDENTIALS=true
```

## API Structure

### Versioning
All API endpoints should be versioned:
- Base path: `/api/v1`

### Standard Response Format
```json
{
  "success": true,
  "data": {},
  "error": null,
  "timestamp": "2025-11-16T12:00:00Z"
}
```

### Error Response Format
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Pod not found in namespace default"
  },
  "timestamp": "2025-11-16T12:00:00Z"
}
```

## Development

### Setup
```bash
cd backend
go mod download
go mod tidy
```

### Run Development Server
```bash
go run cmd/server/main.go
```

### Run with Hot Reload (using air)
```bash
air
```

### Testing
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/services/...
```

### Linting
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

## Building

### Development Build
```bash
go build -o bin/server cmd/server/main.go
```

### Production Build
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" \
  -o bin/server \
  cmd/server/main.go
```

## Dependencies to Add

### Core
```bash
go get github.com/gofiber/fiber/v2  # or github.com/gin-gonic/gin
go get github.com/spf13/viper
go get go.uber.org/zap
go get github.com/go-playground/validator/v10
```

### Kubernetes
```bash
go get k8s.io/client-go
go get k8s.io/apimachinery
go get k8s.io/api
```

### Utilities
```bash
go get github.com/joho/godotenv
go get github.com/google/uuid
```

## Code Conventions

### Error Handling
- Always wrap errors with context
- Use custom error types for business logic errors
- Log errors at appropriate levels
- Return meaningful error messages to clients

### Logging
- Use structured logging (JSON format)
- Include request IDs in logs
- Log at appropriate levels (debug, info, warn, error)
- Don't log sensitive information

### Context Usage
- Always pass context through function calls
- Use context for cancellation and timeouts
- Store request-scoped data in context

### Testing
- Write unit tests for services
- Use table-driven tests
- Mock external dependencies (K8s client)
- Aim for >80% coverage

## Best Practices

1. **Dependency Injection**: Pass dependencies explicitly
2. **Interface Usage**: Define interfaces for services
3. **Error Wrapping**: Use `fmt.Errorf("context: %w", err)`
4. **Nil Checks**: Always check for nil pointers
5. **Defer Usage**: Use defer for cleanup operations
6. **Goroutine Safety**: Ensure thread-safe operations
7. **Resource Cleanup**: Always close resources (defer close())

## Future Enhancements
- [ ] Authentication & Authorization
- [ ] Rate limiting
- [ ] Request validation
- [ ] API documentation (Swagger)
- [ ] Metrics & monitoring (Prometheus)
- [ ] Health checks
- [ ] Database integration (if needed)
- [ ] Caching layer (Redis)
