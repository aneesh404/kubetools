# Getting Started with KubeTools Development

## Overview
This guide will help you set up your development environment and start contributing to KubeTools.

## Prerequisites

### Required
- **Go**: Version 1.21 or higher
  ```bash
  go version  # Should show 1.21+
  ```
- **Node.js**: Version 18 or higher
  ```bash
  node --version  # Should show v18+
  ```
- **kubectl**: Kubernetes CLI tool
  ```bash
  kubectl version --client
  ```
- **Git**: Version control

### Recommended
- **Docker Desktop**: For local Kubernetes cluster
- **minikube** or **kind**: Alternative local K8s options
- **Air**: Go hot-reload tool
  ```bash
  go install github.com/cosmtrek/air@latest
  ```

## Initial Setup

### 1. Clone the Repository
```bash
git clone <repository-url>
cd kubetools
```

### 2. Set Up Kubernetes Access
Ensure you have a kubeconfig file configured:
```bash
kubectl cluster-info
kubectl get namespaces
```

If you need a local cluster:
```bash
# Using minikube
minikube start

# Or using kind
kind create cluster
```

## Backend Setup

### 1. Navigate to Backend Directory
```bash
cd backend
```

### 2. Install Dependencies
```bash
# Download and install Go dependencies
go mod download
go mod tidy
```

### 3. Install Core Dependencies
```bash
# Web framework (choose one)
go get github.com/gofiber/fiber/v2
# OR
# go get github.com/gin-gonic/gin

# Kubernetes client
go get k8s.io/client-go@latest
go get k8s.io/apimachinery@latest
go get k8s.io/api@latest

# Configuration and logging
go get github.com/spf13/viper
go get go.uber.org/zap
go get github.com/joho/godotenv
```

### 4. Configure Environment
```bash
cp .env.example .env
```

Edit `.env` with your settings:
```bash
PORT=8080
KUBECONFIG=/path/to/your/.kube/config
LOG_LEVEL=debug
CORS_ORIGINS=http://localhost:5173
```

### 5. Run Backend
```bash
# Direct run
go run cmd/server/main.go

# Or with hot reload (if air is installed)
air
```

Backend should now be running at `http://localhost:8080`

## Frontend Setup

### 1. Navigate to Frontend Directory
```bash
cd frontend
```

### 2. Install Dependencies
```bash
npm install
```

### 3. Initialize shadcn/ui
```bash
npx shadcn-ui@latest init
```

Follow the prompts:
- TypeScript: Yes
- Style: Default
- Base color: Slate (or your preference)
- CSS variables: Yes

### 4. Install Initial UI Components
```bash
npx shadcn-ui@latest add button
npx shadcn-ui@latest add card
npx shadcn-ui@latest add dialog
npx shadcn-ui@latest add input
npx shadcn-ui@latest add label
npx shadcn-ui@latest add table
npx shadcn-ui@latest add badge
npx shadcn-ui@latest add skeleton
```

### 5. Install Additional Dependencies
```bash
# API and state management
npm install @tanstack/react-query axios

# Routing
npm install react-router-dom

# Forms and validation
npm install react-hook-form zod @hookform/resolvers

# Icons
npm install lucide-react
```

### 6. Configure Environment
```bash
cp .env.example .env.local
```

Edit `.env.local`:
```bash
VITE_API_URL=http://localhost:8080/api/v1
VITE_APP_NAME=KubeTools
```

### 7. Run Frontend
```bash
npm run dev
```

Frontend should now be running at `http://localhost:5173`

## Verify Installation

### Backend Health Check
```bash
curl http://localhost:8080/api/v1/health
```

Expected response:
```json
{
  "success": true,
  "data": {
    "status": "healthy"
  }
}
```

### Frontend Access
Open your browser to `http://localhost:5173` and verify the app loads.

## Development Workflow

### Running Both Applications
Open two terminal windows:

**Terminal 1 (Backend):**
```bash
cd backend
go run cmd/server/main.go
# Or: air (for hot reload)
```

**Terminal 2 (Frontend):**
```bash
cd frontend
npm run dev
```

### Making Changes

#### Backend Changes
1. Edit Go files in `backend/internal/`
2. If using `air`, changes auto-reload
3. Otherwise, restart the server
4. Test your changes with curl or the frontend

#### Frontend Changes
1. Edit files in `frontend/src/`
2. Vite will hot-reload automatically
3. Check browser for updates
4. Use React DevTools for debugging

### Testing

#### Backend Tests
```bash
cd backend
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/services/...
```

#### Frontend Tests
```bash
cd frontend
npm test

# With coverage
npm test -- --coverage
```

### Code Quality

#### Backend Linting
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

#### Frontend Linting
```bash
npm run lint

# Auto-fix issues
npm run lint -- --fix
```

## Common Tasks

### Adding a New API Endpoint

1. Define the handler in `backend/internal/api/handlers/`
2. Add the route in `backend/internal/api/routes/`
3. Implement business logic in `backend/internal/services/`
4. Add corresponding service call in frontend `services/`
5. Create a custom hook if needed in `frontend/hooks/`
6. Use the hook in your component

### Adding a New Page

1. Create page component in `frontend/src/pages/`
2. Add route in `App.tsx`
3. Add navigation link in sidebar/header
4. Create necessary API services
5. Build feature-specific components

### Working with Kubernetes Resources

1. Use `k8s.io/client-go` in backend
2. Create service methods in `backend/internal/k8s/`
3. Expose via API handlers
4. Consume in frontend via TanStack Query

## Troubleshooting

### Backend Issues

**Port already in use:**
```bash
lsof -ti:8080 | xargs kill -9
```

**Kubeconfig not found:**
```bash
export KUBECONFIG=~/.kube/config
# Or set in .env file
```

**Dependencies issues:**
```bash
go mod tidy
go clean -modcache
go mod download
```

### Frontend Issues

**Port 5173 in use:**
Edit `vite.config.ts` to change port:
```typescript
export default defineConfig({
  server: {
    port: 5174
  }
})
```

**Module not found:**
```bash
rm -rf node_modules package-lock.json
npm install
```

**Vite cache issues:**
```bash
rm -rf node_modules/.vite
npm run dev
```

## Next Steps

1. Read through `CLAUDE.md` for project context
2. Review `ARCHITECTURE.md` for system design
3. Check `backend/CLAUDE.md` for backend guidelines
4. Check `frontend/CLAUDE.md` for frontend guidelines
5. Start with a small feature or bug fix
6. Write tests for your changes
7. Submit a pull request

## Useful Commands

### Backend
```bash
go run cmd/server/main.go    # Run server
go test ./...                 # Run tests
go build -o bin/server cmd/server/main.go  # Build
golangci-lint run            # Lint
go mod tidy                  # Clean dependencies
```

### Frontend
```bash
npm run dev                  # Dev server
npm run build                # Production build
npm run preview              # Preview build
npm test                     # Run tests
npm run lint                 # Lint code
```

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Kubernetes client-go](https://github.com/kubernetes/client-go)
- [Fiber Framework](https://gofiber.io/)
- [React Documentation](https://react.dev/)
- [shadcn/ui](https://ui.shadcn.com/)
- [TanStack Query](https://tanstack.com/query/latest)
- [Vite](https://vitejs.dev/)

## Getting Help

- Check existing documentation in `docs/`
- Review CLAUDE.md files for context
- Search GitHub issues
- Ask in team chat/discussions

Happy coding!
