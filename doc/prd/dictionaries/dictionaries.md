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
