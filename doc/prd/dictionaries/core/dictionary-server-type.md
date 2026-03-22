# Server Type

## Overview
Defines how a server collects log data. The type is inferred from which configuration fields are set.

## Values

| Value | Label | Description | Sort Order | Notes |
|-------|-------|-------------|------------|-------|
| command | Single Command | Execute a shell command and tail its output | 1 | Set `command` field |
| commands | Multiple Commands | Execute multiple commands in parallel | 2 | Set `commands` field (newline-separated) |
| command_gen | Command Generator | Run a command that generates other commands dynamically | 3 | Set `command_gen` field; workers are recreated when output changes |
| file | File/Directory Watch | Monitor files or directories for changes | 4 | Set `file` field with path and filter options |
| manual | Manual Input | Accept data via API only | 5 | No command or file config; data written via Server.Write() |
