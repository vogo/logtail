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
	"fmt"
	"sync/atomic"
	"time"

	"github.com/vogo/logtail/internal/consts"
)

// nolint:gochecknoglobals //ignore this
// statistic duration, default one day.
var statisticDuration = time.Hour * 24

func SetTransStatisticDuration(d time.Duration) {
	statisticDuration = d
}

// Counter is a utility to count the count of transfer times.
type Counter struct {
	count        int32
	countStartAt time.Time
	countEndAt   time.Time
}

// CountReset reset the counter and return the statistic message.
func (c *Counter) CountReset() string {
	count := atomic.SwapInt32(&c.count, 0)

	countResult := fmt.Sprintf("[logtail statistics] Count: %d, Time Peroid: %s ~ %s",
		count,
		c.countStartAt.Format(consts.FormatDateTime),
		time.Now().Format(consts.FormatDateTime))

	c.countStartAt = time.Now()
	c.countEndAt = c.countStartAt.Add(statisticDuration)

	return countResult
}

// CountIncr increase the counter.
func (c *Counter) CountIncr() {
	atomic.AddInt32(&c.count, 1)
}

// CountStat return statistic message if exceeding the end of the time range.
func (c *Counter) CountStat() (string, bool) {
	if time.Now().Before(c.countEndAt) {
		return "", false
	}

	return c.CountReset(), true
}
