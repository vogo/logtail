# Model: Config

## Description
Top-level configuration structure parsed from JSON/YAML files or CLI flags.

## Current Struct Definition

```go
type Config struct {
    file                   string
    Port                   int
    LogLevel               string
    DefaultFormat          *match.Format
    StatisticPeriodMinutes int
    Transfers              map[string]*TransferConfig
    Routers                map[string]*RouterConfig
    Servers                map[string]*ServerConfig
}
```

## No structural changes required for this requirement.

---

# Model: TransferConfig

## Description
Configuration for a single Transfer instance.

## Current Struct Definition

```go
type TransferConfig struct {
    Name   string `json:"-"`
    Type   string `json:"type"`
    URL    string `json:"url,omitempty"`
    Dir    string `json:"dir,omitempty"`
    Prefix string `json:"prefix,omitempty"`
}
```

## Changes (2026-03-22-001)

### New Fields

| Field | Type | JSON Key | Default | Description |
|-------|------|----------|---------|-------------|
| `MaxIdleConns` | `int` | `max_idle_conns` | 2 | Max idle connections per host in HTTP transport |
| `IdleConnTimeout` | `string` | `idle_conn_timeout` | `"90s"` | Idle connection timeout (Go duration string) |
| `RateLimit` | `float64` | `rate_limit` | 0 (disabled) | Max requests per second (token bucket rate); applies to ding/lark types |
| `RateBurst` | `int` | `rate_burst` | 1 | Token bucket burst size; applies to ding/lark types |
| `BatchSize` | `int` | `batch_size` | 1 (no batching) | Lines to accumulate before sending; applies to webhook type |
| `BatchTimeout` | `string` | `batch_timeout` | `"1s"` | Max wait before flushing partial batch; applies to webhook type |

### Updated Struct Definition

```go
type TransferConfig struct {
    Name            string  `json:"-"`
    Type            string  `json:"type"`
    URL             string  `json:"url,omitempty"`
    Dir             string  `json:"dir,omitempty"`
    Prefix          string  `json:"prefix,omitempty"`
    MaxIdleConns    int     `json:"max_idle_conns,omitempty"`
    IdleConnTimeout string  `json:"idle_conn_timeout,omitempty"`
    RateLimit       float64 `json:"rate_limit,omitempty"`
    RateBurst       int     `json:"rate_burst,omitempty"`
    BatchSize       int     `json:"batch_size,omitempty"`
    BatchTimeout    string  `json:"batch_timeout,omitempty"`
}
```

---

# Model: RouterConfig

## Description
Configuration for a single Router instance.

## Current Struct Definition

```go
type RouterConfig struct {
    Name      string
    Matchers  []*MatcherConfig
    Transfers []string
}
```

## Changes (2026-03-21-001)

### New Fields

| Field | Type | JSON Key | Default | Description |
|-------|------|----------|---------|-------------|
| `BufferSize` | `int` | `buffer_size` | 0 (use default 16) | Channel buffer size for this router |
| `BlockingMode` | `bool` | `blocking_mode` | false | Whether to block on full buffer instead of dropping |

### Updated Struct Definition

```go
type RouterConfig struct {
    Name         string           `json:"-"`
    Matchers     []*MatcherConfig `json:"matchers"`
    Transfers    []string         `json:"transfers"`
    BufferSize   int              `json:"buffer_size,omitempty"`
    BlockingMode bool             `json:"blocking_mode,omitempty"`
}
```
