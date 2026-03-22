# Tailer

## Overview
Central orchestrator that owns and manages the complete log tailing pipeline. Coordinates the lifecycle of all servers and transfers.

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| config | Pipeline configuration | reference to Config | Yes | Immutable after creation |
| servers | Active log source instances | map of text → Server | Yes | Key is server name |
| transfers | Active destination instances | map of text → Transfer | Yes | Key is transfer name |

## State Definitions

| State | Description | Notes |
|-------|-------------|-------|
| Created | Tailer instantiated with config, not yet started | Validates config on creation |
| Running | All transfers initialized, servers spawned and producing logs | Normal operating state |
| Stopped | All servers and transfers shut down | Graceful shutdown complete |

## State Transitions

| Current State | Trigger Action | Target State | Preconditions | Post Actions |
|---------------|----------------|--------------|---------------|--------------|
| Created | Start | Running | Valid config | Initialize transfers, add servers |
| Running | Stop | Stopped | — | Stop all servers, stop all transfers |

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| Config | Uses (1:1) | Configuration for the pipeline |
| Server | Owns (1:N) | Active log source instances |
| Transfer | Owns (1:N) | Active destination instances |
