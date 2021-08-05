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

func TestWildcardMatch(t *testing.T) {
	t.Parallel()

	assert.False(t, logtail.WildcardMatch("!!!!", nil))
	assert.True(t, logtail.WildcardMatch("", nil))
	assert.True(t, logtail.WildcardMatch("", []byte("abcd")))

	assert.False(t, logtail.WildcardMatch("!!!!", []byte("a")))
	assert.False(t, logtail.WildcardMatch("!!!!", []byte("abcd")))
	assert.False(t, logtail.WildcardMatch("!!!!", []byte("123a")))

	assert.True(t, logtail.WildcardMatch("!!!!", []byte("1234")))
	assert.True(t, logtail.WildcardMatch("!!!!", []byte("1234abcd")))
	assert.True(t, logtail.WildcardMatch("!!!!-!!-!!", []byte("2021-01-01")))
	assert.False(t, logtail.WildcardMatch("!!!!-!!-!!", []byte("2021001001")))

	assert.False(t, logtail.WildcardMatch("~~~~", []byte("1234abcd")))
	assert.True(t, logtail.WildcardMatch("~~~~", []byte("abcd")))
	assert.True(t, logtail.WildcardMatch("~~~~", []byte("abcd1234")))

	assert.True(t, logtail.WildcardMatch("????", []byte("1234")))
	assert.True(t, logtail.WildcardMatch("????", []byte("abcd")))
}
