# Router

## Overview
Receives log data from workers, applies matchers to filter lines, and dispatches matched lines to transfers. Operates as a buffered channel-based pipeline stage with configurable overflow behavior.

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| id | Unique identifier | text | Yes | Auto-generated |
| name | Display name | text | Yes | From RouterConfig |
| source | Server name that feeds this router | text | Yes | Set when router is attached to a server |
| matchers | Filter conditions | list of Matcher | No | Empty = match all lines |
| transfers | Destinations for matched lines | list of Transfer | Yes | |
| buffer_size | Channel buffer capacity | number | Yes | Default: 16 |
| blocking_mode | Overflow handling strategy | enum (Router Receive Mode) | Yes | Default: non-blocking |
| drop_count | Messages dropped due to buffer overflow | number | Read-only | Incremented atomically; queryable via stats API |

## State Definitions

| State | Description | Notes |
|-------|-------------|-------|
| Active | Receiving, filtering, and dispatching log data | Normal state |
| Stopped | No longer processing | Shutdown |

## State Transitions

| Current State | Trigger Action | Target State | Preconditions | Post Actions |
|---------------|----------------|--------------|---------------|--------------|
| Active | Stop | Stopped | — | Close channel |

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| Worker | Receives from (N:1) | Gets log data from workers |
| Matcher | Contains (0:N) | Filter conditions (AND logic) |
| Transfer | Dispatches to (1:N) | Sends matched lines |
| RouterConfig | Configured by | Configuration source |
