# Process: Startup Flow

## Current Flow (Before Change)

```
main()
  |-> starter.Start()
  |     |-> conf.ParseConfig()
  |     |-> tail.NewTailer(config)
  |     |-> go StartTailer(tailer)     // fire-and-forget goroutine
  |     |     |-> tail.DefaultTailer = tailer  // sets global
  |     |     |-> tailer.Start()
  |     |     |     |-> StartTransfers()
  |     |     |     |-> AddServer() for each server
  |     |     |     \-> return error (only caught by panic in goroutine)
  |     \-> return tailer
  |-> webapi.StartWebAPI(tailer)
  |-> handleSignal()
  \-> starter.StopLogtail()             // accesses DefaultTailer global
```

**Problems:**
1. `StartTailer()` sets `tail.DefaultTailer` global -- coupling
2. Startup runs in a goroutine -- caller cannot detect failure
3. `StopLogtail()` depends on `DefaultTailer` global

## Target Flow (After Change)

```
main()
  |-> starter.Start()
  |     |-> conf.ParseConfig()
  |     |-> tail.NewTailer(config)
  |     |-> StartTailer(tailer)           // synchronous call, returns error
  |     |     |-> tailer.Start()
  |     |     |     |-> StartTransfers()
  |     |     |     |-> AddServer() for each server
  |     |     |     \-> return error
  |     |     \-> return error to caller
  |     \-> return (tailer, error)
  |-> handle error from Start()
  |-> webapi.StartWebAPI(tailer)
  |-> handleSignal()
  \-> tailer.Stop()                       // direct call on instance, no global
```

**Changes:**
1. `starter.Start()` returns `(*tail.Tailer, error)` -- startup is synchronous
2. No global `DefaultTailer` -- `StopLogtail()` is replaced by `tailer.Stop()`
3. `starter.StartTailer()` no longer sets any global; it calls `tailer.Start()` and returns the error
4. `main()` handles startup errors explicitly

## Alternative: Async with Completion Channel

If synchronous startup is undesirable (e.g., servers take a long time to start), an alternative is:

```go
func Start() (*tail.Tailer, <-chan error)
```

The returned channel receives `nil` on success or an error on failure. The caller can `select` on it. This approach is more flexible but adds complexity. The synchronous approach is recommended for simplicity since `tailer.Start()` already returns promptly (servers run in their own goroutines internally).
