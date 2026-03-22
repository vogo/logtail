# Model: Tailer

## Description
Central orchestrator that owns Servers and Transfers. Manages the complete lifecycle of the log tailing pipeline.

## Struct Definition

```go
type Tailer struct {
    lock      sync.Mutex
    Config    *conf.Config
    Servers   map[string]*serve.Server
    Transfers map[string]trans.Transfer
}
```

## Lifecycle

1. Created via `NewTailer(config)` -- validates config, initializes maps
2. Started via `Start()` -- initializes transfers, adds servers; returns error on failure
3. Stopped via `Stop()` -- stops all servers and transfers

## Changes (2026-03-21-001)

- **Remove global singleton**: `DefaultTailer` global variable is eliminated. Tailer instances are created and owned by the caller.
- **Explicit lifecycle**: Callers create, start, and stop Tailer instances directly. No package-level state.
