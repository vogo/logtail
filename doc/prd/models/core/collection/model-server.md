# Server

## Overview
Manages a log source and its workers. Each server corresponds to a configured log source and produces log data that flows through routers to transfers.

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| name | Unique identifier | text | Yes | Matches ServerConfig name |
| workers | Active worker instances | list of Worker | Yes | One per command or watched file |
| routers | Processing pipelines attached to this server | list of Router | Yes | Built from config references |

## State Definitions

| State | Description | Notes |
|-------|-------------|-------|
| Created | Server instantiated, workers not yet spawned | |
| Running | Workers active and producing log data | |
| Stopped | All workers stopped | Graceful shutdown |

## State Transitions

| Current State | Trigger Action | Target State | Preconditions | Post Actions |
|---------------|----------------|--------------|---------------|--------------|
| Created | Start | Running | Valid config | Spawn workers based on server type |
| Running | Stop | Stopped | — | Stop all workers, close routers |
| Running | Add Worker | Running | File watch mode | New file detected, spawn worker |
| Running | Remove Worker | Running | — | File inactive/deleted, stop worker |

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| Tailer | Belongs to | Owned by tailer instance |
| Worker | Owns (1:N) | Worker instances for this server |
| Router | Uses (1:N) | Routers that process this server's output |
| ServerConfig | Configured by | Configuration source |
