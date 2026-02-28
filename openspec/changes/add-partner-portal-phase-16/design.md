# Design: Partner Portal (Phase 16)

## Architecture

The partner portal follows the same architecture as the admin portal:
- **Next.js 15** App Router with TypeScript
- **Route groups**: `(auth)` for public routes, `(protected)` for authenticated routes
- **Middleware**: Cookie-based session verification (`partner_access_token`)
- **API routes**: Server-side handlers for auth operations (login, logout, refresh)
- **API client**: Axios with auto-refresh interceptor and tenant-aware base URL

## Multi-Restaurant Context

Partners may manage multiple restaurants. The auth store includes:
- `activeRestaurantId`: Currently selected restaurant
- `restaurants`: List of accessible restaurants
- Restaurant picker shown after login when user has multiple restaurants
- Branch switcher in sidebar for quick restaurant switching
- All data-fetching hooks scope queries by `activeRestaurantId`

## Real-Time Features

### SSE for Live Orders
- EventSource connection to `/api/v1/partner/orders/stream`
- Reconnection with exponential backoff
- Custom `useSSE` hook manages connection lifecycle
- Dashboard and order board subscribe to new-order events

### Audio Notifications
- `useAudioNotification` hook plays notification sound on new orders
- Uses Web Audio API with user-gesture activation
- Configurable per notification preferences

## Navigation Structure

```
Sidebar:
├── Dashboard          /
├── Restaurants        /restaurants
├── Menu               /menu
├── Orders             /orders
├── Riders             /riders
├── Promotions         /promotions
├── Finance
│   ├── Summary        /finance
│   ├── Invoices       /finance/invoices
│   └── Payments       /finance/payments
├── Reports            /reports
├── Content
│   ├── Banners        /content/banners
│   ├── Stories        /content/stories
│   └── Sections       /content/sections
├── Team               /team
└── Settings           /settings
```

## Component Patterns

- **UI primitives**: Button, Input, Select, Badge, Card (same as admin)
- **Form pattern**: React Hook Form + Zod schema + `@hookform/resolvers`
- **Data tables**: Server-side filtering/pagination via query params
- **Drawers/Sheets**: Slide-in panels for create/edit forms and detail views
- **Charts**: Recharts (LineChart, BarChart, PieChart, custom heatmap)
- **Drag-drop**: HTML5 native drag-and-drop for category/banner reorder
- **Kanban**: Custom kanban board using CSS grid columns

## API Integration

All partner API calls go through the Axios `apiClient` instance.
Base URL pattern: `NEXT_PUBLIC_API_BASE_URL` (default: `/api/v1`)
Partner endpoints are prefixed with `/partner/`.

### Key Endpoints Used:
- `GET /partner/dashboard` — Dashboard KPIs
- `GET/POST /partner/restaurants` — Restaurant CRUD
- `GET/POST /partner/restaurants/:id/categories` — Categories
- `GET/POST /partner/restaurants/:id/products` — Products
- `GET/PATCH /partner/orders` — Order management
- `GET/POST /partner/riders` — Rider management
- `GET /partner/finance/summary` — Finance summary
- `GET /partner/finance/invoices` — Invoices
- `GET /partner/reports/sales` — Sales reports

## Security

- httpOnly cookies for session tokens
- CSRF protection via `X-Request-ID` headers
- Tenant isolation enforced by backend (all partner endpoints are scoped)
- Restaurant-level access controlled by user's assigned restaurants
- Invitation tokens expire after 72 hours
