# logtail - Architecture

## System Architecture

```
Config --> Tailer --> Servers --> Workers --> Routers --> Transfers
                                              |
                                          Matchers (filter)
```

## Core Modules

### Configuration (`internal/conf`)
Parses JSON/YAML config files and CLI flags. Validates configuration integrity. Provides config accessor functions to other modules.

### Tailer (`internal/tail`)
Central orchestrator that owns Servers and Transfers. Created via `NewTailer(config)` and started via `Start()`. Lifecycle is managed by the caller (no global singleton). Returns startup completion signal.

### Server (`internal/serve`)
Manages Workers that execute commands or watch directories. Each Server corresponds to a configured log source. Workers read output and dispatch to Routers.

### Worker (`internal/work`)
Platform-specific process management. Executes shell commands or watches file changes. Writes output into Routers for filtering and transfer.

### Router (`internal/route`)
Receives log data from Workers. Applies Matchers to filter lines. Dispatches matched lines to Transfers. Supports configurable channel buffer size, optional blocking mode, and exposes drop counters for observability.

### Matcher (`internal/match`)
Log line filtering: contains-match and wildcard-based format matching. Custom wildcard syntax: `?`=any byte, `~`=alpha, `!`=digit.

### Transfer (`internal/trans`)
Interface for log output destinations. Implementations: console, file, webhook, DingTalk (`ding`), Lark. Includes transfer statistics counting.

### Web API (`internal/webapi`)
HTTP API for runtime configuration management and websocket log streaming. Receives Tailer instance via dependency injection.

### Starter (`internal/starter`)
Bootstrap logic: parses config, creates Tailer, starts it with completion signaling, and provides stop functionality. No global state.

## Data Flow

1. **Config** is parsed and validated
2. **Tailer** is created with the config and started (caller receives startup completion signal)
3. **Transfers** are initialized from config
4. **Servers** are created, each spawning **Workers**
5. Workers produce log data and send to **Routers** via `Receive()`
6. Routers buffer data in channels (configurable size), apply **Matchers**, and dispatch to **Transfers**
7. If channel is full: non-blocking mode increments drop counter; blocking mode waits

## Key Design Decisions

- **No global singleton**: Tailer is instantiated and passed explicitly to all consumers
- **Startup signaling**: Callers can detect whether startup succeeded before proceeding
- **Backpressure-aware pipeline**: Routers track dropped messages and support configurable buffering behavior
