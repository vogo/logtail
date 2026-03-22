# Batcher

## Overview
Batch aggregation component that accumulates log data and flushes it in batches to reduce network call frequency. Used by webhook transfers when batch_size > 1.

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| batch_size | Number of lines to accumulate before flushing | number | Yes | Threshold trigger |
| timeout | Maximum wait time before flushing a partial batch | duration | Yes | Time trigger |
| buffer | Accumulated log lines pending flush | list of log entries | Internal | Each entry stores data and source |

## State Definitions

| State | Description | Notes |
|-------|-------------|-------|
| Collecting | Accumulating log lines in buffer | Normal state |
| Stopped | Flushed remaining data, no longer accepting input | |

## Flush Triggers
1. **Count threshold**: Buffer reaches batch_size → immediate flush
2. **Timeout**: Timer expires since first buffered entry → flush partial batch
3. **Stop**: Explicit stop request → flush remaining data

## State Transitions

| Current State | Trigger Action | Target State | Preconditions | Post Actions |
|---------------|----------------|--------------|---------------|--------------|
| Collecting | Add (buffer full) | Collecting | — | Flush buffer, reset timer |
| Collecting | Timer fires | Collecting | Buffer non-empty | Flush buffer |
| Collecting | Stop | Stopped | — | Flush remaining, cancel timer |

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| Transfer (Webhook) | Belongs to | Used by webhook transfer for batch aggregation |
