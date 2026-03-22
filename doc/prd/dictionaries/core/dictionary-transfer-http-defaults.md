# Transfer HTTP Configuration Defaults

## Overview
Default values for HTTP-based transfer configuration parameters. These apply to webhook, DingTalk, and Lark transfer types.

## Values

| Parameter | Default Value | Description | Sort Order | Notes |
|-----------|---------------|-------------|------------|-------|
| max_idle_conns | 2 | Max idle connections per host in HTTP transport | 1 | Per-transfer client isolation |
| idle_conn_timeout | 90s | Duration before idle connections are closed | 2 | Go duration string format |
| rate_limit | 0 (disabled) | Requests per second; 0 means no rate limiting | 3 | Applies to ding/lark types |
| rate_burst | 1 | Token bucket burst allowance | 4 | Only effective when rate_limit > 0 |
| batch_size | 1 (disabled) | Lines per batch; 1 means send individually | 5 | Applies to webhook type |
| batch_timeout | 1s | Max wait before flushing a partial batch | 6 | Only effective when batch_size > 1 |
