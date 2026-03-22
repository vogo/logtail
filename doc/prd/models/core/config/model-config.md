# Config

## Overview
Top-level configuration for a logtail instance. Parsed from JSON/YAML config files or CLI flags. Defines all servers, routers, and transfers that make up the log pipeline.

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| port | HTTP API listening port | number | No | If not set, web API is not started |
| log_level | Logging verbosity level | text | No | Default: INFO |
| default_format | Global log line format definition | reference to FormatConfig | No | Applied to all servers unless overridden |
| statistic_period_minutes | Interval for reporting transfer statistics | number | No | 0 = no periodic reporting |
| servers | Collection of log sources | map of text → ServerConfig | No | Key is server name |
| routers | Collection of log processing pipelines | map of text → RouterConfig | No | Key is router name |
| transfers | Collection of log destinations | map of text → TransferConfig | No | Key is transfer name |

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| ServerConfig | Contains (1:N) | Each server defines a log source |
| RouterConfig | Contains (1:N) | Each router defines a processing pipeline |
| TransferConfig | Contains (1:N) | Each transfer defines a destination |
| FormatConfig | Contains (0:1) | Optional default format for log line recognition |
