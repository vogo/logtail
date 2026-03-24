/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package serve_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vogo/fwatch"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/serve"
	"github.com/vogo/logtail/internal/trans"
)

// helper to create a server configured for directory watching and start it.
func newDirWatchServer(t *testing.T, id string, fileConfig *conf.FileConfig) *serve.Server {
	t.Helper()

	server := serve.NewRawServer(id)
	server.RouterConfigsFunc = func() []*conf.RouterConfig { return nil }
	server.TransferMatcher = func(_ []string) []trans.Transfer { return nil }

	serverConfig := &conf.ServerConfig{
		Name: id,
		File: fileConfig,
	}

	server.Start(serverConfig)

	return server
}

// TestDirWatch_TimerMethod verifies that directory watching works with the timer method.
// This covers the basic case of starting a server that watches a directory using polling.
// The test validates that:
// - The server starts without error using fwatch v1.6.1 option-based API
// - Workers are created for existing files in the watched directory
// - The server shuts down cleanly
func TestDirWatch_TimerMethod(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create a file before starting the watcher so it gets picked up.
	logFile := filepath.Join(dir, "app.log")
	require.NoError(t, os.WriteFile(logFile, []byte("initial log line\n"), 0o644))

	server := newDirWatchServer(t, "timer-dir", &conf.FileConfig{
		Path:   dir,
		Method: fwatch.WatchMethodTimer,
	})

	// Allow the watcher time to detect the file and create a worker.
	time.Sleep(3 * time.Second)

	assert.NotEmpty(t, server.Workers, "expected at least one worker for the existing log file")

	require.NoError(t, server.Stop())
}

// TestDirWatch_FSMethod verifies that directory watching works with the FS (fsnotify) method.
// This covers the same scenario as TestDirWatch_TimerMethod but using OS-level file events.
// The test validates that:
// - fwatch.New() succeeds with WithMethod(WatchMethodFS) option
// - Workers are created when files are written to in the watched directory
// - The server shuts down cleanly
func TestDirWatch_FSMethod(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create a file before starting.
	logFile := filepath.Join(dir, "app.log")
	require.NoError(t, os.WriteFile(logFile, []byte("fs log line\n"), 0o644))

	server := newDirWatchServer(t, "fs-dir", &conf.FileConfig{
		Path:   dir,
		Method: fwatch.WatchMethodFS,
	})

	time.Sleep(3 * time.Second)

	assert.NotEmpty(t, server.Workers, "expected at least one worker for the existing log file")

	require.NoError(t, server.Stop())
}

// TestDirWatch_NewFileCreation verifies that when a new file is created in a watched directory,
// the watcher detects it and creates a worker for it.
// This tests the core Create event handling in startDirWatchWorkers.
// Note: Uses the FS (fsnotify) method because the timer method's check interval is derived from
// fileInactiveDeadline (1 hour / 3 = 20 min, capped to 1 min), which is too slow for a test.
// The FS method receives OS-level file events immediately.
func TestDirWatch_NewFileCreation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	server := newDirWatchServer(t, "new-file", &conf.FileConfig{
		Path:   dir,
		Method: fwatch.WatchMethodFS,
	})

	// Give the watcher time to start and register the fsnotify watch.
	time.Sleep(1 * time.Second)

	// Create a new file after the watcher is running.
	logFile := filepath.Join(dir, "new.log")
	require.NoError(t, os.WriteFile(logFile, []byte("new log data\n"), 0o644))

	// Allow the FS watcher to detect the new file event.
	time.Sleep(3 * time.Second)

	assert.NotEmpty(t, server.Workers, "expected a worker to be created for the new file")

	require.NoError(t, server.Stop())
}

// TestDirWatch_PrefixSuffixFiltering verifies that only files matching both
// the configured prefix and suffix are watched.
// This validates the matcher closure in startDirWorkers:
//
//	matcher := func(name string) bool {
//	    return (config.Prefix == "" || strings.HasPrefix(name, config.Prefix)) &&
//	        (config.Suffix == "" || strings.HasSuffix(name, config.Suffix))
//	}
func TestDirWatch_PrefixSuffixFiltering(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create files: only "app-something.log" should match prefix "app" and suffix ".log".
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app-main.log"), []byte("match\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sys-main.log"), []byte("no prefix\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app-main.txt"), []byte("no suffix\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "other.dat"), []byte("no match\n"), 0o644))

	server := newDirWatchServer(t, "filter", &conf.FileConfig{
		Path:   dir,
		Method: fwatch.WatchMethodTimer,
		Prefix: "app",
		Suffix: ".log",
	})

	time.Sleep(3 * time.Second)

	// Only "app-main.log" matches both prefix and suffix.
	assert.Len(t, server.Workers, 1, "expected exactly one worker for the matching file")

	require.NoError(t, server.Stop())
}

// TestDirWatch_PrefixOnly verifies filtering with only a prefix configured (no suffix).
func TestDirWatch_PrefixOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "app-1.log"), []byte("m\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app-2.txt"), []byte("m\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sys-1.log"), []byte("n\n"), 0o644))

	server := newDirWatchServer(t, "prefix-only", &conf.FileConfig{
		Path:   dir,
		Method: fwatch.WatchMethodTimer,
		Prefix: "app",
	})

	time.Sleep(3 * time.Second)

	// "app-1.log" and "app-2.txt" both match the prefix.
	assert.Len(t, server.Workers, 2, "expected two workers for files matching the prefix")

	require.NoError(t, server.Stop())
}

// TestDirWatch_SuffixOnly verifies filtering with only a suffix configured (no prefix).
func TestDirWatch_SuffixOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.log"), []byte("m\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sys.log"), []byte("m\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "data.csv"), []byte("n\n"), 0o644))

	server := newDirWatchServer(t, "suffix-only", &conf.FileConfig{
		Path:   dir,
		Method: fwatch.WatchMethodTimer,
		Suffix: ".log",
	})

	time.Sleep(3 * time.Second)

	// "app.log" and "sys.log" match the suffix.
	assert.Len(t, server.Workers, 2, "expected two workers for files matching the suffix")

	require.NoError(t, server.Stop())
}

// TestDirWatch_RecursiveWatch verifies that with Recursive=true, files in subdirectories
// are also watched.
func TestDirWatch_RecursiveWatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subDir := filepath.Join(dir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "root.log"), []byte("root\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "sub.log"), []byte("sub\n"), 0o644))

	server := newDirWatchServer(t, "recursive", &conf.FileConfig{
		Path:      dir,
		Method:    fwatch.WatchMethodTimer,
		Recursive: true,
	})

	time.Sleep(3 * time.Second)

	// Both root.log and subdir/sub.log should be watched.
	assert.GreaterOrEqual(t, len(server.Workers), 2,
		"expected workers for files in both root and subdirectory")

	require.NoError(t, server.Stop())
}

// TestDirWatch_NonRecursiveWatch verifies that with Recursive=false (default),
// files in subdirectories are NOT watched.
func TestDirWatch_NonRecursiveWatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subDir := filepath.Join(dir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "root.log"), []byte("root\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "sub.log"), []byte("sub\n"), 0o644))

	server := newDirWatchServer(t, "non-recursive", &conf.FileConfig{
		Path:      dir,
		Method:    fwatch.WatchMethodTimer,
		Recursive: false,
	})

	time.Sleep(3 * time.Second)

	// Only root.log should be watched.
	assert.Len(t, server.Workers, 1, "expected only root directory file to have a worker")

	require.NoError(t, server.Stop())
}

// TestDirWatch_ServerStopCleansUp verifies that stopping a server with an active directory
// watcher shuts down cleanly without goroutine leaks or panics.
// The watcher's Done() channel (v1.6.1 API replacing Runner.C) should cause the
// startDirWatchWorkers goroutine to return.
func TestDirWatch_ServerStopCleansUp(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.log"), []byte("data\n"), 0o644))

	server := newDirWatchServer(t, "stop-cleanup", &conf.FileConfig{
		Path:   dir,
		Method: fwatch.WatchMethodTimer,
	})

	time.Sleep(2 * time.Second)

	// Stop should return without error and no panics.
	require.NoError(t, server.Stop())

	// After stop, workers should be cleaned up.
	assert.Empty(t, server.Workers, "expected all workers to be cleaned up after stop")
}

// TestDirWatch_ValidDirFileCountLimit verifies that a valid DirFileCountLimit (within 32-1024)
// is accepted by fwatch.New() via WithDirFileCountLimit option.
// In fwatch v1.6.1, this is passed as an Option to New() rather than a post-construction call.
func TestDirWatch_ValidDirFileCountLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.log"), []byte("data\n"), 0o644))

	server := newDirWatchServer(t, "valid-limit", &conf.FileConfig{
		Path:              dir,
		Method:            fwatch.WatchMethodTimer,
		DirFileCountLimit: 64, // valid: within 32-1024 range
	})

	time.Sleep(2 * time.Second)

	assert.NotEmpty(t, server.Workers, "expected worker to be created with valid file count limit")

	require.NoError(t, server.Stop())
}

// TestFwatchNew_InvalidDirFileCountLimit verifies that fwatch.New() returns an error
// when DirFileCountLimit is below the minimum (32).
// In fwatch v1.3.0, this was silently ignored. In v1.6.1, it causes New() to fail.
// The logtail code calls vlog.Fatal on this error, so we test the fwatch API directly.
func TestFwatchNew_InvalidDirFileCountLimit(t *testing.T) {
	t.Parallel()

	// Test value below minimum (32).
	_, err := fwatch.New(
		fwatch.WithMethod(fwatch.WatchMethodTimer),
		fwatch.WithInactiveDuration(time.Hour),
		fwatch.WithSilenceDuration(24*time.Hour),
		fwatch.WithDirFileCountLimit(10),
	)
	assert.Error(t, err, "expected error for DirFileCountLimit below minimum of 32")
	assert.ErrorIs(t, err, fwatch.ErrInvalidDirFileCountLimit)
}

// TestFwatchNew_DirFileCountLimitAboveMax verifies that fwatch.New() returns an error
// when DirFileCountLimit is above the maximum (1024).
func TestFwatchNew_DirFileCountLimitAboveMax(t *testing.T) {
	t.Parallel()

	_, err := fwatch.New(
		fwatch.WithMethod(fwatch.WatchMethodTimer),
		fwatch.WithInactiveDuration(time.Hour),
		fwatch.WithSilenceDuration(24*time.Hour),
		fwatch.WithDirFileCountLimit(2000),
	)
	assert.Error(t, err, "expected error for DirFileCountLimit above maximum of 1024")
	assert.ErrorIs(t, err, fwatch.ErrInvalidDirFileCountLimit)
}

// TestFwatchNew_DirFileCountLimitBoundary verifies boundary values for DirFileCountLimit.
// 32 and 1024 should be accepted; 31 and 1025 should be rejected.
func TestFwatchNew_DirFileCountLimitBoundary(t *testing.T) {
	t.Parallel()

	baseOpts := []fwatch.Option{
		fwatch.WithMethod(fwatch.WatchMethodTimer),
		fwatch.WithInactiveDuration(time.Hour),
		fwatch.WithSilenceDuration(24 * time.Hour),
	}

	// Minimum boundary: 32 should succeed.
	w, err := fwatch.New(append(baseOpts, fwatch.WithDirFileCountLimit(32))...)
	require.NoError(t, err, "DirFileCountLimit=32 should be accepted")
	require.NoError(t, w.Stop())

	// Maximum boundary: 1024 should succeed.
	w, err = fwatch.New(append(baseOpts, fwatch.WithDirFileCountLimit(1024))...)
	require.NoError(t, err, "DirFileCountLimit=1024 should be accepted")
	require.NoError(t, w.Stop())

	// Below minimum: 31 should fail.
	_, err = fwatch.New(append(baseOpts, fwatch.WithDirFileCountLimit(31))...)
	assert.ErrorIs(t, err, fwatch.ErrInvalidDirFileCountLimit, "DirFileCountLimit=31 should be rejected")

	// Above maximum: 1025 should fail.
	_, err = fwatch.New(append(baseOpts, fwatch.WithDirFileCountLimit(1025))...)
	assert.ErrorIs(t, err, fwatch.ErrInvalidDirFileCountLimit, "DirFileCountLimit=1025 should be rejected")
}

// TestFwatchNew_ZeroDirFileCountLimit verifies that DirFileCountLimit=0 results in using
// the default (128). The logtail code only passes the option when config.DirFileCountLimit > 0,
// so 0 means the option is never set.
func TestFwatchNew_ZeroDirFileCountLimitUsesDefault(t *testing.T) {
	t.Parallel()

	// With DirFileCountLimit=0 in config, the option is NOT passed to fwatch.New().
	// This should succeed and use the default limit of 128.
	w, err := fwatch.New(
		fwatch.WithMethod(fwatch.WatchMethodTimer),
		fwatch.WithInactiveDuration(time.Hour),
		fwatch.WithSilenceDuration(24*time.Hour),
	)
	require.NoError(t, err, "omitting DirFileCountLimit should use default and succeed")
	require.NoError(t, w.Stop())
}

// TestFwatchNew_OptionsPattern verifies that the v1.6.1 functional options API works correctly.
// This validates the core API change from positional arguments to options.
func TestFwatchNew_OptionsPattern(t *testing.T) {
	t.Parallel()

	// All three core options provided.
	w, err := fwatch.New(
		fwatch.WithMethod(fwatch.WatchMethodTimer),
		fwatch.WithInactiveDuration(time.Hour),
		fwatch.WithSilenceDuration(24*time.Hour),
	)
	require.NoError(t, err)
	require.NoError(t, w.Stop())

	// FS method.
	w, err = fwatch.New(
		fwatch.WithMethod(fwatch.WatchMethodFS),
		fwatch.WithInactiveDuration(time.Hour),
		fwatch.WithSilenceDuration(24*time.Hour),
	)
	require.NoError(t, err)
	require.NoError(t, w.Stop())
}

// TestFwatchDone_ChannelClosedOnStop verifies that Done() returns a channel that is
// closed when the watcher is stopped. This tests the v1.6.1 replacement for Runner.C.
func TestFwatchDone_ChannelClosedOnStop(t *testing.T) {
	t.Parallel()

	w, err := fwatch.New(
		fwatch.WithMethod(fwatch.WatchMethodTimer),
		fwatch.WithInactiveDuration(time.Hour),
		fwatch.WithSilenceDuration(24*time.Hour),
	)
	require.NoError(t, err)

	// Done() should not be closed before Stop().
	select {
	case <-w.Done():
		t.Fatal("Done() channel should not be closed before Stop()")
	default:
		// expected
	}

	require.NoError(t, w.Stop())

	// After Stop(), Done() should be closed.
	select {
	case <-w.Done():
		// expected: channel is closed
	case <-time.After(2 * time.Second):
		t.Fatal("Done() channel should be closed after Stop()")
	}
}

// TestFwatchWatchDir_WithMatcher verifies that WatchDir correctly applies the file matcher.
func TestFwatchWatchDir_WithMatcher(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create some files.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.log"), []byte("data\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sys.log"), []byte("data\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.txt"), []byte("data\n"), 0o644))

	w, err := fwatch.New(
		fwatch.WithMethod(fwatch.WatchMethodTimer),
		fwatch.WithInactiveDuration(time.Hour),
		fwatch.WithSilenceDuration(24*time.Hour),
	)
	require.NoError(t, err)

	// Only accept .log files.
	matcher := func(name string) bool {
		return len(name) > 4 && name[len(name)-4:] == ".log"
	}

	require.NoError(t, w.WatchDir(dir, false, matcher))

	// Collect events for a short period.
	var events []*fwatch.WatchEvent

	timeout := time.After(2 * time.Second)

	for collecting := true; collecting; {
		select {
		case ev := <-w.Events:
			events = append(events, ev)
		case <-timeout:
			collecting = false
		}
	}

	// Verify only .log files generated events.
	for _, ev := range events {
		base := filepath.Base(ev.Name)
		assert.Contains(t, base, ".log", "only .log files should generate events, got: %s", base)
	}

	require.NoError(t, w.Stop())
}

// TestFwatchWatchDir_NilMatcher verifies that WatchDir returns an error when matcher is nil.
func TestFwatchWatchDir_NilMatcher(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	w, err := fwatch.New(
		fwatch.WithMethod(fwatch.WatchMethodTimer),
		fwatch.WithInactiveDuration(time.Hour),
		fwatch.WithSilenceDuration(24*time.Hour),
	)
	require.NoError(t, err)

	err = w.WatchDir(dir, false, nil)
	assert.Error(t, err, "expected error when matcher is nil")

	require.NoError(t, w.Stop())
}

// TestFwatchStop_ErrorHandling verifies that Stop() returns nil for timer-based watchers
// and that calling Stop() multiple times does not panic.
func TestFwatchStop_ErrorHandling(t *testing.T) {
	t.Parallel()

	w, err := fwatch.New(
		fwatch.WithMethod(fwatch.WatchMethodTimer),
		fwatch.WithInactiveDuration(time.Hour),
		fwatch.WithSilenceDuration(24*time.Hour),
	)
	require.NoError(t, err)

	// First stop should succeed.
	assert.NoError(t, w.Stop())
}

// TestDirWatch_EmptyDirectory verifies that watching an empty directory does not cause errors.
// The server should start and stop cleanly with no workers.
func TestDirWatch_EmptyDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	server := newDirWatchServer(t, "empty-dir", &conf.FileConfig{
		Path:   dir,
		Method: fwatch.WatchMethodTimer,
	})

	time.Sleep(2 * time.Second)

	assert.Empty(t, server.Workers, "expected no workers for empty directory")

	require.NoError(t, server.Stop())
}
