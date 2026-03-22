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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail/internal/trans"
)

func TestCounterIncr(t *testing.T) {
	t.Parallel()

	c := &trans.Counter{}

	c.CountIncr()
	c.CountIncr()
	c.CountIncr()

	result := c.CountReset()
	assert.Contains(t, result, "Count: 3")
}

func TestCounterReset(t *testing.T) {
	t.Parallel()

	c := &trans.Counter{}
	c.CountIncr()

	result := c.CountReset()
	assert.Contains(t, result, "Count: 1")
	assert.Contains(t, result, "[logtail statistics]")

	// After reset, count should be 0
	result2 := c.CountReset()
	assert.Contains(t, result2, "Count: 0")
}

func TestCounterStat_BeforeExpiry(t *testing.T) {
	t.Parallel()

	trans.SetTransStatisticDuration(time.Hour)

	c := &trans.Counter{}
	c.CountReset() // initialize time range

	c.CountIncr()

	msg, expired := c.CountStat()
	assert.False(t, expired)
	assert.Empty(t, msg)
}

func TestCounterStat_AfterExpiry(t *testing.T) {
	t.Parallel()

	trans.SetTransStatisticDuration(time.Millisecond)

	c := &trans.Counter{}
	c.CountReset() // initialize time range

	c.CountIncr()

	time.Sleep(2 * time.Millisecond)

	msg, expired := c.CountStat()
	assert.True(t, expired)
	assert.Contains(t, msg, "Count: 1")
}

func TestSetTransStatisticDuration(t *testing.T) {
	t.Parallel()

	trans.SetTransStatisticDuration(time.Minute * 5)
	// no panic is good enough
}
