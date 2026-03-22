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

package trans

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiterAllowWithinBurst(t *testing.T) {
	t.Parallel()

	rl := newRateLimiter(10, 5)
	defer rl.Stop()

	// Should allow burst of 5
	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow(), "should allow request %d within burst", i)
	}
}

func TestRateLimiterDenyBeyondBurst(t *testing.T) {
	t.Parallel()

	rl := newRateLimiter(10, 3)
	defer rl.Stop()

	// Exhaust burst
	for i := 0; i < 3; i++ {
		assert.True(t, rl.Allow())
	}

	// Should deny beyond burst
	assert.False(t, rl.Allow(), "should deny request beyond burst")
}

func TestRateLimiterRefill(t *testing.T) {
	t.Parallel()

	rl := newRateLimiter(10, 1)
	defer rl.Stop()

	// Exhaust the single token
	assert.True(t, rl.Allow())
	assert.False(t, rl.Allow())

	// Wait for refill (rate=10/s, tick every 100ms)
	time.Sleep(200 * time.Millisecond)

	// Should have refilled
	assert.True(t, rl.Allow(), "should allow after refill")
}

func TestRateLimiterDefaultBurst(t *testing.T) {
	t.Parallel()

	// burst=0 should default to 1
	rl := newRateLimiter(1, 0)
	defer rl.Stop()

	assert.True(t, rl.Allow())
	assert.False(t, rl.Allow())
}

func TestRateLimiterLowRate(t *testing.T) {
	t.Parallel()

	// rate < 1: one token every ~2 seconds
	rl := newRateLimiter(0.5, 1)
	defer rl.Stop()

	assert.True(t, rl.Allow())
	assert.False(t, rl.Allow())
}

func TestRateLimiterStop(t *testing.T) {
	t.Parallel()

	rl := newRateLimiter(10, 5)
	rl.Stop()

	// After stop, Allow should still work (just no more refills)
	// Consume remaining tokens
	allowed := 0
	for i := 0; i < 10; i++ {
		if rl.Allow() {
			allowed++
		}
	}

	assert.Equal(t, 5, allowed)
}

func TestRateLimiterDoubleStop(t *testing.T) {
	t.Parallel()

	rl := newRateLimiter(10, 5)

	// Double stop should not panic.
	rl.Stop()
	rl.Stop()
}
