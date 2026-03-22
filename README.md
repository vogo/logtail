# logtail

A log tailing utility that monitors command output or log files, filters log lines by matching rules, and transfers matched logs to various destinations.

[![codecov](https://codecov.io/gh/vogo/logtail/branch/master/graph/badge.svg)](https://codecov.io/gh/vogo/logtail)
[![GoDoc](https://godoc.org/github.com/vogo/logtail?status.svg)](https://godoc.org/github.com/vogo/logtail)
![license](https://img.shields.io/badge/license-Apache--2.0-green.svg)

## Features

- **Command tailing** — run a command and continuously tail its stdout
- **File watching** — watch files or directories (including subdirectories) for new log content
- **Log filtering** — filter log lines using `contains` / `not_contains` matchers
- **Log format** — recognize multi-line log entries using configurable prefix patterns
- **Multiple transfers** — route matched logs to console, file, webhook, DingTalk, or Lark
- **Web API** — runtime configuration and websocket-based log streaming
- **Multiple servers** — run multiple tailing sources concurrently with independent routers

## Architecture

![](https://github.com/vogo/vogo.github.io/raw/master/logtail/logtail-architecture-v2.png)

The core pipeline: **Config → Servers → Workers → Routes → Transfers**

- **Server** — defines a log source (command or file) and which routers to use
- **Router** — defines matchers (filtering rules) and which transfers receive matched lines
- **Transfer** — defines the output destination (console, file, webhook, DingTalk, Lark)

## Installation

```bash
go install github.com/vogo/logtail@master
```

Or download a binary from the [release page](https://github.com/vogo/logtail/releases/).

## Quick Start

### Using a config file

```bash
logtail -file config.json
```

### Using CLI flags

```bash
# Tail a command and send matched lines to DingTalk
logtail -cmd "tail -f /var/log/app.log" -match-contains ERROR -ding-url https://oapi.dingtalk.com/robot/send?access_token=xxx

# Tail a command and send matched lines to a webhook
logtail -cmd "tail -f /var/log/app.log" -match-contains ERROR -webhook-url https://example.com/webhook
```

### Using the Web API

```bash
# Start logtail with the web API on port 54321
logtail -port 54321
```

Then open `http://<server-ip>:54321/manage` to configure servers, routers, and transfers via the web interface. Browse `http://<server-ip>:54321` to view all tailing logs in real time.

## Configuration

The config file uses JSON format with three main sections: `transfers`, `routers`, and `servers`.

### Example: tail a command, filter ERROR lines, print to console

```json
{
  "transfers": {
    "console": {
      "type": "console"
    }
  },
  "routers": {
    "error-router": {
      "matchers": [
        {
          "contains": ["ERROR"],
          "not_contains": ["IgnoredError"]
        }
      ],
      "transfers": ["console"]
    }
  },
  "servers": {
    "app-log": {
      "command": "tail -f /var/log/app/app.log",
      "routers": ["error-router"]
    }
  }
}
```

### Example: tail a command, write matched lines to file

```json
{
  "transfers": {
    "file-out": {
      "type": "file",
      "dir": "/var/log/logtail-output"
    }
  },
  "routers": {
    "all": {
      "transfers": ["file-out"]
    }
  },
  "servers": {
    "app-log": {
      "command": "tail -f /var/log/app/app.log",
      "routers": ["all"]
    }
  }
}
```

### Example: watch log directory, send ERROR to DingTalk

```json
{
  "port": 54321,
  "default_format": {
    "prefix": "!!!!-!!-!!"
  },
  "transfers": {
    "ding-alarm": {
      "type": "ding",
      "prefix": "LOG ERROR ",
      "url": "https://oapi.dingtalk.com/robot/send?access_token=xxx"
    }
  },
  "routers": {
    "error-router": {
      "matchers": [
        {
          "contains": ["ERROR"],
          "not_contains": ["IgnoredError"]
        }
      ],
      "transfers": ["ding-alarm"]
    }
  },
  "servers": {
    "app-service": {
      "file": {
        "path": "/var/log/app-service/",
        "recursive": true,
        "suffix": ".log",
        "method": "timer"
      },
      "routers": ["error-router"]
    }
  }
}
```

### Example: watch log directory, send ERROR to Lark

```json
{
  "port": 54321,
  "default_format": {
    "prefix": "!!!!-!!-!!"
  },
  "transfers": {
    "lark-alarm": {
      "type": "lark",
      "prefix": "Log Alarm",
      "url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx"
    }
  },
  "routers": {
    "error-router": {
      "matchers": [
        {
          "contains": ["ERROR"],
          "not_contains": ["Invalid", "NotFound"]
        }
      ],
      "transfers": ["lark-alarm"]
    }
  },
  "servers": {
    "app-service": {
      "file": {
        "path": "/var/log/app-service/",
        "recursive": true,
        "suffix": ".log",
        "method": "timer",
        "dir_file_count_limit": 256
      },
      "routers": ["error-router"]
    }
  }
}
```

### Example: multiple servers with different routers

```json
{
  "transfers": {
    "console": { "type": "console" },
    "file-out": { "type": "file", "dir": "/var/log/logtail-output" }
  },
  "routers": {
    "error-to-console": {
      "matchers": [{ "contains": ["ERROR"] }],
      "transfers": ["console"]
    },
    "warn-to-file": {
      "matchers": [{ "contains": ["WARN"] }],
      "transfers": ["file-out"]
    }
  },
  "servers": {
    "app1": {
      "command": "tail -f /var/log/app1/app1.log",
      "routers": ["error-to-console"]
    },
    "app2": {
      "command": "tail -f /var/log/app2/app2.log",
      "routers": ["warn-to-file"]
    }
  }
}
```

## Config Reference

### Top-level fields

| Field | Type | Description |
|-------|------|-------------|
| `port` | int | Web API port (enables web UI and websocket streaming) |
| `log_level` | string | Log level: `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `default_format` | object | Global log format for multi-line log recognition |
| `statistic_period_minutes` | int | Statistics reporting interval in minutes |
| `transfers` | map | Transfer definitions (keyed by name) |
| `routers` | map | Router definitions (keyed by name) |
| `servers` | map | Server definitions (keyed by name) |

### Server config

| Field | Type | Description |
|-------|------|-------------|
| `command` | string | Single command to tail |
| `commands` | string | Multiple commands (newline-separated) |
| `command_gen` | string | Command that generates commands to tail |
| `file` | object | File/directory watch config (see below) |
| `format` | object | Per-server log format (overrides `default_format`) |
| `routers` | []string | List of router names to route output through |

### File config

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | File or directory path to watch |
| `method` | string | Watch method: `os` (filesystem events) or `timer` (polling) |
| `prefix` | string | Only watch files with this prefix |
| `suffix` | string | Only watch files with this suffix |
| `recursive` | bool | Include files in subdirectories |
| `dir_file_count_limit` | int | Skip directories with more files than this limit |

### Router config

| Field | Type | Description |
|-------|------|-------------|
| `matchers` | []object | List of matchers (all must match for a line to pass) |
| `transfers` | []string | List of transfer names to send matched lines to |
| `buffer_size` | int | Router buffer size |
| `blocking_mode` | bool | Block when buffer is full instead of dropping |

### Matcher config

| Field | Type | Description |
|-------|------|-------------|
| `contains` | []string | Line must contain ALL of these substrings |
| `not_contains` | []string | Line must NOT contain ANY of these substrings |

### Transfer config

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Transfer type: `console`, `file`, `webhook`, `ding`, `lark` |
| `url` | string | Webhook/DingTalk/Lark URL |
| `dir` | string | Output directory (for `file` type) |
| `prefix` | string | Message prefix (for webhook/ding/lark) |
| `max_idle_conns` | int | HTTP connection pool: max idle connections |
| `idle_conn_timeout` | string | HTTP connection pool: idle connection timeout (e.g., `90s`) |
| `rate_limit` | float | Rate limiting: requests per second |
| `rate_burst` | int | Rate limiting: burst size |
| `batch_size` | int | Batch aggregation: number of messages per batch |
| `batch_timeout` | string | Batch aggregation: max wait time before sending (e.g., `5s`) |

## Log Format

Configure log format to recognize multi-line log entries. The `prefix` field is a wildcard pattern matching the start of a new log record.

Wildcard syntax:
- `!` — matches one digit (`0`-`9`)
- `~` — matches one letter (`a`-`z`, `A`-`Z`)
- `?` — matches any single byte
- Any other character — must match exactly

Example: `!!!!-!!-!!` matches date prefixes like `2024-01-15`.

```json
{
  "default_format": {
    "prefix": "!!!!-!!-!! !!:!!:!!"
  }
}
```

With this format, log entries like:

```
2024-01-15 10:30:45 ERROR something failed
  at com.example.App.main(App.java:10)
  at com.example.App.run(App.java:5)
2024-01-15 10:30:46 INFO recovered
```

are correctly recognized as two entries — the ERROR entry includes its stack trace lines.

## Command Examples

Useful commands for tailing with logtail:

```bash
# Tail a local log file
tail -f /usr/local/myapp/myapp.log

# K8s: tail logs for a single pod
kubectl logs --tail 10 -f $(kubectl get pods --selector=app=myapp -o jsonpath='{.items[*].metadata.name}')

# K8s: tail logs for a deployment (multiple pods)
kubectl logs --tail 10 -f deployment/$(kubectl get deployments --selector=project-name=myapp -o jsonpath='{.items[*].metadata.name}')
```

## Development

```bash
make format       # Run goimports, gofmt, gofumpt
make check        # License header check + golangci-lint
make test         # Run unit tests with coverage
make integration  # Run integration tests
make build        # format + check + test + package
```
