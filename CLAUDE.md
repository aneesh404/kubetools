# KubeTools - Project Overview

## Project Description
KubeTools is a comprehensive web application designed to make Kubernetes developers' lives easier by providing essential tools and utilities in one accessible platform.

## Architecture
This is a monorepo containing both frontend and backend applications:

### Backend (Go)
- **Location**: `./backend`
- **Language**: Go
- **Framework**: TBD (Fiber or Gin recommended)
- **Primary Libraries**:
  - client-go: Kubernetes client library
  - Viper: Configuration management
  - Zap/Logrus: Logging
  - Swagger: API documentation

### Frontend (React + shadcn)
- **Location**: `./frontend`
- **Framework**: React with TypeScript
- **Build Tool**: Vite
- **UI Library**: shadcn/ui
- **State Management**: TanStack Query for API state
- **Styling**: Tailwind CSS (via shadcn)

## Project Structure
```
kubetools/
├── backend/              # Go backend application
│   ├── cmd/
│   │   └── server/      # Main application entry point
│   ├── internal/        # Private application code
│   │   ├── api/         # API layer
│   │   │   ├── handlers/   # HTTP request handlers
│   │   │   ├── middleware/ # HTTP middleware
│   │   │   └── routes/     # Route definitions
│   │   ├── config/      # Configuration management
│   │   ├── k8s/         # Kubernetes client logic
│   │   ├── models/      # Data models
│   │   ├── services/    # Business logic
│   │   └── utils/       # Utility functions
│   └── pkg/             # Public libraries (if needed)
├── frontend/            # React frontend application
│   ├── src/
│   │   ├── components/  # React components
│   │   │   ├── ui/         # shadcn UI components
│   │   │   └── features/   # Feature-specific components
│   │   ├── hooks/       # Custom React hooks
│   │   ├── lib/         # Utility libraries
│   │   ├── pages/       # Page components
│   │   ├── services/    # API client services
│   │   ├── types/       # TypeScript type definitions
│   │   └── styles/      # Global styles
│   └── public/          # Static assets
├── docs/                # Project documentation
├── scripts/             # Build and deployment scripts
└── .github/workflows/   # CI/CD workflows
```

## Development Workflow

### Backend Development
1. Navigate to `./backend`
2. Run `go mod tidy` to install dependencies
3. Run `go run cmd/server/main.go` to start the server
4. Backend typically runs on `http://localhost:8080`

### Frontend Development
1. Navigate to `./frontend`
2. Run `npm install` to install dependencies
3. Run `npm run dev` to start development server
4. Frontend typically runs on `http://localhost:5173`

## Key Conventions

### Go Code Style
- Follow standard Go conventions and `gofmt` formatting
- Use meaningful package names
- Keep handlers thin, move business logic to services
- Use dependency injection for testability

### React Code Style
- Use functional components with hooks
- Follow shadcn/ui patterns for component structure
- Organize components by feature when they're feature-specific
- Keep reusable UI components in `components/ui/`

### API Conventions
- RESTful API design
- Use proper HTTP status codes
- Version API endpoints (e.g., `/api/v1/...`)
- Document all endpoints with OpenAPI/Swagger

### Git Workflow
- Use conventional commits (feat:, fix:, docs:, etc.)
- Feature branches from main
- PR-based workflow

## Environment Variables

### Backend
- `PORT`: Server port (default: 8080)
- `KUBECONFIG`: Path to kubeconfig file
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `CORS_ORIGINS`: Allowed CORS origins

### Frontend
- `VITE_API_URL`: Backend API URL
- `VITE_APP_NAME`: Application name

## Getting Started
See individual README files in `backend/` and `frontend/` directories for detailed setup instructions.

## Testing
- Backend: `go test ./...` in backend directory
- Frontend: `npm test` in frontend directory

## Building for Production
- Backend: `go build -o bin/server cmd/server/main.go`
- Frontend: `npm run build`

## Additional Resources
- [Kubernetes client-go](https://github.com/kubernetes/client-go)
- [shadcn/ui Documentation](https://ui.shadcn.com)
- [Go Best Practices](https://go.dev/doc/effective_go)
- [React Best Practices](https://react.dev/learn)
