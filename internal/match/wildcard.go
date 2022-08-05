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

import "github.com/vogo/logtail/internal/util"

// WildcardMatch -  finds whether the bytes match/satisfies the pattern wildcard.
// supports:
// - '?' as one byte char
// - '~' as one alphabet char
// - '!' as one number char
// NOT support '*' for none or many char.
//nolint:varnamelen //ignore this.
func WildcardMatch(pattern string, data []byte) bool {
	var p, b byte

	for i, j := 0, 0; i < len(pattern); i++ {
		if j >= len(data) {
			return false
		}

		p = pattern[i]
		b = data[j]

		switch p {
		case '?':
			if len(data) == 0 {
				return false
			}
		case '~':
			if !util.IsAlphabetChar(b) {
				return false
			}
		case '!':
			if !util.IsNumberChar(b) {
				return false
			}
		default:
			if len(data) == 0 || b != p {
				return false
			}
		}

		j++
	}

	return true
}
