# Router Receive Mode

## Overview
Determines how a router handles incoming data when its buffer channel is full.

## Values

| Value | Label | Description | Sort Order | Notes |
|-------|-------|-------------|------------|-------|
| false | Non-blocking (default) | Drop data when channel is full; increment drop counter | 1 | Default behavior; prevents pipeline backpressure |
| true | Blocking | Block sender until space is available; respect shutdown signal | 2 | Use for reliability-critical routes where no data loss is acceptable |
