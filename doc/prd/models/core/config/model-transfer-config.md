# TransferConfig

## Overview
Configuration for a log destination. Defines where and how filtered logs are delivered.

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| name | Unique identifier | text | Yes | Used as map key in Config |
| type | Destination type | enum (Transfer Type) | Yes | console, file, webhook, ding, lark |
| url | HTTP endpoint URL | text | Conditional | Required for webhook, ding, lark types |
| dir | Output directory path | text | Conditional | Required for file type |
| prefix | Custom message prefix | text | No | Used by ding, lark types; defaults to system hostname |
| max_idle_conns | Max idle HTTP connections per host | number | No | Default: 2; applies to HTTP types |
| idle_conn_timeout | Idle connection timeout | duration (text) | No | Default: 90s; Go duration format |
| rate_limit | Max requests per second | number (decimal) | No | Default: 0 (disabled); applies to ding, lark |
| rate_burst | Rate limiter burst size | number | No | Default: 1; effective only when rate_limit > 0 |
| batch_size | Lines per batch | number | No | Default: 1 (no batching); applies to webhook |
| batch_timeout | Max batch wait time | duration (text) | No | Default: 1s; effective only when batch_size > 1 |

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| Config | Belongs to | Part of top-level configuration |
