# FormatConfig

## Overview
Defines the expected format prefix of log lines. Used to identify log record boundaries in multi-line output.

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| prefix | Wildcard pattern for log line prefix | text | Yes | Uses custom wildcard syntax: `?`=any byte, `~`=alpha, `!`=digit |

## Usage
When data arrives, each line is checked against the prefix pattern. Lines matching the pattern start a new log record. Lines not matching are treated as continuation lines of the current record.

## Example
A prefix of `!!!!-!!-!!` matches ISO date prefixes like `2024-01-15`, grouping stack traces and multi-line messages with the originating log line.

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| Config | Belongs to (0:1) | Global default format |
| ServerConfig | Belongs to (0:1) | Server-specific format override |
