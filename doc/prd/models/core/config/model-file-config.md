# FileConfig

## Overview
Configuration for file and directory watching. Defines which files to monitor and how to detect changes.

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| path | File or directory path to watch | text | Yes | |
| method | Change detection method | enum (File Watch Method) | No | Default: os |
| recursive | Whether to monitor subdirectories | boolean | No | Default: false |
| prefix | Filename prefix filter | text | No | Only watch files starting with this prefix |
| suffix | Filename suffix filter | text | No | Only watch files ending with this suffix |
| dir_file_count_limit | Max files per directory | number | No | Skip directories exceeding this limit |

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| ServerConfig | Belongs to | Part of server file configuration |
