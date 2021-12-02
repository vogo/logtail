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

package logtail

import "fmt"

// Format the log format.
type Format struct {
	Prefix string `json:"prefix"` // the wildcard of the line prefix of a log record
}

// PrefixMatch whether the given data has a prefix of a new record.
func (f *Format) PrefixMatch(data []byte) bool {
	return WildcardMatch(f.Prefix, data)
}

// String format string info.
func (f *Format) String() string {
	return fmt.Sprintf("format{prefix:%s}", f.Prefix)
}

// WildcardMatch -  finds whether the bytes match/satisfies the pattern wildcard.
// supports:
// - '?' as one byte char
// - '~' as one alphabet char
// - '!' as one number char
// NOT support '*' for none or many char.
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
			if !isAlphabetChar(b) {
				return false
			}
		case '!':
			if !isNumberChar(b) {
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
