//go:build integration

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

package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	buildOnce  sync.Once
	binaryPath string
	buildErr   error
)

// projectRoot returns the absolute path to the project root.
// Determined from the location of this source file (integrations/helper/helper.go).
func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// helper.go -> helper/ -> integrations/ -> project root
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

// BuildBinary compiles the logtail binary once per test run.
// Returns the absolute path to the compiled binary.
// Calls t.Fatal if compilation fails.
func BuildBinary(t *testing.T) string {
	t.Helper()

	buildOnce.Do(func() {
		dir := filepath.Join(os.TempDir(), fmt.Sprintf("logtail-integration-test-%d", os.Getpid()))

		if err := os.MkdirAll(dir, 0o755); err != nil {
			buildErr = fmt.Errorf("create binary dir: %w", err)
			return
		}

		binaryPath = filepath.Join(dir, "logtail")
		root := projectRoot()

		cmd := exec.Command("go", "build", "-o", binaryPath, ".")
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			buildErr = fmt.Errorf("build logtail binary: %w", err)
			return
		}
	})

	if buildErr != nil {
		t.Fatal(buildErr)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(filepath.Dir(binaryPath))
	})

	return binaryPath
}

// TempDir creates a temporary directory with the given name prefix.
// Registers t.Cleanup to remove the directory after the test.
func TempDir(t *testing.T, name string) string {
	t.Helper()

	dir, err := os.MkdirTemp(os.TempDir(), name)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	return dir
}

// WriteConfig marshals a config map to JSON and writes it to dir/config.json.
// Returns the absolute path to the config file.
func WriteConfig(t *testing.T, dir string, config map[string]any) string {
	t.Helper()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

// safeBuf is a thread-safe buffer that implements io.Writer.
type safeBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.buf.Write(p)
}

func (s *safeBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.buf.String()
}

// LogtailProcess wraps an exec.Cmd for the logtail binary.
type LogtailProcess struct {
	cmd    *exec.Cmd
	stdout *safeBuf
	stderr *safeBuf
	done   chan struct{}
}

// RunLogtail starts the logtail binary with the given arguments.
func RunLogtail(t *testing.T, binary string, args ...string) *LogtailProcess {
	t.Helper()

	cmd := exec.Command(binary, args...)
	stdoutBuf := &safeBuf{}
	stderrBuf := &safeBuf{}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	proc := &LogtailProcess{
		cmd:    cmd,
		stdout: stdoutBuf,
		stderr: stderrBuf,
		done:   make(chan struct{}),
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	go func() {
		_ = cmd.Wait()
		close(proc.done)
	}()

	return proc
}

// Stdout returns the captured stdout content so far. Thread-safe.
func (p *LogtailProcess) Stdout() string {
	return p.stdout.String()
}

// Stderr returns the captured stderr content so far. Thread-safe.
func (p *LogtailProcess) Stderr() string {
	return p.stderr.String()
}

// Stop sends interrupt signal to the process and waits up to 5 seconds for exit.
func (p *LogtailProcess) Stop() {
	_ = p.cmd.Process.Signal(os.Interrupt)

	select {
	case <-p.done:
	case <-time.After(5 * time.Second):
		_ = p.cmd.Process.Kill()
		<-p.done
	}
}

// Wait waits for the process to exit within the given timeout.
func (p *LogtailProcess) Wait(timeout time.Duration) error {
	select {
	case <-p.done:
		return nil
	case <-time.After(timeout):
		_ = p.cmd.Process.Kill()
		<-p.done

		return fmt.Errorf("process did not exit within %s", timeout)
	}
}

// ExitCode returns the process exit code. Returns -1 if the process has not exited.
func (p *LogtailProcess) ExitCode() int {
	if p.cmd.ProcessState == nil {
		return -1
	}

	return p.cmd.ProcessState.ExitCode()
}

// WaitForStdoutContains polls proc.Stdout() until it contains the expected substring or timeout expires.
func WaitForStdoutContains(t *testing.T, proc *LogtailProcess, expected string, timeout time.Duration) {
	t.Helper()

	deadline := time.After(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if strings.Contains(proc.Stdout(), expected) {
			return
		}

		select {
		case <-deadline:
			t.Errorf("timed out waiting for stdout to contain %q, got:\n%s", expected, proc.Stdout())

			return
		case <-ticker.C:
		}
	}
}

// AssertStdoutContains asserts proc.Stdout() contains the expected substring.
func AssertStdoutContains(t *testing.T, proc *LogtailProcess, expected string) {
	t.Helper()

	out := proc.Stdout()
	if !strings.Contains(out, expected) {
		t.Errorf("expected stdout to contain %q, got:\n%s", expected, out)
	}
}

// vlogLinePattern matches vlog output lines (e.g., "2026/03/22 09:36:09.348 INFO ...")
// and also continuation lines that are part of multi-line vlog messages.
var vlogLinePattern = regexp.MustCompile(`^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}\.\d{3} `)

// filterNonLogLines returns only lines from output that are not vlog framework lines.
// This is used for "not contains" assertions to avoid false positives from vlog messages
// that echo command text.
func filterNonLogLines(output string) string {
	lines := strings.Split(output, "\n")
	var filtered []string
	inLogBlock := false

	for _, line := range lines {
		if vlogLinePattern.MatchString(line) {
			inLogBlock = true
			continue
		}

		// Lines starting with whitespace after a vlog line are continuations
		// (e.g., multi-line command text in worker log messages)
		if inLogBlock && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") || line == "'") {
			continue
		}

		inLogBlock = false
		filtered = append(filtered, line)
	}

	return strings.Join(filtered, "\n")
}

// AssertStdoutNotContains asserts that non-vlog lines in proc.Stdout()
// do NOT contain the substring. Vlog framework lines are excluded because
// they may echo command text in worker log messages.
func AssertStdoutNotContains(t *testing.T, proc *LogtailProcess, unexpected string) {
	t.Helper()

	out := filterNonLogLines(proc.Stdout())
	if strings.Contains(out, unexpected) {
		t.Errorf("expected stdout (non-log lines) NOT to contain %q, got:\n%s", unexpected, out)
	}
}

// AssertStderrContains asserts proc.Stderr() contains the expected substring.
func AssertStderrContains(t *testing.T, proc *LogtailProcess, expected string) {
	t.Helper()

	out := proc.Stderr()
	if !strings.Contains(out, expected) {
		t.Errorf("expected stderr to contain %q, got:\n%s", expected, out)
	}
}

// AssertFileContains reads the file at path and asserts it contains expected.
func AssertFileContains(t *testing.T, path string, expected string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}

	if !strings.Contains(string(data), expected) {
		t.Errorf("expected file %s to contain %q, got:\n%s", path, expected, string(data))
	}
}

// AssertFileExists asserts at least one file with the given prefix exists in dir.
// Returns the path of the first matching file.
func AssertFileExists(t *testing.T, dir string, prefix string) string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) {
			return filepath.Join(dir, entry.Name())
		}
	}

	t.Fatalf("no file with prefix %q found in %s", prefix, dir)

	return ""
}

// WaitForFileWithPrefix polls until a file with the given prefix appears in dir or timeout expires.
// Returns the path of the first matching file.
func WaitForFileWithPrefix(t *testing.T, dir string, prefix string, timeout time.Duration) string {
	t.Helper()

	deadline := time.After(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		entries, err := os.ReadDir(dir)
		if err == nil {
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), prefix) {
					return filepath.Join(dir, entry.Name())
				}
			}
		}

		select {
		case <-deadline:
			t.Fatalf("no file with prefix %q found in %s within %s", prefix, dir, timeout)

			return ""
		case <-ticker.C:
		}
	}
}

// WaitForFile polls every 100ms until the file exists or timeout expires.
func WaitForFile(t *testing.T, path string, timeout time.Duration) {
	t.Helper()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	deadline := time.After(timeout)

	for {
		select {
		case <-deadline:
			t.Fatalf("file %s did not appear within %s", path, timeout)
			return
		case <-ticker.C:
			if _, err := os.Stat(path); err == nil {
				return
			}
		}
	}
}

// AppendToFile opens the file in append mode and writes the content.
// Creates the file if it does not exist.
func AppendToFile(t *testing.T, path string, content string) {
	t.Helper()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
}
