# Process: Data Pipeline (Router Receive and Dispatch)

## Current Flow (Before Change)

```
Worker produces data
  |-> Router.Receive(data)
  |     |-> select:
  |     |     case <-Runner.C: return          // stopped
  |     |     case Channel <- data:            // buffered (cap=16, fixed)
  |     |     default:                         // SILENT DROP -- no logging, no counter
  |     \-> (recover from panic on closed channel)
  |
  |-> (consumer goroutine reads from Channel)
  |     |-> Router.Route(data)
  |     |     |-> Matchers filter
  |     |     \-> Trans(data) to all Transfers
```

**Problems:**
1. Channel buffer size is hardcoded to 16
2. Data drops are completely silent -- no counter, no log
3. No option to block instead of drop for reliability-critical use cases

## Target Flow (After Change)

```
Worker produces data
  |-> Router.Receive(data)
  |     |-> if BlockingMode:
  |     |     select:
  |     |       case <-Runner.C: return
  |     |       case Channel <- data:          // blocks until space
  |     |   else:
  |     |     select:
  |     |       case <-Runner.C: return
  |     |       case Channel <- data:
  |     |       default:
  |     |         DropCount.Add(1)             // atomic increment
  |     \-> (recover from panic on closed channel)
  |
  |-> (consumer goroutine reads from Channel)
  |     |-> Router.Route(data)
  |     |     |-> Matchers filter
  |     |     \-> Trans(data) to all Transfers
```

**Changes:**
1. Channel created with configurable `BufferSize` (default 16 for backward compatibility)
2. `BlockingMode` controls whether Receive blocks or drops
3. `DropCount` (atomic.Int64) incremented on every drop
4. Drop statistics queryable via web API endpoint

## Configuration Example

```yaml
routers:
  my-router:
    matchers:
      - contains: ["ERROR"]
    transfers: ["webhook-1"]
    buffer_size: 64
    blocking_mode: false
```

## Observability

Drop counts are exposed:
- Via web API: `GET /manage/stats` returns per-router drop counts
- Via periodic logging: when `statistic_period_minutes > 0`, drop counts are included in the statistics output
