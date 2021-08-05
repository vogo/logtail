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

package logtail_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail"
)

func TestMatch(t *testing.T) {
	t.Parallel()

	data := []byte(`2020-12-25 14:54:38.523  ERROR exception occurs`)

	assert.True(t, logtail.NewContainsMatcher("ERROR", true).Match(data))
	assert.True(t, logtail.NewContainsMatcher("exception", true).Match(data))

	assert.False(t, logtail.NewContainsMatcher("ERROR", false).Match(data))
	assert.False(t, logtail.NewContainsMatcher("exception", false).Match(data))
	assert.False(t, logtail.NewContainsMatcher("WARN", true).Match(data))
}

func TestMatch2(t *testing.T) {
	t.Parallel()

	data := []byte(`2020-12-25 14:54:38.523  错误 error 异常 exception 数据找不到信息`)

	assert.True(t, logtail.NewContainsMatcher("error", true).Match(data))
	assert.True(t, logtail.NewContainsMatcher("exception", true).Match(data))
	assert.True(t, logtail.NewContainsMatcher("错误", true).Match(data))
	assert.True(t, logtail.NewContainsMatcher("异常", true).Match(data))

	assert.False(t, logtail.NewContainsMatcher("error", false).Match(data))
	assert.False(t, logtail.NewContainsMatcher("exception", false).Match(data))
	assert.False(t, logtail.NewContainsMatcher("错误", false).Match(data))
	assert.False(t, logtail.NewContainsMatcher("异常", false).Match(data))
	assert.False(t, logtail.NewContainsMatcher("找不到", false).Match(data))

	assert.True(t, logtail.NewContainsMatcher("没问题", false).Match(data))
}
