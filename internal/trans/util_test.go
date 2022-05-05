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

package trans_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail/internal/trans"
)

func TestEscapeLimitJsonBytes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []byte(`ab`), trans.EscapeLimitJSONBytes([]byte(`abcd`), 2))
	assert.Equal(t, []byte(`abcd`), trans.EscapeLimitJSONBytes([]byte(`abcd`), 4))
	assert.Equal(t, []byte(``), trans.EscapeLimitJSONBytes([]byte(`你好世界`), 2))
	assert.Equal(t, []byte(`你`), trans.EscapeLimitJSONBytes([]byte(`你好世界`), 3))
	assert.Equal(t, []byte(`你`), trans.EscapeLimitJSONBytes([]byte(`你好世界`), 4))
	assert.Equal(t, []byte(`你好`), trans.EscapeLimitJSONBytes([]byte(`你好世界`), 6))
	assert.Equal(t, []byte(`你好`), trans.EscapeLimitJSONBytes([]byte(`你好世界`), 8))
	assert.Equal(t, []byte(`你好世`), trans.EscapeLimitJSONBytes([]byte(`你好世界`), 9))
	assert.Equal(t, []byte(`你好世`), trans.EscapeLimitJSONBytes([]byte(`你好世界`), 10))
	assert.Equal(t, []byte(`你好世界`), trans.EscapeLimitJSONBytes([]byte(`你好世界`), 12))
	assert.Equal(t, []byte(`你好世界abc`), trans.EscapeLimitJSONBytes([]byte(`你好世界abc`), 16))

	assert.Equal(t, []byte(`ab\"cd`), trans.EscapeLimitJSONBytes([]byte(`ab"cd`), 6))
	assert.Equal(t, []byte(`ab\\\"cd`), trans.EscapeLimitJSONBytes([]byte(`ab\"cd`), 6))
	assert.Equal(t, []byte(`ab\\\\\\\"cd`), trans.EscapeLimitJSONBytes([]byte(`ab\\\"cd`), 10))
	assert.Equal(t, []byte(`ab\tcd`), trans.EscapeLimitJSONBytes([]byte(`ab	cd`), 8))
	assert.Equal(t, []byte(`ab\ncd`), trans.EscapeLimitJSONBytes([]byte("ab\ncd"), 8))
	assert.Equal(t, []byte(`abc\n`), trans.EscapeLimitJSONBytes([]byte("abc\nd"), 4))
	assert.Equal(t, []byte(`abc\n`), trans.EscapeLimitJSONBytes([]byte("abc\nd"), 4))

	assert.Equal(t, []byte(`test 操作异常`), trans.EscapeLimitJSONBytes([]byte("test 操作异常"), 1024))
}

func TestEscapeLog(t *testing.T) {
	s := `2022-04-28 13:25:00.263 ERROR [Thread-2742] [xxljob-d077a2cc-4a5e-440e-bf99-b3530e322772] c.h.t.c.c.i.MyService - 结果:失败,TradeNo=[202204280131042516], resultCode=[1], errCode=[not exists],message=[不存在]`

	larkTextMessageDataPrefix := `{"msg_type":"text","content":{"text":"`
	prefix := `[Log Alarm my-service]: `
	capacity := 1024 - len(larkTextMessageDataPrefix) - len(prefix)
	b := trans.EscapeLimitJSONBytes([]byte(s), capacity)
	println(string(b))
}
