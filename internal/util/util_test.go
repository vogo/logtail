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

package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail/internal/util"
)

func TestIsNumberChar(t *testing.T) {
	t.Parallel()

	assert.True(t, util.IsNumberChar('0'))
	assert.True(t, util.IsNumberChar('5'))
	assert.True(t, util.IsNumberChar('9'))
	assert.False(t, util.IsNumberChar('a'))
	assert.False(t, util.IsNumberChar('Z'))
	assert.False(t, util.IsNumberChar(' '))
}

func TestIsAlphabetChar(t *testing.T) {
	t.Parallel()

	assert.True(t, util.IsAlphabetChar('a'))
	assert.True(t, util.IsAlphabetChar('z'))
	assert.True(t, util.IsAlphabetChar('A'))
	assert.True(t, util.IsAlphabetChar('Z'))
	assert.False(t, util.IsAlphabetChar('0'))
	assert.False(t, util.IsAlphabetChar(' '))
	assert.False(t, util.IsAlphabetChar('@'))
}

func TestIndexLineEnd(t *testing.T) {
	t.Parallel()

	data := []byte("hello\nworld\r\nfoo")

	idx := util.IndexLineEnd(data, len(data), 0)
	assert.Equal(t, 5, idx) // position of \n

	idx = util.IndexLineEnd(data, len(data), 6)
	assert.Equal(t, 11, idx) // position of \r

	idx = util.IndexLineEnd(data, len(data), 13)
	assert.Equal(t, 16, idx) // end of data
}

func TestIgnoreLineEnd(t *testing.T) {
	t.Parallel()

	data := []byte("hello\n\n\nworld")

	idx := util.IgnoreLineEnd(data, len(data), 5)
	assert.Equal(t, 8, idx) // skips three \n

	data2 := []byte("hello\r\nworld")
	idx = util.IgnoreLineEnd(data2, len(data2), 5)
	assert.Equal(t, 7, idx) // skips \r\n
}

func TestFollowRetryTailCommand(t *testing.T) {
	t.Parallel()

	cmd := util.FollowRetryTailCommand("/var/log/app.log")
	assert.Equal(t, "tail -F /var/log/app.log", cmd)
}

func TestAllStacks(t *testing.T) {
	t.Parallel()

	stacks := util.AllStacks()
	assert.NotEmpty(t, stacks)
	assert.Contains(t, string(stacks), "goroutine")
}
