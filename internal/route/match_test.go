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

package route_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/route"
)

func TestNewMatchers(t *testing.T) {
	t.Parallel()

	matchers, err := route.NewMatchers(nil)
	assert.NoError(t, err)
	assert.Nil(t, matchers)

	matchers, err = route.NewMatchers([]*conf.MatcherConfig{
		{Contains: []string{"ERROR"}, NotContains: []string{"IGNORE"}},
	})
	assert.NoError(t, err)
	assert.Len(t, matchers, 2)
}

func TestBuildMatchers(t *testing.T) {
	t.Parallel()

	matchers := route.BuildMatchers([]*conf.MatcherConfig{
		{Contains: []string{"ERROR", "WARN"}},
		{NotContains: []string{"DEBUG"}},
	})

	assert.Len(t, matchers, 3)
}

func TestBuildMatcher(t *testing.T) {
	t.Parallel()

	matchers := route.BuildMatcher(&conf.MatcherConfig{
		Contains:    []string{"ERROR"},
		NotContains: []string{"IGNORE", "SKIP"},
	})

	assert.Len(t, matchers, 3)

	// First is contains matcher
	assert.True(t, matchers[0].Match([]byte("ERROR happened")))
	assert.False(t, matchers[0].Match([]byte("INFO happened")))

	// Second and third are not_contains matchers
	assert.False(t, matchers[1].Match([]byte("IGNORE this")))
	assert.True(t, matchers[1].Match([]byte("ERROR real")))
}

func TestBuildMatcher_Empty(t *testing.T) {
	t.Parallel()

	matchers := route.BuildMatcher(&conf.MatcherConfig{})
	assert.Empty(t, matchers)
}
