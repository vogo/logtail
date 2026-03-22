# File Watch Method

## Overview
Defines how logtail detects changes in watched files and directories.

## Values

| Value | Label | Description | Sort Order | Notes |
|-------|-------|-------------|------------|-------|
| os | OS Events | Use operating system file system event notifications | 1 | More efficient; recommended for most cases |
| timer | Polling | Periodically check files for modifications | 2 | Fallback for environments where OS events are unreliable |
