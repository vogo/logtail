# logtail

## Product Positioning
- Log tailing and aggregation utility
- Collects logs from multiple sources (shell commands, file/directory watching, dynamic command generation)
- Filters logs using pattern matching
- Routes matched logs to multiple destinations simultaneously
- Provides real-time web-based monitoring and runtime management

## Target Users
- DevOps engineers monitoring application logs in production
- Developers needing log aggregation and filtering during development
- System administrators forwarding log events to alerting systems (DingTalk, Lark, webhooks)
- Platform teams managing Kubernetes pod logs

## Core Value Proposition
- Lightweight single-binary deployment with no external dependencies
- Flexible pipeline: any log source → filter → multiple destinations
- Real-time web UI and WebSocket streaming for live monitoring
- Runtime configuration via HTTP API without service restart
- Multi-line log record support via format prefix pattern matching
- Built-in integration with popular messaging platforms (DingTalk, Lark)

## Product Boundary
### In Scope
- Tailing command output (single, multiple, dynamically generated)
- Watching files and directories for log changes
- Pattern-based log filtering (contains/not-contains, wildcard format matching)
- Routing filtered logs to: console, files, HTTP webhooks, DingTalk, Lark
- HTTP API for runtime configuration management
- WebSocket-based real-time log streaming
- Batch aggregation and rate limiting for HTTP destinations
- Transfer statistics and pipeline observability

### Out of Scope
- Log storage or indexing (not a log database)
- Log parsing or structured extraction (not a log parser)
- User authentication or access control
- Distributed deployment or clustering
- Log compression or archiving
