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

const DoubleSize = 2

func EscapeLimitJSONBytes(bytes []byte, capacity int) []byte {
	size := len(bytes)
	num := capacity

	if size < num {
		num = size
	}

	jsonData := make([]byte, num*DoubleSize)

	srcIndex := 0
	dstIndex := 0
	from := 0

	for ; srcIndex < num; srcIndex++ {
		switch char := bytes[srcIndex]; char {
		case '\n', '\t', '"', '\\':
			copy(jsonData[dstIndex:], bytes[from:srcIndex])

			dstIndex += srcIndex - from

			if srcIndex < num {
				jsonData[dstIndex] = '\\'
				dstIndex++

				jsonData[dstIndex] = toEscapeChar(char)
				dstIndex++
			}

			from = srcIndex + 1
		}
	}

	copy(jsonData[dstIndex:], bytes[from:srcIndex])
	dstIndex += srcIndex - from

	// srcIndex < size means not reach the end of the bytes
	if srcIndex < size {
		// remove uncompleted utf8 bytes
		for i := dstIndex - 1; i >= 0 && jsonData[i]&0xC0 == 0x80; i-- {
			dstIndex = i - 1
		}
	}

	return jsonData[:dstIndex]
}

func toEscapeChar(c byte) byte {
	switch c {
	case '\n':
		return 'n'
	case '\t':
		return 't'
	case '"':
		return '"'
	case '\\':
		return '\\'
	default:
		return ' '
	}
}
