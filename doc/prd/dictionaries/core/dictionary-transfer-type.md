# Transfer Type

## Overview
Defines the available destination types for log delivery.

## Values

| Value | Label | Description | Sort Order | Notes |
|-------|-------|-------------|------------|-------|
| console | Console | Output to stdout | 1 | Default, no additional config needed |
| file | File | Write to rotating local files | 2 | Requires `dir` config |
| webhook | Webhook | HTTP POST to endpoint | 3 | Requires `url` config; supports batching |
| ding | DingTalk | DingTalk bot webhook | 4 | Requires `url` config; supports rate limiting |
| lark | Lark | Lark/Feishu bot webhook | 5 | Requires `url` config; supports rate limiting |
