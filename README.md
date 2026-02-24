# KubeTools

A comprehensive web application designed to make Kubernetes developers' lives easier by providing essential tools and utilities in one accessible platform.

## Project Structure

This is a monorepo containing both frontend and backend applications:

- **Backend** (`./backend`): Go-based REST API server
- **Frontend** (`./frontend`): React + TypeScript web application with shadcn/ui

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- kubectl and kubeconfig configured

### Backend Setup
```bash
cd backend
go mod download
cp .env.example .env
go run cmd/server/main.go
```

Backend runs on `http://localhost:8080`

### Frontend Setup
```bash
cd frontend
npm install
cp .env.example .env.local
npm run dev
```

Frontend runs on `http://localhost:5173`

## Documentation

For detailed information about each component:
- [Main Documentation](./CLAUDE.md)
- [Backend Documentation](./backend/CLAUDE.md)
- [Frontend Documentation](./frontend/CLAUDE.md)

## Tech Stack

### Backend
- Go 1.21+
- Fiber/Gin (web framework)
- client-go (Kubernetes client)
- Viper (configuration)
- Zap (logging)

### Frontend
- React 18+ with TypeScript
- Vite (build tool)
- shadcn/ui (UI components)
- TanStack Query (data fetching)
- Tailwind CSS (styling)

## Development

### Running Tests
```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && npm test
```

### Building for Production
```bash
# Backend
cd backend && go build -o bin/server cmd/server/main.go

# Frontend
cd frontend && npm run build
```

## Contributing

1. Create a feature branch
2. Make your changes
3. Write/update tests
4. Submit a pull request

## License

MIT
