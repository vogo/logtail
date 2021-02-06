package logtail_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail"
)

func TestEscapeLimitJsonBytes(t *testing.T) {
	assert.Equal(t, []byte(`ab`), logtail.EscapeLimitJSONBytes([]byte(`abcd`), 2))
	assert.Equal(t, []byte(`abcd`), logtail.EscapeLimitJSONBytes([]byte(`abcd`), 4))
	assert.Equal(t, []byte(`你好`), logtail.EscapeLimitJSONBytes([]byte(`你好世界`), 8))
	assert.Equal(t, []byte(`你好`), logtail.EscapeLimitJSONBytes([]byte(`你好世界`), 9))

	assert.Equal(t, []byte(`ab\"cd`), logtail.EscapeLimitJSONBytes([]byte(`ab"cd`), 6))
	assert.Equal(t, []byte(`ab\tcd`), logtail.EscapeLimitJSONBytes([]byte(`ab	cd`), 8))
	assert.Equal(t, []byte(`ab\ncd`), logtail.EscapeLimitJSONBytes([]byte("ab\ncd"), 8))
	assert.Equal(t, []byte(`abc\n`), logtail.EscapeLimitJSONBytes([]byte("abc\nd"), 4))
	assert.Equal(t, []byte(`abc\n`), logtail.EscapeLimitJSONBytes([]byte("abc\nd"), 4))

	assert.Equal(t, []byte(`test 操作异常`), logtail.EscapeLimitJSONBytes([]byte("test 操作异常"), 1024))
}
