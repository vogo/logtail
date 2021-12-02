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

package transfer

const DoubleSize = 2

func EscapeLimitJSONBytes(b []byte, capacity int) []byte {
	size := len(b)
	num := capacity

	if size < num {
		num = size
	}

	t := make([]byte, num*DoubleSize)

	index := 0
	from := 0

	for i := 0; i < num; i++ {
		for ; i < num && b[i] != '\n' && b[i] != '\t' && b[i] != '"'; i++ {
		}

		copy(t[index:], b[from:i])
		index += i - from
		from = i + 1

		if i < num {
			fillJSONEscape(t, &index, b[i])
		}
	}

	// from <= size means not reach the end of the bytes
	if from <= size {
		// remove uncompleted utf8 bytes
		for i := index - 1; i >= 0 && t[i]&0xC0 == 0x80; i-- {
			index = i - 1
		}
	}

	return t[:index]
}

func fillJSONEscape(t []byte, index *int, b byte) {
	t[*index] = '\\'
	*index++

	switch b {
	case '\n':
		t[*index] = 'n'
	case '\t':
		t[*index] = 't'
	case '"':
		t[*index] = '"'
	}

	*index++
}
