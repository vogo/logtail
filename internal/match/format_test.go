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
