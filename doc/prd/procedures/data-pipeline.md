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

---

## Changes (2026-03-22-001): HTTP Transfer Optimization

### Transfer Stage — Current Flow

```
Router.Route(data)
  |-> Transfer.Trans(source, data)
       |-> httpTrans(url, data...)        // uses http.Post (default client)
            |-> http.Post(url, ...)       // new connection possible each time
```

**Problems:**
1. `httpTrans()` uses `http.Post()` without a configured transport — no explicit connection pool tuning
2. No client-side rate limiting for DingTalk/Lark APIs that enforce server-side rate limits
3. WebhookTransfer sends one HTTP request per matched log line, causing excessive network calls under high throughput

### Transfer Stage — Target Flow

```
Router.Route(data)
  |-> Transfer.Trans(source, data)
       |-> [webhook with batching] Batcher.Add(source, data)
       |     |-> [on count threshold or timeout] Batcher.Flush()
       |          |-> httpTransWithClient(client, url, batchedData)
       |-> [ding/lark with rate limit] limiter.Allow() check
       |     |-> if denied: drop + warn log
       |-> httpTransWithClient(client, url, data...)
            |-> client.Post(url, ...)     // reuses connections via configured transport
```

### New Component: Batcher

```
Batcher.Add(source, data):
  -> append to internal buffer
  -> if len(buffer) >= batchSize: Flush()
  -> if timer not started: start timer(batchTimeout)

Batcher timer fires:
  -> Flush()

Batcher.Flush():
  -> if buffer empty: return
  -> combine buffer entries into single payload
  -> call transferFunc(source, combinedData)
  -> reset buffer and timer

Batcher.Stop():
  -> Flush() (drain remaining)
  -> stop timer
```

### Key Design Decisions

- **Client per transfer**: Each transfer instance gets its own `http.Client` to isolate connection pools
- **Backward compatible defaults**: Batching disabled (batch_size=1), rate limiting disabled (rate_limit=0), transport defaults match Go stdlib
- **Non-blocking rate limiting**: Rate-exceeded messages are dropped with a warning log, not queued (preserves pipeline throughput)
- **Flush on stop**: Pending batches are flushed during `Transfer.Stop()` before closing the HTTP client
