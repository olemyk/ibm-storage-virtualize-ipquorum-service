# IP Quorum Management Platform - Web Dashboard

Modern React + TypeScript web dashboard for managing IBM Storage Virtualize IP Quorum instances.

## Features

- 🎨 Modern, responsive UI with Tailwind CSS
- 📊 Real-time instance monitoring
- 🔐 Secure authentication with JWT
- 📈 Interactive charts and metrics
- 🚀 Fast development with Vite
- 💪 Type-safe with TypeScript
- 🔄 Real-time updates with React Query

## Tech Stack

- **Framework:** React 18
- **Language:** TypeScript
- **Build Tool:** Vite
- **Styling:** Tailwind CSS
- **State Management:** Zustand
- **Data Fetching:** TanStack React Query
- **HTTP Client:** Axios
- **Routing:** React Router v6
- **Charts:** Recharts
- **Date Handling:** date-fns

## Prerequisites

- Node.js 18+ and npm/yarn/pnpm
- Running IP Quorum Management Platform API (port 8080)

## Installation

```bash
# Install dependencies
npm install

# or with yarn
yarn install

# or with pnpm
pnpm install
```

## Development

```bash
# Start development server (port 3000)
npm run dev

# Type checking
npm run type-check

# Linting
npm run lint
```

The development server will proxy API requests to `http://localhost:8080`.

## Building for Production

```bash
# Build for production
npm run build

# Preview production build
npm run preview
```

The built files will be in the `dist/` directory.

## Project Structure

```
web/
├── src/
│   ├── components/       # Reusable UI components
│   │   ├── layout/      # Layout components (Header, Sidebar, etc.)
│   │   ├── instances/   # Instance-related components
│   │   ├── auth/        # Authentication components
│   │   └── common/      # Common UI components (Button, Card, etc.)
│   ├── pages/           # Page components
│   │   ├── Dashboard.tsx
│   │   ├── Instances.tsx
│   │   ├── Login.tsx
│   │   └── Settings.tsx
│   ├── services/        # API services
│   │   ├── api.ts       # Axios instance
│   │   ├── auth.ts      # Authentication service
│   │   └── instances.ts # Instance management service
│   ├── types/           # TypeScript type definitions
│   │   ├── instance.ts
│   │   ├── auth.ts
│   │   └── api.ts
│   ├── utils/           # Utility functions
│   │   ├── format.ts    # Formatting helpers
│   │   └── validation.ts # Validation helpers
│   ├── styles/          # Global styles
│   │   └── index.css
│   ├── App.tsx          # Main App component
│   └── main.tsx         # Entry point
├── public/              # Static assets
├── index.html           # HTML template
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

## Key Features

### 1. Dashboard
- Overview of all instances
- Health status summary
- Recent activity
- Quick actions

### 2. Instance Management
- List all instances
- Create new instances
- Start/Stop/Restart instances
- View instance details
- Monitor instance health
- View logs

### 3. Authentication
- Secure login
- JWT token management
- Auto-refresh tokens
- Role-based access control

### 4. Monitoring
- Real-time health checks
- Performance metrics
- Historical data
- Alert notifications

### 5. Settings
- User profile
- System configuration
- API settings
- Theme preferences

## API Integration

The dashboard communicates with the backend API at `http://localhost:8080/api/v1`.

### Authentication

```typescript
// Login
POST /api/v1/auth/login
{
  "username": "admin",
  "password": "changeme"
}

// Response
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "username": "admin",
    "role": "admin"
  }
}
```

### Instance Operations

```typescript
// List instances
GET /api/v1/instances

// Create instance
POST /api/v1/instances

// Start instance
POST /api/v1/instances/:id/start

// Get instance status
GET /api/v1/instances/:id/status
```

## State Management

### Authentication State (Zustand)

```typescript
interface AuthState {
  user: User | null
  token: string | null
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  isAuthenticated: boolean
}
```

### Data Fetching (React Query)

```typescript
// Fetch instances
const { data, isLoading, error } = useQuery({
  queryKey: ['instances'],
  queryFn: fetchInstances,
  refetchInterval: 30000, // Refetch every 30s
})
```

## Styling

### Tailwind CSS

The dashboard uses Tailwind CSS for styling with a custom theme:

```javascript
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      colors: {
        primary: '#0066cc',
        secondary: '#6c757d',
        success: '#28a745',
        danger: '#dc3545',
        warning: '#ffc107',
      },
    },
  },
}
```

### Component Example

```tsx
<div className="bg-white rounded-lg shadow-md p-6">
  <h2 className="text-2xl font-bold text-gray-800 mb-4">
    Instances
  </h2>
  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
    {instances.map(instance => (
      <InstanceCard key={instance.id} instance={instance} />
    ))}
  </div>
</div>
```

## Environment Variables

Create a `.env` file in the web directory:

```env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
VITE_REFRESH_INTERVAL=30000
```

## Testing

```bash
# Run tests
npm test

# Run tests with coverage
npm test -- --coverage

# Run tests in watch mode
npm test -- --watch
```

## Deployment

### Docker

```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Static Hosting

The built files in `dist/` can be served by any static file server:

- Nginx
- Apache
- Caddy
- Netlify
- Vercel
- GitHub Pages

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Performance

- Code splitting with React.lazy()
- Image optimization
- Lazy loading
- Service Worker for offline support
- Gzip compression

## Security

- XSS protection
- CSRF tokens
- Secure HTTP headers
- Content Security Policy
- JWT token storage in httpOnly cookies (recommended)

## Accessibility

- WCAG 2.1 Level AA compliance
- Keyboard navigation
- Screen reader support
- High contrast mode
- Focus indicators

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## License

Same as the main project

## Support

For issues and questions:
- GitHub Issues
- Documentation
- API Specification

## Roadmap

- [ ] WebSocket support for real-time updates
- [ ] Dark mode
- [ ] Multi-language support (i18n)
- [ ] Advanced filtering and search
- [ ] Bulk operations
- [ ] Export data (CSV, JSON)
- [ ] Custom dashboards
- [ ] Mobile app (React Native)

## Credits

Built with ❤️ using modern web technologies.