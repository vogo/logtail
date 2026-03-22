# ServerConfig

## Overview
Configuration for a single log source. Defines how logs are collected — by executing commands, watching files, or accepting manual input.

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| name | Unique identifier | text | Yes | Used as map key in Config |
| command | Shell command to execute and tail | text | No | Mutually exclusive with commands, command_gen, file |
| commands | Multiple commands to execute in parallel | text | No | Newline-separated; mutually exclusive with others |
| command_gen | Command that generates other commands dynamically | text | No | Output changes trigger worker recreation |
| file | File/directory watch configuration | reference to FileConfig | No | Mutually exclusive with command fields |
| format | Log line format specific to this server | reference to FormatConfig | No | Overrides global default_format |
| routers | List of router names to process logs from this server | list of text | Yes | References RouterConfig names |

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| Config | Belongs to | Part of top-level configuration |
| FileConfig | Contains (0:1) | File watching configuration |
| FormatConfig | Contains (0:1) | Server-specific format override |
| RouterConfig | References (N:M) | Routers that process this server's logs |
