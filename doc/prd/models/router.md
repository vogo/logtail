# Model: Router

## Description
Receives log data from Workers, applies matchers to filter lines, and dispatches matched lines to Transfers. Operates as a buffered channel-based pipeline stage.

## Current Struct Definition

```go
type Router struct {
    Lock      sync.Mutex
    Runner    *gorun.Runner
    ID        string
    Name      string
    Source    string
    Channel   chan []byte
    Matchers  []match.Matcher
    Transfers []trans.Transfer
}
```

## Changes (2026-03-21-001)

### New Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `DropCount` | `atomic.Int64` | 0 | Count of messages dropped due to full channel buffer |
| `BufferSize` | `int` | 16 | Channel buffer size (configurable via config) |
| `BlockingMode` | `bool` | false | If true, `Receive()` blocks when channel is full instead of dropping |

### Updated Struct Definition

```go
type Router struct {
    Lock         sync.Mutex
    Runner       *gorun.Runner
    ID           string
    Name         string
    Source       string
    Channel      chan []byte
    Matchers     []match.Matcher
    Transfers    []trans.Transfer
    DropCount    atomic.Int64
    BufferSize   int
    BlockingMode bool
}
```

### Behavior Changes

**`Receive(data []byte)`** -- Updated logic:

```
if BlockingMode:
    select:
        case <-Runner.C: return  (respect shutdown)
        case Channel <- data:    (block until space available)
else (default, backward compatible):
    select:
        case <-Runner.C: return
        case Channel <- data:
        default:
            DropCount.Add(1)     (NEW: increment counter instead of silent drop)
```

**`BuildRouter()`** -- Updated to accept configurable buffer size:
- Reads `BufferSize` from `RouterConfig`; defaults to `DefaultChannelBufferSize` (16) if zero
- Reads `BlockingMode` from `RouterConfig`; defaults to false
