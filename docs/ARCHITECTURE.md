# KubeTools Architecture

## Overview
KubeTools follows a modern client-server architecture with a clear separation between frontend and backend concerns.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Browser                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         React Frontend (Vite + TypeScript)            │  │
│  │                                                        │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────────┐   │  │
│  │  │  Pages   │  │Components│  │ TanStack Query   │   │  │
│  │  └──────────┘  └──────────┘  └──────────────────┘   │  │
│  │                                                        │  │
│  │  ┌───────────────────────────────────────────────┐   │  │
│  │  │         shadcn/ui + Tailwind CSS              │   │  │
│  │  └───────────────────────────────────────────────┘   │  │
│  └────────────────────┬───────────────────────────────┘  │
└───────────────────────┼──────────────────────────────────┘
                        │
                        │ HTTP/REST API
                        │ (JSON)
                        │
┌───────────────────────┼──────────────────────────────────┐
│                       │         Server                    │
│  ┌────────────────────▼─────────────────────────────┐   │
│  │         Go Backend (Fiber/Gin)                    │   │
│  │                                                    │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────┐   │   │
│  │  │ Handlers │  │ Services │  │  K8s Client  │   │   │
│  │  └──────────┘  └──────────┘  └──────┬───────┘   │   │
│  │                                      │            │   │
│  │  ┌────────────────────────────┐     │            │   │
│  │  │       Middleware           │     │            │   │
│  │  │  (CORS, Auth, Logging)     │     │            │   │
│  │  └────────────────────────────┘     │            │   │
│  └──────────────────────────────────────┼───────────┘   │
└────────────────────────────────────────┼────────────────┘
                                         │
                                         │ client-go
                                         │
                        ┌────────────────▼───────────────┐
                        │   Kubernetes Cluster           │
                        │                                │
                        │  ┌──────┐  ┌──────────────┐  │
                        │  │ Pods │  │ Deployments  │  │
                        │  └──────┘  └──────────────┘  │
                        │                                │
                        │  ┌──────────┐  ┌──────────┐  │
                        │  │Services  │  │ConfigMaps│  │
                        │  └──────────┘  └──────────┘  │
                        └────────────────────────────────┘
```

## Frontend Architecture

### Layer Structure
```
┌─────────────────────────────────────────┐
│         Presentation Layer              │
│    (Pages, Components, UI)              │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│         State Management                │
│   (TanStack Query, React State)         │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│         Services Layer                  │
│   (API Clients, Business Logic)         │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│         Network Layer                   │
│        (Axios, HTTP)                    │
└─────────────────────────────────────────┘
```

### Data Flow
1. User interacts with UI Component
2. Component triggers action (button click, form submit)
3. Action calls custom hook or service
4. TanStack Query manages cache and API calls
5. API service makes HTTP request
6. Response data flows back through layers
7. Component re-renders with new data

### Component Hierarchy
```
App
├── Layout
│   ├── Header
│   ├── Sidebar
│   └── Main Content
│       ├── Dashboard (Page)
│       ├── Pods (Page)
│       │   ├── PodList (Feature Component)
│       │   │   ├── PodCard (UI Component)
│       │   │   └── PodFilters (UI Component)
│       │   └── PodDetails (Feature Component)
│       └── Deployments (Page)
└── ErrorBoundary
```

## Backend Architecture

### Layer Structure
```
┌─────────────────────────────────────────┐
│         API Layer                       │
│    (Handlers, Routes, Middleware)       │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│         Service Layer                   │
│    (Business Logic, Validation)         │
└───────────────┬─────────────────────────┘
                │
┌───────────────▼─────────────────────────┐
│         Integration Layer               │
│    (K8s Client, External APIs)          │
└─────────────────────────────────────────┘
```

### Request Flow
1. HTTP request received by router
2. Middleware chain processes request (CORS, auth, logging)
3. Router dispatches to appropriate handler
4. Handler validates input and calls service
5. Service contains business logic
6. Service uses K8s client to interact with cluster
7. Response flows back through layers
8. Handler formats response (JSON)

### Module Structure
```
backend/
├── cmd/server/          # Application entry point
│   └── main.go          # Server initialization
│
├── internal/api/        # HTTP layer
│   ├── handlers/        # Request handlers (thin)
│   ├── middleware/      # HTTP middleware
│   └── routes/          # Route definitions
│
├── internal/services/   # Business logic (thick)
│   └── k8s_service.go
│
├── internal/k8s/        # K8s integration
│   ├── client.go        # Client initialization
│   └── resources.go     # Resource operations
│
└── internal/config/     # Configuration
    └── config.go
```

## API Design

### RESTful Endpoints
```
GET    /api/v1/health              # Health check
GET    /api/v1/namespaces          # List namespaces
GET    /api/v1/pods/:namespace     # List pods in namespace
GET    /api/v1/pods/:namespace/:name # Get pod details
DELETE /api/v1/pods/:namespace/:name # Delete pod
GET    /api/v1/deployments/:namespace # List deployments
```

### Request/Response Format
```json
// Success Response
{
  "success": true,
  "data": { ... },
  "error": null,
  "timestamp": "2025-11-16T12:00:00Z"
}

// Error Response
{
  "success": false,
  "data": null,
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Pod not found"
  },
  "timestamp": "2025-11-16T12:00:00Z"
}
```

## Data Flow Example: Listing Pods

### Frontend
1. User navigates to Pods page
2. `Pods.tsx` component mounts
3. `usePods()` hook is called
4. TanStack Query checks cache
5. If not cached, calls `k8sService.getPods(namespace)`
6. Axios makes GET request to `/api/v1/pods/default`

### Backend
7. Fiber router receives request
8. CORS middleware processes request
9. Logger middleware logs request
10. Handler validates namespace parameter
11. Handler calls `k8sService.ListPods(namespace)`
12. Service uses K8s client to query cluster
13. K8s client returns pod list
14. Service formats data
15. Handler returns JSON response

### Frontend (continued)
16. Axios receives response
17. TanStack Query caches response
18. Data flows to component
19. Component renders pod list
20. UI updates with pod cards

## Security Considerations

### Authentication (Future)
- JWT-based authentication
- Token stored in httpOnly cookies
- Refresh token rotation

### Authorization
- RBAC integration with K8s
- Namespace-based access control
- Action-level permissions

### CORS
- Configured allowed origins
- Credentials support
- Preflight handling

### Input Validation
- Backend: Validate all inputs
- Frontend: Form validation with Zod
- Sanitize user inputs

## Scalability Considerations

### Frontend
- Code splitting by route
- Virtual scrolling for large lists
- Image optimization
- Asset caching

### Backend
- Stateless design
- Connection pooling
- Rate limiting
- Horizontal scaling ready

## Deployment Architecture

### Development
```
Frontend: localhost:5173 (Vite dev server)
Backend:  localhost:8080 (Go binary)
K8s:      Local cluster (minikube/kind)
```

### Production (Future)
```
Frontend: CDN (Vercel/Netlify)
Backend:  Container (Docker + K8s)
K8s:      Target cluster(s)
```

## Error Handling

### Frontend
- Error boundaries for React errors
- TanStack Query error handling
- User-friendly error messages
- Retry logic for failed requests

### Backend
- Centralized error handling
- Structured logging
- Error wrapping with context
- Proper HTTP status codes

## Monitoring & Observability (Future)

### Metrics
- Request latency
- Error rates
- K8s API call latency
- Resource usage

### Logging
- Structured JSON logs
- Request/response logging
- Error tracking
- Correlation IDs

### Tracing
- Distributed tracing
- Request flow visualization
- Performance bottleneck identification
