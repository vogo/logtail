# Worker

## Overview
Executes a single command or tails a single file, reads its output, and dispatches log records to routers. Handles multi-line log grouping via format prefix matching.

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| id | Unique identifier | text | Yes | Auto-generated |
| command | Shell command being executed | text | Conditional | Set for command-based workers |
| file_path | File being tailed | text | Conditional | Set for file-based workers |
| routers | Routers receiving this worker's output | list of Router | Yes | Inherited from parent server |

## State Definitions

| State | Description | Notes |
|-------|-------------|-------|
| Running | Actively reading output and dispatching to routers | Normal operating state |
| Failed | Command exited or file read error | May retry for dynamic workers |
| Stopped | Manually stopped or server shutting down | No retry |

## State Transitions

| Current State | Trigger Action | Target State | Preconditions | Post Actions |
|---------------|----------------|--------------|---------------|--------------|
| Running | Command exits | Failed | — | Wait 10 seconds, then retry (dynamic workers only) |
| Running | Stop request | Stopped | — | Kill subprocess, close resources |
| Failed | Retry timer | Running | Dynamic worker | Re-execute command |
| Failed | Stop request | Stopped | — | Cancel retry |

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| Server | Belongs to | Owned by parent server |
| Router | Dispatches to (1:N) | Sends log records to routers |
| Format | Uses (0:1) | Format for multi-line grouping |
