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

package util

import "fmt"

const DefaultMapSize = 4

func IsNumberChar(b byte) bool {
	return b >= '0' && b <= '9'
}

func IsAlphabetChar(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func isLineEnd(b byte) bool {
	return b == '\n' || b == '\r'
}

func IndexLineEnd(bytes []byte, length, index *int) {
	for ; *index < *length && !isLineEnd(bytes[*index]); *index++ {
	}
}

func IgnoreLineEnd(bytes []byte, length, index *int) {
	for ; *index < *length && isLineEnd(bytes[*index]); *index++ {
	}
}

// FollowRetryTailCommand flag `-F` is same as `--follow=name --retry`
func FollowRetryTailCommand(f string) string {
	return fmt.Sprintf("tail -F %s", f)
}
