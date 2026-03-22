# logtail - Applications and Pages

N/A -- logtail is a CLI tool and library. It exposes an HTTP API for runtime management but has no dedicated UI application.

## API Endpoints (Existing)

- `GET /` -- Server list
- `GET /index/<server-id>` -- Server index page (minimal HTML)
- `GET /tail/<server-id>` -- WebSocket tailing
- `POST /manage/<operation>` -- Runtime management (add/remove servers, transfers, routers)

## API Changes (2026-03-21-001)

### New Endpoint: Pipeline Statistics

- `GET /manage/stats` -- Returns JSON with per-router pipeline statistics including drop counts

**Response Example:**
```json
{
  "routers": {
    "router-1": {
      "drop_count": 42,
      "buffer_size": 64,
      "blocking_mode": false
    }
  }
}
```

No other API changes required. The existing management endpoints continue to work as before.
