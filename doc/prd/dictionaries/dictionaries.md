# logtail - Dictionaries

## Router Receive Mode

| Value | Label | Description |
|-------|-------|-------------|
| `false` (default) | Non-blocking | Drop data when channel is full; increment drop counter |
| `true` | Blocking | Block sender when channel is full; respect shutdown signal |

Used by: `RouterConfig.BlockingMode`, `Router.BlockingMode`

## Transfer Types

| Value | Constant | Description |
|-------|----------|-------------|
| `webhook` | `trans.TypeWebhook` | HTTP webhook POST |
| `ding` | `trans.TypeDing` | DingTalk bot webhook |
| `lark` | `trans.TypeLark` | Lark/Feishu bot webhook |
| `file` | `trans.TypeFile` | Write to local files |
| `console` | `trans.TypeConsole` | Write to stdout |

Used by: `TransferConfig.Type`

## Server Types

| Value | Description |
|-------|-------------|
| `command` | Single shell command |
| `commands` | Multiple commands (newline-separated) |
| `command_gen` | Command that generates other commands |
| `file` | File/directory watching |

Used by: `ServerConfig`

## Transfer HTTP Config Defaults (2026-03-22-001)

| Parameter | Default Value | Description |
|-----------|---------------|-------------|
| `max_idle_conns` | 2 | Max idle connections per host in HTTP transport |
| `idle_conn_timeout` | `"90s"` | Duration before idle connections are closed |
| `rate_limit` | 0 (disabled) | Requests per second; 0 means no rate limiting |
| `rate_burst` | 1 | Token bucket burst allowance |
| `batch_size` | 1 (disabled) | Lines per batch; 1 means send individually |
| `batch_timeout` | `"1s"` | Max wait before flushing a partial batch |

Used by: `TransferConfig`
