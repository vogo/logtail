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
	"sync"
	"sync/atomic"
	"time"
)

const tokenScale int64 = 1000

// rateLimiter is a simple token-bucket rate limiter using atomic operations.
type rateLimiter struct {
	tokens    atomic.Int64 // current token count (scaled by tokenScale for sub-unit precision)
	maxTokens int64        // burst size (scaled)
	refillQty int64        // tokens added per tick (scaled)
	ticker    *time.Ticker
	done      chan struct{}
	stopOnce  sync.Once
}

// newRateLimiter creates a rate limiter. rate is events/second, burst is max burst size.
func newRateLimiter(rate float64, burst int) *rateLimiter {
	if burst <= 0 {
		burst = 1
	}

	maxTokens := int64(burst) * tokenScale

	r := &rateLimiter{
		maxTokens: maxTokens,
		done:      make(chan struct{}),
	}

	r.tokens.Store(maxTokens)

	// Calculate ticker interval and refill quantity.
	// For rate < 1, tick less frequently and add 1 token per tick.
	// For rate >= 1, tick every 100ms and add rate/10 tokens per tick.
	var interval time.Duration

	if rate < 1 {
		interval = time.Duration(float64(time.Second) / rate)
		r.refillQty = tokenScale
	} else {
		interval = 100 * time.Millisecond
		r.refillQty = int64(rate * float64(tokenScale) / 10)

		r.refillQty = max(r.refillQty, 1)
	}

	r.ticker = time.NewTicker(interval)

	go r.refillLoop()

	return r
}

func (r *rateLimiter) refillLoop() {
	for {
		select {
		case <-r.done:
			return
		case <-r.ticker.C:
			for {
				current := r.tokens.Load()
				newVal := min(current+r.refillQty, r.maxTokens)

				if r.tokens.CompareAndSwap(current, newVal) {
					break
				}
			}
		}
	}
}

// Allow returns true if a token is available (non-blocking).
// Uses CAS to atomically check and decrement, avoiding spurious denials
// that can occur with subtract-then-restore under high concurrency.
func (r *rateLimiter) Allow() bool {
	for {
		current := r.tokens.Load()
		if current < tokenScale {
			return false
		}

		if r.tokens.CompareAndSwap(current, current-tokenScale) {
			return true
		}
	}
}

// Stop stops the refill ticker. Safe to call multiple times.
func (r *rateLimiter) Stop() {
	r.stopOnce.Do(func() {
		r.ticker.Stop()
		close(r.done)
	})
}
