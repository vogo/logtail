# Model: Transfer (Interface)

## Description
Interface for log output destinations. All transfer types implement the `Transfer` interface.

## Interface Definition

```go
type Transfer interface {
    Name() string
    Trans(source string, data ...[]byte) error
    Start() error
    Stop() error
}
```

---

# Model: WebhookTransfer

## Description
Sends log data to an HTTP webhook endpoint via POST request.

## Current Struct Definition

```go
type WebhookTransfer struct {
    id     string
    url    string
    prefix string
}
```

## Changes (2026-03-22-001)

### New Fields

| Field | Type | Description |
|-------|------|-------------|
| `client` | `*http.Client` | Dedicated HTTP client with configured transport for connection pooling |
| `batcher` | `*Batcher` | Optional batch aggregation component; nil when batching is disabled |

### Updated Struct Definition

```go
type WebhookTransfer struct {
    id      string
    url     string
    prefix  string
    client  *http.Client
    batcher *Batcher
}
```

### Behavior Changes

- **`Trans()`**: If batcher is non-nil, delegates to `batcher.Add()` instead of calling `httpTrans()` directly
- **`Start()`**: Initializes the HTTP client with configured transport; starts batcher goroutine if batching is enabled
- **`Stop()`**: Flushes pending batch; closes idle connections via `client.CloseIdleConnections()`
- **`NewWebhookTransfer()`**: Accepts `TransferConfig` to configure `MaxIdleConns`, `IdleConnTimeout`, `BatchSize`, `BatchTimeout`

---

# Model: DingTransfer

## Description
Sends log alert messages to DingTalk bot webhook.

## Current Struct Definition

```go
type DingTransfer struct {
    Counter
    id           string
    url          string
    prefix       []byte
    transferring int32
}
```

## Changes (2026-03-22-001)

### New Fields

| Field | Type | Description |
|-------|------|-------------|
| `client` | `*http.Client` | Dedicated HTTP client with configured transport |
| `limiter` | `*rate.Limiter` | Token bucket rate limiter; nil when rate limiting is disabled |

### Updated Struct Definition

```go
type DingTransfer struct {
    Counter
    id           string
    url          string
    prefix       []byte
    transferring int32
    client       *http.Client
    limiter      *rate.Limiter
}
```

### Behavior Changes

- **`execTrans()`**: Uses `client.Post()` instead of `httpTrans()` with default client; checks `limiter.Allow()` before sending
- **`Start()`**: Initializes the HTTP client and rate limiter
- **`Stop()`**: Closes idle connections
- **`NewDingTransfer()`**: Accepts `TransferConfig` to configure transport and rate limit parameters

---

# Model: LarkTransfer

## Description
Sends log alert messages to Lark/Feishu bot webhook.

## Current Struct Definition

```go
type LarkTransfer struct {
    Counter
    id           string
    url          string
    prefix       []byte
    transferring int32
}
```

## Changes (2026-03-22-001)

### New Fields

| Field | Type | Description |
|-------|------|-------------|
| `client` | `*http.Client` | Dedicated HTTP client with configured transport |
| `limiter` | `*rate.Limiter` | Token bucket rate limiter; nil when rate limiting is disabled |

### Updated Struct Definition

```go
type LarkTransfer struct {
    Counter
    id           string
    url          string
    prefix       []byte
    transferring int32
    client       *http.Client
    limiter      *rate.Limiter
}
```

### Behavior Changes

- Same changes as DingTransfer (see above)

---

# Model: Batcher (New)

## Description
Batch aggregation component that accumulates log data and flushes it in batches to reduce HTTP call frequency. Used by WebhookTransfer when batch_size > 1.

## Struct Definition

```go
type Batcher struct {
    mu          sync.Mutex
    buffer      [][]byte
    source      string
    batchSize   int
    timeout     time.Duration
    timer       *time.Timer
    transferFn  func(source string, data ...[]byte) error
    stopped     bool
}
```

## Methods

| Method | Description |
|--------|-------------|
| `NewBatcher(batchSize, timeout, transferFn)` | Creates a new Batcher with the given thresholds and transfer callback |
| `Add(source string, data []byte)` | Adds data to buffer; triggers Flush() if batch size is reached |
| `Flush()` | Sends all buffered data as a single transfer call; resets buffer and timer |
| `Stop()` | Flushes remaining data and marks the batcher as stopped |
