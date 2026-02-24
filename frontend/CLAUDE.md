# Frontend - KubeTools Web Application

## Overview
Modern React-based web application for Kubernetes tooling, built with TypeScript, Vite, and shadcn/ui.

## Tech Stack
- **Framework**: React 18+ with TypeScript
- **Build Tool**: Vite
- **UI Components**: shadcn/ui
- **Styling**: Tailwind CSS
- **State Management**: TanStack Query (React Query) for server state
- **Routing**: React Router v6
- **Forms**: React Hook Form + Zod validation
- **HTTP Client**: Axios
- **Icons**: Lucide React

## Directory Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── ui/              # shadcn/ui components
│   │   │   ├── button.tsx
│   │   │   ├── card.tsx
│   │   │   ├── dialog.tsx
│   │   │   └── ...
│   │   ├── features/        # Feature-specific components
│   │   │   ├── pods/
│   │   │   ├── deployments/
│   │   │   └── namespaces/
│   │   └── layout/          # Layout components
│   │       ├── Header.tsx
│   │       ├── Sidebar.tsx
│   │       └── Layout.tsx
│   ├── hooks/               # Custom React hooks
│   │   ├── useApi.ts
│   │   ├── useK8s.ts
│   │   └── useDebounce.ts
│   ├── lib/                 # Utility libraries
│   │   ├── api.ts           # Axios instance
│   │   ├── utils.ts         # Helper functions
│   │   └── constants.ts
│   ├── pages/               # Page components
│   │   ├── Dashboard.tsx
│   │   ├── Pods.tsx
│   │   └── NotFound.tsx
│   ├── services/            # API client services
│   │   ├── k8sService.ts
│   │   └── types.ts
│   ├── types/               # TypeScript types
│   │   ├── k8s.ts
│   │   └── api.ts
│   ├── styles/              # Global styles
│   │   └── globals.css
│   ├── App.tsx              # Root component
│   ├── main.tsx             # Entry point
│   └── vite-env.d.ts        # Vite types
├── public/                  # Static assets
│   ├── favicon.ico
│   └── robots.txt
├── index.html               # HTML template
├── package.json             # Dependencies
├── tsconfig.json            # TypeScript config
├── tsconfig.node.json       # TypeScript config for Vite
├── vite.config.ts           # Vite configuration
├── tailwind.config.js       # Tailwind configuration
├── postcss.config.js        # PostCSS configuration
├── components.json          # shadcn/ui config
├── .env.example             # Example environment variables
└── .eslintrc.cjs            # ESLint configuration
```

## Setup & Installation

### Prerequisites
- Node.js 18+ and npm/pnpm/yarn

### Initial Setup
```bash
cd frontend
npm install
```

### Development Server
```bash
npm run dev
# Opens at http://localhost:5173
```

### Build for Production
```bash
npm run build
# Output in ./dist
```

### Preview Production Build
```bash
npm run preview
```

## Configuration

### Environment Variables
Create a `.env.local` file:
```bash
VITE_API_URL=http://localhost:8080/api/v1
VITE_APP_NAME=KubeTools
VITE_APP_VERSION=1.0.0
```

### shadcn/ui Setup
Components are initialized with:
```bash
npx shadcn-ui@latest init
```

Add components as needed:
```bash
npx shadcn-ui@latest add button
npx shadcn-ui@latest add card
npx shadcn-ui@latest add dialog
```

## Key Patterns

### API Integration
Use TanStack Query for all API calls:

```typescript
// In services/k8sService.ts
export const k8sService = {
  getPods: (namespace: string) =>
    api.get(`/pods/${namespace}`)
}

// In components
const { data, isLoading, error } = useQuery({
  queryKey: ['pods', namespace],
  queryFn: () => k8sService.getPods(namespace)
})
```

### Component Structure
```typescript
// Feature component example
interface PodListProps {
  namespace: string
}

export function PodList({ namespace }: PodListProps) {
  const { data: pods, isLoading } = usePods(namespace)

  if (isLoading) return <Skeleton />

  return (
    <div className="space-y-4">
      {pods?.map(pod => (
        <PodCard key={pod.name} pod={pod} />
      ))}
    </div>
  )
}
```

### Routing
```typescript
// App.tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="pods" element={<Pods />} />
          <Route path="deployments" element={<Deployments />} />
          <Route path="*" element={<NotFound />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
```

### State Management
- **Server State**: TanStack Query
- **UI State**: React useState/useReducer
- **Global State**: Context API (if needed)
- **Form State**: React Hook Form

### Styling with Tailwind
Follow shadcn/ui patterns:
```typescript
import { cn } from '@/lib/utils'

function Component({ className }: { className?: string }) {
  return (
    <div className={cn(
      "default-classes",
      className
    )}>
      Content
    </div>
  )
}
```

## Code Organization

### Component Guidelines
1. **UI Components** (`components/ui/`): Pure, reusable UI elements from shadcn
2. **Feature Components** (`components/features/`): Business logic components
3. **Page Components** (`pages/`): Route-level components
4. **Layout Components** (`components/layout/`): App structure components

### File Naming
- Components: PascalCase (e.g., `PodList.tsx`)
- Utilities: camelCase (e.g., `formatDate.ts`)
- Hooks: camelCase with 'use' prefix (e.g., `usePods.ts`)
- Types: PascalCase (e.g., `Pod.ts`)

### Import Order
1. External dependencies (react, etc.)
2. Internal components
3. Hooks
4. Utils/lib
5. Types
6. Styles

## TypeScript Guidelines

### Type Definitions
```typescript
// types/k8s.ts
export interface Pod {
  name: string
  namespace: string
  status: PodStatus
  createdAt: string
}

export type PodStatus = 'Running' | 'Pending' | 'Failed' | 'Succeeded'
```

### Props Types
```typescript
interface ComponentProps {
  required: string
  optional?: number
  children?: React.ReactNode
}
```

## Testing

### Unit Tests (Vitest)
```bash
npm run test
```

### E2E Tests (Playwright - future)
```bash
npm run test:e2e
```

## Dependencies to Install

### Core
```json
{
  "react": "^18.2.0",
  "react-dom": "^18.2.0",
  "react-router-dom": "^6.20.0"
}
```

### UI & Styling
```json
{
  "@radix-ui/react-*": "latest",
  "tailwindcss": "^3.4.0",
  "class-variance-authority": "^0.7.0",
  "clsx": "^2.0.0",
  "tailwind-merge": "^2.0.0",
  "lucide-react": "^0.300.0"
}
```

### Data Fetching & Forms
```json
{
  "@tanstack/react-query": "^5.17.0",
  "axios": "^1.6.0",
  "react-hook-form": "^7.49.0",
  "zod": "^3.22.0",
  "@hookform/resolvers": "^3.3.0"
}
```

### Dev Dependencies
```json
{
  "@types/react": "^18.2.0",
  "@types/react-dom": "^18.2.0",
  "@typescript-eslint/eslint-plugin": "^6.0.0",
  "@typescript-eslint/parser": "^6.0.0",
  "@vitejs/plugin-react": "^4.2.0",
  "typescript": "^5.3.0",
  "vite": "^5.0.0",
  "eslint": "^8.56.0",
  "autoprefixer": "^10.4.0",
  "postcss": "^8.4.0"
}
```

## Best Practices

### Performance
- Use React.memo for expensive components
- Implement virtual scrolling for large lists
- Code splitting with lazy loading
- Optimize images (WebP format)
- Use useMemo/useCallback appropriately

### Accessibility
- Use semantic HTML
- Add ARIA labels where needed
- Ensure keyboard navigation
- Maintain proper heading hierarchy
- Test with screen readers

### Error Handling
```typescript
// Use error boundaries
<ErrorBoundary fallback={<ErrorFallback />}>
  <Component />
</ErrorBoundary>

// Handle query errors
const { error } = useQuery(...)
if (error) return <ErrorMessage error={error} />
```

### Code Quality
- Run ESLint before commits
- Use TypeScript strict mode
- Write meaningful component names
- Keep components small and focused
- Document complex logic

## Future Enhancements
- [ ] Dark mode support
- [ ] Internationalization (i18n)
- [ ] PWA capabilities
- [ ] Advanced filtering & search
- [ ] Real-time updates (WebSocket)
- [ ] Export functionality
- [ ] User preferences persistence
- [ ] Keyboard shortcuts
