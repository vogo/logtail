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
