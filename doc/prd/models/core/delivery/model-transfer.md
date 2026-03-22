# Transfer

## Overview
Delivers log data to a configured destination. Each transfer type implements the same lifecycle interface but with different delivery mechanisms.

## Transfer Types

| Type | Description | Key Attributes |
|------|-------------|---------------|
| Console | Output to stdout | No additional config |
| File | Write to rotating local files | dir (output directory); auto-rotates at 8MB |
| Webhook | HTTP POST to endpoint | url; supports connection pooling, batching |
| DingTalk | DingTalk bot messaging | url, prefix; 1024 byte message limit, 5s throttle, rate limiting |
| Lark | Lark/Feishu bot messaging | url, prefix; 1024 byte message limit, 5s throttle, rate limiting |

## Common Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| name | Unique identifier | text | Yes | From TransferConfig |
| type | Destination type | enum (Transfer Type) | Yes | |

## State Definitions

| State | Description | Notes |
|-------|-------------|-------|
| Created | Transfer instantiated, not started | |
| Active | Accepting and delivering log data | |
| Stopped | Resources released, no longer accepting data | |

## State Transitions

| Current State | Trigger Action | Target State | Preconditions | Post Actions |
|---------------|----------------|--------------|---------------|--------------|
| Created | Start | Active | — | Initialize resources (HTTP client, file handles) |
| Active | Stop | Stopped | — | Flush pending data, close resources |

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| Tailer | Belongs to | Owned by tailer instance |
| Router | Receives from (N:M) | Gets matched log lines from routers |
| Batcher | Contains (0:1) | Optional batch aggregation (webhook only) |
| TransferConfig | Configured by | Configuration source |
