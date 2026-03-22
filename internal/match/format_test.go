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

package match_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail/internal/match"
)

func TestSplitFirstLog(t *testing.T) {
	t.Parallel()

	s := `2022-06-02 17:17:01.542  INFO abc
2022-06-02 17:17:01.545  INFO def
2022-06-02 17:17:01.546 ERROR ghk`

	format := &match.Format{Prefix: "!!!!-!!-!!"}
	first, remain := match.SplitFirstLog(format, []byte(s))

	assert.Equal(t, []byte("2022-06-02 17:17:01.542  INFO abc\n"), first)

	first, remain = match.SplitFirstLog(format, remain)

	assert.Equal(t, []byte("2022-06-02 17:17:01.545  INFO def\n"), first)
	assert.Equal(t, []byte("2022-06-02 17:17:01.546 ERROR ghk"), remain)
}

func TestSplitFirstLog_Multiline(t *testing.T) {
	t.Parallel()

	s := "2022-06-02 ERROR something\n  stack line 1\n  stack line 2\n2022-06-02 INFO ok\n"
	format := &match.Format{Prefix: "!!!!-!!-!!"}

	first, remain := match.SplitFirstLog(format, []byte(s))
	assert.Equal(t, "2022-06-02 ERROR something\n  stack line 1\n  stack line 2\n", string(first))
	assert.Equal(t, "2022-06-02 INFO ok\n", string(remain))
}

func TestSplitFirstLog_NilFormat(t *testing.T) {
	t.Parallel()

	s := "line1\n  continuation\nline2\n"
	first, remain := match.SplitFirstLog(nil, []byte(s))
	assert.Equal(t, "line1\n  continuation\n", string(first))
	assert.Equal(t, "line2\n", string(remain))
}

func TestSplitFollowingLog(t *testing.T) {
	t.Parallel()

	s := "  stack1\n  stack2\n2022-06-02 next entry\n"
	format := &match.Format{Prefix: "!!!!-!!-!!"}

	following, remain := match.SplitFollowingLog(format, []byte(s))
	assert.Equal(t, "  stack1\n  stack2\n", string(following))
	assert.Equal(t, "2022-06-02 next entry\n", string(remain))
}

func TestSplitFollowingLog_NilFormat(t *testing.T) {
	t.Parallel()

	s := "\tcontinuation\nline2\n"
	following, remain := match.SplitFollowingLog(nil, []byte(s))
	assert.Equal(t, "\tcontinuation\n", string(following))
	assert.Equal(t, "line2\n", string(remain))
}

func TestFormatString(t *testing.T) {
	t.Parallel()

	f := &match.Format{Prefix: "!!!!-!!-!!"}
	assert.Equal(t, "format{prefix:!!!!-!!-!!}", f.String())
}

func TestFormatPrefixMatch(t *testing.T) {
	t.Parallel()

	f := &match.Format{Prefix: "!!!!-!!-!!"}
	assert.True(t, f.PrefixMatch([]byte("2024-01-15 something")))
	assert.False(t, f.PrefixMatch([]byte("  not a date")))
	assert.False(t, f.PrefixMatch([]byte("abcd-ef-gh")))
}

func TestIndexToLineStart(t *testing.T) {
	t.Parallel()

	format := &match.Format{Prefix: "!!!!-!!-!!"}

	// Already at line start
	data := []byte("2024-01-15 something")
	result := match.IndexToLineStart(format, data)
	assert.Equal(t, data, result)

	// Not at line start, needs to skip to next
	data2 := []byte("  continuation\n2024-01-15 next")
	result2 := match.IndexToLineStart(format, data2)
	assert.Equal(t, "2024-01-15 next", string(result2))

	// Nil format returns data as-is
	result3 := match.IndexToLineStart(nil, data)
	assert.Equal(t, data, result3)
}

func TestIndexToLineStart_NoMatch(t *testing.T) {
	t.Parallel()

	format := &match.Format{Prefix: "!!!!-!!-!!"}
	data := []byte("  no match here\n  still no match")
	result := match.IndexToLineStart(format, data)
	assert.Nil(t, result)
}

func TestIsFollowingLine(t *testing.T) {
	t.Parallel()

	format := &match.Format{Prefix: "!!!!-!!-!!"}

	assert.True(t, match.IsFollowingLine(format, []byte("  indented")))
	assert.True(t, match.IsFollowingLine(format, []byte("not a date")))
	assert.False(t, match.IsFollowingLine(format, []byte("2024-01-15 a date")))

	// Nil format: only spaces/tabs are following
	assert.True(t, match.IsFollowingLine(nil, []byte(" space")))
	assert.True(t, match.IsFollowingLine(nil, []byte("\ttab")))
	assert.False(t, match.IsFollowingLine(nil, []byte("normal")))
}
