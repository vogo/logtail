# logtail - Product Overview

## Summary

logtail is a Go-based log tailing utility that supports tailing command output, watching files/directories, websocket streaming, log matching/filtering, and transferring logs to various destinations (console, file, webhook, DingTalk, Lark).

## Core Value

Provide a lightweight, configurable, and extensible log tailing pipeline that can:
- Tail output from shell commands, generated command lists, or watched file/directory paths
- Filter log lines using contains-match and wildcard-based format matching
- Route matched lines to multiple transfer destinations simultaneously
- Be managed at runtime via HTTP API and websocket streaming

## Target Users

- DevOps engineers monitoring application logs
- Developers needing log aggregation and filtering
- System administrators forwarding log events to alerting systems (DingTalk, Lark, webhooks)

## Design Principles

- **Dependency injection over globals**: Components receive their dependencies explicitly, enabling testability and multi-instance usage
- **Observable pipelines**: Data flow through the system is measurable; drops and errors are visible
- **Backward compatibility**: New features use sensible defaults that preserve existing behavior
- **Graceful lifecycle**: All components support clean startup signaling and shutdown
