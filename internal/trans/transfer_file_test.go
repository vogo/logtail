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

package trans_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vogo/logtail/internal/trans"
)

func TestFileTransfer_Name(t *testing.T) {
	t.Parallel()

	ft := trans.NewFileTransfer("test-file", t.TempDir())
	assert.Equal(t, "test-file", ft.Name())
}

func TestFileTransfer_StartStopLifecycle(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ft := trans.NewFileTransfer("lifecycle", dir)

	err := ft.Start()
	require.NoError(t, err)

	// Send data and wait for async processing
	for range 5 {
		_ = ft.Trans("source", []byte("hello world"))
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)

	err = ft.Stop()
	assert.NoError(t, err)

	// Wait for file submission
	time.Sleep(200 * time.Millisecond)

	// Check that a file was created in the directory
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	// Check content
	found := false

	for _, entry := range entries {
		data, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
		require.NoError(t, readErr)

		if len(data) > 0 {
			assert.Contains(t, string(data), "hello world")

			found = true
		}
	}

	assert.True(t, found, "expected at least one file with content")
}

func TestFileTransfer_TransAfterStop(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ft := trans.NewFileTransfer("stopped", dir)

	require.NoError(t, ft.Start())
	require.NoError(t, ft.Stop())

	time.Sleep(50 * time.Millisecond)

	// Should not panic
	err := ft.Trans("source", []byte("data"))
	assert.NoError(t, err)
}
