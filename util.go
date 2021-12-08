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

const defaultMapSize = 4

func isNumberChar(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlphabetChar(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

// nolint:varnamelen //ignore this.
func indexToNextLineStart(format *Format, message []byte) []byte {
	l := len(message)
	i := 0

	for i < l {
		indexLineEnd(message, &l, &i)
		ignoreLineEnd(message, &l, &i)

		if format == nil || format.PrefixMatch(message[i:]) {
			return message[i:]
		}
	}

	return nil
}

func indexToLineStart(format *Format, data []byte) []byte {
	if format == nil || format.PrefixMatch(data) {
		return data
	}

	return indexToNextLineStart(format, data)
}

func isFollowingLine(format *Format, bytes []byte) bool {
	if format == nil {
		format = defaultRunner.Config.DefaultFormat
	}

	if format != nil {
		return !format.PrefixMatch(bytes)
	}

	return bytes[0] == ' ' || bytes[0] == '\t'
}

func isLineEnd(b byte) bool {
	return b == '\n' || b == '\r'
}

func indexLineEnd(bytes []byte, length, index *int) {
	for ; *index < *length && !isLineEnd(bytes[*index]); *index++ {
	}
}

func ignoreLineEnd(bytes []byte, length, index *int) {
	for ; *index < *length && isLineEnd(bytes[*index]); *index++ {
	}
}

//  flag `-F` is same as `--follow=name --retry`
func followRetryTailCommand(f string) string {
	return fmt.Sprintf("tail -F %s", f)
}
