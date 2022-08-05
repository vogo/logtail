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

package match

type ContainsMatcher struct {
	contains bool
	pattern  string
	plen     int
	kmp      []int
}

func NewContainsMatcher(pattern string, contains bool) *ContainsMatcher {
	if pattern == "" {
		panic("pattern nil")
	}

	containsMatcher := &ContainsMatcher{
		contains: contains,
		pattern:  pattern,
		plen:     len(pattern),
		kmp:      make([]int, len(pattern)+1),
	}

	containsMatcher.kmp[0] = -1

	//nolint:varnamelen //ignore this
	for i := 1; i < containsMatcher.plen; i++ {
		j := containsMatcher.kmp[i-1]
		for j > -1 && containsMatcher.pattern[j+1] != containsMatcher.pattern[i] {
			j = containsMatcher.kmp[j]
		}

		if containsMatcher.pattern[j+1] == containsMatcher.pattern[i] {
			j++
		}

		containsMatcher.kmp[i] = j
	}

	return containsMatcher
}

func (cm *ContainsMatcher) Match(bytes []byte) bool {
	length := len(bytes)

	if length == 0 {
		return false
	}

	//nolint:varnamelen //ignore this
	j := -1

	for i := 0; i < length; i++ {
		for j > -1 && cm.pattern[j+1] != bytes[i] {
			j = cm.kmp[j]
		}

		if cm.pattern[j+1] == bytes[i] {
			j++
		}

		if j+1 == cm.plen {
			return cm.contains
		}
	}

	return !cm.contains
}
