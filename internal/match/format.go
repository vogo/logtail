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

import (
	"fmt"

	"github.com/vogo/logtail/internal/util"
)

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

//nolint:varnamelen //ignore this.
func indexToNextLineStart(format *Format, message []byte) []byte {
	l := len(message)
	i := 0

	for i < l {
		i = util.IndexLineEnd(message, l, i)
		i = util.IgnoreLineEnd(message, l, i)

		if format == nil || format.PrefixMatch(message[i:]) {
			return message[i:]
		}
	}

	return nil
}

func IndexToLineStart(format *Format, data []byte) []byte {
	if format == nil || format.PrefixMatch(data) {
		return data
	}

	return indexToNextLineStart(format, data)
}

func IsFollowingLine(format *Format, bytes []byte) bool {
	if format != nil {
		return !format.PrefixMatch(bytes)
	}

	return bytes[0] == ' ' || bytes[0] == '\t'
}

//nolint:nonamedreturns //ignore this.
func SplitFirstLog(format *Format, data []byte) (first, remain []byte) {
	length := len(data)
	index := util.IndexLineEnd(data, length, 0)
	index = util.IgnoreLineEnd(data, length, index)

	if index < length {
		for index < length && IsFollowingLine(format, data[index:]) {
			index = util.IndexLineEnd(data, length, index)
			index = util.IgnoreLineEnd(data, length, index)
		}
	}

	return data[:index], data[index:]
}

//nolint:nonamedreturns //ignore this.
func SplitFollowingLog(format *Format, data []byte) (following, remain []byte) {
	length := len(data)
	index := 0

	for index < length && IsFollowingLine(format, data[index:]) {
		index = util.IndexLineEnd(data, length, index)
		index = util.IgnoreLineEnd(data, length, index)
	}

	return data[:index], data[index:]
}
