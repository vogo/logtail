# RouterConfig

## Overview
Configuration for a log processing pipeline. Defines matchers for filtering and transfers for delivery.

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| name | Unique identifier | text | Yes | Used as map key in Config |
| matchers | List of match conditions (AND logic) | list of MatcherConfig | No | Empty = match all lines |
| transfers | List of transfer names for matched lines | list of text | Yes | References TransferConfig names |
| buffer_size | Channel buffer size | number | No | Default: 16 |
| blocking_mode | Buffer overflow handling strategy | enum (Router Receive Mode) | No | Default: non-blocking |

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| Config | Belongs to | Part of top-level configuration |
| MatcherConfig | Contains (0:N) | Filter conditions |
| TransferConfig | References (N:M) | Destinations for matched lines |
