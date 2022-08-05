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

//nolint:gomnd,funlen //ignore this
func EscapeLimitJSONBytes(bytes []byte, capacity int) []byte {
	if size := len(bytes); size < capacity {
		capacity = size
	}

	var jsonData []byte

	srcIndex := 0
	dstIndex := 0
	from := 0

CHECKLOOP:
	for ; srcIndex < capacity; srcIndex++ {
		switch char := bytes[srcIndex]; char {
		case '\n', '\t', '"', '\\':
			jsonData = make([]byte, capacity+capacity/2)

			break CHECKLOOP
		}
	}

	if srcIndex < capacity {
		for ; srcIndex < capacity; srcIndex++ {
			switch char := bytes[srcIndex]; char {
			case '\n', '\t', '"', '\\':
				copy(jsonData[dstIndex:], bytes[from:srcIndex])

				dstIndex += srcIndex - from

				jsonData[dstIndex] = '\\'
				dstIndex++

				jsonData[dstIndex] = toEscapeChar(char)
				dstIndex++

				from = srcIndex + 1
			}
		}

		copy(jsonData[dstIndex:], bytes[from:srcIndex])
		dstIndex += srcIndex - from
	} else {
		jsonData = bytes[:capacity]
		dstIndex = capacity
	}

	// remove latest uncompleted utf8 bytes
UTF8LOOP:
	for idx, followCount := dstIndex-1, 0; idx >= 0; idx-- {
		switch jsonData[idx] & 0xC0 {
		case 0x80:
			followCount++
		case 0xC0:
			if followCount == 0 || utf8FollowSize(jsonData[idx]) != followCount {
				dstIndex = idx
			}

			break UTF8LOOP
		default:
			break UTF8LOOP
		}
	}

	return jsonData[:dstIndex]
}

//nolint:gomnd //ignore this
func utf8FollowSize(utf8LeadByte byte) int {
	switch {
	case utf8LeadByte&0x20 == 0:
		return 1
	case utf8LeadByte&0x10 == 0:
		return 2
	case utf8LeadByte&0x08 == 0:
		return 3
	case utf8LeadByte&0x04 == 0:
		return 4
	case utf8LeadByte&0x02 == 0:
		return 5
	default:
		return 0
	}
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
