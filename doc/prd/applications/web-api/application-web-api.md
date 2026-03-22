# Web API Application

## Application Overview
HTTP-based management interface and real-time log streaming service. Provides REST endpoints for runtime configuration of the log pipeline and WebSocket endpoints for live log tailing.

## Target User Roles
- Operator (any caller with network access; no authentication)

## Page Framework

### Navigation Structure
Minimal HTML pages with direct URL navigation. No formal navigation pattern — pages are accessed via URL paths.

### Page Map

| Page | Path | Description | Module |
|------|------|-------------|--------|
| Server List | `/` | Lists all active servers with links | Management |
| Server View | `/index/{server-id}` | Single server index page with embedded UI | Management |
| Log Stream | `/tail/{server-id}` | WebSocket real-time log streaming | Management |
| Management Console | `/manage` | Runtime configuration management interface | Management |

### API Endpoints

| Endpoint | Method | Description | Module |
|----------|--------|-------------|--------|
| `/manage/server/types` | GET | List available server types | Management |
| `/manage/server/list` | GET | List all configured servers | Management |
| `/manage/server/add` | POST | Add a new file-watch server | Management |
| `/manage/server/delete` | POST | Remove a server by name | Management |
| `/manage/router/list` | GET | List all configured routers | Management |
| `/manage/router/add` | POST | Create a new router | Management |
| `/manage/router/delete` | POST | Remove a router by name | Management |
| `/manage/transfer/types` | GET | List available transfer types | Management |
| `/manage/transfer/list` | GET | List all configured transfers | Management |
| `/manage/transfer/add` | POST | Create a new transfer | Management |
| `/manage/transfer/delete` | POST | Remove a transfer by name | Management |
| `/manage/stats` | GET | Get pipeline statistics (per-router drop counts) | Management |

### Global Components
None — minimal HTML pages with no shared UI framework.
